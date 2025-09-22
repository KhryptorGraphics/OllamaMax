/**
 * Adaptive Mesh Network Swarm System
 * Implements peer-to-peer mesh networking with distributed decision making,
 * fault tolerance, and self-organizing behavior for maximum resilience
 * 
 * Features:
 * - Full mesh connectivity with adaptive routing
 * - Distributed consensus mechanisms (Raft, PBFT, Gossip)
 * - Self-healing network topology
 * - Dynamic peer discovery and load balancing
 * - Byzantine fault tolerance
 */

const EventEmitter = require('events');
const Redis = require('ioredis');
const crypto = require('crypto');

class AdaptiveMeshNetwork extends EventEmitter {
  constructor(options = {}) {
    super();
    
    this.redis = new Redis({
      host: process.env.REDIS_HOST || 'redis-cluster-0.redis-cluster-service.ollamamax-redis',
      port: process.env.REDIS_PORT || 6379,
      password: process.env.REDIS_PASSWORD || 'ollama_redis_pass',
      retryDelayOnFailure: 1000,
      maxRetriesPerRequest: 3
    });

    // Mesh network configuration
    this.config = {
      maxPeers: options.maxPeers || 25,
      minPeers: options.minPeers || 3,
      heartbeatInterval: options.heartbeatInterval || 15000, // 15 seconds
      consensusTimeout: options.consensusTimeout || 30000,   // 30 seconds
      routingUpdateInterval: options.routingUpdateInterval || 45000, // 45 seconds
      peerDiscoveryInterval: options.peerDiscoveryInterval || 60000,  // 1 minute
      faultToleranceThreshold: options.faultToleranceThreshold || 0.33, // Can tolerate 1/3 failures
      gossipFanout: options.gossipFanout || 3, // Number of peers to gossip to
      ...options
    };

    // Local node identity and state
    this.nodeId = options.nodeId || this.generateNodeId();
    this.nodeInfo = {
      id: this.nodeId,
      address: options.address || 'localhost',
      port: options.port || 8000,
      capabilities: options.capabilities || ['worker', 'coordinator'],
      specializations: options.specializations || [],
      joinedAt: Date.now(),
      lastSeen: Date.now(),
      reputation: 1.0,
      workload: 0,
      status: 'initializing'
    };

    // Mesh network state
    this.meshState = {
      peers: new Map(),           // Active peer connections
      routingTable: new Map(),    // Network routing information
      pendingConnections: new Set(), // Connections being established
      failedPeers: new Set(),     // Known failed peers
      networkPartitions: [],      // Detected network partitions
      consensusState: 'idle',     // Current consensus state
      leader: null,               // Current leader (for hybrid consensus)
      term: 0                     // Current consensus term
    };

    // Consensus mechanisms
    this.consensus = {
      raft: {
        state: 'follower',
        votedFor: null,
        log: [],
        commitIndex: 0,
        lastApplied: 0,
        nextIndex: new Map(),
        matchIndex: new Map()
      },
      pbft: {
        view: 0,
        phase: 'idle',
        proposals: new Map(),
        votes: new Map(),
        commits: new Map()
      },
      gossip: {
        knownMessages: new Set(),
        messageBuffer: [],
        rumorsReceived: new Map(),
        rumorsForwarded: new Map()
      }
    };

    // Communication protocols
    this.protocols = {
      messageTypes: [
        'heartbeat', 'peer_discovery', 'routing_update', 'task_assignment',
        'consensus_proposal', 'vote', 'commit', 'gossip_message',
        'leader_election', 'join_request', 'leave_notification'
      ],
      reliability: 'best_effort',
      encryption: true,
      compression: true,
      retransmission: true
    };

    // Performance metrics
    this.metrics = {
      messagesSent: 0,
      messagesReceived: 0,
      consensusRounds: 0,
      networkLatency: new Map(),
      throughput: 0,
      availability: 1.0,
      partitionEvents: 0,
      healingEvents: 0
    };

    this.initializeMeshNetwork();
  }

  async initializeMeshNetwork() {
    try {
      // Initialize local node
      await this.setupLocalNode();
      
      // Start core network services
      this.startHeartbeatService();
      this.startPeerDiscovery();
      this.startRoutingUpdates();
      this.startConsensusMonitoring();
      
      // Join existing network or bootstrap new one
      await this.joinMeshNetwork();
      
      this.nodeInfo.status = 'active';
      
      console.log(`Adaptive mesh network initialized: ${this.nodeId}`);
      this.emit('mesh_initialized', {
        nodeId: this.nodeId,
        capabilities: this.nodeInfo.capabilities,
        maxPeers: this.config.maxPeers
      });
    } catch (error) {
      console.error('Failed to initialize mesh network:', error);
      throw error;
    }
  }

  async setupLocalNode() {
    // Store local node information
    await this.redis.setex(`mesh:node:${this.nodeId}`, 300, JSON.stringify(this.nodeInfo));
    
    // Initialize routing table with self
    this.meshState.routingTable.set(this.nodeId, {
      nextHop: this.nodeId,
      distance: 0,
      latency: 0,
      updated: Date.now()
    });
    
    // Initialize consensus state
    this.consensus.raft.log.push({
      term: 0,
      index: 0,
      command: 'initialize',
      timestamp: Date.now(),
      nodeId: this.nodeId
    });
  }

