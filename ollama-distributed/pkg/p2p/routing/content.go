package routing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

const (
	// Content routing protocols
	ContentRoutingProtocol = protocol.ID("/ollamacron/content-routing/1.0.0")
	ContentRequestProtocol = protocol.ID("/ollamacron/content-request/1.0.0")
	ContentProvideProtocol = protocol.ID("/ollamacron/content-provide/1.0.0")
	
	// DHT content keys
	ContentKeyPrefix = "/ollamacron/content/"
	ModelKeyPrefix   = "/ollamacron/models/"
	DataKeyPrefix    = "/ollamacron/data/"
)

// ContentRouter manages content routing and discovery
type ContentRouter struct {
	host           host.Host
	dht            *dht.IpfsDHT
	
	// Content storage
	contentStore   *ContentStore
	
	// Routing table
	routingTable   *RoutingTable
	
	// Provider management
	providers      map[string][]peer.ID
	providersMux   sync.RWMutex
	
	// Content discovery
	discovery      *ContentDiscovery
	
	// Request tracking
	activeRequests map[string]*ContentRequest
	requestsMux    sync.RWMutex
	
	// Configuration
	config         *ContentRouterConfig
	
	// Metrics
	metrics        *ContentRouterMetrics
	
	// Lifecycle
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// ContentStore manages local and remote content references
type ContentStore struct {
	// Local content
	localContent   map[string]*ContentMetadata
	localMux       sync.RWMutex
	
	// Remote references
	remoteContent  map[string]*RemoteContent
	remoteMux      sync.RWMutex
	
	// Cache management
	cache          *ContentCache
	
	// Storage backend
	storage        Storage
	
	// Indexing
	index          *ContentIndex
}

// ContentMetadata represents content metadata
type ContentMetadata struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	Size           int64             `json:"size"`
	Checksum       string            `json:"checksum"`
	
	// Model-specific metadata
	ModelType      string            `json:"model_type"`
	Architecture   string            `json:"architecture"`
	Parameters     int64             `json:"parameters"`
	Quantization   string            `json:"quantization"`
	
	// Availability
	Providers      []peer.ID         `json:"providers"`
	Replicas       int               `json:"replicas"`
	
	// Access control
	AccessLevel    string            `json:"access_level"`
	RequiredAuth   bool              `json:"required_auth"`
	
	// Versioning
	Version        string            `json:"version"`
	ParentID       string            `json:"parent_id,omitempty"`
	
	// Timestamps
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	LastAccessed   time.Time         `json:"last_accessed"`
	
	// Tags and labels
	Tags           map[string]string `json:"tags"`
	Labels         []string          `json:"labels"`
}

// RemoteContent represents remote content reference
type RemoteContent struct {
	Metadata       *ContentMetadata
	Providers      []peer.ID
	LastUpdated    time.Time
	RetrievalCost  int64
	Availability   float64
}

// ContentRequest represents a content request
type ContentRequest struct {
	ID             string
	ContentID      string
	RequestorID    peer.ID
	Priority       int
	Timeout        time.Duration
	CreatedAt      time.Time
	Status         RequestStatus
	Progress       float64
	Providers      []peer.ID
	FailedProviders []peer.ID
}

// RequestStatus represents request status
type RequestStatus int

const (
	RequestStatusPending RequestStatus = iota
	RequestStatusActive
	RequestStatusCompleted
	RequestStatusFailed
	RequestStatusCancelled
)

// ContentRouterConfig holds configuration
type ContentRouterConfig struct {
	MaxProviders       int           `json:"max_providers"`
	ProviderTimeout    time.Duration `json:"provider_timeout"`
	RequestTimeout     time.Duration `json:"request_timeout"`
	CacheSize          int           `json:"cache_size"`
	CacheTTL           time.Duration `json:"cache_ttl"`
	ReplicationFactor  int           `json:"replication_factor"`
	EnableCaching      bool          `json:"enable_caching"`
	EnableIndexing     bool          `json:"enable_indexing"`
}

