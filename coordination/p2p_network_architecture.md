# P2P Network Architecture for Ollamacron

## Executive Summary

This document presents the comprehensive P2P network architecture design for Ollamacron, a distributed AI model execution platform. The architecture leverages libp2p Go implementation patterns optimized for 2025, focusing on robust NAT traversal, intelligent peer discovery, secure communication channels, and efficient resource advertisement.

## 1. Core Architecture Overview

### 1.1 Network Topology
- **Hybrid Topology**: Combines mesh and hierarchical structures
- **Adaptive Mesh**: Dynamic mesh formation based on node capabilities
- **Super Nodes**: High-capacity nodes serving as regional coordinators
- **Edge Nodes**: Resource-constrained nodes for lightweight operations

### 1.2 Key Components
1. **Peer Discovery Engine**: Multi-strategy discovery with DHT integration
2. **NAT Traversal System**: Hole punching + relay infrastructure
3. **Secure Channel Framework**: Multi-layered security with noise protocol
4. **Resource Advertisement Protocol**: Efficient capability broadcasting
5. **Content Routing Engine**: DHT-based model and data distribution

## 2. LibP2P Go Implementation Design

### 2.1 Host Configuration

```go
// Enhanced Node Configuration
type NodeConfig struct {
    // Network Settings
    Listen              []string            `yaml:"listen"`
    AnnounceAddresses   []string            `yaml:"announce_addresses"`
    NoAnnounceAddresses []string            `yaml:"no_announce_addresses"`
    
    // Security
    PrivateKey          string              `yaml:"private_key"`
    EnableTLS           bool                `yaml:"enable_tls"`
    EnableNoise         bool                `yaml:"enable_noise"`
    
    // NAT Traversal
    EnableNATService    bool                `yaml:"enable_nat_service"`
    EnableHolePunching  bool                `yaml:"enable_hole_punching"`
    EnableAutoRelay     bool                `yaml:"enable_auto_relay"`
    StaticRelays        []string            `yaml:"static_relays"`
    ForceReachability   string              `yaml:"force_reachability"` // public/private
    
    // DHT Settings
    EnableDHT           bool                `yaml:"enable_dht"`
    DHTMode             string              `yaml:"dht_mode"` // client/server/auto
    BootstrapPeers      []string            `yaml:"bootstrap_peers"`
    
    // Connection Management
    ConnMgrLow          int                 `yaml:"conn_mgr_low"`
    ConnMgrHigh         int                 `yaml:"conn_mgr_high"`
    ConnMgrGrace        time.Duration       `yaml:"conn_mgr_grace"`
    
    // Resource Management
    MaxMemory           int64               `yaml:"max_memory"`
    MaxCPU              float64             `yaml:"max_cpu"`
    MaxGPU              int                 `yaml:"max_gpu"`
    
    // Ollamacron Specific
    NodeType            string              `yaml:"node_type"` // edge/standard/super
    ModelCapabilities   []string            `yaml:"model_capabilities"`
    ResourceTags        map[string]string   `yaml:"resource_tags"`
}
```

### 2.2 Enhanced Host Initialization

```go
func (n *Node) initHost() error {
    // Load or generate private key
    priv, err := n.loadOrGenerateKey()
    if err != nil {
        return fmt.Errorf("failed to initialize key: %w", err)
    }
    
    // Build listen addresses
    listenAddrs := make([]multiaddr.Multiaddr, 0, len(n.config.Listen))
    for _, addr := range n.config.Listen {
        maddr, err := multiaddr.NewMultiaddr(addr)
        if err != nil {
            continue
        }
        listenAddrs = append(listenAddrs, maddr)
    }
    
    // Configure transports
    transports := []libp2p.Option{
        libp2p.Transport(tcp.NewTCPTransport),
        libp2p.Transport(ws.New),
        libp2p.Transport(webtransport.New),
    }
    
    // Configure security
    security := []libp2p.Option{
        libp2p.Security(noise.ID, noise.New),
        libp2p.Security(tls.ID, tls.New),
    }
    
    // Configure NAT traversal
    natOptions := []libp2p.Option{
        libp2p.EnableNATService(),
        libp2p.EnableHolePunching(),
    }
    
    if n.config.EnableAutoRelay {
        natOptions = append(natOptions, libp2p.EnableAutoRelayWithStaticRelays(
            n.parseStaticRelays(),
        ))
    }
    
    // Configure connection manager
    connMgr, err := connmgr.NewConnManager(
        n.config.ConnMgrLow,
        n.config.ConnMgrHigh,
        connmgr.WithGracePeriod(n.config.ConnMgrGrace),
    )
    if err != nil {
        return fmt.Errorf("failed to create connection manager: %w", err)
    }
    
    // Build host options
    opts := []libp2p.Option{
        libp2p.Identity(priv),
        libp2p.ListenAddrs(listenAddrs...),
        libp2p.ConnectionManager(connMgr),
        libp2p.EnableRelay(),
    }
    
    // Add transport options
    opts = append(opts, transports...)
    opts = append(opts, security...)
    opts = append(opts, natOptions...)
    
    // Create host
    host, err := libp2p.New(opts...)
    if err != nil {
        return fmt.Errorf("failed to create libp2p host: %w", err)
    }
    
    n.host = host
    return nil
}
```

