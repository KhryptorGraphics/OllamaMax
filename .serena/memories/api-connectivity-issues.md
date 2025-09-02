# Critical API Connectivity Issues Identified

## 🚨 Root Cause Analysis

### 1. **Port Mismatch Issue**
**Problem**: Frontend is calling mixed API endpoints:
- Basic nodes API: `http://localhost:13000/api/nodes` (WRONG PORT)
- Enhanced nodes API: `http://localhost:13100/api/nodes/detailed` (CORRECT PORT)
- Models API: `http://localhost:13100/api/models` (CORRECT PORT)

**Location**: `web-interface/app.js:321`
```javascript
// INCORRECT - should be 13100, not 13000
const response = await fetch('http://localhost:13000/api/nodes');
```

### 2. **Missing DOM Element References**
**Problem**: HTML template references non-existent DOM elements:
- `enhancedNodesGrid` referenced but should be `enhancedNodesContainer` 
- Missing container element for enhanced nodes display

### 3. **WebSocket Connection Issues**
**Problem**: Chat page WebSocket trying to connect to wrong endpoint
- Default: `ws://localhost:13100/chat`
- But WebSocket server may not be properly initialized

### 4. **P2P Model Migration Logic**
**Problem**: Frontend has P2P controls but backend API endpoints incomplete:
- `propagateModel()` function calls `/api/models/propagate`
- Settings have `p2pEnabled` checkbox but no backend implementation

## 🔧 Required Fixes

1. **Fix port mismatch in nodes API call**
2. **Correct DOM element references** 
3. **Complete WebSocket server implementation**
4. **Implement missing P2P model migration backend endpoints**
5. **Add proper error handling for all API calls**
6. **Complete enhanced node management backend APIs**