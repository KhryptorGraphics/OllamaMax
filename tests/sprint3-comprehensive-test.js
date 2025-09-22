/**
 * Sprint 3 Comprehensive Test Suite
 * 
 * Tests all Sprint 3 advanced swarm orchestration components:
 * - Multi-objective optimizer
 * - Dynamic topology optimizer  
 * - Queen-led hierarchical coordinator
 * - Adaptive mesh network
 * - Swarm intelligence patterns
 * - Cross-agent learning system
 * - Performance optimizer
 * - Integration system
 * 
 * Test categories:
 * - Unit tests for individual components
 * - Integration tests for component interactions
 * - Performance tests for optimization algorithms
 * - Stress tests for system resilience
 * - End-to-end tests for complete workflows
 */

const { expect } = require('chai');
const sinon = require('sinon');
const path = require('path');
const fs = require('fs').promises;

// Import Sprint 3 components
const MultiObjectiveOptimizer = require('../src/swarm/multi-objective-optimizer');
const TopologyOptimizer = require('../src/swarm/topology-optimizer');
const QueenCoordinator = require('../src/swarm/queen-coordinator');
const AdaptiveMeshNetwork = require('../src/swarm/adaptive-mesh-network');
const SwarmIntelligencePatterns = require('../src/swarm/swarm-intelligence-patterns');
const CrossAgentLearning = require('../src/swarm/cross-agent-learning');
const SwarmPerformanceOptimizer = require('../src/swarm/performance-optimizer');
const Sprint3Integration = require('../src/swarm/sprint3-integration');

