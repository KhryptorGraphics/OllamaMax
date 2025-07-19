package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/network"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	
	"github.com/ollama/ollama-distributed/pkg/config"
)

const (
	// Advertisement protocols
	ResourceAdvertisementProtocol = protocol.ID("/ollamacron/resource-advertisement/1.0.0")
	ResourceDiscoveryProtocol     = protocol.ID("/ollamacron/resource-discovery/1.0.0")
	
	// DHT keys
	ResourceKeyPrefix = "/ollamacron/resources/"
	ModelKeyPrefix    = "/ollamacron/models/"
	NodeKeyPrefix     = "/ollamacron/nodes/"
)

// ResourceAdvertiser manages resource advertisement and discovery
type ResourceAdvertiser struct {
	host           host.Host
	dht            *dht.IpfsDHT
	
	// Resource information
	capabilities   *config.NodeCapabilities
	resources      *config.ResourceMetrics
	
	// Advertisement management
	advertisements map[string]*Advertisement
	advMux         sync.RWMutex
	
	// Update channels
	capabilityUpdates chan *config.NodeCapabilities
	resourceUpdates   chan *config.ResourceMetrics
	
	// Subscriptions
	subscriptions     map[string]*ResourceSubscription
	subMux           sync.RWMutex
	
	// Discovery cache
	discoveryCache    *DiscoveryCache
	
	// Configuration
	config           *AdvertiserConfig
	
	// Metrics
	metrics          *AdvertiserMetrics
	
	// Lifecycle
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

// Advertisement represents a resource advertisement
type Advertisement struct {
	ID            string                `json:"id"`
	NodeID        peer.ID               `json:"node_id"`
	Capabilities  *config.NodeCapabilities `json:"capabilities"`
	Resources     *config.ResourceMetrics  `json:"resources"`
	Timestamp     time.Time             `json:"timestamp"`
	TTL           time.Duration         `json:"ttl"`
	Version       int                   `json:"version"`
	Signature     []byte                `json:"signature,omitempty"`
	
	// Advertisement metadata
	Priority      int                   `json:"priority"`
	Tags          map[string]string     `json:"tags"`
	Geo           *GeographicInfo       `json:"geo,omitempty"`
	Availability  *AvailabilityInfo     `json:"availability"`
}

// GeographicInfo contains geographic information
type GeographicInfo struct {
	Country     string  `json:"country"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
}

// AvailabilityInfo contains availability information
type AvailabilityInfo struct {
	Online        bool              `json:"online"`
	LastSeen      time.Time         `json:"last_seen"`
	Uptime        time.Duration     `json:"uptime"`
	MaintenanceWindow *TimeWindow   `json:"maintenance_window,omitempty"`
	ServiceLevel  string            `json:"service_level"`
}

// TimeWindow represents a time window
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ResourceSubscription represents a resource subscription
type ResourceSubscription struct {
	ID            string
	Query         *ResourceQuery
	Callback      func(*Advertisement)
	CreatedAt     time.Time
	LastMatched   time.Time
	MatchCount    int
}

// AdvertiserConfig holds advertiser configuration
type AdvertiserConfig struct {
	AdvertisementInterval time.Duration `json:"advertisement_interval"`
	TTL                   time.Duration `json:"ttl"`
	MaxAdvertisements     int           `json:"max_advertisements"`
	EnableSigning         bool          `json:"enable_signing"`
	EnableGeoLocation     bool          `json:"enable_geo_location"`
	Priority              int           `json:"priority"`
	Tags                  map[string]string `json:"tags"`
}

// AdvertiserMetrics tracks advertiser metrics
type AdvertiserMetrics struct {
	AdvertisementsSent     int
	AdvertisementsReceived int
	DiscoveryQueries       int
	SubscriptionMatches    int
	CacheHits              int
	CacheMisses            int
	LastAdvertisement      time.Time
	LastDiscovery          time.Time
	StartTime              time.Time
}

// NewResourceAdvertiser creates a new resource advertiser
func NewResourceAdvertiser(ctx context.Context, host host.Host, dht *dht.IpfsDHT, config *AdvertiserConfig) (*ResourceAdvertiser, error) {
	if config == nil {
		config = DefaultAdvertiserConfig()
	}
	
	ctx, cancel := context.WithCancel(ctx)
	
	ra := &ResourceAdvertiser{
		host:              host,
		dht:               dht,
		config:            config,
		advertisements:    make(map[string]*Advertisement),
		capabilityUpdates: make(chan *config.NodeCapabilities, 10),
		resourceUpdates:   make(chan *config.ResourceMetrics, 10),
		subscriptions:     make(map[string]*ResourceSubscription),
		metrics: &AdvertiserMetrics{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Initialize discovery cache
	ra.discoveryCache = NewDiscoveryCache(1000, 5*time.Minute)
	
	// Setup protocol handlers
	ra.setupProtocolHandlers()
	
	return ra, nil
}

// setupProtocolHandlers sets up protocol handlers
func (ra *ResourceAdvertiser) setupProtocolHandlers() {
	ra.host.SetStreamHandler(ResourceAdvertisementProtocol, ra.handleResourceAdvertisement)
	ra.host.SetStreamHandler(ResourceDiscoveryProtocol, ra.handleResourceDiscovery)
}

// Start starts the resource advertiser
func (ra *ResourceAdvertiser) Start() {
	log.Printf("Starting resource advertiser")
	
	// Start periodic advertisement
	ra.wg.Add(1)
	go ra.advertisementTask()
	
	// Start update processing
	ra.wg.Add(1)
	go ra.updateProcessor()
	
	// Start metrics collection
	ra.wg.Add(1)
	go ra.metricsTask()
	
	// Start cache cleanup
	ra.wg.Add(1)
	go ra.cacheCleanupTask()
	
	log.Printf("Resource advertiser started")
}

// advertisementTask handles periodic advertisements
func (ra *ResourceAdvertiser) advertisementTask() {
	defer ra.wg.Done()
	
	// Initial advertisement
	ra.advertiseResources()
	
	ticker := time.NewTicker(ra.config.AdvertisementInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ra.ctx.Done():
			return
		case <-ticker.C:
			ra.advertiseResources()
		}
	}
}

// updateProcessor processes capability and resource updates
func (ra *ResourceAdvertiser) updateProcessor() {
	defer ra.wg.Done()
	
	for {
		select {
		case <-ra.ctx.Done():
			return
		case caps := <-ra.capabilityUpdates:
			ra.capabilities = caps
			ra.advertiseResources()
		case metrics := <-ra.resourceUpdates:
			ra.resources = metrics
			ra.advertiseResources()
		}
	}
}

// advertiseResources advertises current resources
func (ra *ResourceAdvertiser) advertiseResources() {
	if ra.capabilities == nil {
		return
	}
	
	// Create advertisement
	ad := &Advertisement{
		ID:           fmt.Sprintf("%s-%d", ra.host.ID(), time.Now().UnixNano()),
		NodeID:       ra.host.ID(),
		Capabilities: ra.capabilities,
		Resources:    ra.resources,
		Timestamp:    time.Now(),
		TTL:          ra.config.TTL,
		Version:      1,
		Priority:     ra.config.Priority,
		Tags:         ra.config.Tags,
		Availability: &AvailabilityInfo{
			Online:       true,
			LastSeen:     time.Now(),
			Uptime:       time.Since(ra.metrics.StartTime),
			ServiceLevel: "standard",
		},
	}
	
	// Add geographic info if enabled
	if ra.config.EnableGeoLocation {
		ad.Geo = ra.getGeographicInfo()
	}
	
	// Sign advertisement if enabled
	if ra.config.EnableSigning {
		signature, err := ra.signAdvertisement(ad)
		if err != nil {
			log.Printf("Failed to sign advertisement: %v", err)
		} else {
			ad.Signature = signature
		}
	}
	
	// Store advertisement
	ra.advMux.Lock()
	ra.advertisements[ad.ID] = ad
	ra.advMux.Unlock()
	
	// Publish to DHT
	if err := ra.publishAdvertisement(ad); err != nil {
		log.Printf("Failed to publish advertisement: %v", err)
		return
	}
	
	// Broadcast to interested peers
	ra.broadcastAdvertisement(ad)
	
	// Check subscriptions
	ra.checkSubscriptions(ad)
	
	ra.metrics.AdvertisementsSent++
	ra.metrics.LastAdvertisement = time.Now()
	
	log.Printf("Advertised resources: %d models, %d GPUs, %.2f CPU", 
		len(ra.capabilities.SupportedModels), 
		len(ra.capabilities.GPUs), 
		float64(ra.capabilities.CPUCores))
}

// publishAdvertisement publishes advertisement to DHT
func (ra *ResourceAdvertiser) publishAdvertisement(ad *Advertisement) error {
	// Serialize advertisement
	data, err := json.Marshal(ad)
	if err != nil {
		return fmt.Errorf("failed to serialize advertisement: %w", err)
	}
	
	// Store in DHT with multiple keys for discoverability
	keys := []string{
		fmt.Sprintf("%s%s", ResourceKeyPrefix, ad.NodeID),
		fmt.Sprintf("%s%s", NodeKeyPrefix, ad.NodeID),
	}
	
	// Add model-specific keys
	for _, model := range ad.Capabilities.SupportedModels {
		keys = append(keys, fmt.Sprintf("%s%s", ModelKeyPrefix, model))
	}
	
	// Store with all keys
	for _, key := range keys {
		if err := ra.dht.PutValue(ra.ctx, key, data); err != nil {
			log.Printf("Failed to store advertisement with key %s: %v", key, err)
			continue
		}
	}
	
	return nil
}

// broadcastAdvertisement broadcasts advertisement to connected peers
func (ra *ResourceAdvertiser) broadcastAdvertisement(ad *Advertisement) {
	// Get connected peers
	peers := ra.host.Network().Peers()
	
	// Broadcast to random subset to avoid flooding
	maxBroadcast := 10
	if len(peers) > maxBroadcast {
		// Randomly select peers
		selected := make([]peer.ID, maxBroadcast)
		for i := 0; i < maxBroadcast; i++ {
			selected[i] = peers[i]
		}
		peers = selected
	}
	
	// Send to selected peers
	for _, peerID := range peers {
		go ra.sendAdvertisement(peerID, ad)
	}
}

// sendAdvertisement sends advertisement to a specific peer
func (ra *ResourceAdvertiser) sendAdvertisement(peerID peer.ID, ad *Advertisement) {
	stream, err := ra.host.NewStream(ra.ctx, peerID, ResourceAdvertisementProtocol)
	if err != nil {
		log.Printf("Failed to create stream to peer %s: %v", peerID, err)
		return
	}
	defer stream.Close()
	
	// Send advertisement
	data, err := json.Marshal(ad)
	if err != nil {
		log.Printf("Failed to serialize advertisement: %v", err)
		return
	}
	
	if _, err := stream.Write(data); err != nil {
		log.Printf("Failed to send advertisement to peer %s: %v", peerID, err)
		return
	}
}

// SetCapabilities updates node capabilities
func (ra *ResourceAdvertiser) SetCapabilities(caps *config.NodeCapabilities) {
	select {
	case ra.capabilityUpdates <- caps:
	default:
		// Channel full, update directly
		ra.capabilities = caps
		go ra.advertiseResources()
	}
}

// SetResourceMetrics updates resource metrics
func (ra *ResourceAdvertiser) SetResourceMetrics(metrics *config.ResourceMetrics) {
	select {
	case ra.resourceUpdates <- metrics:
	default:
		// Channel full, update directly
		ra.resources = metrics
		go ra.advertiseResources()
	}
}

// Subscribe creates a subscription for resource updates
func (ra *ResourceAdvertiser) Subscribe(query *ResourceQuery, callback func(*Advertisement)) string {
	ra.subMux.Lock()
	defer ra.subMux.Unlock()
	
	subscriptionID := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	subscription := &ResourceSubscription{
		ID:        subscriptionID,
		Query:     query,
		Callback:  callback,
		CreatedAt: time.Now(),
	}
	
	ra.subscriptions[subscriptionID] = subscription
	
	log.Printf("Created resource subscription: %s", subscriptionID)
	return subscriptionID
}

// Unsubscribe removes a resource subscription
func (ra *ResourceAdvertiser) Unsubscribe(subscriptionID string) {
	ra.subMux.Lock()
	defer ra.subMux.Unlock()
	
	delete(ra.subscriptions, subscriptionID)
	log.Printf("Removed resource subscription: %s", subscriptionID)
}

// checkSubscriptions checks if advertisement matches any subscriptions
func (ra *ResourceAdvertiser) checkSubscriptions(ad *Advertisement) {
	ra.subMux.RLock()
	defer ra.subMux.RUnlock()
	
	for _, sub := range ra.subscriptions {
		if ra.matchesQuery(ad, sub.Query) {
			sub.LastMatched = time.Now()
			sub.MatchCount++
			ra.metrics.SubscriptionMatches++
			
			// Call callback
			go sub.Callback(ad)
		}
	}
}

// Protocol handlers

// handleResourceAdvertisement handles incoming resource advertisements
func (ra *ResourceAdvertiser) handleResourceAdvertisement(stream network.Stream) {
	defer stream.Close()
	
	// Read advertisement
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		log.Printf("Failed to read advertisement: %v", err)
		return
	}
	
	var ad Advertisement
	if err := json.Unmarshal(buf[:n], &ad); err != nil {
		log.Printf("Failed to unmarshal advertisement: %v", err)
		return
	}
	
	// Verify advertisement if signed
	if ad.Signature != nil {
		if err := ra.verifyAdvertisement(&ad); err != nil {
			log.Printf("Advertisement verification failed: %v", err)
			return
		}
	}
	
	// Store in cache
	ra.discoveryCache.Store(&ad)
	
	// Check subscriptions
	ra.checkSubscriptions(&ad)
	
	ra.metrics.AdvertisementsReceived++
	
	log.Printf("Received resource advertisement from: %s", ad.NodeID)
}

// handleResourceDiscovery handles resource discovery requests
func (ra *ResourceAdvertiser) handleResourceDiscovery(stream network.Stream) {
	defer stream.Close()
	
	// Read query
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		log.Printf("Failed to read discovery query: %v", err)
		return
	}
	
	var query ResourceQuery
	if err := json.Unmarshal(buf[:n], &query); err != nil {
		log.Printf("Failed to unmarshal discovery query: %v", err)
		return
	}
	
	// Find matching resources
	results := ra.findMatchingResources(&query)
	
	// Send results
	response := &DiscoveryResponse{
		QueryID:   query.ID,
		Results:   results,
		Timestamp: time.Now(),
	}
	
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to serialize discovery response: %v", err)
		return
	}
	
	if _, err := stream.Write(data); err != nil {
		log.Printf("Failed to send discovery response: %v", err)
		return
	}
	
	ra.metrics.DiscoveryQueries++
	
	log.Printf("Processed discovery query: %s", query.ID)
}

// findMatchingResources finds resources matching a query
func (ra *ResourceAdvertiser) findMatchingResources(query *ResourceQuery) []*Advertisement {
	var results []*Advertisement
	
	// Check cache first
	cached := ra.discoveryCache.Find(query)
	if len(cached) > 0 {
		ra.metrics.CacheHits++
		return cached
	}
	
	ra.metrics.CacheMisses++
	
	// Query DHT
	results = ra.queryDHT(query)
	
	// Store in cache
	for _, ad := range results {
		ra.discoveryCache.Store(ad)
	}
	
	return results
}

// queryDHT queries DHT for matching resources
func (ra *ResourceAdvertiser) queryDHT(query *ResourceQuery) []*Advertisement {
	var results []*Advertisement
	
	// Search by model types
	for _, modelType := range query.ModelTypes {
		key := fmt.Sprintf("%s%s", ModelKeyPrefix, modelType)
		
		val, err := ra.dht.GetValue(ra.ctx, key)
		if err != nil {
			continue
		}
		
		var ad Advertisement
		if err := json.Unmarshal(val, &ad); err != nil {
			continue
		}
		
		if ra.matchesQuery(&ad, query) {
			results = append(results, &ad)
		}
	}
	
	return results
}

// matchesQuery checks if advertisement matches query
func (ra *ResourceAdvertiser) matchesQuery(ad *Advertisement, query *ResourceQuery) bool {
	// Check model types
	if len(query.ModelTypes) > 0 {
		hasModel := false
		for _, modelType := range query.ModelTypes {
			for _, supportedModel := range ad.Capabilities.SupportedModels {
				if supportedModel == modelType {
					hasModel = true
					break
				}
			}
			if hasModel {
				break
			}
		}
		if !hasModel {
			return false
		}
	}
	
	// Check CPU requirements
	if query.MinCPU > 0 && ad.Capabilities.CPUCores < query.MinCPU {
		return false
	}
	
	// Check memory requirements
	if query.MinMemory > 0 && ad.Capabilities.Memory < query.MinMemory {
		return false
	}
	
	// Check GPU requirements
	if query.RequiredGPU && len(ad.Capabilities.GPUs) == 0 {
		return false
	}
	
	// Check latency requirements
	if query.MaxLatency > 0 && ad.Capabilities.Latency > query.MaxLatency {
		return false
	}
	
	// Check price requirements
	if query.MaxPrice > 0 && ad.Capabilities.PricePerToken > query.MaxPrice {
		return false
	}
	
	return true
}

// Utility methods

// getGeographicInfo returns geographic information
func (ra *ResourceAdvertiser) getGeographicInfo() *GeographicInfo {
	// TODO: Implement geographic info detection
	return &GeographicInfo{
		Country:  "Unknown",
		Region:   "Unknown",
		City:     "Unknown",
		Timezone: "UTC",
	}
}

// signAdvertisement signs an advertisement
func (ra *ResourceAdvertiser) signAdvertisement(ad *Advertisement) ([]byte, error) {
	// Create signature payload
	payload := fmt.Sprintf("%s:%s:%d", ad.NodeID, ad.ID, ad.Timestamp.Unix())
	
	// Sign with host's private key
	privKey := ra.host.Peerstore().PrivKey(ra.host.ID())
	return privKey.Sign([]byte(payload))
}

// verifyAdvertisement verifies an advertisement signature
func (ra *ResourceAdvertiser) verifyAdvertisement(ad *Advertisement) error {
	// Get peer's public key
	pubKey, err := ad.NodeID.ExtractPublicKey()
	if err != nil {
		return fmt.Errorf("failed to extract public key: %w", err)
	}
	
	// Verify signature
	payload := fmt.Sprintf("%s:%s:%d", ad.NodeID, ad.ID, ad.Timestamp.Unix())
	valid, err := pubKey.Verify([]byte(payload), ad.Signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	
	if !valid {
		return fmt.Errorf("invalid signature")
	}
	
	return nil
}

// metricsTask collects metrics
func (ra *ResourceAdvertiser) metricsTask() {
	defer ra.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ra.ctx.Done():
			return
		case <-ticker.C:
			ra.updateMetrics()
		}
	}
}

// updateMetrics updates metrics
func (ra *ResourceAdvertiser) updateMetrics() {
	// TODO: Implement metrics collection
}

// cacheCleanupTask cleans up expired cache entries
func (ra *ResourceAdvertiser) cacheCleanupTask() {
	defer ra.wg.Done()
	
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ra.ctx.Done():
			return
		case <-ticker.C:
			ra.discoveryCache.Cleanup()
		}
	}
}

// GetMetrics returns advertiser metrics
func (ra *ResourceAdvertiser) GetMetrics() *AdvertiserMetrics {
	return ra.metrics
}

// Stop stops the resource advertiser
func (ra *ResourceAdvertiser) Stop() {
	log.Printf("Stopping resource advertiser")
	ra.cancel()
	ra.wg.Wait()
	log.Printf("Resource advertiser stopped")
}

// DefaultAdvertiserConfig returns default configuration
func DefaultAdvertiserConfig() *AdvertiserConfig {
	return &AdvertiserConfig{
		AdvertisementInterval: 30 * time.Second,
		TTL:                   5 * time.Minute,
		MaxAdvertisements:     100,
		EnableSigning:         true,
		EnableGeoLocation:     false,
		Priority:              1,
		Tags:                  make(map[string]string),
	}
}