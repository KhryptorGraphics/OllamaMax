# AI UI Generation Prompt for Distributed Llama Chat Interface

## Prompt for v0, Lovable, or Similar AI UI Tools

Create a modern, responsive chat interface for a distributed Llama AI inference system with the following specifications:

### Core Requirements

Build a real-time chat application with:
- WebSocket-based messaging for instant communication
- Support for multiple distributed inference nodes
- Live streaming of AI responses character by character
- Node health monitoring and automatic failover
- Performance metrics visualization

### Visual Design

**Color Scheme:**
- Primary: #667eea (purple)
- Secondary: #764ba2 (deep purple)
- Success: #48bb78 (green)
- Warning: #ed8936 (orange)
- Error: #e53e3e (red)
- Background: Linear gradient from #667eea to #764ba2
- Card backgrounds: rgba(255,255,255,0.95) with backdrop blur

**Typography:**
- Font: Inter or system-ui
- Headers: Bold, 2rem/1.5rem/1.25rem
- Body: Regular, 1rem
- Code: 'Fira Code' or monospace

### Layout Structure

```
+----------------------------------+
|         Header Bar               |
|  [Chat] [Nodes] [Settings]       |
+----------------------------------+
|                                  |
|     Message Thread Area          |
|     - User messages (right)      |
|     - AI responses (left)        |
|     - Streaming indicators       |
|                                  |
+----------------------------------+
|  Status Bar                      |
|  Node: llama-01 | Queue: 3       |
+----------------------------------+
|  [Input field............] [Send]|
+----------------------------------+
```

### Components Needed

1. **ChatMessage Component**
```jsx
<div class="message {sender}">
  <div class="avatar">{icon}</div>
  <div class="content">
    <div class="header">
      <span class="name">{sender}</span>
      <span class="node-badge">{node}</span>
      <span class="timestamp">{time}</span>
    </div>
    <div class="text">{message}</div>
    <div class="streaming-cursor" v-if="streaming">▊</div>
  </div>
</div>
```

2. **NodeCard Component**
```jsx
<div class="node-card {status}">
  <div class="node-header">
    <h3>{nodeName}</h3>
    <span class="status-indicator"></span>
  </div>
  <div class="metrics">
    <div>Requests: {requestCount}/s</div>
    <div>Latency: {latency}ms</div>
    <div>Memory: {memory}%</div>
  </div>
  <canvas class="sparkline">{recentActivity}</canvas>
</div>
```

3. **Input Area Component**
```jsx
<div class="input-area">
  <select class="model-selector">
    <option>llama2-7b</option>
    <option>llama2-13b</option>
    <option>codellama</option>
  </select>
  <textarea 
    placeholder="Type your message..."
    onKeyDown={handleEnterKey}
  />
  <button class="send-button">
    <svg><!-- send icon --></svg>
  </button>
</div>
```

### Interactive Features

1. **Real-time Updates**
   - WebSocket connection indicator (green/red dot)
   - Typing indicators when AI is processing
   - Live streaming text with character-by-character appearance
   - Auto-scroll to bottom on new messages

2. **Node Dashboard View**
   - Grid layout of node cards
   - Real-time performance charts using Chart.js or similar
   - Drag-and-drop to reorder node priority
   - Click to view detailed logs

3. **Smart Features**
   - Automatic reconnection on disconnect
   - Message queue visualization
   - Copy code blocks with syntax highlighting
   - Markdown rendering for AI responses
   - File upload for context (PDF, TXT, MD)

### API Integration Code

```javascript
class DistributedLlamaClient {
  constructor() {
    this.ws = null;
    this.nodes = [];
    this.activeNode = null;
    this.messageQueue = [];
  }

  connect() {
    this.ws = new WebSocket('ws://localhost:13000/chat');
    
    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleMessage(data);
    };
    
    this.ws.onerror = () => {
      this.reconnect();
    };
  }

  sendMessage(message) {
    const payload = {
      type: 'inference',
      content: message,
      model: this.selectedModel,
      timestamp: Date.now()
    };
    
    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(payload));
    } else {
      this.messageQueue.push(payload);
    }
  }

  streamResponse(callback) {
    // Handle streaming response chunks
  }

  getNodeStatus() {
    return fetch('http://localhost:13000/api/nodes')
      .then(res => res.json());
  }
}
```

### Responsive Design

**Mobile (< 768px):**
- Single column layout
- Collapsible node dashboard
- Bottom sheet for settings
- Larger touch targets (44x44px minimum)

**Tablet (768px - 1024px):**
- Two-column layout with sidebar
- Floating action buttons
- Modal overlays for node details

**Desktop (> 1024px):**
- Three-panel layout
- Persistent sidebar
- Hover states and tooltips
- Keyboard shortcuts (Ctrl+Enter to send)

### Accessibility

- ARIA labels for all interactive elements
- Role="alert" for error messages
- Live regions for streaming text
- Keyboard navigation support
- High contrast mode support
- Screen reader announcements for status changes

### Animation Specifications

```css
/* Message appearance */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Typing indicator */
@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}

/* Streaming cursor */
@keyframes blink {
  50% { opacity: 0; }
}
```

### State Management

```javascript
const appState = {
  messages: [],
  nodes: [],
  activeNode: null,
  connectionStatus: 'connecting',
  streamingMessage: null,
  queueLength: 0,
  selectedModel: 'llama2-7b',
  performanceMetrics: {
    latency: [],
    throughput: [],
    memoryUsage: []
  }
};
```

### Error States

- Connection lost: Red banner with reconnection countdown
- Node offline: Strikethrough node card with "Offline" badge
- Rate limited: Orange warning with reset timer
- Invalid input: Red border with error message below

### Sample Test Data

```json
{
  "nodes": [
    {
      "id": "node-1",
      "name": "llama-01",
      "status": "healthy",
      "load": 45,
      "memory": 67,
      "requestsPerSecond": 12
    },
    {
      "id": "node-2",
      "name": "llama-02",
      "status": "warning",
      "load": 89,
      "memory": 92,
      "requestsPerSecond": 8
    }
  ],
  "messages": [
    {
      "id": "msg-1",
      "sender": "user",
      "content": "Explain quantum computing",
      "timestamp": "2025-09-02T10:30:00Z"
    },
    {
      "id": "msg-2",
      "sender": "ai",
      "content": "Quantum computing is...",
      "node": "llama-01",
      "timestamp": "2025-09-02T10:30:02Z",
      "streaming": true
    }
  ]
}
```

### Performance Optimizations

- Virtual scrolling for message history > 100 messages
- Debounced input for search
- Lazy loading for node dashboard
- WebSocket message batching
- CSS containment for better paint performance
- Web Workers for heavy computations

---

Use this specification to generate a complete, production-ready chat interface with all components, styling, and interactivity included. The interface should feel modern, responsive, and professional while maintaining excellent performance across all devices.