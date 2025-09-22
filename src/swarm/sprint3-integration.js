/**
 * Sprint 3 Integration Module
 * 
 * Orchestrates the deployment and integration of all Sprint 3 advanced swarm components:
 * - Multi-objective optimizer
 * - Dynamic topology optimizer
 * - Queen-led hierarchical coordinator
 * - Adaptive mesh network
 * - Swarm intelligence patterns
 * - Cross-agent learning system
 * - Performance optimizer
 * 
 * This module provides a unified interface for initializing, coordinating, and
 * managing all Sprint 3 components as an integrated swarm orchestration system.
 */

const EventEmitter = require('events');
const path = require('path');
const fs = require('fs').promises;

// Import Sprint 3 components
const MultiObjectiveOptimizer = require('./multi-objective-optimizer');
const TopologyOptimizer = require('./topology-optimizer');
const QueenCoordinator = require('./queen-coordinator');
const AdaptiveMeshNetwork = require('./adaptive-mesh-network');
const SwarmIntelligencePatterns = require('./swarm-intelligence-patterns');
const CrossAgentLearning = require('./cross-agent-learning');
const SwarmPerformanceOptimizer = require('./performance-optimizer');

class Sprint3Integration extends EventEmitter {
    constructor(config = {}) {
        super();
        
        this.config = {
            // Integration configuration
            enableAllComponents: config.enableAllComponents !== false,
            coordinationMode: config.coordinationMode || 'hierarchical', // hierarchical, mesh, hybrid
            
            // Component-specific configurations
            multiObjectiveConfig: config.multiObjectiveConfig || {},
            topologyConfig: config.topologyConfig || {},
            queenConfig: config.queenConfig || {},
            meshConfig: config.meshConfig || {},
            patternsConfig: config.patternsConfig || {},
            learningConfig: config.learningConfig || {},
            performanceConfig: config.performanceConfig || {},
            
            // Global swarm parameters
            maxAgents: config.maxAgents || 50,
            defaultTopology: config.defaultTopology || 'hierarchical',
            learningEnabled: config.learningEnabled !== false,
            performanceOptimization: config.performanceOptimization !== false,
            
            // Coordination intervals
            coordinationInterval: config.coordinationInterval || 10000, // 10 seconds
            syncInterval: config.syncInterval || 5000, // 5 seconds
            
            ...config
        };
        
        // Component instances
        this.components = {
            multiObjective: null,
            topology: null,
            queen: null,
            mesh: null,
            patterns: null,
            learning: null,
            performance: null
        };
        
        // Integration state
        this.state = {
            initialized: false,
            active: false,
            currentTopology: this.config.defaultTopology,
            coordinationMode: this.config.coordinationMode,
            swarmMetrics: new Map(),
            componentStatuses: new Map(),
            lastSync: null
        };
        
        // Coordination system
        this.coordinator = null;
        this.syncActive = false;
        
        // Event routing map for inter-component communication
        this.eventRouting = new Map();
        
        this.setupEventRouting();
    }
    
    async initialize() {
        console.log('🚀 Initializing Sprint 3 Advanced Swarm Orchestration...');
        
        try {
            // Initialize components in dependency order
            await this.initializeComponents();
            
            // Setup inter-component communication
            await this.setupInterComponentCommunication();
            
            // Initialize coordination system
            await this.initializeCoordination();
            
            // Start synchronization loop
            await this.startSynchronization();
            
            this.state.initialized = true;
            this.state.active = true;
            
            console.log('✅ Sprint 3 Integration initialized successfully');
            console.log(`📊 Active components: ${Object.values(this.components).filter(c => c).length}/7`);
            console.log(`🏗️ Coordination mode: ${this.state.coordinationMode}`);
            console.log(`🌐 Default topology: ${this.state.currentTopology}`);
            
            this.emit('initialized', {
                timestamp: Date.now(),
                components: Object.keys(this.components).filter(k => this.components[k]),
                mode: this.state.coordinationMode,
                topology: this.state.currentTopology
            });
            
            return this.getIntegrationStatus();
            
        } catch (error) {
            console.error('❌ Failed to initialize Sprint 3 Integration:', error);
            throw error;
        }
    }
    