// ContentRouterMetrics tracks routing metrics
type ContentRouterMetrics struct {
	ContentPublished   int
	ContentRequests    int
	ContentProvided    int
	CacheHits          int
	CacheMisses        int
	ProviderQueries    int
	SuccessfulRoutes   int
	FailedRoutes       int
	AverageLatency     time.Duration
	StartTime          time.Time
}

// NewContentRouter creates a new content router
func NewContentRouter(ctx context.Context, host host.Host, dht *dht.IpfsDHT, config *ContentRouterConfig) (*ContentRouter, error) {
	if config == nil {
		config = DefaultContentRouterConfig()
	}
	
	ctx, cancel := context.WithCancel(ctx)
	
	cr := &ContentRouter{
		host:           host,
		dht:            dht,
		providers:      make(map[string][]peer.ID),
		activeRequests: make(map[string]*ContentRequest),
		config:         config,
		metrics: &ContentRouterMetrics{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Initialize content store
	contentStore, err := NewContentStore(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize content store: %w", err)
	}
	cr.contentStore = contentStore
	
	// Initialize routing table
	cr.routingTable = NewRoutingTable(host.ID())
	
	// Initialize content discovery
	cr.discovery = NewContentDiscovery(host, dht, config)
	
	// Setup protocol handlers
	cr.setupProtocolHandlers()
	
	return cr, nil
}

// setupProtocolHandlers sets up protocol handlers
func (cr *ContentRouter) setupProtocolHandlers() {
	cr.host.SetStreamHandler(ContentRoutingProtocol, cr.handleContentRouting)
	cr.host.SetStreamHandler(ContentRequestProtocol, cr.handleContentRequest)
	cr.host.SetStreamHandler(ContentProvideProtocol, cr.handleContentProvide)
}

// Start starts the content router
func (cr *ContentRouter) Start() {
	log.Printf("Starting content router")
	
	// Start discovery
	cr.discovery.Start()
	
	// Start provider management
	cr.wg.Add(1)
	go cr.providerManagementTask()
	
	// Start request processing
	cr.wg.Add(1)
	go cr.requestProcessingTask()
	
	// Start metrics collection
	cr.wg.Add(1)
	go cr.metricsTask()
	
	// Start cache cleanup
	cr.wg.Add(1)
	go cr.cacheCleanupTask()
	
	log.Printf("Content router started")
}

// PublishContent publishes content to the network
func (cr *ContentRouter) PublishContent(ctx context.Context, content *ContentMetadata) error {
	log.Printf("Publishing content: %s", content.ID)
	
	// Store content locally
	cr.contentStore.StoreLocal(content)
	
	// Announce to DHT
	key := fmt.Sprintf("%s%s", ContentKeyPrefix, content.ID)
	data, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}
	
	if err := cr.dht.PutValue(ctx, key, data); err != nil {
		return fmt.Errorf("failed to publish to DHT: %w", err)
	}
	
	// Become provider
	contentHash, err := cr.calculateContentHash(content.ID)
	if err != nil {
		return fmt.Errorf("failed to calculate content hash: %w", err)
	}
	
	if err := cr.dht.Provide(ctx, contentHash, true); err != nil {
		return fmt.Errorf("failed to announce as provider: %w", err)
	}
	
	// Update routing table
	cr.routingTable.AddRoute(content.ID, cr.host.ID())
	
	// Update metrics
	cr.metrics.ContentPublished++
	
	log.Printf("Published content: %s", content.ID)
	return nil
}

// FindContent finds content in the network
func (cr *ContentRouter) FindContent(ctx context.Context, contentID string) (*ContentMetadata, []peer.ID, error) {
	log.Printf("Finding content: %s", contentID)
	
	// Check local store first
	if content, exists := cr.contentStore.GetLocal(contentID); exists {
		return content, []peer.ID{cr.host.ID()}, nil
	}
	
	// Check cache
	if cr.config.EnableCaching {
		if cached, exists := cr.contentStore.GetCached(contentID); exists {
			cr.metrics.CacheHits++
			return cached.Metadata, cached.Providers, nil
		}
		cr.metrics.CacheMisses++
	}
	
	// Query DHT
	key := fmt.Sprintf("%s%s", ContentKeyPrefix, contentID)
	val, err := cr.dht.GetValue(ctx, key)
	if err != nil {
		return nil, nil, fmt.Errorf("content not found in DHT: %w", err)
	}
	
	var content ContentMetadata
	if err := json.Unmarshal(val, &content); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	
	// Find providers
	providers, err := cr.findProviders(ctx, contentID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find providers: %w", err)
	}
	
	// Cache the result
	if cr.config.EnableCaching {
		cr.contentStore.CacheRemote(contentID, &content, providers)
	}
	
	log.Printf("Found content: %s with %d providers", contentID, len(providers))
	return &content, providers, nil
}

// findProviders finds providers for content
func (cr *ContentRouter) findProviders(ctx context.Context, contentID string) ([]peer.ID, error) {
	contentHash, err := cr.calculateContentHash(contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate content hash: %w", err)
	}
	
	providersChan := cr.dht.FindProvidersAsync(ctx, contentHash, cr.config.MaxProviders)
	
	var providerIDs []peer.ID
	for provider := range providersChan {
		providerIDs = append(providerIDs, provider.ID)
	}
	
	cr.metrics.ProviderQueries++
	return providerIDs, nil
}

// RequestContent requests content from the network
func (cr *ContentRouter) RequestContent(ctx context.Context, contentID string, priority int) (*ContentRequest, error) {
	log.Printf("Requesting content: %s with priority %d", contentID, priority)
	
	// Create request
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	request := &ContentRequest{
		ID:          requestID,
		ContentID:   contentID,
		RequestorID: cr.host.ID(),
		Priority:    priority,
		Timeout:     cr.config.RequestTimeout,
		CreatedAt:   time.Now(),
		Status:      RequestStatusPending,
		Progress:    0,
	}
	
	// Store request
	cr.requestsMux.Lock()
	cr.activeRequests[requestID] = request
	cr.requestsMux.Unlock()
	
	// Find providers
	providers, err := cr.findProviders(ctx, contentID)
	if err != nil {
		request.Status = RequestStatusFailed
		return request, fmt.Errorf("failed to find providers: %w", err)
	}
	
	request.Providers = providers
	request.Status = RequestStatusActive
	
	// Start request processing
	go cr.processContentRequest(request)
	
	cr.metrics.ContentRequests++
	return request, nil
}

// processContentRequest processes a content request
func (cr *ContentRouter) processContentRequest(request *ContentRequest) {
	log.Printf("Processing content request: %s", request.ID)
	
	ctx, cancel := context.WithTimeout(cr.ctx, request.Timeout)
	defer cancel()
	
	// Try providers in order
	for _, providerID := range request.Providers {
		if cr.ctx.Err() != nil {
			request.Status = RequestStatusCancelled
			return
		}
		
		// Skip failed providers
		if cr.isFailedProvider(request, providerID) {
			continue
		}
		
		// Request from provider
		if err := cr.requestFromProvider(ctx, request, providerID); err != nil {
			log.Printf("Failed to request from provider %s: %v", providerID, err)
			request.FailedProviders = append(request.FailedProviders, providerID)
			continue
		}
		
		// Success
		request.Status = RequestStatusCompleted
		request.Progress = 1.0
		cr.metrics.SuccessfulRoutes++
		log.Printf("Completed content request: %s", request.ID)
		return
	}
	
	// All providers failed
	request.Status = RequestStatusFailed
	cr.metrics.FailedRoutes++
	log.Printf("Failed content request: %s", request.ID)
}

// requestFromProvider requests content from a specific provider
func (cr *ContentRouter) requestFromProvider(ctx context.Context, request *ContentRequest, providerID peer.ID) error {
	// Create stream to provider
	stream, err := cr.host.NewStream(ctx, providerID, ContentRequestProtocol)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()
	
	// Send request
	requestMsg := &ContentRequestMessage{
		RequestID: request.ID,
		ContentID: request.ContentID,
		Priority:  request.Priority,
		Timestamp: time.Now(),
	}
	
	data, err := json.Marshal(requestMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	
	// Read response
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	var response ContentResponseMessage
	if err := json.Unmarshal(buf[:n], &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("provider rejected request: %s", response.Message)
	}
	
	// Update progress
	request.Progress = 0.5
	
	// Handle content transfer (simplified)
	// In a real implementation, this would involve chunked transfers
	if response.ContentData != nil {
		// Process content data
		request.Progress = 1.0
	}
	
	return nil
}

// Protocol handlers

// handleContentRouting handles content routing requests
func (cr *ContentRouter) handleContentRouting(stream network.Stream) {
	defer stream.Close()
	
	// Read routing request
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		log.Printf("Failed to read routing request: %v", err)
		return
	}
	
	var routingMsg RoutingMessage
	if err := json.Unmarshal(buf[:n], &routingMsg); err != nil {
		log.Printf("Failed to unmarshal routing request: %v", err)
		return
	}
	
	// Process routing request
	response := cr.processRoutingRequest(&routingMsg)
	
	// Send response
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal routing response: %v", err)
		return
	}
	
	if _, err := stream.Write(data); err != nil {
		log.Printf("Failed to send routing response: %v", err)
		return
	}
}