## 3. Peer Discovery Mechanism

### 3.1 Multi-Strategy Discovery

```go
type DiscoveryEngine struct {
    host        host.Host
    dht         *dht.IpfsDHT
    mdns        discovery.Discovery
    bootstrap   *BootstrapDiscovery
    rendezvous  *RendezvousDiscovery
    
    // Discovery strategies
    strategies  []DiscoveryStrategy
    
    // Peer cache
    peerCache   *PeerCache
    
    // Events
    peerFound   chan peer.AddrInfo
    peerLost    chan peer.ID
}

type DiscoveryStrategy interface {
    FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error)
    Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error)
}

// DHT-based discovery
func (d *DiscoveryEngine) startDHTDiscovery() {
    // Use DHT for peer discovery
    routingDiscovery := drouting.NewRoutingDiscovery(d.dht)
    
    // Advertise our capabilities
    go func() {
        for {
            ttl, err := routingDiscovery.Advertise(context.Background(), "ollamacron")
            if err != nil {
                log.Printf("Failed to advertise: %v", err)
                time.Sleep(30 * time.Second)
                continue
            }
            
            select {
            case <-time.After(ttl):
            case <-d.ctx.Done():
                return
            }
        }
    }()
    
    // Discover peers
    go func() {
        for {
            peerChan, err := routingDiscovery.FindPeers(
                context.Background(),
                "ollamacron",
                discovery.Limit(50),
            )
            if err != nil {
                log.Printf("Failed to find peers: %v", err)
                time.Sleep(30 * time.Second)
                continue
            }
            
            for peer := range peerChan {
                select {
                case d.peerFound <- peer:
                case <-d.ctx.Done():
                    return
                }
            }
        }
    }()
}
```

### 3.2 Bootstrap Discovery

```go
type BootstrapDiscovery struct {
    host            host.Host
    bootstrapPeers  []peer.AddrInfo
    minPeers        int
    maxPeers        int
    
    // Connection tracking
    connections     map[peer.ID]*ConnectionInfo
    connectionsMux  sync.RWMutex
}

func (b *BootstrapDiscovery) Start(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            b.ensureConnections(ctx)
        }
    }
}

func (b *BootstrapDiscovery) ensureConnections(ctx context.Context) {
    connected := len(b.host.Network().Peers())
    
    if connected < b.minPeers {
        // Connect to more bootstrap peers
        for _, peer := range b.bootstrapPeers {
            if connected >= b.maxPeers {
                break
            }
            
            if b.host.Network().Connectedness(peer.ID) == network.Connected {
                continue
            }
            
            go func(p peer.AddrInfo) {
                ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
                defer cancel()
                
                if err := b.host.Connect(ctx, p); err != nil {
                    log.Printf("Failed to connect to bootstrap peer %s: %v", p.ID, err)
                }
            }(peer)
            
            connected++
        }
    }
}
```

## 4. NAT Traversal Strategy

### 4.1 Advanced NAT Traversal