    async initializeComponents() {
        console.log('🔧 Initializing Sprint 3 components...');
        
        // 1. Multi-objective optimizer (foundation for other optimizations)
        if (this.config.enableAllComponents) {
            console.log('  📈 Initializing Multi-objective Optimizer...');
            this.components.multiObjective = new MultiObjectiveOptimizer({
                ...this.config.multiObjectiveConfig,
                maxAgents: this.config.maxAgents
            });
            await this.components.multiObjective.initialize();
        }
        
        // 2. Topology optimizer (manages network structure)
        if (this.config.enableAllComponents) {
            console.log('  🕸️ Initializing Topology Optimizer...');
            this.components.topology = new TopologyOptimizer({
                ...this.config.topologyConfig,
                defaultTopology: this.config.defaultTopology,
                maxAgents: this.config.maxAgents
            });
            await this.components.topology.initialize();
        }
        
        // 3. Queen coordinator (strategic leadership)
        if (this.config.coordinationMode === 'hierarchical' || this.config.coordinationMode === 'hybrid') {
            console.log('  👑 Initializing Queen Coordinator...');
            this.components.queen = new QueenCoordinator({
                ...this.config.queenConfig,
                swarmSize: this.config.maxAgents,
                topologyOptimizer: this.components.topology
            });
            await this.components.queen.initialize();
        }
        
        // 4. Adaptive mesh network (peer-to-peer coordination)
        if (this.config.coordinationMode === 'mesh' || this.config.coordinationMode === 'hybrid') {
            console.log('  🌐 Initializing Adaptive Mesh Network...');
            this.components.mesh = new AdaptiveMeshNetwork({
                ...this.config.meshConfig,
                maxPeers: this.config.maxAgents,
                topologyOptimizer: this.components.topology
            });
            await this.components.mesh.initialize();
        }
        
        // 5. Swarm intelligence patterns (nature-inspired optimization)
        if (this.config.enableAllComponents) {
            console.log('  🐜 Initializing Swarm Intelligence Patterns...');
            this.components.patterns = new SwarmIntelligencePatterns({
                ...this.config.patternsConfig,
                swarmSize: this.config.maxAgents,
                topologyOptimizer: this.components.topology
            });
            await this.components.patterns.initialize();
        }
        
        // 6. Cross-agent learning (distributed learning)
        if (this.config.learningEnabled) {
            console.log('  🧠 Initializing Cross-agent Learning...');
            this.components.learning = new CrossAgentLearning({
                ...this.config.learningConfig,
                swarmSize: this.config.maxAgents,
                components: {
                    multiObjective: this.components.multiObjective,
                    topology: this.components.topology,
                    patterns: this.components.patterns
                }
            });
            await this.components.learning.initialize();
        }
        
        // 7. Performance optimizer (integrates and optimizes all components)
        if (this.config.performanceOptimization) {
            console.log('  ⚡ Initializing Performance Optimizer...');
            this.components.performance = new SwarmPerformanceOptimizer({
                ...this.config.performanceConfig,
                multiObjectiveOptimizer: this.components.multiObjective,
                topologyOptimizer: this.components.topology,
                queenCoordinator: this.components.queen,
                meshNetwork: this.components.mesh,
                swarmPatterns: this.components.patterns,
                crossAgentLearning: this.components.learning
            });
            await this.components.performance.initialize();
        }
        
        console.log('✅ All components initialized successfully');
    }
    
