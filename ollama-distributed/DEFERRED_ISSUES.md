# Deferred Issues Log

This document tracks issues that were identified but deferred to be fixed at the appropriate time during the implementation phases.

## Phase 1 Deferred Issues

### 1. Import Path Conflicts (Partially Resolved)
**Status**: Deferred to Phase 2
**Files Affected**:
- `pkg/scheduler/partitioning/strategies.go` - User reverted to ollama imports
- `test_enhanced_scheduler_components_main.go` - Multiple undefined types

**Issue**: Auto-formatting keeps reverting integration imports back to ollama imports. The code uses types like `api.Options`, `server.Model`, `ggml.GGML`, `discover.GpuInfoList` that need proper integration stubs.

**Resolution Plan**: 
- Phase 2: Create comprehensive integration stubs that match all ollama types
- Implement proper type aliases in pkg/integration
- Update all references to use integration types consistently

### 2. Fault Tolerance Type Mismatches
**Status**: Deferred to Phase 2
**Files Affected**:
- `pkg/scheduler/fault_tolerance/self_healing_engine.go`
- `pkg/scheduler/fault_tolerance/enhanced_fault_tolerance.go`
- `pkg/scheduler/fault_tolerance/predictive_detection.go`

**Issue**: Multiple undefined types and interface mismatches:
- `SelfHealingEngine` vs `SelfHealingEngineImpl`
- `HealingAttempt` vs `HealingAttemptImpl`
- `Fault` vs `FaultDetection`
- Missing method implementations

**Resolution Plan**:
- Phase 2: Refactor fault tolerance types for consistency
- Implement missing interfaces and methods
- Add proper type definitions and aliases

### 3. Test File Compilation Issues
**Status**: Deferred to Phase 5
**Files Affected**:
- `tests/unit/scheduler_test.go`
- `tests/unit/actual_api_test.go`
- Various test files with undefined types

**Issue**: Test files reference non-existent types and methods, causing compilation failures.

**Resolution Plan**:
- Phase 5: Comprehensive test suite implementation
- Create proper test fixtures and mocks
- Implement missing test infrastructure

### 4. Enhanced Scheduler Components
**Status**: Deferred to Phase 2
**Files Affected**:
- `test_enhanced_scheduler_components_main.go`

**Issue**: File contains advanced scheduler features with many undefined types and methods. Appears to be prototype code for enhanced partitioning.

**Resolution Plan**:
- Phase 2: Implement basic scheduler first
- Phase 4: Add enhanced features and optimizations
- Consider moving to separate enhancement package

## Tracking Guidelines

When deferring an issue:
1. Document the specific files and line numbers affected
2. Describe the root cause of the issue
3. Specify which phase should address it
4. Provide a clear resolution plan
5. Update this log when issues are resolved

## Resolution Status

- [ ] Import Path Conflicts - Target: Phase 2
- [ ] Fault Tolerance Types - Target: Phase 2  
- [ ] Test Compilation - Target: Phase 5
- [ ] Enhanced Scheduler - Target: Phase 2/4

## Notes

- Focus on getting core packages (config, types, integration, p2p, consensus) building first
- Defer complex features until basic functionality is working
- Prioritize issues that block other development
- Some prototype/experimental code may need significant refactoring