// handleContentRequest handles incoming content requests
func (cr *ContentRouter) handleContentRequest(stream network.Stream) {
	defer stream.Close()
	
	peerID := stream.Conn().RemotePeer()
	
	// Read request
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		log.Printf("Failed to read content request: %v", err)
		return
	}
	
	var request ContentRequestMessage
	if err := json.Unmarshal(buf[:n], &request); err != nil {
		log.Printf("Failed to unmarshal content request: %v", err)
		return
	}
	
	// Process request
	response := cr.processIncomingContentRequest(&request, peerID)
	
	// Send response
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal content response: %v", err)
		return
	}
	
	if _, err := stream.Write(data); err != nil {
		log.Printf("Failed to send content response: %v", err)
		return
	}
	
	cr.metrics.ContentProvided++
}

// handleContentProvide handles content provision announcements
func (cr *ContentRouter) handleContentProvide(stream network.Stream) {
	defer stream.Close()
	
	peerID := stream.Conn().RemotePeer()
	
	// Read provision announcement
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		log.Printf("Failed to read provision announcement: %v", err)
		return
	}
	
	var announcement ProvisionAnnouncement
	if err := json.Unmarshal(buf[:n], &announcement); err != nil {
		log.Printf("Failed to unmarshal provision announcement: %v", err)
		return
	}
	
	// Process announcement
	cr.processProvisionAnnouncement(&announcement, peerID)
}