```go
type NATTraversalManager struct {
    host           host.Host
    autoRelay      *autorelay.AutoRelay
    holePuncher    *holepunch.Service
    natService     *natmgr.NATManager
    
    // Relay management
    relayManager   *RelayManager
    
    // NAT detection
    natDetector    *NATDetector
    
    // Statistics
    stats          *NATStats
}

type RelayManager struct {
    staticRelays   []peer.AddrInfo
    discoveredRelays []peer.AddrInfo
    activeRelays   map[peer.ID]*RelayConnection
    
    // Quality metrics
    relayQuality   map[peer.ID]*RelayQuality
    
    // Selection algorithm
    selector       RelaySelector
}

type RelayQuality struct {
    Latency        time.Duration
    Bandwidth      int64
    Reliability    float64
    LastTested     time.Time
}

func (r *RelayManager) SelectOptimalRelay(ctx context.Context) (peer.ID, error) {
    // Score relays based on quality metrics
    scores := make(map[peer.ID]float64)
    
    for relayID, quality := range r.relayQuality {
        score := r.calculateRelayScore(quality)
        scores[relayID] = score
    }
    
    // Select best relay
    var bestRelay peer.ID
    var bestScore float64
    
    for relayID, score := range scores {
        if score > bestScore {
            bestScore = score
            bestRelay = relayID
        }
    }
    
    if bestRelay == "" {
        return "", fmt.Errorf("no suitable relay found")
    }
    
    return bestRelay, nil
}

func (r *RelayManager) calculateRelayScore(quality *RelayQuality) float64 {
    // Weight factors
    const (
        latencyWeight     = 0.4
        bandwidthWeight   = 0.3
        reliabilityWeight = 0.3
    )
    
    // Normalize metrics
    latencyScore := 1.0 - float64(quality.Latency.Milliseconds())/1000.0
    bandwidthScore := math.Min(float64(quality.Bandwidth)/1000000, 1.0)
    reliabilityScore := quality.Reliability
    
    // Calculate weighted score
    return latencyWeight*latencyScore + 
           bandwidthWeight*bandwidthScore + 
           reliabilityWeight*reliabilityScore
}
```

### 4.2 NAT Detection and Classification

```go
type NATDetector struct {
    host           host.Host
    natService     *natmgr.NATManager
    
    // Detection results
    natType        NATType
    externalAddr   multiaddr.Multiaddr
    lastDetection  time.Time
    
    // Detection strategies
    strategies     []NATDetectionStrategy
}

type NATType int

const (
    NATTypeUnknown NATType = iota
    NATTypeNone
    NATTypeFullCone
    NATTypeRestrictedCone
    NATTypePortRestricted
    NATTypeSymmetric
)

func (n *NATDetector) DetectNATType(ctx context.Context) (NATType, error) {
    // Use multiple detection strategies
    results := make([]NATType, 0, len(n.strategies))
    
    for _, strategy := range n.strategies {
        result, err := strategy.Detect(ctx)
        if err != nil {
            continue
        }
        results = append(results, result)
    }
    
    // Aggregate results
    natType := n.aggregateResults(results)
    n.natType = natType
    n.lastDetection = time.Now()
    
    return natType, nil
}
```

## 5. Secure Channel Architecture

### 5.1 Multi-Layer Security

```go
type SecurityManager struct {
    host           host.Host
    
    // Key management
    keyManager     *KeyManager
    
    // Protocol handlers
    protocols      map[protocol.ID]*SecureProtocol
    
    // Authentication
    authManager    *AuthManager
    
    // Encryption
    encryptionMgr  *EncryptionManager
}

type SecureProtocol struct {
    ID             protocol.ID
    Handler        network.StreamHandler
    
    // Security settings
    RequireAuth    bool
    RequireEncryption bool
    
    // Rate limiting
    RateLimit      *RateLimit
    
    // Access control
    AccessControl  *AccessControl
}

type KeyManager struct {
    // Node identity
    privateKey     crypto.PrivKey
    publicKey      crypto.PubKey
    
    // Session keys
    sessionKeys    map[peer.ID]*SessionKey
    sessionKeysMux sync.RWMutex
    
    // Key rotation
    rotationInterval time.Duration
    lastRotation     time.Time
}

type SessionKey struct {
    Key            []byte
    CreatedAt      time.Time
    ExpiresAt      time.Time
    Algorithm      string
}

func (s *SecurityManager) EstablishSecureChannel(ctx context.Context, peerID peer.ID) (*SecureChannel, error) {
    // Create new stream
    stream, err := s.host.NewStream(ctx, peerID, SecureChannelProtocol)
    if err != nil {
        return nil, fmt.Errorf("failed to create stream: %w", err)
    }
    
    // Perform handshake
    sessionKey, err := s.performHandshake(ctx, stream)
    if err != nil {
        stream.Close()
        return nil, fmt.Errorf("handshake failed: %w", err)
    }
    
    // Create secure channel
    channel := &SecureChannel{
        Stream:     stream,
        SessionKey: sessionKey,
        Peer:       peerID,
    }
    
    return channel, nil
}

func (s *SecurityManager) performHandshake(ctx context.Context, stream network.Stream) (*SessionKey, error) {
    // Implement noise protocol handshake
    handshake := noise.NewHandshake(s.keyManager.privateKey)
    
    // Exchange keys and establish session
    sessionKey, err := handshake.Execute(ctx, stream)
    if err != nil {
        return nil, fmt.Errorf("handshake execution failed: %w", err)
    }
    
    return sessionKey, nil
}
```

