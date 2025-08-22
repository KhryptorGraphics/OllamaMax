package routing

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ContentDiscovery manages content discovery in the network
type ContentDiscovery struct {
	host   host.Host
	dht    *dht.IpfsDHT
	config *ContentRouterConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// RoutingTable manages peer routing information
type RoutingTable struct {
	localPeerID peer.ID
	routes      map[string][]peer.ID
	mu          sync.RWMutex
}

// NewContentDiscovery creates a new content discovery instance
func NewContentDiscovery(host host.Host, dht *dht.IpfsDHT, config *ContentRouterConfig) *ContentDiscovery {
	ctx, cancel := context.WithCancel(context.Background())
	return &ContentDiscovery{
		host:   host,
		dht:    dht,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// NewRoutingTable creates a new routing table
func NewRoutingTable(localPeerID peer.ID) *RoutingTable {
	return &RoutingTable{
		localPeerID: localPeerID,
		routes:      make(map[string][]peer.ID),
	}
}

// ContentDiscovery methods

// Start starts the content discovery service
func (cd *ContentDiscovery) Start() {
	log.Printf("Starting content discovery service")
	// TODO: Implement discovery service startup logic
}

// Stop stops the content discovery service
func (cd *ContentDiscovery) Stop() {
	log.Printf("Stopping content discovery service")
	cd.cancel()
}

// DiscoverContent discovers content in the network
func (cd *ContentDiscovery) DiscoverContent(ctx context.Context, contentID string) ([]peer.ID, error) {
	// TODO: Implement content discovery logic
	// This would use DHT to find providers of the content
	return []peer.ID{}, nil
}

// AnnounceContent announces content availability
func (cd *ContentDiscovery) AnnounceContent(ctx context.Context, contentID string) error {
	// TODO: Implement content announcement logic
	// This would announce to the DHT that we provide this content
	return nil
}

// RoutingTable methods

// AddRoute adds a route to a peer for specific content
func (rt *RoutingTable) AddRoute(contentID string, peerID peer.ID) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.routes[contentID] == nil {
		rt.routes[contentID] = make([]peer.ID, 0)
	}

	// Check if peer already exists in routes
	for _, existingPeer := range rt.routes[contentID] {
		if existingPeer == peerID {
			return // Already exists
		}
	}

	rt.routes[contentID] = append(rt.routes[contentID], peerID)
}

// RemoveRoute removes a route to a peer for specific content
func (rt *RoutingTable) RemoveRoute(contentID string, peerID peer.ID) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	routes, exists := rt.routes[contentID]
	if !exists {
		return
	}

	for i, existingPeer := range routes {
		if existingPeer == peerID {
			rt.routes[contentID] = append(routes[:i], routes[i+1:]...)
			break
		}
	}

	// Remove empty route entries
	if len(rt.routes[contentID]) == 0 {
		delete(rt.routes, contentID)
	}
}

// GetRoutes returns routes for specific content
func (rt *RoutingTable) GetRoutes(contentID string) []peer.ID {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	routes, exists := rt.routes[contentID]
	if !exists {
		return []peer.ID{}
	}

	// Return a copy to avoid race conditions
	result := make([]peer.ID, len(routes))
	copy(result, routes)
	return result
}

// GetAllRoutes returns all routes
func (rt *RoutingTable) GetAllRoutes() map[string][]peer.ID {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	result := make(map[string][]peer.ID)
	for contentID, routes := range rt.routes {
		routesCopy := make([]peer.ID, len(routes))
		copy(routesCopy, routes)
		result[contentID] = routesCopy
	}
	return result
}

// ClearRoutes clears all routes for a specific content
func (rt *RoutingTable) ClearRoutes(contentID string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	delete(rt.routes, contentID)
}

// ClearAllRoutes clears all routes
func (rt *RoutingTable) ClearAllRoutes() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.routes = make(map[string][]peer.ID)
}

// HasRoute checks if a route exists for specific content and peer
func (rt *RoutingTable) HasRoute(contentID string, peerID peer.ID) bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	routes, exists := rt.routes[contentID]
	if !exists {
		return false
	}

	for _, existingPeer := range routes {
		if existingPeer == peerID {
			return true
		}
	}
	return false
}

// GetRouteCount returns the number of routes for specific content
func (rt *RoutingTable) GetRouteCount(contentID string) int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	routes, exists := rt.routes[contentID]
	if !exists {
		return 0
	}
	return len(routes)
}

// GetTotalRouteCount returns the total number of routes
func (rt *RoutingTable) GetTotalRouteCount() int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	total := 0
	for _, routes := range rt.routes {
		total += len(routes)
	}
	return total
}

// Protocol handlers for ContentRouter

// handleContentRouting handles content routing requests
func (cr *ContentRouter) handleContentRouting(stream network.Stream) {
	defer stream.Close()

	peerID := stream.Conn().RemotePeer()
	log.Printf("Handling content routing request from peer: %s", peerID)

	// Read routing message
	var message RoutingMessage
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&message); err != nil {
		log.Printf("Failed to decode routing message: %v", err)
		return
	}

	// Process routing request
	routes := cr.routingTable.GetRoutes(message.ContentID)

	// Send response
	response := &RoutingResponse{
		MessageID: message.MessageID,
		Success:   len(routes) > 0,
		Routes:    routes,
		Timestamp: time.Now(),
	}

	if len(routes) == 0 {
		response.Message = "No routes found"
	}

	encoder := json.NewEncoder(stream)
	if err := encoder.Encode(response); err != nil {
		log.Printf("Failed to send routing response: %v", err)
		return
	}

	log.Printf("Sent routing response with %d routes", len(routes))
}