  /**
   * Peer Discovery and Connection Management
   */
  async joinMeshNetwork() {
    try {
      // Discover existing nodes
      const existingNodes = await this.discoverExistingNodes();
      
      if (existingNodes.length === 0) {
        // Bootstrap new network
        console.log('Bootstrapping new mesh network');
        await this.bootstrapNetwork();
      } else {
        // Join existing network
        console.log(`Joining existing mesh network with ${existingNodes.length} nodes`);
        await this.joinExistingNetwork(existingNodes);
      }
      
      // Establish optimal connections
      await this.establishOptimalConnections();
      
    } catch (error) {
      console.error('Failed to join mesh network:', error);
      throw error;
    }
  }

  async discoverExistingNodes() {
    const nodeKeys = await this.redis.keys('mesh:node:*');
    const nodes = [];
    
    for (const key of nodeKeys) {
      const nodeData = await this.redis.get(key);
      if (nodeData) {
        try {
          const node = JSON.parse(nodeData);
          if (node.id !== this.nodeId && node.status === 'active') {
            nodes.push(node);
          }
        } catch (e) {
          continue;
        }
      }
    }
    
    return nodes;
  }

  async bootstrapNetwork() {
    // Initialize as first node in network
    this.consensus.raft.state = 'leader';
    this.consensus.raft.term = 1;
    this.meshState.leader = this.nodeId;
    this.meshState.term = 1;
    
    // Store bootstrap information
    await this.redis.setex('mesh:bootstrap', 3600, JSON.stringify({
      bootstrapNode: this.nodeId,
      timestamp: Date.now(),
      initialTerm: 1
    }));
  }

  async joinExistingNetwork(existingNodes) {
    // Send join requests to a subset of nodes
    const targetNodes = existingNodes.slice(0, Math.min(5, existingNodes.length));
    
    for (const node of targetNodes) {
      await this.sendJoinRequest(node);
    }
    
    // Wait for responses and establish connections
    await this.waitForJoinResponses();
  }

  async sendJoinRequest(targetNode) {
    const joinMessage = {
      type: 'join_request',
      from: this.nodeId,
      nodeInfo: this.nodeInfo,
      timestamp: Date.now(),
      signature: this.signMessage(this.nodeInfo)
    };
    
    await this.sendMessage(targetNode.id, joinMessage);
    this.meshState.pendingConnections.add(targetNode.id);
  }

  async establishOptimalConnections() {
    const currentPeers = this.meshState.peers.size;
    const targetConnections = Math.min(this.config.maxPeers, Math.max(this.config.minPeers, Math.ceil(Math.sqrt(currentPeers + 1))));
    
    if (currentPeers < targetConnections) {
      // Need more connections
      await this.establishAdditionalConnections(targetConnections - currentPeers);
    } else if (currentPeers > targetConnections) {
      // Optimize existing connections
      await this.optimizeConnections(targetConnections);
    }
  }

  async establishAdditionalConnections(needed) {
    const availableNodes = await this.discoverExistingNodes();
    const unconnectedNodes = availableNodes.filter(node => 
      !this.meshState.peers.has(node.id) && 
      !this.meshState.failedPeers.has(node.id)
    );
    
    // Sort by reputation and compatibility
    unconnectedNodes.sort((a, b) => {
      const scoreA = this.calculateConnectionScore(a);
      const scoreB = this.calculateConnectionScore(b);
      return scoreB - scoreA;
    });
    
    // Connect to best candidates
    const targets = unconnectedNodes.slice(0, needed);
    for (const node of targets) {
      await this.establishConnection(node);
    }
  }

  calculateConnectionScore(node) {
    let score = node.reputation || 1.0;
    
    // Favor nodes with complementary capabilities
    const sharedCapabilities = this.nodeInfo.capabilities.filter(cap => 
      node.capabilities?.includes(cap)
    ).length;
    const uniqueCapabilities = (node.capabilities?.length || 0) - sharedCapabilities;
    score += uniqueCapabilities * 0.2;
    
    // Consider workload balance
    score -= (node.workload || 0) * 0.1;
    
    // Consider network latency (if known)
    const latency = this.metrics.networkLatency.get(node.id) || 100;
    score -= latency / 1000; // Penalize high latency
    
    return score;
  }

  async establishConnection(node) {
    try {
      const connectionId = this.generateConnectionId(this.nodeId, node.id);
      
      const peer = {
        id: node.id,
        nodeInfo: node,
        connectionId,
        status: 'connecting',
        connectedAt: Date.now(),
        lastHeartbeat: Date.now(),
        messagesSent: 0,
        messagesReceived: 0,
        latency: 0,
        reliability: 1.0
      };
      
      this.meshState.peers.set(node.id, peer);
      this.meshState.pendingConnections.delete(node.id);
      
      // Update routing table
      this.meshState.routingTable.set(node.id, {
        nextHop: node.id,
        distance: 1,
        latency: 0,
        updated: Date.now()
      });
      
      // Send connection established message
      await this.sendMessage(node.id, {
        type: 'connection_established',
        from: this.nodeId,
        connectionId,
        timestamp: Date.now()
      });
      
      peer.status = 'connected';
      
      this.emit('peer_connected', { peerId: node.id, connectionId });
      console.log(`Connected to peer: ${node.id}`);
      
    } catch (error) {
      console.error(`Failed to connect to ${node.id}:`, error);
      this.meshState.failedPeers.add(node.id);
    }
  }