## 6. Resource Advertisement Protocol

### 6.1 Capability Broadcasting

```go
type ResourceAdvertiser struct {
    host           host.Host
    dht            *dht.IpfsDHT
    
    // Resource information
    capabilities   *NodeCapabilities
    resources      *ResourceMetrics
    
    // Advertisement management
    advertisements map[string]*Advertisement
    advMux         sync.RWMutex
    
    // Update channels
    capabilityUpdates chan *NodeCapabilities
    resourceUpdates   chan *ResourceMetrics
}

type NodeCapabilities struct {
    // Compute resources
    CPUCores       int              `json:"cpu_cores"`
    Memory         int64            `json:"memory"`
    Storage        int64            `json:"storage"`
    GPUs           []*GPUInfo       `json:"gpus"`
    
    // AI capabilities
    SupportedModels []string         `json:"supported_models"`
    ModelFormats    []string         `json:"model_formats"`
    Quantizations   []string         `json:"quantizations"`
    
    // Network capabilities
    Bandwidth      int64            `json:"bandwidth"`
    Latency        time.Duration    `json:"latency"`
    Reliability    float64          `json:"reliability"`
    
    // Availability
    Uptime         time.Duration    `json:"uptime"`
    LoadAverage    float64          `json:"load_average"`
    
    // Pricing
    PricePerToken  float64          `json:"price_per_token"`
    PricePerHour   float64          `json:"price_per_hour"`
}

type GPUInfo struct {
    Model          string           `json:"model"`
    Memory         int64            `json:"memory"`
    ComputeCapability string        `json:"compute_capability"`
    Utilization    float64          `json:"utilization"`
}

func (r *ResourceAdvertiser) Start(ctx context.Context) {
    // Start periodic advertisement
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            r.advertiseResources(ctx)
        case caps := <-r.capabilityUpdates:
            r.capabilities = caps
            r.advertiseResources(ctx)
        case metrics := <-r.resourceUpdates:
            r.resources = metrics
            r.advertiseResources(ctx)
        }
    }
}

func (r *ResourceAdvertiser) advertiseResources(ctx context.Context) {
    // Create advertisement
    ad := &Advertisement{
        NodeID:       r.host.ID(),
        Capabilities: r.capabilities,
        Resources:    r.resources,
        Timestamp:    time.Now(),
        TTL:          5 * time.Minute,
    }
    
    // Serialize advertisement
    data, err := json.Marshal(ad)
    if err != nil {
        log.Printf("Failed to serialize advertisement: %v", err)
        return
    }
    
    // Store in DHT
    key := fmt.Sprintf("/ollamacron/resources/%s", r.host.ID())
    if err := r.dht.PutValue(ctx, key, data); err != nil {
        log.Printf("Failed to store advertisement in DHT: %v", err)
        return
    }
    
    // Broadcast to interested peers
    r.broadcastAdvertisement(ctx, ad)
}
```

### 6.2 Resource Discovery

```go
type ResourceDiscovery struct {
    host           host.Host
    dht            *dht.IpfsDHT
    
    // Resource cache
    resourceCache  *ResourceCache
    
    // Query interface
    queries        chan *ResourceQuery
    responses      chan *ResourceResponse
    
    // Subscription management
    subscriptions  map[string]*ResourceSubscription
    subMux         sync.RWMutex
}

type ResourceQuery struct {
    // Filter criteria
    ModelTypes     []string          `json:"model_types"`
    MinCPU         int               `json:"min_cpu"`
    MinMemory      int64             `json:"min_memory"`
    RequiredGPU    bool              `json:"required_gpu"`
    MaxLatency     time.Duration     `json:"max_latency"`
    MaxPrice       float64           `json:"max_price"`
    
    // Geographic preferences
    PreferredRegions []string         `json:"preferred_regions"`
    
    // Response options
    MaxResults     int               `json:"max_results"`
    SortBy         string            `json:"sort_by"`
}

func (r *ResourceDiscovery) FindResources(ctx context.Context, query *ResourceQuery) ([]*NodeCapabilities, error) {
    // Check cache first
    cached := r.resourceCache.Find(query)
    if len(cached) > 0 {
        return cached, nil
    }
    
    // Query DHT
    results := make([]*NodeCapabilities, 0)
    
    // Search by model types
    for _, modelType := range query.ModelTypes {
        key := fmt.Sprintf("/ollamacron/models/%s", modelType)
        
        vals, err := r.dht.GetValues(ctx, key, 16)
        if err != nil {
            continue
        }
        
        for _, val := range vals {
            var caps NodeCapabilities
            if err := json.Unmarshal(val, &caps); err != nil {
                continue
            }
            
            if r.matchesQuery(&caps, query) {
                results = append(results, &caps)
            }
        }
    }
    
    // Sort results
    r.sortResults(results, query.SortBy)
    
    // Limit results
    if query.MaxResults > 0 && len(results) > query.MaxResults {
        results = results[:query.MaxResults]
    }
    
    // Cache results
    r.resourceCache.Store(query, results)
    
    return results, nil
}
```