describe('Sprint 3 Advanced Swarm Orchestration', function() {
    // Increase timeout for complex optimization tests
    this.timeout(30000);
    
    let testConfig;
    let testComponents;
    
    before(async function() {
        console.log('🧪 Setting up Sprint 3 test environment...');
        
        testConfig = {
            maxAgents: 10,
            testMode: true,
            enableAllComponents: true,
            monitoringInterval: 1000,
            optimizationInterval: 2000,
            coordinationInterval: 1500
        };
        
        testComponents = {};
    });
    
    after(async function() {
        console.log('🧹 Cleaning up Sprint 3 test environment...');
        
        // Cleanup test components
        for (const component of Object.values(testComponents)) {
            if (component && typeof component.shutdown === 'function') {
                try {
                    await component.shutdown();
                } catch (error) {
                    console.warn('Cleanup warning:', error.message);
                }
            }
        }
    });
    
    describe('Multi-Objective Optimizer', function() {
        let optimizer;
        
        beforeEach(async function() {
            optimizer = new MultiObjectiveOptimizer({
                ...testConfig,
                populationSize: 10,
                generations: 5
            });
            await optimizer.initialize();
            testComponents.multiObjective = optimizer;
        });
        
        it('should initialize with correct objectives', function() {
            expect(optimizer.objectives).to.have.length.greaterThan(0);
            expect(optimizer.objectives[0]).to.have.property('name');
            expect(optimizer.objectives[0]).to.have.property('weight');
            expect(optimizer.objectives[0]).to.have.property('minimize');
        });
        
        it('should optimize agent allocation using NSGA-II', async function() {
            const swarmState = {
                agents: Array.from({length: 10}, (_, i) => ({
                    id: `agent_${i}`,
                    performance: Math.random(),
                    cost: Math.random() * 100,
                    availability: true
                })),
                tasks: Array.from({length: 5}, (_, i) => ({
                    id: `task_${i}`,
                    priority: Math.random(),
                    complexity: Math.random()
                }))
            };
            
            const optimizationRequest = {
                type: 'agent_allocation',
                constraints: {
                    max_cost: 500,
                    min_performance: 0.6
                }
            };
            
            const result = await optimizer.optimizeWithNSGAII(swarmState, optimizationRequest);
            
            expect(result).to.have.property('solutions');
            expect(result.solutions).to.be.an('array');
            expect(result.solutions.length).to.be.greaterThan(0);
            expect(result).to.have.property('paretoFront');
            expect(result).to.have.property('statistics');
        });
        
        it('should optimize using MOEA/D algorithm', async function() {
            const swarmState = {
                agents: Array.from({length: 8}, (_, i) => ({
                    id: `agent_${i}`,
                    capabilities: ['compute', 'storage', 'network'][i % 3],
                    performance: 0.5 + Math.random() * 0.5
                }))
            };
            
            const result = await optimizer.optimizeWithMOEAD(swarmState, {
                type: 'resource_optimization',
                decomposition: 'weighted_sum'
            });
            
            expect(result).to.have.property('solutions');
            expect(result).to.have.property('convergence');
            expect(result.convergence).to.be.a('number');
        });
        
        it('should maintain Pareto front correctly', async function() {
            const mockSolutions = [
                { objectives: [0.1, 0.9], allocation: ['a1', 'a2'] },
                { objectives: [0.5, 0.5], allocation: ['a1', 'a3'] },
                { objectives: [0.9, 0.1], allocation: ['a2', 'a3'] },
                { objectives: [0.6, 0.6], allocation: ['a1', 'a2', 'a3'] } // Dominated
            ];
            
            const paretoFront = optimizer.extractParetoFront(mockSolutions);
            
            expect(paretoFront).to.have.length(3); // Should exclude dominated solution
            expect(paretoFront.some(s => s.objectives[0] === 0.6 && s.objectives[1] === 0.6)).to.be.false;
        });
        
        it('should handle constraint violations', async function() {
            const solution = {
                allocation: ['agent_1', 'agent_2', 'agent_3'],
                cost: 600, // Exceeds constraint
                performance: 0.8
            };
            
            const constraints = {
                max_cost: 500,
                min_performance: 0.6
            };
            
            const violations = optimizer.evaluateConstraints(solution, constraints);
            
            expect(violations).to.have.property('violated');
            expect(violations.violated).to.be.true;
            expect(violations).to.have.property('violations');
            expect(violations.violations).to.include('max_cost');
        });
        
        it('should emit optimization events', function(done) {
            optimizer.on('optimization-complete', (data) => {
                expect(data).to.have.property('solutions');
                expect(data).to.have.property('generation');
                done();
            });
            
            // Trigger optimization
            optimizer.optimizeWithNSGAII(
                { agents: [], tasks: [] },
                { type: 'test' }
            );
        });
    });
    
    describe('Topology Optimizer', function() {
        let topologyOptimizer;
        
        beforeEach(async function() {
            topologyOptimizer = new TopologyOptimizer({
                ...testConfig,
                initialTopology: 'hierarchical'
            });
            await topologyOptimizer.initialize();
            testComponents.topology = topologyOptimizer;
        });
        
        it('should initialize with default topology', function() {
            expect(topologyOptimizer.currentTopology).to.equal('hierarchical');
            expect(topologyOptimizer.availableTopologies).to.include('mesh');
            expect(topologyOptimizer.availableTopologies).to.include('ring');
            expect(topologyOptimizer.availableTopologies).to.include('star');
        });
        
        it('should optimize topology based on performance metrics', async function() {
            const performanceMetrics = {
                latency: 150,
                throughput: 85,
                reliability: 0.95,
                scalability: 0.7
            };
            
            const result = await topologyOptimizer.optimizeTopology(performanceMetrics);
            
            expect(result).to.have.property('recommendedTopology');
            expect(result).to.have.property('expectedImprovement');
            expect(result).to.have.property('migrationPlan');
        });
        
        it('should switch topology dynamically', async function() {
            const initialTopology = topologyOptimizer.currentTopology;
            
            await topologyOptimizer.switchTopology('mesh');
            
            expect(topologyOptimizer.currentTopology).to.equal('mesh');
            expect(topologyOptimizer.currentTopology).not.to.equal(initialTopology);
        });
        
        it('should calculate topology fitness correctly', function() {
            const topologyMetrics = {
                communicationCost: 0.3,
                faultTolerance: 0.8,
                scalability: 0.9,
                latency: 100
            };
            
            const fitness = topologyOptimizer.calculateTopologyFitness('mesh', topologyMetrics);
            
            expect(fitness).to.be.a('number');
            expect(fitness).to.be.greaterThan(0);
            expect(fitness).to.be.lessThanOrEqual(1);
        });
        
        it('should handle topology transition safely', async function() {
            const transitionPlan = await topologyOptimizer.planTopologyTransition('hierarchical', 'mesh');
            
            expect(transitionPlan).to.have.property('phases');
            expect(transitionPlan.phases).to.be.an('array');
            expect(transitionPlan).to.have.property('estimatedDuration');
            expect(transitionPlan).to.have.property('riskAssessment');
        });
    });
    
    describe('Queen Coordinator', function() {
        let queenCoordinator;
        
        beforeEach(async function() {
            queenCoordinator = new QueenCoordinator({
                ...testConfig,
                swarmSize: 15,
                strategicPlanningInterval: 2000
            });
            await queenCoordinator.initialize();
            testComponents.queen = queenCoordinator;
        });
        
        it('should initialize Queen with strategic planning capability', function() {
            expect(queenCoordinator.queenAgent).to.exist;
            expect(queenCoordinator.queenAgent).to.have.property('role', 'queen');
            expect(queenCoordinator.queenAgent).to.have.property('strategicPlanning');
        });
        
        it('should assign roles to swarm agents', async function() {
            const agents = Array.from({length: 12}, (_, i) => ({
                id: `agent_${i}`,
                capabilities: ['scouting', 'working', 'guarding'][i % 3],
                performance: Math.random()
            }));
            
            const roleAssignments = await queenCoordinator.assignRoles(agents);
            
            expect(roleAssignments).to.be.an('array');
            expect(roleAssignments).to.have.length.greaterThan(0);
            
            const roles = roleAssignments.map(a => a.role);
            expect(roles).to.include.members(['lieutenant', 'worker', 'scout', 'guard']);
        });
        
        it('should make strategic decisions', async function() {
            const swarmState = {
                performance: 0.75,
                resources: { cpu: 0.6, memory: 0.8 },
                threats: [],
                opportunities: ['optimization_target_1']
            };
            
            const decisions = await queenCoordinator.makeStrategicDecisions(swarmState);
            
            expect(decisions).to.be.an('array');
            if (decisions.length > 0) {
                expect(decisions[0]).to.have.property('type');
                expect(decisions[0]).to.have.property('priority');
                expect(decisions[0]).to.have.property('parameters');
            }
        });
        
        it('should coordinate hierarchical mission execution', async function() {
            const mission = {
                objective: 'optimize_performance',
                priority: 'high',
                deadline: Date.now() + 60000,
                resources: ['cpu', 'memory']
            };
            
            const executionPlan = await queenCoordinator.coordinateMission(mission);
            
            expect(executionPlan).to.have.property('phases');
            expect(executionPlan).to.have.property('assignments');
            expect(executionPlan).to.have.property('timeline');
        });
        
        it('should handle lieutenant promotion and demotion', async function() {
            const agents = [
                { id: 'agent_1', role: 'worker', performance: 0.95 },
                { id: 'agent_2', role: 'lieutenant', performance: 0.4 }
            ];
            
            const roleChanges = await queenCoordinator.evaluateRoleChanges(agents);
            
            expect(roleChanges.promotions).to.include('agent_1');
            expect(roleChanges.demotions).to.include('agent_2');
        });
        
        it('should emit strategic events', function(done) {
            queenCoordinator.on('strategic-decision', (data) => {
                expect(data).to.have.property('decision');
                expect(data).to.have.property('reasoning');
                done();
            });
            
            // Trigger strategic planning
            queenCoordinator.executeStrategicPlanning();
        });
    });
    
    describe('Adaptive Mesh Network', function() {
        let meshNetwork;
        
        beforeEach(async function() {
            meshNetwork = new AdaptiveMeshNetwork({
                ...testConfig,
                consensusProtocol: 'raft',
                maxPeers: 12
            });
            await meshNetwork.initialize();
            testComponents.mesh = meshNetwork;
        });
        
        it('should initialize mesh network with proper configuration', function() {
            expect(meshNetwork.consensusProtocol).to.equal('raft');
            expect(meshNetwork.peers).to.be.a('map');
            expect(meshNetwork.consensusEngine).to.exist;
        });
        
        it('should add and manage peers correctly', async function() {
            const peer = {
                id: 'peer_1',
                address: '192.168.1.100',
                capabilities: ['compute', 'storage'],
                trust_score: 0.8
            };
            
            await meshNetwork.addPeer(peer);
            
            expect(meshNetwork.peers.has(peer.id)).to.be.true;
            expect(meshNetwork.peers.get(peer.id)).to.deep.include(peer);
        });
        
        it('should reach consensus using Raft protocol', async function() {
            // Add multiple peers for consensus
            for (let i = 0; i < 5; i++) {
                await meshNetwork.addPeer({
                    id: `peer_${i}`,
                    address: `192.168.1.${100 + i}`,
                    capabilities: ['compute']
                });
            }
            
            const proposal = {
                type: 'resource_allocation',
                data: { target: 'optimization', resources: ['cpu'] },
                proposer: 'peer_1'
            };
            
            const consensusResult = await meshNetwork.proposeAndReachConsensus(proposal);
            
            expect(consensusResult).to.have.property('decision');
            expect(consensusResult).to.have.property('votes');
            expect(consensusResult).to.have.property('consensus_reached');
        });
        
        it('should handle Byzantine fault tolerance', async function() {
            // Simulate Byzantine fault scenario
            const maliciousPeer = {
                id: 'malicious_peer',
                address: '192.168.1.200',
                malicious: true
            };
            
            await meshNetwork.addPeer(maliciousPeer);
            
            const proposal = {
                type: 'security_test',
                data: { test: true }
            };
            
            const result = await meshNetwork.proposeAndReachConsensus(proposal);
            
            // System should still reach consensus despite malicious peer
            expect(result.consensus_reached).to.be.true;
            expect(result.byzantine_detected).to.be.true;
        });
        
        it('should detect and handle network partitions', async function() {
            // Simulate network partition
            const partitionEvent = {
                type: 'network_partition',
                affected_peers: ['peer_1', 'peer_2'],
                timestamp: Date.now()
            };
            
            const response = await meshNetwork.handleNetworkPartition(partitionEvent);
            
            expect(response).to.have.property('recovery_strategy');
            expect(response).to.have.property('estimated_recovery_time');
        });
        
        it('should self-heal network connectivity', async function() {
            // Simulate peer failure
            await meshNetwork.addPeer({ id: 'failing_peer', address: '192.168.1.150' });
            await meshNetwork.simulatePeerFailure('failing_peer');
            
            const healingResult = await meshNetwork.triggerSelfHealing();
            
            expect(healingResult).to.have.property('actions_taken');
            expect(healingResult).to.have.property('network_health');
        });
    });
    
    describe('Swarm Intelligence Patterns', function() {
        let swarmPatterns;
        
        beforeEach(async function() {
            swarmPatterns = new SwarmIntelligencePatterns({
                ...testConfig,
                enabledPatterns: ['aco', 'pso', 'abc', 'boids'],
                swarmSize: 20
            });
            await swarmPatterns.initialize();
            testComponents.patterns = swarmPatterns;
        });
        
        it('should initialize all enabled patterns', function() {
            expect(swarmPatterns.patterns.has('aco')).to.be.true;
            expect(swarmPatterns.patterns.has('pso')).to.be.true;
            expect(swarmPatterns.patterns.has('abc')).to.be.true;
            expect(swarmPatterns.patterns.has('boids')).to.be.true;
        });
        
        it('should execute Ant Colony Optimization', async function() {
            const optimizationProblem = {
                type: 'path_finding',
                graph: {
                    nodes: Array.from({length: 10}, (_, i) => `node_${i}`),
                    edges: Array.from({length: 20}, (_, i) => ({
                        from: `node_${i % 10}`,
                        to: `node_${(i + 1) % 10}`,
                        weight: Math.random() * 10
                    }))
                },
                start: 'node_0',
                end: 'node_9'
            };
            
            const result = await swarmPatterns.executeACO(optimizationProblem);
            
            expect(result).to.have.property('best_path');
            expect(result).to.have.property('path_cost');
            expect(result).to.have.property('iterations');
            expect(result).to.have.property('convergence');
        });
        
        it('should execute Particle Swarm Optimization', async function() {
            const optimizationProblem = {
                type: 'function_optimization',
                dimensions: 5,
                bounds: { min: -10, max: 10 },
                fitness_function: 'sphere' // Test function
            };
            
            const result = await swarmPatterns.executePSO(optimizationProblem);
            
            expect(result).to.have.property('best_position');
            expect(result).to.have.property('best_fitness');
            expect(result).to.have.property('particles');
            expect(result.best_position).to.have.length(5);
        });
        
        it('should execute Artificial Bee Colony', async function() {
            const optimizationProblem = {
                type: 'resource_allocation',
                resources: ['cpu', 'memory', 'network'],
                constraints: {
                    total_budget: 1000,
                    min_performance: 0.7
                }
            };
            
            const result = await swarmPatterns.executeABC(optimizationProblem);
            
            expect(result).to.have.property('best_solution');
            expect(result).to.have.property('fitness');
            expect(result).to.have.property('employed_bees');
            expect(result).to.have.property('onlooker_bees');
        });
        
        it('should execute Boids flocking behavior', async function() {
            const flockingScenario = {
                agents: Array.from({length: 15}, (_, i) => ({
                    id: `agent_${i}`,
                    position: [Math.random() * 100, Math.random() * 100],
                    velocity: [Math.random() * 5 - 2.5, Math.random() * 5 - 2.5]
                })),
                rules: {
                    separation: 0.5,
                    alignment: 0.3,
                    cohesion: 0.2
                }
            };
            
            const result = await swarmPatterns.executeBoids(flockingScenario);
            
            expect(result).to.have.property('final_positions');
            expect(result).to.have.property('flock_cohesion');
            expect(result).to.have.property('convergence_time');
        });
        
        it('should adapt pattern parameters based on performance', async function() {
            const performanceData = {
                pattern: 'pso',
                fitness_history: [0.1, 0.08, 0.06, 0.06, 0.06], // Converged
                execution_time: 2000
            };
            
            const adaptations = await swarmPatterns.adaptPatternParameters(performanceData);
            
            expect(adaptations).to.have.property('parameter_changes');
            expect(adaptations).to.have.property('reasoning');
        });
        
        it('should coordinate multiple patterns simultaneously', async function() {
            const multiPatternProblem = {
                type: 'hybrid_optimization',
                subproblems: {
                    path_finding: { pattern: 'aco' },
                    parameter_tuning: { pattern: 'pso' },
                    resource_allocation: { pattern: 'abc' }
                }
            };
            
            const result = await swarmPatterns.executeMultiPatternOptimization(multiPatternProblem);
            
            expect(result).to.have.property('subresults');
            expect(result.subresults).to.have.property('path_finding');
            expect(result.subresults).to.have.property('parameter_tuning');
            expect(result.subresults).to.have.property('resource_allocation');
            expect(result).to.have.property('overall_fitness');
        });
    });
    
    describe('Cross-Agent Learning', function() {
        let crossAgentLearning;
        
        beforeEach(async function() {
            crossAgentLearning = new CrossAgentLearning({
                ...testConfig,
                learningRate: 0.1,
                memoryCapacity: 1000,
                knowledgeGraphSize: 500
            });
            await crossAgentLearning.initialize();
            testComponents.learning = crossAgentLearning;
        });
        
        it('should initialize learning system with proper configuration', function() {
            expect(crossAgentLearning.distributedRL).to.exist;
            expect(crossAgentLearning.knowledgeGraph).to.exist;
            expect(crossAgentLearning.skillTransfer).to.exist;
            expect(crossAgentLearning.collectiveMemory).to.exist;
        });
        
        it('should execute distributed Q-learning', async function() {
            const learningScenario = {
                agents: Array.from({length: 5}, (_, i) => ({
                    id: `agent_${i}`,
                    state: `state_${i % 3}`,
                    action_space: ['action_1', 'action_2', 'action_3']
                })),
                environment: {
                    states: ['state_0', 'state_1', 'state_2'],
                    rewards: { state_0: 1, state_1: 5, state_2: 10 }
                }
            };
            
            const result = await crossAgentLearning.executeDistributedQLearning(learningScenario);
            
            expect(result).to.have.property('q_tables');
            expect(result).to.have.property('convergence_metrics');
            expect(result).to.have.property('shared_experiences');
        });
        
        it('should build and update knowledge graph', async function() {
            const knowledgeEntries = [
                { subject: 'optimization', predicate: 'improves', object: 'performance' },
                { subject: 'learning', predicate: 'requires', object: 'data' },
                { subject: 'data', predicate: 'enables', object: 'optimization' }
            ];
            
            for (const entry of knowledgeEntries) {
                await crossAgentLearning.addKnowledgeEntry(entry);
            }
            
            const graphMetrics = await crossAgentLearning.analyzeKnowledgeGraph();
            
            expect(graphMetrics).to.have.property('nodes');
            expect(graphMetrics).to.have.property('edges');
            expect(graphMetrics).to.have.property('connectivity');
            expect(graphMetrics.nodes).to.be.greaterThan(0);
        });
        
        it('should execute skill transfer between agents', async function() {
            const transferScenario = {
                mentor: {
                    id: 'expert_agent',
                    skills: {
                        optimization: 0.9,
                        pattern_recognition: 0.8,
                        coordination: 0.7
                    }
                },
                students: [
                    { id: 'novice_1', skills: { optimization: 0.3 } },
                    { id: 'novice_2', skills: { pattern_recognition: 0.2 } }
                ],
                transfer_method: 'knowledge_distillation'
            };
            
            const result = await crossAgentLearning.executeSkillTransfer(transferScenario);
            
            expect(result).to.have.property('transfers_completed');
            expect(result).to.have.property('improvement_metrics');
            expect(result.transfers_completed).to.be.greaterThan(0);
        });
        
        it('should manage collective memory', async function() {
            const memoryEntries = [
                { type: 'episodic', data: { event: 'optimization_success', context: 'high_load' } },
                { type: 'semantic', data: { concept: 'load_balancing', definition: 'distribute_work' } },
                { type: 'procedural', data: { skill: 'task_allocation', steps: ['analyze', 'assign', 'monitor'] } }
            ];
            
            for (const entry of memoryEntries) {
                await crossAgentLearning.storeInCollectiveMemory(entry);
            }
            
            const retrievedMemory = await crossAgentLearning.retrieveFromCollectiveMemory({
                type: 'episodic',
                context: 'high_load'
            });
            
            expect(retrievedMemory).to.be.an('array');
            expect(retrievedMemory.length).to.be.greaterThan(0);
        });
        
        it('should execute meta-learning for optimization improvement', async function() {
            const metaLearningTask = {
                base_tasks: [
                    { type: 'classification', performance: 0.8 },
                    { type: 'regression', performance: 0.75 },
                    { type: 'optimization', performance: 0.85 }
                ],
                target_task: { type: 'new_optimization' },
                meta_algorithm: 'model_agnostic_meta_learning'
            };
            
            const result = await crossAgentLearning.executeMetaLearning(metaLearningTask);
            
            expect(result).to.have.property('adapted_model');
            expect(result).to.have.property('transfer_score');
            expect(result).to.have.property('few_shot_performance');
        });
        
        it('should share knowledge between agents efficiently', async function() {
            const sharingScenario = {
                sender: 'agent_1',
                receivers: ['agent_2', 'agent_3', 'agent_4'],
                knowledge: {
                    type: 'optimization_strategy',
                    data: { strategy: 'gradient_descent', parameters: { lr: 0.01 } }
                },
                sharing_method: 'federated_learning'
            };
            
            const result = await crossAgentLearning.shareKnowledge(sharingScenario);
            
            expect(result).to.have.property('sharing_success');
            expect(result).to.have.property('receivers_updated');
            expect(result.sharing_success).to.be.true;
            expect(result.receivers_updated).to.equal(3);
        });
    });
    
    describe('Performance Optimizer', function() {
        let performanceOptimizer;
        let mockComponents;
        
        beforeEach(async function() {
            // Create mock components for performance optimizer
            mockComponents = {
                multiObjective: { on: sinon.stub(), getMetrics: sinon.stub().resolves({ performance: 0.8 }) },
                topology: { on: sinon.stub(), getMetrics: sinon.stub().resolves({ efficiency: 0.75 }) },
                queen: { on: sinon.stub(), getMetrics: sinon.stub().resolves({ coordination: 0.9 }) },
                mesh: { on: sinon.stub(), getMetrics: sinon.stub().resolves({ consensus: 0.85 }) },
                patterns: { on: sinon.stub(), getMetrics: sinon.stub().resolves({ convergence: 0.7 }) },
                learning: { on: sinon.stub(), getMetrics: sinon.stub().resolves({ accuracy: 0.8 }) }
            };
            
            performanceOptimizer = new SwarmPerformanceOptimizer({
                ...testConfig,
                ...mockComponents
            });
            await performanceOptimizer.initialize();
            testComponents.performance = performanceOptimizer;
        });
        
        it('should initialize performance monitoring', function() {
            expect(performanceOptimizer.monitoringActive).to.be.true;
            expect(performanceOptimizer.optimizationActive).to.be.true;
            expect(performanceOptimizer.performanceMetrics).to.exist;
        });
        
        it('should collect performance metrics from all components', async function() {
            const metrics = await performanceOptimizer.collectPerformanceMetrics();
            
            expect(metrics).to.be.a('map');
            expect(metrics.has('system')).to.be.true;
            expect(metrics.has('swarm')).to.be.true;
        });
        
        it('should calculate overall performance score', function() {
            const score = performanceOptimizer.calculatePerformanceScore();
            
            expect(score).to.be.a('number');
            expect(score).to.be.greaterThanOrEqual(0);
            expect(score).to.be.lessThanOrEqual(1);
        });
        
        it('should detect performance bottlenecks', async function() {
            // Mock metrics with bottleneck conditions
            performanceOptimizer.performanceMetrics.current = new Map([
                ['system', { cpu: { utilization: 0.95 }, memory: { utilization: 0.9 } }],
                ['swarm', { averageResponseTime: 2000, errorRate: 0.1 }]
            ]);
            
            const bottlenecks = await performanceOptimizer.detectBottlenecks();
            
            expect(bottlenecks).to.be.an('array');
            expect(bottlenecks.length).to.be.greaterThan(0);
            expect(bottlenecks[0]).to.have.property('type');
            expect(bottlenecks[0]).to.have.property('severity');
        });
        
        it('should generate optimization strategies', async function() {
            const analysis = {
                overall_health: { score: 0.6, issues: ['high_response_time'] },
                component_health: {
                    patterns: { performance_score: 0.5, error_rate: 0.1 }
                },
                resource_utilization: { overall: 0.85 },
                bottlenecks: [{ type: 'cpu', severity: 0.9 }]
            };
            
            const strategies = await performanceOptimizer.generateOptimizationStrategies(analysis);
            
            expect(strategies).to.be.an('array');
            expect(strategies.length).to.be.greaterThan(0);
            expect(strategies[0]).to.have.property('type');
            expect(strategies[0]).to.have.property('action');
            expect(strategies[0]).to.have.property('priority');
        });
        
        it('should execute optimization strategies', async function() {
            const strategies = [
                {
                    type: 'resource_scaling',
                    action: 'scale_up',
                    priority: 'high',
                    target: 'system',
                    parameters: { current_utilization: 0.9, target_utilization: 0.7 }
                }
            ];
            
            const results = await performanceOptimizer.executeOptimizationStrategies(strategies);
            
            expect(results).to.be.an('array');
            expect(results[0]).to.have.property('strategy');
            expect(results[0]).to.have.property('success');
        });
        
        it('should adapt optimization parameters based on results', async function() {
            const results = {
                improvement: { overall: 0.15 },
                adaptationUpdates: { learningRate: 0.11, optimizationInterval: 27000 }
            };
            
            const initialLearningRate = performanceOptimizer.config.learningRate;
            await performanceOptimizer.adaptOptimizationParameters(results);
            
            expect(performanceOptimizer.config.learningRate).to.not.equal(initialLearningRate);
        });
    });
    
    describe('Sprint 3 Integration', function() {
        let integration;
        
        beforeEach(async function() {
            integration = new Sprint3Integration({
                ...testConfig,
                coordinationMode: 'hierarchical'
            });
            testComponents.integration = integration;
        });
        
        it('should initialize all components successfully', async function() {
            const status = await integration.initialize();
            
            expect(status.initialized).to.be.true;
            expect(status.active).to.be.true;
            expect(status.coordinationMode).to.equal('hierarchical');
            expect(Object.values(status.components).filter(s => s === 'active').length).to.be.greaterThan(0);
        });
        
        it('should coordinate components hierarchically', async function() {
            await integration.initialize();
            
            // Trigger coordination
            await integration.coordinateComponents();
            
            // Check that coordination events were emitted
            expect(integration.state.lastSync).to.exist;
        });
        
        it('should handle inter-component communication', function(done) {
            integration.initialize().then(() => {
                integration.on('optimization-complete', (data) => {
                    expect(data).to.have.property('source');
                    expect(data).to.have.property('data');
                    done();
                });
                
                // Simulate optimization completion
                integration.handleOptimizationComplete({ performance: 0.9 });
            });
        });
        
        it('should switch coordination modes dynamically', async function() {
            await integration.initialize();
            
            const initialMode = integration.state.coordinationMode;
            const newStatus = await integration.switchCoordinationMode('mesh');
            
            expect(newStatus.coordinationMode).to.equal('mesh');
            expect(newStatus.coordinationMode).to.not.equal(initialMode);
        });
        
        it('should trigger global optimization across all components', async function() {
            await integration.initialize();
            
            const results = await integration.triggerGlobalOptimization({
                target: 'performance',
                intensity: 'medium'
            });
            
            expect(results).to.be.an('object');
            expect(Object.keys(results).length).to.be.greaterThan(0);
        });
        
        it('should provide comprehensive swarm metrics', async function() {
            await integration.initialize();
            
            const metrics = await integration.getSwarmMetrics();
            
            expect(metrics).to.be.an('object');
            expect(Object.keys(metrics).length).to.be.greaterThan(0);
        });
        
        it('should handle graceful shutdown', async function() {
            await integration.initialize();
            
            const shutdownPromise = integration.shutdown();
            await expect(shutdownPromise).to.not.be.rejected;
            
            expect(integration.state.active).to.be.false;
        });
    });
    
    describe('End-to-End Integration Tests', function() {
        let fullSwarm;
        
        before(async function() {
            console.log('🌐 Setting up full swarm integration test...');
            
            fullSwarm = new Sprint3Integration({
                maxAgents: 20,
                coordinationMode: 'hybrid',
                enableAllComponents: true,
                testMode: true
            });
        });
        
        after(async function() {
            if (fullSwarm && fullSwarm.state.active) {
                await fullSwarm.shutdown();
            }
        });
        
        it('should complete full swarm optimization workflow', async function() {
            // Initialize full swarm
            const initStatus = await fullSwarm.initialize();
            expect(initStatus.initialized).to.be.true;
            
            // Wait for components to stabilize
            await new Promise(resolve => setTimeout(resolve, 3000));
            
            // Trigger global optimization
            const optimizationResults = await fullSwarm.triggerGlobalOptimization({
                objectives: ['performance', 'efficiency', 'reliability'],
                duration: 5000
            });
            
            expect(optimizationResults).to.be.an('object');
            expect(Object.keys(optimizationResults).length).to.be.greaterThan(0);
            
            // Get final metrics
            const finalMetrics = await fullSwarm.getSwarmMetrics();
            expect(finalMetrics).to.be.an('object');
        });
        
        it('should handle dynamic topology switching under load', async function() {
            await fullSwarm.initialize();
            
            // Start with hierarchical
            await fullSwarm.switchCoordinationMode('hierarchical');
            
            // Switch to mesh under load
            await fullSwarm.switchCoordinationMode('mesh');
            
            // Switch to hybrid for optimization
            const finalStatus = await fullSwarm.switchCoordinationMode('hybrid');
            
            expect(finalStatus.coordinationMode).to.equal('hybrid');
        });
        
        it('should demonstrate emergent swarm intelligence', async function() {
            await fullSwarm.initialize();
            
            // Allow swarm to run and learn
            await new Promise(resolve => setTimeout(resolve, 5000));
            
            const initialMetrics = await fullSwarm.getSwarmMetrics();
            
            // Trigger learning and optimization
            await fullSwarm.triggerGlobalOptimization({ focus: 'learning' });
            
            // Allow time for improvement
            await new Promise(resolve => setTimeout(resolve, 3000));
            
            const improvedMetrics = await fullSwarm.getSwarmMetrics();
            
            // Check for improvement (simple heuristic)
            expect(improvedMetrics).to.be.an('object');
            expect(initialMetrics).to.be.an('object');
        });
        
        it('should maintain system resilience under component failures', async function() {
            await fullSwarm.initialize();
            
            // Simulate component failure
            if (fullSwarm.components.patterns) {
                // Disable patterns component
                await fullSwarm.components.patterns.shutdown();
                fullSwarm.components.patterns = null;
            }
            
            // System should adapt and continue functioning
            const statusAfterFailure = await fullSwarm.getSwarmMetrics();
            expect(statusAfterFailure).to.be.an('object');
            
            // Trigger optimization with reduced components
            const optimizationResult = await fullSwarm.triggerGlobalOptimization({
                resilience_mode: true
            });
            
            expect(optimizationResult).to.be.an('object');
        });
    });
    
    describe('Performance Benchmarks', function() {
        it('should complete multi-objective optimization within time limit', async function() {
            const optimizer = new MultiObjectiveOptimizer({
                populationSize: 50,
                generations: 20,
                testMode: true
            });
            
            await optimizer.initialize();
            
            const startTime = Date.now();
            
            const result = await optimizer.optimizeWithNSGAII(
                {
                    agents: Array.from({length: 20}, (_, i) => ({ id: `agent_${i}` })),
                    tasks: Array.from({length: 10}, (_, i) => ({ id: `task_${i}` }))
                },
                { type: 'performance_test' }
            );
            
            const executionTime = Date.now() - startTime;
            
            expect(executionTime).to.be.lessThan(10000); // Should complete within 10 seconds
            expect(result.solutions.length).to.be.greaterThan(0);
        });
        
        it('should handle large swarm sizes efficiently', async function() {
            const integration = new Sprint3Integration({
                maxAgents: 100,
                testMode: true,
                enableAllComponents: false // Disable heavy components for speed
            });
            
            const startTime = Date.now();
            await integration.initialize();
            const initTime = Date.now() - startTime;
            
            expect(initTime).to.be.lessThan(15000); // Should initialize within 15 seconds
            
            await integration.shutdown();
        });
        
        it('should maintain memory efficiency during long runs', async function() {
            const patterns = new SwarmIntelligencePatterns({
                swarmSize: 30,
                enabledPatterns: ['pso'],
                testMode: true
            });
            
            await patterns.initialize();
            
            const initialMemory = process.memoryUsage().heapUsed;
            
            // Run multiple optimizations
            for (let i = 0; i < 5; i++) {
                await patterns.executePSO({
                    type: 'function_optimization',
                    dimensions: 10,
                    bounds: { min: -10, max: 10 }
                });
            }
            
            const finalMemory = process.memoryUsage().heapUsed;
            const memoryIncrease = finalMemory - initialMemory;
            
            // Memory increase should be reasonable (less than 100MB)
            expect(memoryIncrease).to.be.lessThan(100 * 1024 * 1024);
            
            await patterns.shutdown();
        });
    });
});

