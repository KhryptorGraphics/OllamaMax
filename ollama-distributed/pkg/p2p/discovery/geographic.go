package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// GeographicInfo represents geographic information about a peer
type GeographicInfo struct {
	Country     string    `json:"country"`
	Region      string    `json:"region"`
	City        string    `json:"city"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Timezone    string    `json:"timezone"`
	ISP         string    `json:"isp"`
	LastUpdated time.Time `json:"last_updated"`
}

// GeographicDetector detects geographic information for peers
type GeographicDetector struct {
	mu            sync.RWMutex
	peerLocations map[peer.ID]*GeographicInfo
	localInfo     *GeographicInfo
	httpClient    *http.Client

	// Configuration
	geoAPIURL    string
	cacheTimeout time.Duration
}

// NewGeographicDetector creates a new geographic detector
func NewGeographicDetector() *GeographicDetector {
	return &GeographicDetector{
		peerLocations: make(map[peer.ID]*GeographicInfo),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		geoAPIURL:    "http://ip-api.com/json/",
		cacheTimeout: 24 * time.Hour, // Cache for 24 hours
	}
}

// DetectLocalLocation detects the geographic location of the local node
func (gd *GeographicDetector) DetectLocalLocation(ctx context.Context) error {
	info, err := gd.getLocationFromAPI(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to detect local location: %w", err)
	}

	gd.mu.Lock()
	gd.localInfo = info
	gd.mu.Unlock()

	return nil
}

// DetectPeerLocation detects the geographic location of a peer
func (gd *GeographicDetector) DetectPeerLocation(ctx context.Context, peerID peer.ID, addrs []multiaddr.Multiaddr) error {
	// Extract IP address from multiaddrs
	ip := gd.extractIPFromAddrs(addrs)
	if ip == "" {
		return fmt.Errorf("no valid IP address found for peer %s", peerID)
	}

	// Check cache first
	gd.mu.RLock()
	if cached, exists := gd.peerLocations[peerID]; exists {
		if time.Since(cached.LastUpdated) < gd.cacheTimeout {
			gd.mu.RUnlock()
			return nil // Use cached data
		}
	}
	gd.mu.RUnlock()

	// Fetch new location data
	info, err := gd.getLocationFromAPI(ctx, ip)
	if err != nil {
		return fmt.Errorf("failed to detect location for peer %s: %w", peerID, err)
	}

	gd.mu.Lock()
	gd.peerLocations[peerID] = info
	gd.mu.Unlock()

	return nil
}

// GetLocalLocation returns the local node's geographic information
func (gd *GeographicDetector) GetLocalLocation() *GeographicInfo {
	gd.mu.RLock()
	defer gd.mu.RUnlock()
	return gd.localInfo
}

// GetPeerLocation returns a peer's geographic information
func (gd *GeographicDetector) GetPeerLocation(peerID peer.ID) *GeographicInfo {
	gd.mu.RLock()
	defer gd.mu.RUnlock()
	return gd.peerLocations[peerID]
}

// GetNearbyPeers returns peers sorted by geographic distance
func (gd *GeographicDetector) GetNearbyPeers(maxDistance float64) []peer.ID {
	gd.mu.RLock()
	defer gd.mu.RUnlock()

	if gd.localInfo == nil {
		return nil
	}

	type peerDistance struct {
		peerID   peer.ID
		distance float64
	}

	var peers []peerDistance

	for peerID, info := range gd.peerLocations {
		distance := gd.calculateDistance(
			gd.localInfo.Latitude, gd.localInfo.Longitude,
			info.Latitude, info.Longitude,
		)

		if distance <= maxDistance {
			peers = append(peers, peerDistance{
				peerID:   peerID,
				distance: distance,
			})
		}
	}

	// Sort by distance
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].distance < peers[j].distance
	})

	result := make([]peer.ID, len(peers))
	for i, p := range peers {
		result[i] = p.peerID
	}

	return result
}

// GetPeersByRegion returns peers in the same region
func (gd *GeographicDetector) GetPeersByRegion(region string) []peer.ID {
	gd.mu.RLock()
	defer gd.mu.RUnlock()

	var peers []peer.ID

	for peerID, info := range gd.peerLocations {
		if info.Region == region {
			peers = append(peers, peerID)
		}
	}

	return peers
}

// extractIPFromAddrs extracts an IP address from multiaddrs
func (gd *GeographicDetector) extractIPFromAddrs(addrs []multiaddr.Multiaddr) string {
	for _, addr := range addrs {
		// Parse the multiaddr to extract IP
		components := addr.Protocols()
		for i, proto := range components {
			if proto.Code == multiaddr.P_IP4 || proto.Code == multiaddr.P_IP6 {
				if i+1 < len(components) {
					// Get the IP value
					value, err := addr.ValueForProtocol(proto.Code)
					if err == nil {
						ip := net.ParseIP(value)
						if ip != nil && !ip.IsLoopback() && !ip.IsPrivate() {
							return ip.String()
						}
					}
				}
			}
		}
	}
	return ""
}

// getLocationFromAPI fetches location data from IP geolocation API
func (gd *GeographicDetector) getLocationFromAPI(ctx context.Context, ip string) (*GeographicInfo, error) {
	url := gd.geoAPIURL
	if ip != "" {
		url += ip
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := gd.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Status     string  `json:"status"`
		Country    string  `json:"country"`
		RegionName string  `json:"regionName"`
		City       string  `json:"city"`
		Lat        float64 `json:"lat"`
		Lon        float64 `json:"lon"`
		Timezone   string  `json:"timezone"`
		ISP        string  `json:"isp"`
		Message    string  `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	if apiResponse.Status != "success" {
		return nil, fmt.Errorf("API error: %s", apiResponse.Message)
	}

	return &GeographicInfo{
		Country:     apiResponse.Country,
		Region:      apiResponse.RegionName,
		City:        apiResponse.City,
		Latitude:    apiResponse.Lat,
		Longitude:   apiResponse.Lon,
		Timezone:    apiResponse.Timezone,
		ISP:         apiResponse.ISP,
		LastUpdated: time.Now(),
	}, nil
}

// calculateDistance calculates the distance between two geographic points using Haversine formula
func (gd *GeographicDetector) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