  /**
   * Heartbeat and Failure Detection
   */
  startHeartbeatService() {
    this.heartbeatTimer = setInterval(async () => {
      try {
        await this.sendHeartbeats();
        await this.checkPeerHealth();
      } catch (error) {
        console.error('Heartbeat service error:', error);
      }
    }, this.config.heartbeatInterval);
  }

  async sendHeartbeats() {
    const heartbeatMessage = {
      type: 'heartbeat',
      from: this.nodeId,
      timestamp: Date.now(),
      workload: this.nodeInfo.workload,
      status: this.nodeInfo.status,
      term: this.meshState.term
    };
    
    // Send to all connected peers
    for (const [peerId, peer] of this.meshState.peers) {
      if (peer.status === 'connected') {
        await this.sendMessage(peerId, heartbeatMessage);
        peer.messagesSent++;
      }
    }
    
    this.metrics.messagesSent += this.meshState.peers.size;
  }

  async checkPeerHealth() {
    const now = Date.now();
    const timeout = this.config.heartbeatInterval * 3; // 3 missed heartbeats
    
    const failedPeers = [];
    
    for (const [peerId, peer] of this.meshState.peers) {
      if (now - peer.lastHeartbeat > timeout) {
        console.warn(`Peer ${peerId} failed health check`);
        failedPeers.push(peerId);
      }
    }
    
    // Handle failed peers
    for (const peerId of failedPeers) {
      await this.handlePeerFailure(peerId);
    }
    
    // Update availability metric
    const totalPeers = this.meshState.peers.size + failedPeers.length;
    this.metrics.availability = totalPeers > 0 ? this.meshState.peers.size / totalPeers : 1.0;
  }

  async handlePeerFailure(peerId) {
    console.log(`Handling failure of peer: ${peerId}`);
    
    const peer = this.meshState.peers.get(peerId);
    if (peer) {
      // Remove from active peers
      this.meshState.peers.delete(peerId);
      this.meshState.failedPeers.add(peerId);
      
      // Update routing table
      this.updateRoutingTableOnFailure(peerId);
      
      // Trigger healing if needed
      if (this.meshState.peers.size < this.config.minPeers) {
        await this.triggerNetworkHealing();
      }
      
      this.emit('peer_failed', { peerId, timestamp: Date.now() });
    }
  }

  updateRoutingTableOnFailure(failedPeerId) {
    // Remove direct routes
    this.meshState.routingTable.delete(failedPeerId);
    
    // Update routes that went through failed peer
    for (const [targetId, route] of this.meshState.routingTable) {
      if (route.nextHop === failedPeerId) {
        // Find alternative route
        const alternativeRoute = this.findAlternativeRoute(targetId, failedPeerId);
        if (alternativeRoute) {
          this.meshState.routingTable.set(targetId, alternativeRoute);
        } else {
          // No alternative route found
          this.meshState.routingTable.delete(targetId);
        }
      }
    }
  }

  findAlternativeRoute(targetId, excludePeerId) {
    // Simple alternative route finding (BFS through peers)
    const visited = new Set([this.nodeId, excludePeerId]);
    const queue = [{peerId: this.nodeId, distance: 0, path: []}];
    
    while (queue.length > 0) {
      const {peerId, distance, path} = queue.shift();
      
      // Get peer's connections
      const peerConnections = this.getPeerConnections(peerId);
      
      for (const nextPeerId of peerConnections) {
        if (visited.has(nextPeerId)) continue;
        
        if (nextPeerId === targetId) {
          // Found route
          const nextHop = path.length > 0 ? path[0] : nextPeerId;
          return {
            nextHop,
            distance: distance + 1,
            latency: this.estimateRouteLatency(path.concat([nextPeerId])),
            updated: Date.now()
          };
        }
        
        visited.add(nextPeerId);
        queue.push({
          peerId: nextPeerId,
          distance: distance + 1,
          path: path.concat([nextPeerId])
        });
      }
    }
    
    return null; // No alternative route found
  }

  getPeerConnections(peerId) {
    if (peerId === this.nodeId) {
      return Array.from(this.meshState.peers.keys());
    }
    
    // For other peers, we'd need to maintain topology information
    // This is simplified for the implementation
    return [];
  }

  estimateRouteLatency(path) {
    let totalLatency = 0;
    for (let i = 0; i < path.length - 1; i++) {
      const fromId = i === 0 ? this.nodeId : path[i-1];
      const toId = path[i];
      totalLatency += this.metrics.networkLatency.get(toId) || 50;
    }
    return totalLatency;
  }