    async setupInterComponentCommunication() {
        console.log('🔗 Setting up inter-component communication...');
        
        // Multi-objective optimizer events
        if (this.components.multiObjective) {
            this.components.multiObjective.on('optimization-complete', (data) => {
                this.handleOptimizationComplete(data);
            });
            
            this.components.multiObjective.on('pareto-front-updated', (data) => {
                this.handleParetoFrontUpdated(data);
            });
        }
        
        // Topology optimizer events
        if (this.components.topology) {
            this.components.topology.on('topology-changed', (data) => {
                this.handleTopologyChanged(data);
            });
            
            this.components.topology.on('optimization-complete', (data) => {
                this.handleTopologyOptimizationComplete(data);
            });
        }
        
        // Queen coordinator events
        if (this.components.queen) {
            this.components.queen.on('strategic-decision', (data) => {
                this.handleStrategicDecision(data);
            });
            
            this.components.queen.on('role-assignment', (data) => {
                this.handleRoleAssignment(data);
            });
            
            this.components.queen.on('mission-update', (data) => {
                this.handleMissionUpdate(data);
            });
        }
        
        // Mesh network events
        if (this.components.mesh) {
            this.components.mesh.on('consensus-reached', (data) => {
                this.handleConsensusReached(data);
            });
            
            this.components.mesh.on('peer-discovered', (data) => {
                this.handlePeerDiscovered(data);
            });
            
            this.components.mesh.on('network-partitioned', (data) => {
                this.handleNetworkPartitioned(data);
            });
        }
        
        // Swarm patterns events
        if (this.components.patterns) {
            this.components.patterns.on('pattern-converged', (data) => {
                this.handlePatternConverged(data);
            });
            
            this.components.patterns.on('optimization-update', (data) => {
                this.handlePatternsOptimizationUpdate(data);
            });
        }
        
        // Learning system events
        if (this.components.learning) {
            this.components.learning.on('knowledge-shared', (data) => {
                this.handleKnowledgeShared(data);
            });
            
            this.components.learning.on('skill-transferred', (data) => {
                this.handleSkillTransferred(data);
            });
            
            this.components.learning.on('learning-milestone', (data) => {
                this.handleLearningMilestone(data);
            });
        }
        
        // Performance optimizer events
        if (this.components.performance) {
            this.components.performance.on('performance-optimized', (data) => {
                this.handlePerformanceOptimized(data);
            });
            
            this.components.performance.on('bottleneck-detected', (data) => {
                this.handleBottleneckDetected(data);
            });
        }
        
        console.log('✅ Inter-component communication established');
    }
    
    async initializeCoordination() {
        console.log(`🎯 Initializing ${this.state.coordinationMode} coordination...`);
        
        switch (this.state.coordinationMode) {
            case 'hierarchical':
                this.coordinator = this.components.queen;
                break;
                
            case 'mesh':
                this.coordinator = this.components.mesh;
                break;
                
            case 'hybrid':
                // Use both Queen and Mesh with intelligent switching
                this.coordinator = new HybridCoordinator({
                    queen: this.components.queen,
                    mesh: this.components.mesh,
                    topology: this.components.topology
                });
                break;
                
            default:
                throw new Error(`Unknown coordination mode: ${this.state.coordinationMode}`);
        }
        
        console.log('✅ Coordination system initialized');
    }
    
    async startSynchronization() {
        console.log('🔄 Starting component synchronization...');
        
        this.syncActive = true;
        
        const syncLoop = async () => {
            if (!this.syncActive) return;
            
            try {
                await this.synchronizeComponents();
                this.state.lastSync = Date.now();
                
                this.emit('sync-complete', {
                    timestamp: this.state.lastSync,
                    metrics: this.state.swarmMetrics,
                    statuses: this.state.componentStatuses
                });
                
            } catch (error) {
                console.error('❌ Synchronization error:', error);
                this.emit('sync-error', { error, timestamp: Date.now() });
            }
            
            setTimeout(syncLoop, this.config.syncInterval);
        };
        
        // Start coordination loop
        const coordinationLoop = async () => {
            if (!this.syncActive) return;
            
            try {
                await this.coordinateComponents();
                
                this.emit('coordination-complete', {
                    timestamp: Date.now(),
                    mode: this.state.coordinationMode,
                    coordinator: this.coordinator?.constructor.name
                });
                
            } catch (error) {
                console.error('❌ Coordination error:', error);
                this.emit('coordination-error', { error, timestamp: Date.now() });
            }
            
            setTimeout(coordinationLoop, this.config.coordinationInterval);
        };
        
        syncLoop();
        coordinationLoop();
        
        console.log('✅ Synchronization started');
    }
    