// handleContentRequest handles content requests
func (cr *ContentRouter) handleContentRequest(stream network.Stream) {
	defer stream.Close()

	peerID := stream.Conn().RemotePeer()
	log.Printf("Handling content request from peer: %s", peerID)

	// Read request message
	var request ContentRequestMessage
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&request); err != nil {
		log.Printf("Failed to decode content request: %v", err)
		return
	}

	// Check if we have the content locally
	content, exists := cr.contentStore.GetLocal(request.ContentID)

	response := &ContentResponseMessage{
		RequestID: request.RequestID,
		Success:   exists,
		Timestamp: time.Now(),
	}

	if exists {
		response.ContentMeta = content
		cr.metrics.ContentProvided++
	} else {
		response.Message = "Content not found"
	}

	// Send response
	encoder := json.NewEncoder(stream)
	if err := encoder.Encode(response); err != nil {
		log.Printf("Failed to send content response: %v", err)
		return
	}

	log.Printf("Handled content request for: %s (found: %v)", request.ContentID, exists)
}

// handleContentProvide handles content provision announcements
func (cr *ContentRouter) handleContentProvide(stream network.Stream) {
	defer stream.Close()

	peerID := stream.Conn().RemotePeer()
	log.Printf("Handling content provision from peer: %s", peerID)

	// Read provision announcement
	var announcement ProvisionAnnouncement
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&announcement); err != nil {
		log.Printf("Failed to decode provision announcement: %v", err)
		return
	}

	// Add route to routing table
	cr.routingTable.AddRoute(announcement.ContentID, announcement.Provider)

	// Update providers list
	cr.providersMux.Lock()
	if cr.providers[announcement.ContentID] == nil {
		cr.providers[announcement.ContentID] = make([]peer.ID, 0)
	}

	// Check if provider already exists
	exists := false
	for _, existingProvider := range cr.providers[announcement.ContentID] {
		if existingProvider == announcement.Provider {
			exists = true
			break
		}
	}

	if !exists {
		cr.providers[announcement.ContentID] = append(cr.providers[announcement.ContentID], announcement.Provider)
	}
	cr.providersMux.Unlock()

	log.Printf("Added provider %s for content: %s", announcement.Provider, announcement.ContentID)
}

// Background tasks for ContentRouter

// metricsUpdateTask periodically updates metrics
func (cr *ContentRouter) metricsUpdateTask() {
	defer cr.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.updateMetrics()
		}
	}
}

// requestTimeoutTask handles request timeouts
func (cr *ContentRouter) requestTimeoutTask() {
	defer cr.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.checkRequestTimeouts()
		}
	}
}

// checkRequestTimeouts checks for and handles request timeouts
func (cr *ContentRouter) checkRequestTimeouts() {
	now := time.Now()

	cr.requestsMux.Lock()
	defer cr.requestsMux.Unlock()

	for id, request := range cr.activeRequests {
		if request.Status == RequestStatusPending || request.Status == RequestStatusInProgress {
			if now.Sub(request.CreatedAt) > request.Timeout {
				// Timeout the request
				request.Status = RequestStatusFailed
				request.Error = "Request timeout"
				request.CompletedAt = &now

				// Send timeout response
				response := &ContentResponseMessage{
					RequestID: request.ID,
					Success:   false,
					Message:   "Request timeout",
					Timestamp: now,
				}

				select {
				case request.ResponseChannel <- response:
				default:
				}

				// Remove from active requests
				delete(cr.activeRequests, id)
				cr.metrics.FailedRoutes++

				log.Printf("Request timeout: %s", id)
			}
		}
	}
}

// updateMetrics updates router metrics
func (cr *ContentRouter) updateMetrics() {
	// TODO: Implement metrics collection
	// This could include calculating average latency, updating counters, etc.
}

// GetProviders returns providers for specific content
func (cr *ContentRouter) GetProviders(contentID string) []peer.ID {
	cr.providersMux.RLock()
	defer cr.providersMux.RUnlock()

	providers, exists := cr.providers[contentID]
	if !exists {
		return []peer.ID{}
	}

	// Return a copy to avoid race conditions
	result := make([]peer.ID, len(providers))
	copy(result, providers)
	return result
}

// AddProvider adds a provider for specific content
func (cr *ContentRouter) AddProvider(contentID string, peerID peer.ID) {
	cr.providersMux.Lock()
	defer cr.providersMux.Unlock()

	if cr.providers[contentID] == nil {
		cr.providers[contentID] = make([]peer.ID, 0)
	}

	// Check if provider already exists
	for _, existingProvider := range cr.providers[contentID] {
		if existingProvider == peerID {
			return // Already exists
		}
	}

	cr.providers[contentID] = append(cr.providers[contentID], peerID)
	cr.routingTable.AddRoute(contentID, peerID)
}

// RemoveProvider removes a provider for specific content
func (cr *ContentRouter) RemoveProvider(contentID string, peerID peer.ID) {
	cr.providersMux.Lock()
	defer cr.providersMux.Unlock()

	providers, exists := cr.providers[contentID]
	if !exists {
		return
	}

	for i, existingProvider := range providers {
		if existingProvider == peerID {
			cr.providers[contentID] = append(providers[:i], providers[i+1:]...)
			break
		}
	}

	// Remove empty provider entries
	if len(cr.providers[contentID]) == 0 {
		delete(cr.providers, contentID)
	}

	cr.routingTable.RemoveRoute(contentID, peerID)
}