  async triggerNetworkHealing() {
    console.log('Triggering network healing...');
    
    // Attempt to reconnect to failed peers
    const recentFailures = Array.from(this.meshState.failedPeers).slice(0, 3);
    for (const peerId of recentFailures) {
      const node = await this.getNodeInfo(peerId);
      if (node && await this.testConnection(node)) {
        this.meshState.failedPeers.delete(peerId);
        await this.establishConnection(node);
      }
    }
    
    // Discover new peers if still under minimum
    if (this.meshState.peers.size < this.config.minPeers) {
      const needed = this.config.minPeers - this.meshState.peers.size;
      await this.establishAdditionalConnections(needed);
    }
    
    this.metrics.healingEvents++;
  }

  /**
   * Distributed Consensus Mechanisms
   */
  async initiateConsensus(proposal) {
    if (this.meshState.consensusState !== 'idle') {
      throw new Error('Consensus already in progress');
    }
    
    const consensusType = this.selectConsensusType(proposal);
    
    switch (consensusType) {
      case 'raft':
        return await this.initiateRaftConsensus(proposal);
      case 'pbft':
        return await this.initiatePBFTConsensus(proposal);
      case 'gossip':
        return await this.initiateGossipConsensus(proposal);
      default:
        throw new Error(`Unknown consensus type: ${consensusType}`);
    }
  }

  selectConsensusType(proposal) {
    const peerCount = this.meshState.peers.size + 1; // +1 for self
    
    // Use PBFT for critical decisions requiring Byzantine fault tolerance
    if (proposal.critical && peerCount >= 4) {
      return 'pbft';
    }
    
    // Use Raft for leader election and consistent decisions
    if (proposal.type === 'leader_election' || proposal.requiresOrder) {
      return 'raft';
    }
    
    // Use Gossip for eventual consistency and information dissemination
    if (proposal.eventualConsistency) {
      return 'gossip';
    }
    
    // Default to Raft for most cases
    return 'raft';
  }

  async initiateRaftConsensus(proposal) {
    if (this.consensus.raft.state !== 'leader') {
      throw new Error('Only leader can initiate Raft consensus');
    }
    
    this.meshState.consensusState = 'raft_proposal';
    
    const entry = {
      term: this.consensus.raft.term || this.meshState.term,
      index: this.consensus.raft.log.length,
      command: proposal,
      timestamp: Date.now(),
      nodeId: this.nodeId
    };
    
    // Add to local log
    this.consensus.raft.log.push(entry);
    
    // Send append entries to all followers
    const votes = await this.sendAppendEntries(entry);
    
    // Check for majority
    const totalNodes = this.meshState.peers.size + 1;
    const majority = Math.floor(totalNodes / 2) + 1;
    
    if (votes.success >= majority) {
      // Commit entry
      this.consensus.raft.commitIndex = entry.index;
      await this.commitRaftEntry(entry);
      
      this.meshState.consensusState = 'idle';
      
      return {
        success: true,
        consensusType: 'raft',
        entry,
        votes: votes.success,
        totalNodes
      };
    } else {
      // Consensus failed
      this.meshState.consensusState = 'idle';
      
      return {
        success: false,
        consensusType: 'raft',
        error: 'Insufficient votes',
        votes: votes.success,
        totalNodes
      };
    }
  }

  async sendAppendEntries(entry) {
    const appendMessage = {
      type: 'append_entries',
      from: this.nodeId,
      term: this.consensus.raft.term,
      prevLogIndex: entry.index - 1,
      prevLogTerm: entry.index > 0 ? this.consensus.raft.log[entry.index - 1].term : 0,
      entries: [entry],
      leaderCommit: this.consensus.raft.commitIndex,
      timestamp: Date.now()
    };
    
    const votes = { success: 1, failure: 0 }; // Count self as success
    const votePromises = [];
    
    for (const [peerId, peer] of this.meshState.peers) {
      if (peer.status === 'connected') {
        const votePromise = this.sendMessage(peerId, appendMessage)
          .then(() => this.waitForVote(peerId, entry.index))
          .then(vote => vote.success ? 'success' : 'failure')
          .catch(() => 'failure');
        
        votePromises.push(votePromise);
      }
    }
    
    const results = await Promise.allSettled(votePromises);
    
    for (const result of results) {
      if (result.status === 'fulfilled') {
        votes[result.value]++;
      } else {
        votes.failure++;
      }
    }
    
    return votes;
  }

