package routing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
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
	host host.Host
	dht  *dht.IpfsDHT

	// Content storage
	contentStore *ContentStore

	// Routing table
	routingTable *RoutingTable

	// Provider management
	providers    map[string][]peer.ID
	providersMux sync.RWMutex

	// Content discovery
	discovery *ContentDiscovery

	// Request tracking
	activeRequests map[string]*ContentRequest
	requestsMux    sync.RWMutex

	// Configuration
	config *ContentRouterConfig

	// Metrics
	metrics *ContentRouterMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ContentRouterConfig holds configuration
type ContentRouterConfig struct {
	MaxProviders      int           `json:"max_providers"`
	ProviderTimeout   time.Duration `json:"provider_timeout"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	CacheSize         int           `json:"cache_size"`
	CacheTTL          time.Duration `json:"cache_ttl"`
	ReplicationFactor int           `json:"replication_factor"`
	EnableCaching     bool          `json:"enable_caching"`
	EnableIndexing    bool          `json:"enable_indexing"`
}

// ContentRouterMetrics tracks routing metrics
type ContentRouterMetrics struct {
	ContentPublished int
	ContentRequests  int
	ContentProvided  int
	CacheHits        int
	CacheMisses      int
	SuccessfulRoutes int
	FailedRoutes     int
	AverageLatency   time.Duration
	StartTime        time.Time
}

// ContentRequest represents a content request
type ContentRequest struct {
	ID              string
	ContentID       string
	RequestorID     peer.ID
	Priority        int
	Timeout         time.Duration
	Status          RequestStatus
	CreatedAt       time.Time
	CompletedAt     *time.Time
	Error           string
	ResponseChannel chan *ContentResponseMessage
}

// RequestStatus represents the status of a content request
type RequestStatus int

const (
	RequestStatusPending RequestStatus = iota
	RequestStatusInProgress
	RequestStatusCompleted
	RequestStatusFailed
	RequestStatusCancelled
)

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
	MessageID string    `json:"message_id"`
	Success   bool      `json:"success"`
	Routes    []peer.ID `json:"routes"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
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
	ContentID string        `json:"content_id"`
	Provider  peer.ID       `json:"provider"`
	TTL       time.Duration `json:"ttl"`
	Timestamp time.Time     `json:"timestamp"`
}