// processRoutingRequest processes a routing request
func (cr *ContentRouter) processRoutingRequest(msg *RoutingMessage) *RoutingResponse {
	// Find routes for the requested content
	routes := cr.routingTable.FindRoutes(msg.ContentID)
	
	return &RoutingResponse{
		MessageID: msg.MessageID,
		Success:   len(routes) > 0,
		Routes:    routes,
		Timestamp: time.Now(),
	}
}

// processIncomingContentRequest processes an incoming content request from a peer
func (cr *ContentRouter) processIncomingContentRequest(request *ContentRequestMessage, requesterID peer.ID) *ContentResponseMessage {
	// Check if we have the content
	content, exists := cr.contentStore.GetLocal(request.ContentID)
	if !exists {
		return &ContentResponseMessage{
			RequestID: request.RequestID,
			Success:   false,
			Message:   "Content not found",
			Timestamp: time.Now(),
		}
	}
	
	// TODO: Implement actual content transfer
	// For now, just return metadata
	return &ContentResponseMessage{
		RequestID:   request.RequestID,
		Success:     true,
		ContentMeta: content,
		Timestamp:   time.Now(),
	}
}

// processProvisionAnnouncement processes a provision announcement
func (cr *ContentRouter) processProvisionAnnouncement(announcement *ProvisionAnnouncement, providerID peer.ID) {
	cr.providersMux.Lock()
	defer cr.providersMux.Unlock()
	
	// Add provider to the list
	providers := cr.providers[announcement.ContentID]
	for _, p := range providers {
		if p == providerID {
			return // Already in list
		}
	}
	
	cr.providers[announcement.ContentID] = append(providers, providerID)
	log.Printf("Added provider %s for content %s", providerID, announcement.ContentID)
}