## 7. DHT-Based Content Routing

### 7.1 Content Routing Engine

```go
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
}

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
}

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
    
    CreatedAt      time.Time         `json:"created_at"`
    UpdatedAt      time.Time         `json:"updated_at"`
}

func (c *ContentRouter) PublishContent(ctx context.Context, content *ContentMetadata) error {
    // Store content locally
    c.contentStore.StoreLocal(content)
    
    // Announce to DHT
    key := fmt.Sprintf("/ollamacron/content/%s", content.ID)
    data, err := json.Marshal(content)
    if err != nil {
        return fmt.Errorf("failed to marshal content: %w", err)
    }
    
    if err := c.dht.PutValue(ctx, key, data); err != nil {
        return fmt.Errorf("failed to publish to DHT: %w", err)
    }
    
    // Become provider
    if err := c.dht.Provide(ctx, cid.Cid(content.ID), true); err != nil {
        return fmt.Errorf("failed to announce as provider: %w", err)
    }
    
    return nil
}

func (c *ContentRouter) FindContent(ctx context.Context, contentID string) (*ContentMetadata, []peer.ID, error) {
    // Check local store first
    if content, exists := c.contentStore.GetLocal(contentID); exists {
        return content, []peer.ID{c.host.ID()}, nil
    }
    
    // Query DHT
    key := fmt.Sprintf("/ollamacron/content/%s", contentID)
    val, err := c.dht.GetValue(ctx, key)
    if err != nil {
        return nil, nil, fmt.Errorf("content not found in DHT: %w", err)
    }
    
    var content ContentMetadata
    if err := json.Unmarshal(val, &content); err != nil {
        return nil, nil, fmt.Errorf("failed to unmarshal content: %w", err)
    }
    
    // Find providers
    providers, err := c.dht.FindProviders(ctx, cid.Cid(contentID))
    if err != nil {
        return nil, nil, fmt.Errorf("failed to find providers: %w", err)
    }
    
    var providerIDs []peer.ID
    for provider := range providers {
        providerIDs = append(providerIDs, provider.ID)
    }
    
    return &content, providerIDs, nil
}
```

## 8. Performance Optimization

### 8.1 Connection Management

```go
type ConnectionManager struct {
    host           host.Host
    connMgr        connmgr.ConnManager
    
    // Connection strategies
    strategies     []ConnectionStrategy
    
    // Peer scoring
    peerScorer     *PeerScorer
    
    // Quality metrics
    connections    map[peer.ID]*ConnectionMetrics
    connMux        sync.RWMutex
}

type ConnectionMetrics struct {
    Latency        time.Duration
    Bandwidth      int64
    Reliability    float64
    LastActive     time.Time
    BytesTransferred int64
    ErrorRate      float64
}

type PeerScorer struct {
    scores         map[peer.ID]*PeerScore
    scoresMux      sync.RWMutex
    
    // Scoring algorithms
    algorithms     []ScoringAlgorithm
}

type PeerScore struct {
    Overall        float64
    Latency        float64
    Reliability    float64
    Throughput     float64
    Availability   float64
    
    LastUpdated    time.Time
}

func (p *PeerScorer) UpdateScore(peerID peer.ID, metrics *ConnectionMetrics) {
    p.scoresMux.Lock()
    defer p.scoresMux.Unlock()
    
    score := &PeerScore{
        LastUpdated: time.Now(),
    }
    
    // Calculate individual scores
    score.Latency = p.calculateLatencyScore(metrics.Latency)
    score.Reliability = metrics.Reliability
    score.Throughput = p.calculateThroughputScore(metrics.Bandwidth)
    score.Availability = p.calculateAvailabilityScore(metrics.LastActive)
    
    // Calculate overall score
    score.Overall = (score.Latency + score.Reliability + score.Throughput + score.Availability) / 4.0
    
    p.scores[peerID] = score
}
```