    async synchronizeComponents() {
        // Collect metrics from all components
        const metrics = new Map();
        const statuses = new Map();
        
        for (const [name, component] of Object.entries(this.components)) {
            if (component && typeof component.getMetrics === 'function') {
                try {
                    metrics.set(name, await component.getMetrics());
                    statuses.set(name, 'active');
                } catch (error) {
                    console.warn(`⚠️ Failed to get metrics from ${name}:`, error.message);
                    statuses.set(name, 'error');
                }
            } else if (component) {
                statuses.set(name, 'active');
            } else {
                statuses.set(name, 'disabled');
            }
        }
        
        this.state.swarmMetrics = metrics;
        this.state.componentStatuses = statuses;
        
        // Share metrics across components that need them
        await this.shareMetricsAcrossComponents(metrics);
    }
    
    async shareMetricsAcrossComponents(metrics) {
        // Performance optimizer gets all metrics
        if (this.components.performance && typeof this.components.performance.updateMetrics === 'function') {
            await this.components.performance.updateMetrics(metrics);
        }
        
        // Learning system gets optimization metrics
        if (this.components.learning && typeof this.components.learning.updatePerformanceMetrics === 'function') {
            const optimizationMetrics = new Map();
            if (metrics.has('multiObjective')) optimizationMetrics.set('multiObjective', metrics.get('multiObjective'));
            if (metrics.has('topology')) optimizationMetrics.set('topology', metrics.get('topology'));
            if (metrics.has('patterns')) optimizationMetrics.set('patterns', metrics.get('patterns'));
            
            await this.components.learning.updatePerformanceMetrics(optimizationMetrics);
        }
        
        // Topology optimizer gets network metrics
        if (this.components.topology && typeof this.components.topology.updateNetworkMetrics === 'function') {
            const networkMetrics = new Map();
            if (metrics.has('mesh')) networkMetrics.set('mesh', metrics.get('mesh'));
            if (metrics.has('performance')) networkMetrics.set('performance', metrics.get('performance'));
            
            await this.components.topology.updateNetworkMetrics(networkMetrics);
        }
    }
    
    async coordinateComponents() {
        if (!this.coordinator) return;
        
        // Get coordination decisions from the active coordinator
        let coordinationDecisions;
        
        if (this.coordinator === this.components.queen) {
            coordinationDecisions = await this.coordinateHierarchically();
        } else if (this.coordinator === this.components.mesh) {
            coordinationDecisions = await this.coordinateThroughMesh();
        } else if (this.coordinator instanceof HybridCoordinator) {
            coordinationDecisions = await this.coordinator.coordinate({
                metrics: this.state.swarmMetrics,
                statuses: this.state.componentStatuses
            });
        }
        
        // Execute coordination decisions
        if (coordinationDecisions) {
            await this.executeCoordinationDecisions(coordinationDecisions);
        }
    }
    
    async coordinateHierarchically() {
        // Queen makes strategic decisions based on swarm state
        const swarmState = {
            metrics: this.state.swarmMetrics,
            topology: this.state.currentTopology,
            componentStatuses: this.state.componentStatuses
        };
        
        return await this.components.queen.makeStrategicDecisions(swarmState);
    }
    
    async coordinateThroughMesh() {
        // Mesh network reaches consensus on coordination actions
        const proposal = {
            type: 'coordination',
            metrics: this.state.swarmMetrics,
            timestamp: Date.now()
        };
        
        return await this.components.mesh.proposeAndReachConsensus(proposal);
    }
    
    async executeCoordinationDecisions(decisions) {
        for (const decision of decisions) {
            try {
                await this.executeDecision(decision);
            } catch (error) {
                console.error(`❌ Failed to execute decision ${decision.type}:`, error);
            }
        }
    }
    
    async executeDecision(decision) {
        switch (decision.type) {
            case 'topology_change':
                await this.changeTopology(decision.targetTopology);
                break;
                
            case 'resource_reallocation':
                await this.reallocateResources(decision.allocation);
                break;
                
            case 'optimization_trigger':
                await this.triggerOptimization(decision.target, decision.parameters);
                break;
                
            case 'learning_focus':
                await this.focusLearning(decision.domain, decision.priority);
                break;
                
            default:
                console.warn(`⚠️ Unknown decision type: ${decision.type}`);
        }
    }
    