// Background tasks

// providerManagementTask manages provider information
func (cr *ContentRouter) providerManagementTask() {
	defer cr.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.updateProviderInfo()
		}
	}
}

// requestProcessingTask processes pending requests
func (cr *ContentRouter) requestProcessingTask() {
	defer cr.wg.Done()
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.processTimeouts()
		}
	}
}

// metricsTask collects metrics
func (cr *ContentRouter) metricsTask() {
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

// cacheCleanupTask cleans up expired cache entries
func (cr *ContentRouter) cacheCleanupTask() {
	defer cr.wg.Done()
	
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.contentStore.CleanupCache()
		}
	}
}

// Utility methods

// calculateContentHash calculates hash for content ID
func (cr *ContentRouter) calculateContentHash(contentID string) (cid.Cid, error) {
	// Simple hash calculation - in production use proper content addressing
	mh, err := multihash.Sum([]byte(contentID), multihash.SHA2_256, -1)
	if err != nil {
		return cid.Cid{}, err
	}
	return cid.NewCidV1(cid.Raw, mh), nil
}

// isFailedProvider checks if provider has failed recently
func (cr *ContentRouter) isFailedProvider(request *ContentRequest, providerID peer.ID) bool {
	for _, failed := range request.FailedProviders {
		if failed == providerID {
			return true
		}
	}
	return false
}

// updateProviderInfo updates provider information
func (cr *ContentRouter) updateProviderInfo() {
	// TODO: Implement provider health checking
}

// processTimeouts processes request timeouts
func (cr *ContentRouter) processTimeouts() {
	cr.requestsMux.Lock()
	defer cr.requestsMux.Unlock()
	
	now := time.Now()
	for id, request := range cr.activeRequests {
		if now.Sub(request.CreatedAt) > request.Timeout {
			request.Status = RequestStatusFailed
			delete(cr.activeRequests, id)
			log.Printf("Request timeout: %s", id)
		}
	}
}

// updateMetrics updates router metrics
func (cr *ContentRouter) updateMetrics() {
	// TODO: Implement metrics collection
}

// GetMetrics returns router metrics
func (cr *ContentRouter) GetMetrics() *ContentRouterMetrics {
	return cr.metrics
}

// GetActiveRequests returns active requests
func (cr *ContentRouter) GetActiveRequests() map[string]*ContentRequest {
	cr.requestsMux.RLock()
	defer cr.requestsMux.RUnlock()
	
	requests := make(map[string]*ContentRequest)
	for k, v := range cr.activeRequests {
		requests[k] = v
	}
	return requests
}

// Stop stops the content router
func (cr *ContentRouter) Stop() {
	log.Printf("Stopping content router")
	cr.cancel()
	cr.wg.Wait()
	
	if cr.discovery != nil {
		cr.discovery.Stop()
	}
	
	log.Printf("Content router stopped")
}

// DefaultContentRouterConfig returns default configuration
func DefaultContentRouterConfig() *ContentRouterConfig {
	return &ContentRouterConfig{
		MaxProviders:      10,
		ProviderTimeout:   30 * time.Second,
		RequestTimeout:    60 * time.Second,
		CacheSize:         1000,
		CacheTTL:          10 * time.Minute,
		ReplicationFactor: 3,
		EnableCaching:     true,
		EnableIndexing:    true,
	}
}

// Message types