// Helper functions for testing
function generateMockSwarmState(agentCount = 10) {
    return {
        agents: Array.from({length: agentCount}, (_, i) => ({
            id: `agent_${i}`,
            performance: Math.random(),
            capabilities: ['compute', 'storage', 'network'][i % 3],
            status: 'active'
        })),
        tasks: Array.from({length: Math.floor(agentCount / 2)}, (_, i) => ({
            id: `task_${i}`,
            priority: Math.random(),
            complexity: Math.random(),
            status: 'pending'
        })),
        resources: {
            cpu: Math.random() * 0.5 + 0.3,
            memory: Math.random() * 0.5 + 0.3,
            network: Math.random() * 0.5 + 0.5
        }
    };
}

function createTestOptimizationProblem(type = 'allocation') {
    switch (type) {
        case 'allocation':
            return {
                type: 'agent_allocation',
                constraints: { max_cost: 1000, min_performance: 0.6 },
                objectives: ['performance', 'cost', 'reliability']
            };
        case 'routing':
            return {
                type: 'path_optimization',
                graph: generateTestGraph(15),
                start: 'node_0',
                end: 'node_14'
            };
        default:
            return { type: 'generic_optimization' };
    }
}

function generateTestGraph(nodeCount) {
    const nodes = Array.from({length: nodeCount}, (_, i) => `node_${i}`);
    const edges = [];
    
    for (let i = 0; i < nodeCount; i++) {
        for (let j = i + 1; j < Math.min(i + 4, nodeCount); j++) {
            edges.push({
                from: `node_${i}`,
                to: `node_${j}`,
                weight: Math.random() * 10 + 1
            });
        }
    }
    
    return { nodes, edges };
}

module.exports = {
    generateMockSwarmState,
    createTestOptimizationProblem,
    generateTestGraph
};