    async changeTopology(newTopology) {
        if (this.components.topology) {
            await this.components.topology.changeTopology(newTopology);
            this.state.currentTopology = newTopology;
            
            this.emit('topology-changed', {
                previousTopology: this.state.currentTopology,
                newTopology: newTopology,
                timestamp: Date.now()
            });
        }
    }
    
    async reallocateResources(allocation) {
        if (this.components.multiObjective) {
            await this.components.multiObjective.updateResourceAllocation(allocation);
        }
    }
    
    async triggerOptimization(target, parameters) {
        const component = this.components[target];
        if (component && typeof component.optimize === 'function') {
            await component.optimize(parameters);
        }
    }
    
    async focusLearning(domain, priority) {
        if (this.components.learning) {
            await this.components.learning.focusLearning(domain, priority);
        }
    }
    
    // Event handlers for inter-component communication
    handleOptimizationComplete(data) {
        console.log('📈 Multi-objective optimization completed');
        
        // Inform other components about optimization results
        if (this.components.learning) {
            this.components.learning.recordOptimizationResult(data);
        }
        
        if (this.components.topology) {
            this.components.topology.considerOptimizationResults(data);
        }
        
        this.emit('optimization-complete', { source: 'multiObjective', data });
    }
    
    handleParetoFrontUpdated(data) {
        console.log('📊 Pareto front updated with new solutions');
        
        // Share Pareto front data with learning system
        if (this.components.learning) {
            this.components.learning.updateParetoFront(data);
        }
        
        this.emit('pareto-front-updated', { source: 'multiObjective', data });
    }
    
    handleTopologyChanged(data) {
        console.log(`🕸️ Topology changed: ${data.previousTopology} → ${data.newTopology}`);
        
        this.state.currentTopology = data.newTopology;
        
        // Inform all components about topology change
        this.broadcastToComponents('topology-changed', data, ['topology']);
        
        this.emit('topology-changed', { source: 'topology', data });
    }
    
    handleTopologyOptimizationComplete(data) {
        console.log('🎯 Topology optimization completed');
        
        if (this.components.learning) {
            this.components.learning.recordTopologyOptimization(data);
        }
        
        this.emit('topology-optimization-complete', { source: 'topology', data });
    }
    
    handleStrategicDecision(data) {
        console.log(`👑 Queen made strategic decision: ${data.decision}`);
        
        // Execute strategic decisions
        this.executeStrategicDecision(data);
        
        this.emit('strategic-decision', { source: 'queen', data });
    }
    
    async executeStrategicDecision(data) {
        switch (data.decision) {
            case 'increase_optimization_focus':
                if (this.components.patterns) {
                    await this.components.patterns.increaseOptimizationIntensity();
                }
                break;
                
            case 'improve_coordination':
                if (this.components.mesh) {
                    await this.components.mesh.improveCoordination();
                }
                break;
                
            case 'focus_learning':
                if (this.components.learning) {
                    await this.components.learning.focusLearning(data.learningDomain);
                }
                break;
        }
    }
    
    handleRoleAssignment(data) {
        console.log(`👥 Role assignments updated: ${data.assignments.length} agents`);
        
        // Update component configurations based on role assignments
        this.updateComponentsWithRoles(data.assignments);
        
        this.emit('role-assignment', { source: 'queen', data });
    }
    
    updateComponentsWithRoles(assignments) {
        // Update patterns based on agent roles
        if (this.components.patterns) {
            this.components.patterns.updateAgentRoles(assignments);
        }
        
        // Update learning based on agent specializations
        if (this.components.learning) {
            this.components.learning.updateAgentSpecializations(assignments);
        }
    }
    
    handleMissionUpdate(data) {
        console.log(`🎯 Mission updated: ${data.mission.objective}`);
        
        // Align all components with new mission
        this.alignComponentsWithMission(data.mission);
        
        this.emit('mission-update', { source: 'queen', data });
    }
    