// RoutingMessage represents a routing message
type RoutingMessage struct {
	MessageID string    `json:"message_id"`
	ContentID string    `json:"content_id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// RoutingResponse represents a routing response
type RoutingResponse struct {
	MessageID string     `json:"message_id"`
	Success   bool       `json:"success"`
	Routes    []peer.ID  `json:"routes"`
	Message   string     `json:"message,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// ContentRequestMessage represents a content request
type ContentRequestMessage struct {
	RequestID string    `json:"request_id"`
	ContentID string    `json:"content_id"`
	Priority  int       `json:"priority"`
	Timestamp time.Time `json:"timestamp"`
}

// ContentResponseMessage represents a content response
type ContentResponseMessage struct {
	RequestID   string           `json:"request_id"`
	Success     bool             `json:"success"`
	ContentMeta *ContentMetadata `json:"content_meta,omitempty"`
	ContentData []byte           `json:"content_data,omitempty"`
	Message     string           `json:"message,omitempty"`
	Timestamp   time.Time        `json:"timestamp"`
}

// ProvisionAnnouncement represents a provision announcement
type ProvisionAnnouncement struct {
	ContentID string    `json:"content_id"`
	Provider  peer.ID   `json:"provider"`
	TTL       time.Duration `json:"ttl"`
	Timestamp time.Time `json:"timestamp"`
}

// RoutingTable manages peer routing information
type RoutingTable struct {
	localPeerID peer.ID
	routes      map[string][]peer.ID
	mu          sync.RWMutex
}

// NewRoutingTable creates a new routing table
func NewRoutingTable(localPeerID peer.ID) *RoutingTable {
	return &RoutingTable{
		localPeerID: localPeerID,
		routes:      make(map[string][]peer.ID),
	}
}

// AddRoute adds a route to the routing table
func (rt *RoutingTable) AddRoute(contentID string, peerID peer.ID) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	
	routes := rt.routes[contentID]
	for _, p := range routes {
		if p == peerID {
			return // Already exists
		}
	}
	rt.routes[contentID] = append(routes, peerID)
}

// FindRoutes finds routes for a content ID
func (rt *RoutingTable) FindRoutes(contentID string) []peer.ID {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	
	routes := rt.routes[contentID]
	result := make([]peer.ID, len(routes))
	copy(result, routes)
	return result
}

// ContentDiscovery manages content discovery in the network
type ContentDiscovery struct {
	host   host.Host
	dht    *dht.IpfsDHT
	config *ContentRouterConfig
	ctx    context.Context
	cancel context.CancelFunc
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

// Start starts the content discovery service
func (cd *ContentDiscovery) Start() {
	// TODO: Implement discovery service
}

// Stop stops the content discovery service
func (cd *ContentDiscovery) Stop() {
	cd.cancel()
}

// ContentCache manages cached content
type ContentCache struct {
	cache    map[string]*CacheEntry
	maxSize  int
	ttl      time.Duration
	mu       sync.RWMutex
}

// CacheEntry represents a cache entry
type CacheEntry struct {
	Content   *RemoteContent
	ExpiresAt time.Time
}

// NewContentCache creates a new content cache
func NewContentCache(maxSize int, ttl time.Duration) *ContentCache {
	return &ContentCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get retrieves content from cache
func (cc *ContentCache) Get(contentID string) (*RemoteContent, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	entry, exists := cc.cache[contentID]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Content, true
}

// Put stores content in cache
func (cc *ContentCache) Put(contentID string, content *RemoteContent) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	cc.cache[contentID] = &CacheEntry{
		Content:   content,
		ExpiresAt: time.Now().Add(cc.ttl),
	}
	
	// Simple eviction if cache is full
	if len(cc.cache) > cc.maxSize {
		// Remove oldest entry
		var oldestID string
		var oldestTime time.Time
		for id, entry := range cc.cache {
			if oldestTime.IsZero() || entry.ExpiresAt.Before(oldestTime) {
				oldestID = id
				oldestTime = entry.ExpiresAt
			}
		}
		delete(cc.cache, oldestID)
	}
}

// Cleanup removes expired entries
func (cc *ContentCache) Cleanup() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	now := time.Now()
	for id, entry := range cc.cache {
		if now.After(entry.ExpiresAt) {
			delete(cc.cache, id)
		}
	}
}

// Storage interface for content storage
type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error
	Has(key string) bool
}

// MemoryStorage implements in-memory storage
type MemoryStorage struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewMemoryStorage creates a new memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]byte),
	}
}

// Get retrieves data from storage
func (ms *MemoryStorage) Get(key string) ([]byte, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	data, exists := ms.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return data, nil
}