// NewContentRouter creates a new content router
func NewContentRouter(ctx context.Context, host host.Host, dht *dht.IpfsDHT, config *ContentRouterConfig) (*ContentRouter, error) {
	if config == nil {
		config = DefaultContentRouterConfig()
	}

	routerCtx, cancel := context.WithCancel(ctx)

	cr := &ContentRouter{
		host:           host,
		dht:            dht,
		providers:      make(map[string][]peer.ID),
		activeRequests: make(map[string]*ContentRequest),
		config:         config,
		metrics:        &ContentRouterMetrics{StartTime: time.Now()},
		ctx:            routerCtx,
		cancel:         cancel,
	}

	// Initialize content store
	var err error
	cr.contentStore, err = NewContentStore(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create content store: %w", err)
	}

	// Initialize routing table
	cr.routingTable = NewRoutingTable(host.ID())

	// Initialize discovery
	cr.discovery = NewContentDiscovery(host, dht, config)

	// Set up protocol handlers
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

	// Start background tasks
	cr.wg.Add(3)
	go cr.metricsUpdateTask()
	go cr.cacheCleanupTask()
	go cr.requestTimeoutTask()

	log.Printf("Content router started")
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

// PublishContent publishes content to the network
func (cr *ContentRouter) PublishContent(ctx context.Context, content *ContentMetadata) error {
	// Store locally
	cr.contentStore.StoreLocal(content)

	// Create CID for DHT
	hash, err := multihash.Sum([]byte(content.ID), multihash.SHA2_256, -1)
	if err != nil {
		return fmt.Errorf("failed to create hash: %w", err)
	}

	contentCID := cid.NewCidV1(cid.Raw, hash)

	// Provide to DHT
	err = cr.dht.Provide(ctx, contentCID, true)
	if err != nil {
		return fmt.Errorf("failed to provide content to DHT: %w", err)
	}

	// Update metrics
	cr.metrics.ContentPublished++

	log.Printf("Published content: %s", content.ID)
	return nil
}

// RequestContent requests content from the network
func (cr *ContentRouter) RequestContent(ctx context.Context, contentID string, priority int) (*ContentRequest, error) {
	// Check if content is available locally
	if content, exists := cr.contentStore.GetLocal(contentID); exists {
		// Return immediately if local
		response := &ContentResponseMessage{
			RequestID:   generateRequestID(),
			Success:     true,
			ContentMeta: content,
			Timestamp:   time.Now(),
		}

		request := &ContentRequest{
			ID:              response.RequestID,
			ContentID:       contentID,
			RequestorID:     cr.host.ID(),
			Priority:        priority,
			Status:          RequestStatusCompleted,
			CreatedAt:       time.Now(),
			CompletedAt:     &response.Timestamp,
			ResponseChannel: make(chan *ContentResponseMessage, 1),
		}

		request.ResponseChannel <- response
		return request, nil
	}

	// Check cache
	if _, exists := cr.contentStore.GetCached(contentID); exists {
		cr.metrics.CacheHits++
		// TODO: Implement cache retrieval logic
	} else {
		cr.metrics.CacheMisses++
	}

	// Create request
	request := &ContentRequest{
		ID:              generateRequestID(),
		ContentID:       contentID,
		RequestorID:     cr.host.ID(),
		Priority:        priority,
		Timeout:         cr.config.RequestTimeout,
		Status:          RequestStatusPending,
		CreatedAt:       time.Now(),
		ResponseChannel: make(chan *ContentResponseMessage, 1),
	}

	// Track request
	cr.requestsMux.Lock()
	cr.activeRequests[request.ID] = request
	cr.requestsMux.Unlock()

	// Start content discovery
	go cr.discoverAndRetrieveContent(ctx, request)

	cr.metrics.ContentRequests++
	return request, nil
}

// discoverAndRetrieveContent discovers and retrieves content
func (cr *ContentRouter) discoverAndRetrieveContent(ctx context.Context, request *ContentRequest) {
	request.Status = RequestStatusInProgress

	// Create CID for DHT lookup
	hash, err := multihash.Sum([]byte(request.ContentID), multihash.SHA2_256, -1)
	if err != nil {
		cr.completeRequestWithError(request, fmt.Errorf("failed to create hash: %w", err))
		return
	}

	contentCID := cid.NewCidV1(cid.Raw, hash)

	// Find providers
	providers, err := cr.dht.FindProviders(ctx, contentCID)
	if err != nil {
		cr.completeRequestWithError(request, fmt.Errorf("failed to find providers: %w", err))
		return
	}

	if len(providers) == 0 {
		cr.completeRequestWithError(request, fmt.Errorf("no providers found for content: %s", request.ContentID))
		return
	}

	foundProvider := providers[0].ID

	// Request content from provider
	content, err := cr.requestContentFromPeer(ctx, foundProvider, request.ContentID)
	if err != nil {
		cr.completeRequestWithError(request, fmt.Errorf("failed to retrieve content: %w", err))
		return
	}

	// Complete request successfully
	response := &ContentResponseMessage{
		RequestID:   request.ID,
		Success:     true,
		ContentMeta: content,
		Timestamp:   time.Now(),
	}

	cr.completeRequest(request, response)
}

// requestContentFromPeer requests content from a specific peer
func (cr *ContentRouter) requestContentFromPeer(ctx context.Context, peerID peer.ID, contentID string) (*ContentMetadata, error) {
	// Open stream to peer
	stream, err := cr.host.NewStream(ctx, peerID, ContentRequestProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Send request
	requestMsg := &ContentRequestMessage{
		RequestID: generateRequestID(),
		ContentID: contentID,
		Priority:  1,
		Timestamp: time.Now(),
	}

	encoder := json.NewEncoder(stream)
	if err := encoder.Encode(requestMsg); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	var response ContentResponseMessage
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("request failed: %s", response.Message)
	}

	return response.ContentMeta, nil
}

// completeRequest completes a request successfully
func (cr *ContentRouter) completeRequest(request *ContentRequest, response *ContentResponseMessage) {
	now := time.Now()
	request.Status = RequestStatusCompleted
	request.CompletedAt = &now

	// Send response
	select {
	case request.ResponseChannel <- response:
	default:
	}

	// Update metrics
	cr.metrics.SuccessfulRoutes++

	// Remove from active requests
	cr.requestsMux.Lock()
	delete(cr.activeRequests, request.ID)
	cr.requestsMux.Unlock()
}

// completeRequestWithError completes a request with an error
func (cr *ContentRouter) completeRequestWithError(request *ContentRequest, err error) {
	now := time.Now()
	request.Status = RequestStatusFailed
	request.CompletedAt = &now
	request.Error = err.Error()

	// Send error response
	response := &ContentResponseMessage{
		RequestID: request.ID,
		Success:   false,
		Message:   err.Error(),
		Timestamp: now,
	}

	select {
	case request.ResponseChannel <- response:
	default:
	}

	// Update metrics
	cr.metrics.FailedRoutes++

	// Remove from active requests
	cr.requestsMux.Lock()
	delete(cr.activeRequests, request.ID)
	cr.requestsMux.Unlock()
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

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