### 8.2 Load Balancing

```go
type LoadBalancer struct {
    peers          map[peer.ID]*PeerInfo
    peersMux       sync.RWMutex
    
    // Load balancing strategies
    strategy       LoadBalancingStrategy
    
    // Health monitoring
    healthChecker  *HealthChecker
    
    // Request routing
    router         *RequestRouter
}

type LoadBalancingStrategy interface {
    SelectPeer(peers []*PeerInfo, request *Request) (peer.ID, error)
}

type RoundRobinStrategy struct {
    counter        int64
}

func (r *RoundRobinStrategy) SelectPeer(peers []*PeerInfo, request *Request) (peer.ID, error) {
    if len(peers) == 0 {
        return "", fmt.Errorf("no peers available")
    }
    
    index := atomic.AddInt64(&r.counter, 1) % int64(len(peers))
    return peers[index].ID, nil
}

type WeightedStrategy struct {
    weights        map[peer.ID]float64
    weightsMux     sync.RWMutex
}

func (w *WeightedStrategy) SelectPeer(peers []*PeerInfo, request *Request) (peer.ID, error) {
    w.weightsMux.RLock()
    defer w.weightsMux.RUnlock()
    
    // Calculate total weight
    totalWeight := 0.0
    for _, peer := range peers {
        weight := w.weights[peer.ID]
        totalWeight += weight
    }
    
    if totalWeight == 0 {
        return "", fmt.Errorf("no weighted peers available")
    }
    
    // Random selection based on weights
    r := rand.Float64() * totalWeight
    currentWeight := 0.0
    
    for _, peer := range peers {
        currentWeight += w.weights[peer.ID]
        if r <= currentWeight {
            return peer.ID, nil
        }
    }
    
    return peers[len(peers)-1].ID, nil
}
```

## 9. Implementation Timeline

### Phase 1: Core Infrastructure (Weeks 1-4)
- [ ] Enhanced libp2p host configuration
- [ ] Multi-strategy peer discovery
- [ ] Basic NAT traversal with hole punching
- [ ] Secure channel establishment

### Phase 2: Resource Management (Weeks 5-8)
- [ ] Resource advertisement protocol
- [ ] Content routing engine
- [ ] DHT-based discovery
- [ ] Basic load balancing

### Phase 3: Advanced Features (Weeks 9-12)
- [ ] Advanced NAT traversal strategies
- [ ] Intelligent relay selection
- [ ] Performance optimization
- [ ] Monitoring and analytics

### Phase 4: Production Readiness (Weeks 13-16)
- [ ] Security hardening
- [ ] Scalability testing
- [ ] Documentation and examples
- [ ] Integration testing

## 10. Testing Strategy

### 10.1 Unit Testing
- Component isolation testing
- Protocol compliance testing
- Security mechanism validation
- Performance benchmarking

### 10.2 Integration Testing
- Multi-node network simulation
- NAT traversal scenarios
- Failure recovery testing
- Load balancing validation

### 10.3 Performance Testing
- Latency measurement
- Throughput analysis
- Scalability limits
- Resource utilization monitoring

## 11. Monitoring and Observability

### 11.1 Metrics Collection
- Network topology metrics
- Connection quality metrics
- Resource utilization metrics
- Performance indicators

### 11.2 Alerting
- Connection failures
- NAT traversal issues
- Resource exhaustion
- Security incidents

## 12. Security Considerations

### 12.1 Authentication
- Peer identity verification
- Certificate-based authentication
- Token-based access control
- Rate limiting

### 12.2 Encryption
- End-to-end encryption
- Forward secrecy
- Key rotation
- Secure key exchange

### 12.3 Attack Mitigation
- DDoS protection
- Sybil attack prevention
- Eclipse attack mitigation
- Resource exhaustion protection

## Conclusion

This comprehensive P2P network architecture provides a robust foundation for Ollamacron's distributed AI model execution platform. The design leverages modern libp2p patterns, implements intelligent NAT traversal, and provides efficient resource discovery and content routing capabilities. The modular architecture allows for incremental implementation and future enhancements while maintaining security and performance requirements.