    alignComponentsWithMission(mission) {
        // Align multi-objective optimization with mission objectives
        if (this.components.multiObjective) {
            this.components.multiObjective.updateObjectives(mission.objectives);
        }
        
        // Align patterns with mission requirements
        if (this.components.patterns) {
            this.components.patterns.alignWithMission(mission);
        }
        
        // Focus learning on mission-relevant skills
        if (this.components.learning) {
            this.components.learning.focusOnMissionSkills(mission.requiredSkills);
        }
    }
    
    handleConsensusReached(data) {
        console.log(`🤝 Mesh consensus reached: ${data.decision}`);
        
        // Execute consensus decision
        this.executeConsensusDecision(data);
        
        this.emit('consensus-reached', { source: 'mesh', data });
    }
    
    async executeConsensusDecision(data) {
        // Execute decisions reached by mesh consensus
        if (data.decision === 'optimize_patterns') {
            if (this.components.patterns) {
                await this.components.patterns.optimize(data.parameters);
            }
        }
    }
    
    handlePeerDiscovered(data) {
        console.log(`🌐 New peer discovered: ${data.peerId}`);
        
        // Update topology optimizer with new peer
        if (this.components.topology) {
            this.components.topology.addPeer(data.peerId, data.peerInfo);
        }
        
        this.emit('peer-discovered', { source: 'mesh', data });
    }
    
    handleNetworkPartitioned(data) {
        console.log(`⚠️ Network partition detected: ${data.partitions.length} partitions`);
        
        // Trigger topology optimization to handle partition
        if (this.components.topology) {
            this.components.topology.handleNetworkPartition(data);
        }
        
        this.emit('network-partitioned', { source: 'mesh', data });
    }
    
    handlePatternConverged(data) {
        console.log(`🐜 Pattern converged: ${data.pattern} (fitness: ${data.fitness})`);
        
        // Share convergence results with learning system
        if (this.components.learning) {
            this.components.learning.recordPatternConvergence(data);
        }
        
        this.emit('pattern-converged', { source: 'patterns', data });
    }
    
    handlePatternsOptimizationUpdate(data) {
        console.log(`📊 Patterns optimization update: ${data.improvements} improvements`);
        
        // Update multi-objective optimizer with pattern results
        if (this.components.multiObjective) {
            this.components.multiObjective.integratePatternResults(data);
        }
        
        this.emit('patterns-optimization-update', { source: 'patterns', data });
    }
    
    handleKnowledgeShared(data) {
        console.log(`🧠 Knowledge shared: ${data.sender} → ${data.receiver} (${data.knowledgeType})`);
        
        // Update performance metrics with learning activity
        if (this.components.performance) {
            this.components.performance.recordLearningActivity(data);
        }
        
        this.emit('knowledge-shared', { source: 'learning', data });
    }
    
    handleSkillTransferred(data) {
        console.log(`🎓 Skill transferred: ${data.skill} (success: ${data.success})`);
        
        // Update Queen's strategic planning with skill transfer results
        if (this.components.queen) {
            this.components.queen.recordSkillTransfer(data);
        }
        
        this.emit('skill-transferred', { source: 'learning', data });
    }
    
    handleLearningMilestone(data) {
        console.log(`🏆 Learning milestone reached: ${data.milestone} (score: ${data.score})`);
        
        // Celebrate milestone and adjust optimization strategies
        this.adjustOptimizationForLearning(data);
        
        this.emit('learning-milestone', { source: 'learning', data });
    }
    
    adjustOptimizationForLearning(milestone) {
        // Adjust multi-objective optimization based on learning progress
        if (this.components.multiObjective && milestone.score > 0.8) {
            this.components.multiObjective.increaseLearningWeight();
        }
        
        // Adjust patterns based on learning effectiveness
        if (this.components.patterns) {
            this.components.patterns.adjustBasedOnLearning(milestone);
        }
    }
    
    handlePerformanceOptimized(data) {
        console.log(`⚡ Performance optimized: ${data.improvement}% improvement`);
        
        // All components benefit from performance optimization insights
        this.broadcastToComponents('performance-optimized', data, ['performance']);
        
        this.emit('performance-optimized', { source: 'performance', data });
    }
    