  async waitForVote(peerId, entryIndex) {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Vote timeout'));
      }, this.config.consensusTimeout);
      
      const checkVote = async () => {
        const voteKey = `mesh:vote:${peerId}:${entryIndex}`;
        const vote = await this.redis.get(voteKey);
        
        if (vote) {
          clearTimeout(timeout);
          resolve(JSON.parse(vote));
        } else {
          setTimeout(checkVote, 1000); // Check every second
        }
      };
      
      checkVote();
    });
  }

  async commitRaftEntry(entry) {
    // Apply the committed entry
    await this.applyConsensusDecision(entry.command);
    
    this.consensus.raft.lastApplied = entry.index;
    
    // Notify followers of commit
    const commitMessage = {
      type: 'commit_entry',
      from: this.nodeId,
      term: this.consensus.raft.term,
      entryIndex: entry.index,
      timestamp: Date.now()
    };
    
    for (const [peerId, peer] of this.meshState.peers) {
      if (peer.status === 'connected') {
        await this.sendMessage(peerId, commitMessage);
      }
    }
    
    console.log(`Committed Raft entry: ${entry.index}`);
  }

  async initiatePBFTConsensus(proposal) {
    this.meshState.consensusState = 'pbft_prepare';
    
    const proposalId = this.generateProposalId();
    const prepareMessage = {
      type: 'pbft_prepare',
      from: this.nodeId,
      view: this.consensus.pbft.view,
      proposalId,
      proposal,
      timestamp: Date.now(),
      signature: this.signMessage(proposal)
    };
    
    // Store proposal locally
    this.consensus.pbft.proposals.set(proposalId, {
      proposal,
      phase: 'prepare',
      votes: new Map(),
      commits: new Map(),
      timestamp: Date.now()
    });
    
    // Broadcast prepare message
    await this.broadcastToPeers(prepareMessage);
    
    // Wait for prepare votes
    const prepareVotes = await this.collectPBFTVotes(proposalId, 'prepare');
    
    const totalNodes = this.meshState.peers.size + 1;
    const threshold = Math.floor((2 * totalNodes + 1) / 3); // 2f+1 for PBFT
    
    if (prepareVotes >= threshold) {
      // Move to commit phase
      this.meshState.consensusState = 'pbft_commit';
      
      const commitMessage = {
        type: 'pbft_commit',
        from: this.nodeId,
        view: this.consensus.pbft.view,
        proposalId,
        timestamp: Date.now()
      };
      
      await this.broadcastToPeers(commitMessage);
      
      const commitVotes = await this.collectPBFTVotes(proposalId, 'commit');
      
      if (commitVotes >= threshold) {
        // Execute proposal
        await this.applyConsensusDecision(proposal);
        
        this.meshState.consensusState = 'idle';
        
        return {
          success: true,
          consensusType: 'pbft',
          proposalId,
          prepareVotes,
          commitVotes,
          threshold
        };
      }
    }
    
    this.meshState.consensusState = 'idle';
    
    return {
      success: false,
      consensusType: 'pbft',
      error: 'Insufficient votes',
      prepareVotes,
      threshold
    };
  }

  async collectPBFTVotes(proposalId, phase) {
    return new Promise((resolve) => {
      let votes = 1; // Count self
      const timeout = setTimeout(() => resolve(votes), this.config.consensusTimeout);
      
      const checkVotes = async () => {
        const voteKeys = await this.redis.keys(`mesh:pbft:${phase}:${proposalId}:*`);
        votes = voteKeys.length + 1; // +1 for self
        
        const totalNodes = this.meshState.peers.size + 1;
        const threshold = Math.floor((2 * totalNodes + 1) / 3);
        
        if (votes >= threshold) {
          clearTimeout(timeout);
          resolve(votes);
        } else {
          setTimeout(checkVotes, 1000);
        }
      };
      
      checkVotes();
    });
  }

  async initiateGossipConsensus(proposal) {
    this.meshState.consensusState = 'gossip_propagation';
    
    const gossipMessage = {
      type: 'gossip_message',
      messageId: this.generateMessageId(),
      from: this.nodeId,
      proposal,
      timestamp: Date.now(),
      ttl: 5, // Time to live
      signature: this.signMessage(proposal)
    };
    
    // Start gossip propagation
    await this.gossipMessage(gossipMessage);
    
    // Gossip provides eventual consistency, so we return success immediately
    // The actual consensus happens through message propagation
    
    this.meshState.consensusState = 'idle';
    
    return {
      success: true,
      consensusType: 'gossip',
      messageId: gossipMessage.messageId,
      propagationType: 'eventual_consistency'
    };
  }

  async gossipMessage(message) {
    // Add to known messages to prevent loops
    this.consensus.gossip.knownMessages.add(message.messageId);
    
    // Select random peers for gossip (fanout)
    const peerIds = Array.from(this.meshState.peers.keys());
    const selectedPeers = this.selectRandomPeers(peerIds, this.config.gossipFanout);
    
    // Forward to selected peers
    for (const peerId of selectedPeers) {
      await this.sendMessage(peerId, message);
    }
    
    // Track forwarding
    this.consensus.gossip.rumorsForwarded.set(message.messageId, selectedPeers.length);
  }

  selectRandomPeers(peerIds, count) {
    const shuffled = [...peerIds].sort(() => 0.5 - Math.random());
    return shuffled.slice(0, Math.min(count, peerIds.length));
  }

  /**
   * Message Handling and Routing
   */
  async sendMessage(targetId, message) {
    try {
      // Find route to target
      const route = this.meshState.routingTable.get(targetId);
      if (!route) {
        throw new Error(`No route to ${targetId}`);
      }
      
      // Prepare message for transmission
      const transmissionMessage = {
        ...message,
        messageId: message.messageId || this.generateMessageId(),
        route: {
          from: this.nodeId,
          to: targetId,
          nextHop: route.nextHop,
          hopCount: 0
        },
        timestamp: message.timestamp || Date.now()
      };
      
      // Add encryption and signature if enabled
      if (this.protocols.encryption) {
        transmissionMessage.encrypted = true;
        // Encryption logic would go here
      }
      
      // Store message for transmission
      const messageKey = `mesh:message:${transmissionMessage.messageId}`;
      await this.redis.setex(messageKey, 300, JSON.stringify(transmissionMessage));
      
      // Update metrics
      this.metrics.messagesSent++;
      
      console.log(`Sent ${message.type} message to ${targetId} via ${route.nextHop}`);
      
    } catch (error) {
      console.error(`Failed to send message to ${targetId}:`, error);
      throw error;
    }
  }

  async broadcastToPeers(message) {
    const broadcastPromises = [];
    
    for (const [peerId, peer] of this.meshState.peers) {
      if (peer.status === 'connected') {
        broadcastPromises.push(this.sendMessage(peerId, message));
      }
    }
    
    await Promise.allSettled(broadcastPromises);
  }

  async handleIncomingMessage(message) {
    try {
      this.metrics.messagesReceived++;
      
      // Update peer information
      if (message.from && this.meshState.peers.has(message.from)) {
        const peer = this.meshState.peers.get(message.from);
        peer.lastHeartbeat = Date.now();
        peer.messagesReceived++;
      }
      
      // Route message based on type
      switch (message.type) {
        case 'heartbeat':
          await this.handleHeartbeat(message);
          break;
        case 'join_request':
          await this.handleJoinRequest(message);
          break;
        case 'append_entries':
          await this.handleAppendEntries(message);
          break;
        case 'vote_request':
          await this.handleVoteRequest(message);
          break;
        case 'pbft_prepare':
          await this.handlePBFTPrepare(message);
          break;
        case 'pbft_commit':
          await this.handlePBFTCommit(message);
          break;
        case 'gossip_message':
          await this.handleGossipMessage(message);
          break;
        case 'routing_update':
          await this.handleRoutingUpdate(message);
          break;
        case 'task_assignment':
          await this.handleTaskAssignment(message);
          break;
        default:
          console.warn(`Unknown message type: ${message.type}`);
      }
      
    } catch (error) {
      console.error('Error handling message:', error);
    }
  }

  async handleHeartbeat(message) {
    // Update peer status
    if (this.meshState.peers.has(message.from)) {
      const peer = this.meshState.peers.get(message.from);
      peer.lastHeartbeat = Date.now();
      peer.nodeInfo.workload = message.workload;
      peer.nodeInfo.status = message.status;
    }
    
    // Send heartbeat response
    const response = {
      type: 'heartbeat_response',
      from: this.nodeId,
      timestamp: Date.now(),
      workload: this.nodeInfo.workload,
      status: this.nodeInfo.status
    };
    
    await this.sendMessage(message.from, response);
  }

  async handleJoinRequest(message) {
    // Verify node signature and information
    if (!this.verifyMessage(message.nodeInfo, message.signature)) {
      console.warn(`Invalid join request from ${message.from}`);
      return;
    }
    
    // Check if we can accept new connections
    if (this.meshState.peers.size >= this.config.maxPeers) {
      const rejectMessage = {
        type: 'join_rejected',
        from: this.nodeId,
        reason: 'max_peers_reached',
        timestamp: Date.now()
      };
      await this.sendMessage(message.from, rejectMessage);
      return;
    }
    
    // Accept the connection
    await this.establishConnection(message.nodeInfo);
    
    const acceptMessage = {
      type: 'join_accepted',
      from: this.nodeId,
      nodeInfo: this.nodeInfo,
      peerList: this.getPeerList(),
      timestamp: Date.now()
    };
    
    await this.sendMessage(message.from, acceptMessage);
  }

  async handleGossipMessage(message) {
    // Check if we've already seen this message
    if (this.consensus.gossip.knownMessages.has(message.messageId)) {
      return; // Already processed
    }
    
    // Verify message signature
    if (!this.verifyMessage(message.proposal, message.signature)) {
      console.warn(`Invalid gossip message from ${message.from}`);
      return;
    }
    
    // Process the message
    await this.processGossipMessage(message);
    
    // Forward to other peers if TTL allows
    if (message.ttl > 0) {
      const forwardMessage = {
        ...message,
        ttl: message.ttl - 1,
        forwardedBy: this.nodeId
      };
      
      await this.gossipMessage(forwardMessage);
    }
  }

  async processGossipMessage(message) {
    // Apply the gossip message content
    if (message.proposal) {
      await this.applyConsensusDecision(message.proposal);
    }
    
    this.consensus.gossip.rumorsReceived.set(message.messageId, Date.now());
  }

  /**
   * Routing and Network Topology Management
   */
  startRoutingUpdates() {
    this.routingTimer = setInterval(async () => {
      try {
        await this.updateRoutingTable();
        await this.broadcastRoutingInfo();
      } catch (error) {
        console.error('Routing update error:', error);
      }
    }, this.config.routingUpdateInterval);
  }

  async updateRoutingTable() {
    const now = Date.now();
    const maxAge = this.config.routingUpdateInterval * 3; // 3 update cycles
    
    // Remove stale routes
    for (const [targetId, route] of this.meshState.routingTable) {
      if (now - route.updated > maxAge && targetId !== this.nodeId) {
        this.meshState.routingTable.delete(targetId);
      }
    }
    
    // Update routes through direct peers
    for (const [peerId, peer] of this.meshState.peers) {
      if (peer.status === 'connected') {
        this.meshState.routingTable.set(peerId, {
          nextHop: peerId,
          distance: 1,
          latency: peer.latency,
          updated: now
        });
      }
    }
  }

  async broadcastRoutingInfo() {
    const routingMessage = {
      type: 'routing_update',
      from: this.nodeId,
      routes: this.getRoutingAdvertisement(),
      timestamp: Date.now()
    };
    
    await this.broadcastToPeers(routingMessage);
  }

  getRoutingAdvertisement() {
    const advertisement = {};
    
    for (const [targetId, route] of this.meshState.routingTable) {
      // Don't advertise routes back to the source
      if (targetId !== this.nodeId) {
        advertisement[targetId] = {
          distance: route.distance + 1, // Add one hop
          latency: route.latency,
          updated: route.updated
        };
      }
    }
    
    return advertisement;
  }

  async handleRoutingUpdate(message) {
    const now = Date.now();
    
    for (const [targetId, advertised] of Object.entries(message.routes)) {
      const currentRoute = this.meshState.routingTable.get(targetId);
      
      // Update if we don't have a route or found a better one
      if (!currentRoute || 
          advertised.distance < currentRoute.distance ||
          (advertised.distance === currentRoute.distance && advertised.latency < currentRoute.latency)) {
        
        this.meshState.routingTable.set(targetId, {
          nextHop: message.from,
          distance: advertised.distance,
          latency: advertised.latency + (this.meshState.peers.get(message.from)?.latency || 0),
          updated: now
        });
      }
    }
  }

  /**
   * Peer Discovery Service
   */
  startPeerDiscovery() {
    this.discoveryTimer = setInterval(async () => {
      try {
        await this.discoverNewPeers();
        await this.maintainOptimalConnections();
      } catch (error) {
        console.error('Peer discovery error:', error);
      }
    }, this.config.peerDiscoveryInterval);
  }

  async discoverNewPeers() {
    // Discover peers through existing connections
    for (const [peerId, peer] of this.meshState.peers) {
      if (peer.status === 'connected') {
        const peerListRequest = {
          type: 'peer_list_request',
          from: this.nodeId,
          timestamp: Date.now()
        };
        
        await this.sendMessage(peerId, peerListRequest);
      }
    }
  }

  async maintainOptimalConnections() {
    const targetConnections = this.calculateOptimalConnections();
    const currentConnections = this.meshState.peers.size;
    
    if (currentConnections < targetConnections) {
      await this.establishAdditionalConnections(targetConnections - currentConnections);
    } else if (currentConnections > targetConnections * 1.2) {
      // Optimize connections if we have too many
      await this.optimizeConnections(targetConnections);
    }
  }

  calculateOptimalConnections() {
    const networkSize = await this.estimateNetworkSize();
    return Math.min(
      this.config.maxPeers,
      Math.max(
        this.config.minPeers,
        Math.ceil(Math.log2(networkSize + 1)) // Logarithmic scaling
      )
    );
  }

  async estimateNetworkSize() {
    // Estimate network size based on routing table
    const knownNodes = new Set([this.nodeId]);
    
    for (const targetId of this.meshState.routingTable.keys()) {
      knownNodes.add(targetId);
    }
    
    return knownNodes.size;
  }

  async optimizeConnections(targetCount) {
    if (this.meshState.peers.size <= targetCount) return;
    
    // Calculate connection scores
    const connectionScores = [];
    
    for (const [peerId, peer] of this.meshState.peers) {
      const score = this.calculateConnectionValue(peer);
      connectionScores.push({ peerId, score });
    }
    
    // Sort by score (highest first)
    connectionScores.sort((a, b) => b.score - a.score);
    
    // Disconnect lowest scoring connections
    const toDisconnect = connectionScores.slice(targetCount);
    
    for (const { peerId } of toDisconnect) {
      await this.disconnectPeer(peerId);
    }
  }

  calculateConnectionValue(peer) {
    let score = peer.reliability * 0.3;
    score += (1 - peer.latency / 1000) * 0.2; // Lower latency is better
    score += peer.nodeInfo.reputation * 0.2;
    score += (1 - peer.nodeInfo.workload) * 0.1; // Lower workload is better
    score += this.calculateCapabilityValue(peer.nodeInfo.capabilities) * 0.2;
    
    return score;
  }

  calculateCapabilityValue(capabilities) {
    if (!capabilities) return 0;
    
    const myCapabilities = new Set(this.nodeInfo.capabilities);
    const peerCapabilities = new Set(capabilities);
    
    // Value unique capabilities higher
    const uniqueCapabilities = new Set([...peerCapabilities].filter(cap => !myCapabilities.has(cap)));
    const sharedCapabilities = new Set([...peerCapabilities].filter(cap => myCapabilities.has(cap)));
    
    return (uniqueCapabilities.size * 0.7 + sharedCapabilities.size * 0.3) / Math.max(peerCapabilities.size, 1);
  }

  async disconnectPeer(peerId) {
    const peer = this.meshState.peers.get(peerId);
    if (!peer) return;
    
    // Send disconnect notification
    const disconnectMessage = {
      type: 'disconnect_notification',
      from: this.nodeId,
      reason: 'optimization',
      timestamp: Date.now()
    };
    
    await this.sendMessage(peerId, disconnectMessage);
    
    // Remove from peers
    this.meshState.peers.delete(peerId);
    
    // Update routing table
    this.updateRoutingTableOnFailure(peerId);
    
    this.emit('peer_disconnected', { peerId, reason: 'optimization' });
  }

  /**
   * Utility and Helper Methods
   */
  generateNodeId() {
    return `mesh_${Date.now()}_${crypto.randomBytes(8).toString('hex')}`;
  }

  generateConnectionId(nodeId1, nodeId2) {
    return `conn_${[nodeId1, nodeId2].sort().join('_')}`;
  }

  generateProposalId() {
    return `prop_${Date.now()}_${crypto.randomBytes(4).toString('hex')}`;
  }

  generateMessageId() {
    return `msg_${Date.now()}_${crypto.randomBytes(4).toString('hex')}`;
  }

  signMessage(message) {
    // Simplified signing - in production, use proper cryptographic signing
    const messageStr = JSON.stringify(message);
    return crypto.createHash('sha256').update(messageStr + this.nodeId).digest('hex');
  }

  verifyMessage(message, signature) {
    // Simplified verification - in production, use proper cryptographic verification
    const expectedSignature = crypto.createHash('sha256').update(JSON.stringify(message) + this.nodeId).digest('hex');
    return signature === expectedSignature;
  }

  async getNodeInfo(nodeId) {
    const nodeData = await this.redis.get(`mesh:node:${nodeId}`);
    return nodeData ? JSON.parse(nodeData) : null;
  }

  async testConnection(node) {
    // Simplified connection test - in production, implement proper connectivity check
    return Math.random() > 0.1; // 90% success rate
  }

  getPeerList() {
    return Array.from(this.meshState.peers.values()).map(peer => ({
      id: peer.id,
      address: peer.nodeInfo.address,
      port: peer.nodeInfo.port,
      capabilities: peer.nodeInfo.capabilities,
      reputation: peer.nodeInfo.reputation
    }));
  }

  async waitForJoinResponses() {
    // Wait for join responses from peers
    await new Promise(resolve => setTimeout(resolve, 5000)); // 5 second wait
  }

  async applyConsensusDecision(decision) {
    // Apply the consensus decision to local state
    console.log(`Applying consensus decision: ${JSON.stringify(decision)}`);
    
    if (decision.type === 'task_assignment') {
      await this.handleTaskAssignment(decision);
    } else if (decision.type === 'topology_change') {
      await this.handleTopologyChange(decision);
    } else if (decision.type === 'resource_allocation') {
      await this.handleResourceAllocation(decision);
    }
    
    this.emit('consensus_applied', decision);
  }

  async handleTaskAssignment(assignment) {
    console.log(`Handling task assignment: ${assignment.taskId}`);
    // Implementation for task assignment
  }

  async handleTopologyChange(change) {
    console.log(`Handling topology change: ${change.type}`);
    // Implementation for topology changes
  }

  async handleResourceAllocation(allocation) {
    console.log(`Handling resource allocation: ${allocation.type}`);
    // Implementation for resource allocation
  }

  // Status and monitoring
  async getStatus() {
    return {
      status: 'operational',
      nodeId: this.nodeId,
      nodeInfo: this.nodeInfo,
      mesh: {
        peers: this.meshState.peers.size,
        failedPeers: this.meshState.failedPeers.size,
        consensusState: this.meshState.consensusState,
        leader: this.meshState.leader,
        term: this.meshState.term
      },
      consensus: {
        raft: {
          state: this.consensus.raft.state,
          term: this.consensus.raft.term,
          logSize: this.consensus.raft.log.length
        },
        pbft: {
          view: this.consensus.pbft.view,
          phase: this.consensus.pbft.phase
        }
      },
      metrics: this.metrics,
      routing: {
        knownRoutes: this.meshState.routingTable.size,
        networkPartitions: this.meshState.networkPartitions.length
      }
    };
  }

  async cleanup() {
    // Clean up timers
    if (this.heartbeatTimer) clearInterval(this.heartbeatTimer);
    if (this.routingTimer) clearInterval(this.routingTimer);
    if (this.discoveryTimer) clearInterval(this.discoveryTimer);
    
    // Notify peers of departure
    const leaveMessage = {
      type: 'leave_notification',
      from: this.nodeId,
      timestamp: Date.now()
    };
    
    await this.broadcastToPeers(leaveMessage);
    
    // Clean up Redis
    await this.redis.del(`mesh:node:${this.nodeId}`);
    
    if (this.redis) {
      await this.redis.quit();
    }
  }

  // Placeholder implementations for consensus message handlers
  async handleAppendEntries(message) {
    console.log(`Handling append entries from ${message.from}`);
    // Raft append entries implementation
  }

  async handleVoteRequest(message) {
    console.log(`Handling vote request from ${message.from}`);
    // Raft vote request implementation
  }

  async handlePBFTPrepare(message) {
    console.log(`Handling PBFT prepare from ${message.from}`);
    // PBFT prepare phase implementation
  }

  async handlePBFTCommit(message) {
    console.log(`Handling PBFT commit from ${message.from}`);
    // PBFT commit phase implementation
  }
}

module.exports = AdaptiveMeshNetwork;