// Put stores data in storage
func (ms *MemoryStorage) Put(key string, value []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	ms.data[key] = value
	return nil
}

// Delete removes data from storage
func (ms *MemoryStorage) Delete(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	delete(ms.data, key)
	return nil
}

// Has checks if key exists in storage
func (ms *MemoryStorage) Has(key string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	_, exists := ms.data[key]
	return exists
}

// ContentIndex manages content indexing
type ContentIndex struct {
	index map[string]*IndexEntry
	mu    sync.RWMutex
}

// IndexEntry represents an index entry
type IndexEntry struct {
	ContentID string
	Metadata  *ContentMetadata
	Keywords  []string
	UpdatedAt time.Time
}

// NewContentIndex creates a new content index
func NewContentIndex() *ContentIndex {
	return &ContentIndex{
		index: make(map[string]*IndexEntry),
	}
}

// Add adds content to the index
func (ci *ContentIndex) Add(contentID string, metadata *ContentMetadata) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	
	ci.index[contentID] = &IndexEntry{
		ContentID: contentID,
		Metadata:  metadata,
		Keywords:  extractKeywords(metadata),
		UpdatedAt: time.Now(),
	}
}

// Search searches the index
func (ci *ContentIndex) Search(query string) []*IndexEntry {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	
	var results []*IndexEntry
	for _, entry := range ci.index {
		if matchesQuery(entry, query) {
			results = append(results, entry)
		}
	}
	return results
}

// extractKeywords extracts keywords from metadata
func extractKeywords(metadata *ContentMetadata) []string {
	keywords := []string{metadata.Name, metadata.Type, metadata.ModelType}
	keywords = append(keywords, metadata.Labels...)
	return keywords
}

// matchesQuery checks if entry matches query
func matchesQuery(entry *IndexEntry, query string) bool {
	// Simple substring matching
	for _, keyword := range entry.Keywords {
		if keyword == query {
			return true
		}
	}
	return false
}

// NewContentStore creates a new content store
func NewContentStore(config *ContentRouterConfig) (*ContentStore, error) {
	cache := NewContentCache(config.CacheSize, config.CacheTTL)
	storage := NewMemoryStorage()
	index := NewContentIndex()
	
	return &ContentStore{
		localContent:  make(map[string]*ContentMetadata),
		remoteContent: make(map[string]*RemoteContent),
		cache:         cache,
		storage:       storage,
		index:         index,
	}, nil
}

// StoreLocal stores content locally
func (cs *ContentStore) StoreLocal(content *ContentMetadata) {
	cs.localMux.Lock()
	defer cs.localMux.Unlock()
	
	cs.localContent[content.ID] = content
	if cs.index != nil {
		cs.index.Add(content.ID, content)
	}
}

// GetLocal retrieves local content
func (cs *ContentStore) GetLocal(contentID string) (*ContentMetadata, bool) {
	cs.localMux.RLock()
	defer cs.localMux.RUnlock()
	
	content, exists := cs.localContent[contentID]
	return content, exists
}

// CacheRemote caches remote content
func (cs *ContentStore) CacheRemote(contentID string, metadata *ContentMetadata, providers []peer.ID) {
	cs.remoteMux.Lock()
	defer cs.remoteMux.Unlock()
	
	remote := &RemoteContent{
		Metadata:     metadata,
		Providers:    providers,
		LastUpdated:  time.Now(),
		Availability: float64(len(providers)) / 10.0, // Simple availability metric
	}
	
	cs.remoteContent[contentID] = remote
	if cs.cache != nil {
		cs.cache.Put(contentID, remote)
	}
}

// GetCached retrieves cached content
func (cs *ContentStore) GetCached(contentID string) (*RemoteContent, bool) {
	if cs.cache != nil {
		return cs.cache.Get(contentID)
	}
	
	cs.remoteMux.RLock()
	defer cs.remoteMux.RUnlock()
	
	content, exists := cs.remoteContent[contentID]
	return content, exists
}

// CleanupCache cleans up the cache
func (cs *ContentStore) CleanupCache() {
	if cs.cache != nil {
		cs.cache.Cleanup()
	}
}