    handleBottleneckDetected(data) {
        console.log(`🚨 Bottleneck detected: ${data.component} (severity: ${data.severity})`);
        
        // Route bottleneck to appropriate component for resolution
        this.routeBottleneckForResolution(data);
        
        this.emit('bottleneck-detected', { source: 'performance', data });
    }
    
    routeBottleneckForResolution(bottleneck) {
        switch (bottleneck.component) {
            case 'topology':
                if (this.components.topology) {
                    this.components.topology.resolveBottleneck(bottleneck);
                }
                break;
                
            case 'patterns':
                if (this.components.patterns) {
                    this.components.patterns.resolveBottleneck(bottleneck);
                }
                break;
                
            case 'learning':
                if (this.components.learning) {
                    this.components.learning.resolveBottleneck(bottleneck);
                }
                break;
        }
    }
    
    // Utility method to broadcast events to components
    broadcastToComponents(eventType, data, exclude = []) {
        for (const [name, component] of Object.entries(this.components)) {
            if (component && !exclude.includes(name) && typeof component.emit === 'function') {
                component.emit(eventType, data);
            }
        }
    }
    
    setupEventRouting() {
        // Define event routing patterns for cross-component communication
        this.eventRouting.set('optimization-complete', ['learning', 'topology', 'performance']);
        this.eventRouting.set('topology-changed', ['queen', 'mesh', 'patterns', 'learning', 'performance']);
        this.eventRouting.set('consensus-reached', ['queen', 'topology', 'patterns', 'learning']);
        this.eventRouting.set('pattern-converged', ['learning', 'multiObjective', 'performance']);
        this.eventRouting.set('learning-milestone', ['queen', 'multiObjective', 'patterns', 'performance']);
        this.eventRouting.set('performance-optimized', ['queen', 'mesh', 'multiObjective', 'topology']);
    }
    
    // Public API methods
    getIntegrationStatus() {
        return {
            initialized: this.state.initialized,
            active: this.state.active,
            coordinationMode: this.state.coordinationMode,
            currentTopology: this.state.currentTopology,
            components: Object.fromEntries(this.state.componentStatuses),
            metrics: Object.fromEntries(this.state.swarmMetrics),
            lastSync: this.state.lastSync,
            uptime: this.state.initialized ? Date.now() - this.state.lastSync : 0
        };
    }
    
    async getSwarmMetrics() {
        await this.synchronizeComponents();
        return Object.fromEntries(this.state.swarmMetrics);
    }
    
    async triggerGlobalOptimization(parameters = {}) {
        console.log('🎯 Triggering global swarm optimization...');
        
        const results = {};
        
        // Trigger optimization on all capable components
        if (this.components.multiObjective) {
            results.multiObjective = await this.components.multiObjective.optimize(parameters);
        }
        
        if (this.components.topology) {
            results.topology = await this.components.topology.optimize(parameters);
        }
        
        if (this.components.patterns) {
            results.patterns = await this.components.patterns.optimize(parameters);
        }
        
        if (this.components.performance) {
            results.performance = await this.components.performance.optimize(parameters);
        }
        
        this.emit('global-optimization-complete', {
            timestamp: Date.now(),
            results,
            parameters
        });
        
        return results;
    }
    
    async switchCoordinationMode(newMode) {
        console.log(`🔄 Switching coordination mode: ${this.state.coordinationMode} → ${newMode}`);
        
        this.state.coordinationMode = newMode;
        await this.initializeCoordination();
        
        this.emit('coordination-mode-changed', {
            previousMode: this.state.coordinationMode,
            newMode: newMode,
            timestamp: Date.now()
        });
        
        return this.getIntegrationStatus();
    }
    
    async shutdown() {
        console.log('🛑 Shutting down Sprint 3 Integration...');
        
        this.syncActive = false;
        this.state.active = false;
        
        // Shutdown components in reverse dependency order
        const shutdownOrder = ['performance', 'learning', 'patterns', 'mesh', 'queen', 'topology', 'multiObjective'];
        
        for (const componentName of shutdownOrder) {
            const component = this.components[componentName];
            if (component && typeof component.shutdown === 'function') {
                try {
                    await component.shutdown();
                    console.log(`  ✅ ${componentName} shutdown complete`);
                } catch (error) {
                    console.error(`  ❌ ${componentName} shutdown failed:`, error);
                }
            }
        }
        
        // Save final integration state
        await this.saveIntegrationState();
        
        this.emit('shutdown', { timestamp: Date.now() });
        console.log('✅ Sprint 3 Integration shutdown complete');
    }
    
    async saveIntegrationState() {
        const state = {
            timestamp: Date.now(),
            config: this.config,
            state: this.state,
            metrics: Object.fromEntries(this.state.swarmMetrics),
            componentStatuses: Object.fromEntries(this.state.componentStatuses)
        };
        
        try {
            const statePath = path.join(__dirname, '../../data/sprint3-integration-state.json');
            await fs.writeFile(statePath, JSON.stringify(state, null, 2));
            console.log('💾 Sprint 3 integration state saved');
        } catch (error) {
            console.error('❌ Failed to save integration state:', error);
        }
    }
}

// Hybrid coordinator for mixed hierarchical/mesh coordination
class HybridCoordinator {
    constructor({ queen, mesh, topology }) {
        this.queen = queen;
        this.mesh = mesh;
        this.topology = topology;
        
        this.coordinationStrategy = 'adaptive'; // adaptive, queen-primary, mesh-primary
        this.decisionHistory = [];
    }
    
    async coordinate(context) {
        const { metrics, statuses } = context;
        
        // Determine optimal coordination strategy based on context
        const strategy = this.determineOptimalStrategy(metrics, statuses);
        
        let decisions = [];
        
        if (strategy === 'hierarchical' || strategy === 'queen-primary') {
            // Use Queen for strategic decisions
            const queenDecisions = await this.queen.makeStrategicDecisions(context);
            decisions = decisions.concat(queenDecisions);
            
            // Use mesh for consensus on tactical decisions
            if (queenDecisions.length > 0) {
                const meshConsensus = await this.mesh.proposeAndReachConsensus({
                    type: 'tactical_execution',
                    decisions: queenDecisions
                });
                decisions = decisions.concat(meshConsensus);
            }
        } else {
            // Use mesh for distributed decisions
            const meshDecisions = await this.mesh.proposeAndReachConsensus(context);
            decisions = decisions.concat(meshDecisions);
            
            // Use Queen for strategic oversight
            if (meshDecisions.length > 0) {
                const queenOversight = await this.queen.provideStrategicOversight({
                    decisions: meshDecisions,
                    context
                });
                decisions = decisions.concat(queenOversight);
            }
        }
        
        // Record coordination decision
        this.decisionHistory.push({
            timestamp: Date.now(),
            strategy: strategy,
            context: context,
            decisions: decisions.length
        });
        
        return decisions;
    }
    
    determineOptimalStrategy(metrics, statuses) {
        // Analyze metrics to determine best coordination approach
        const activeComponents = Array.from(statuses.values()).filter(s => s === 'active').length;
        const networkHealth = this.assessNetworkHealth(metrics);
        
        if (activeComponents > 10 && networkHealth > 0.8) {
            return 'mesh-primary'; // Large, healthy network - use mesh
        } else if (networkHealth < 0.5) {
            return 'hierarchical'; // Poor network health - use hierarchy
        } else {
            return 'adaptive'; // Mixed approach
        }
    }
    
    assessNetworkHealth(metrics) {
        // Simple network health assessment
        let healthScore = 0.5; // Base score
        
        if (metrics.has('mesh')) {
            const meshMetrics = metrics.get('mesh');
            if (meshMetrics.connectivity > 0.8) healthScore += 0.2;
            if (meshMetrics.latency < 100) healthScore += 0.1;
            if (meshMetrics.consensus_time < 1000) healthScore += 0.1;
        }
        
        if (metrics.has('topology')) {
            const topoMetrics = metrics.get('topology');
            if (topoMetrics.efficiency > 0.7) healthScore += 0.1;
        }
        
        return Math.min(1, healthScore);
    }
}

module.exports = Sprint3Integration;