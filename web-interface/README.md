# OllamaMax Web Interface

The web interface for the OllamaMax distributed AI platform, providing a modern, responsive dashboard for managing distributed inference nodes, models, and chat interactions.

## Features

### 🚀 Core Functionality
- **Real-time Chat Interface**: WebSocket-based chat with streaming AI responses
- **Distributed Node Management**: Monitor and manage multiple inference nodes
- **Model Management**: Download, propagate, and manage AI models across nodes
- **Settings & Configuration**: Comprehensive settings with localStorage persistence
- **Help & Guidance**: Interactive tour system and contextual help

### 📊 Advanced Monitoring
- **Performance Dashboard**: Real-time charts using Chart.js for system metrics
- **Connection Quality Monitoring**: Latency tracking, packet loss detection, throughput monitoring
- **Enhanced WebSocket Management**: Auto-reconnection, multiple endpoints, connection quality indicators
- **System Status**: Maintenance mode, rate limiting, queue status

### ♿ Accessibility & UX
- **WCAG 2.1 AA Compliance**: Full accessibility support with ARIA labels, keyboard navigation
- **Responsive Design**: Mobile-first approach with breakpoints for all device sizes
- **Dark Mode Support**: Automatic dark theme detection and manual toggle
- **Progressive Enhancement**: Works without JavaScript, enhanced with modern features

### 🛡️ Production Ready
- **Security Features**: Input validation, XSS protection, secure WebSocket connections
- **Error Handling**: Graceful error recovery, connection failure handling, user feedback
- **Performance Optimized**: Lazy loading, efficient DOM updates, memory leak prevention
- **Comprehensive Testing**: Jest unit tests, integration tests, performance benchmarks

## Quick Start

### Development
```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Open http://localhost:3000
```

### Production Build
```bash
# Build for production
npm run build

# Serve built files
npm run serve

# Or use Docker
docker build -t ollamamax-web .
docker run -p 80:80 ollamamax-web
```

### Testing
```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run specific test files
npm test frontend.test.js
```

## Architecture

### File Structure
```
web-interface/
├── index.html              # Main HTML structure
├── app.js                  # JavaScript application logic
├── styles.css              # CSS styling with design system
├── package.json            # Dependencies and scripts
├── webpack.config.js       # Build configuration
├── Dockerfile             # Container configuration
├── nginx.conf             # Production web server config
├── tests/                 # Test suite
│   └── frontend.test.js   # Comprehensive test coverage
├── docs/                  # Documentation
└── dist/                  # Built production files
```

### Technology Stack
- **Frontend**: Vanilla JavaScript ES6+, CSS3, HTML5
- **Build Tool**: Webpack 5 with development server
- **Testing**: Jest with JSDOM for unit tests
- **Linting**: ESLint with comprehensive rules
- **Formatting**: Prettier for code consistency
- **Charts**: Chart.js for data visualization
- **Container**: Docker with Nginx for production

### Design System
- **Colors**: Modern palette with primary, secondary, and semantic colors
- **Typography**: Inter font family with responsive scale
- **Spacing**: 4px base unit with consistent scale
- **Components**: Reusable, accessible UI components
- **Themes**: Light/dark theme support with CSS custom properties

## API Integration

### WebSocket Endpoints
- **Chat**: `ws://localhost:13000/chat` - Main chat interface
- **Node Events**: Real-time node status updates
- **Metrics**: Performance and monitoring data
- **System Status**: Maintenance, rate limiting, queue updates

### HTTP Endpoints
- **Nodes**: `GET /api/nodes/detailed` - Node information
- **Models**: `GET /api/models` - Model inventory
- **Health**: `GET /health` - System health status
- **Metrics**: `GET /metrics` - Prometheus metrics

### Message Types
```javascript
// Chat messages
{ type: 'chat', content: 'Hello AI!' }

// System responses
{ type: 'response', content: 'Hello!', node: 'node-1', streaming: false }
{ type: 'stream_chunk', id: 'msg-1', chunk: 'Hello ', done: false }

// Node updates
{ type: 'node_update', node: 'node-1', status: 'healthy', load: 45 }
{ type: 'queue_update', length: 3, estimatedWaitTime: 1500 }

// Connection quality
{ type: 'ping', timestamp: 123456789 }
{ type: 'pong', timestamp: 123456789 }
{ type: 'connection_quality', latency: 45, packetLoss: 0.02 }

// System status
{ type: 'system_status', status: 'operational' }
{ type: 'system_status', maintenance: { enabled: true, message: 'Planned maintenance' } }
{ type: 'system_status', rateLimit: { remaining: 0, resetTime: 60000 } }
```

## Configuration

### Environment Variables
```bash
# API endpoints
API_ENDPOINT=ws://localhost:13000/chat
BACKUP_ENDPOINTS=ws://localhost:13001/chat,ws://localhost:13002/chat

# Chart.js settings
CHART_UPDATE_INTERVAL=30000  # 30 seconds
CHART_DATA_POINTS=50        # Max data points per chart

# WebSocket settings
WEBSOCKET_TIMEOUT=10000     # 10 seconds
RECONNECT_ATTEMPTS=10
RECONNECT_DELAY=1000        # Base delay in ms
```

### Settings Persistence
Settings are automatically saved to localStorage and include:
- API endpoint configuration
- Chat preferences (streaming, auto-scroll, max tokens, temperature)
- Load balancing strategy
- Dark mode preference
- Node filtering and sorting preferences
- Model propagation settings

### Advanced Configuration
```javascript
// Custom endpoints for load balancing
llamaClient.endpoints = [
    'ws://primary-api:13000/chat',
    'ws://backup-api-1:13000/chat',
    'ws://backup-api-2:13000/chat'
];

// Chart customization
llamaClient.chartOptions = {
    responsive: true,
    animation: false,
    elements: {
        point: { radius: 0 },
        line: { tension: 0.4 }
    }
};
```

## Deployment

### Docker Deployment
```bash
# Build and run
docker build -t ollamamax-web .
docker run -d -p 80:80 --name ollamamax-web ollamamax-web

# With custom configuration
docker run -d -p 80:80 \
    -e API_ENDPOINT=ws://api-server:13000/chat \
    --name ollamamax-web ollamamax-web
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollamamax-web
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ollamamax-web
  template:
    metadata:
      labels:
        app: ollamamax-web
    spec:
      containers:
      - name: web
        image: ollamamax-web:latest
        ports:
        - containerPort: 80
        env:
        - name: API_ENDPOINT
          value: "ws://ollamamax-api:13000/chat"
---
apiVersion: v1
kind: Service
metadata:
  name: ollamamax-web
spec:
  selector:
    app: ollamamax-web
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
```

### Nginx Configuration
The provided `nginx.conf` includes:
- Static file serving with caching
- API proxy with rate limiting
- WebSocket proxy for chat
- Security headers
- Gzip compression
- Health check endpoints

## Monitoring

### Performance Metrics
The web interface tracks and displays:
- **Connection Quality**: Latency, packet loss, throughput
- **System Performance**: CPU, memory, disk usage across nodes
- **Request Metrics**: Response times, queue lengths, error rates
- **Chart.js Visualizations**: Real-time line charts, bar charts, system overview

### Health Checks
- **Frontend Health**: Static file availability, JavaScript errors
- **WebSocket Connectivity**: Connection status, reconnection attempts
- **API Connectivity**: Backend API availability and response times
- **Performance Monitoring**: Memory usage, DOM update performance

### Logging
- **Console Logging**: Structured logging with levels
- **Error Tracking**: Automatic error capture and reporting
- **Performance Logging**: Chart update times, DOM operation times
- **User Actions**: Key user interactions for analytics

## Testing

### Test Coverage
- **Unit Tests**: Jest tests for all major functions (90%+ coverage)
- **Integration Tests**: WebSocket communication, API integration
- **Accessibility Tests**: ARIA compliance, keyboard navigation
- **Performance Tests**: Memory leak detection, DOM update performance
- **Cross-browser Tests**: Chrome, Firefox, Safari, Edge compatibility

### Running Tests
```bash
# All tests
npm test

# With coverage report
npm run test:coverage

# Watch mode for development
npm run test:watch

# Specific test files
npm test frontend.test.js

# Linting and formatting
npm run lint
npm run format
```

### Test Structure
```javascript
// Unit tests
describe('WebSocket Connection Management', () => {
    test('should handle connection failures gracefully', () => {
        // Test implementation
    });
});

// Integration tests
describe('Chat Functionality', () => {
    test('should send and receive messages', () => {
        // Test implementation
    });
});

// Performance tests
describe('Performance Tests', () => {
    test('should handle rapid message sending', () => {
        // Test implementation
    });
});
```

## Security

### Client-side Security
- **XSS Protection**: HTML escaping for all user content
- **Input Validation**: Client-side validation with server-side verification
- **WebSocket Security**: Secure WebSocket (wss://) support
- **CSP Headers**: Content Security Policy via Nginx

### Best Practices
- **No Secrets**: Never store API keys or secrets in client code
- **Input Sanitization**: All user input is escaped before DOM insertion
- **Secure Headers**: X-Frame-Options, X-Content-Type-Options, etc.
- **HTTPS Only**: Production deployment requires HTTPS

### Vulnerability Management
- **Dependency Scanning**: Regular npm audit runs
- **Security Updates**: Automated security patch updates
- **Code Review**: Security-focused code review process
- **Penetration Testing**: Regular security testing

## Troubleshooting

### Common Issues

**WebSocket Connection Failures**
```javascript
// Check console for connection errors
// Verify API endpoint is correct
// Check network connectivity
// Ensure backend server is running
```

**Chart.js Not Loading**
```html
<!-- Ensure Chart.js is loaded before app.js -->
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<script src="app.js"></script>
```

**Performance Issues**
```javascript
// Check for memory leaks in console
// Monitor DOM node count
// Verify chart update intervals
// Check WebSocket message volume
```

### Debug Mode
Enable debug logging:
```javascript
// Add to app initialization
llamaClient.debug = true;
```

### Logs and Monitoring
- **Browser Console**: Real-time debugging information
- **Network Tab**: WebSocket and HTTP request monitoring
- **Performance Tab**: Memory and CPU usage tracking
- **Application Tab**: localStorage and IndexedDB inspection

## Contributing

### Development Setup
1. Fork and clone the repository
2. Install dependencies: `npm install`
3. Start development server: `npm run dev`
4. Make changes and test thoroughly
5. Run tests: `npm test`
6. Submit pull request

### Code Guidelines
- **ESLint**: Follow all linting rules
- **Prettier**: Use consistent formatting
- **Comments**: Document complex logic
- **Tests**: Add tests for new features
- **Accessibility**: Ensure WCAG compliance

### Pull Request Process
1. Create feature branch from `develop`
2. Write comprehensive tests
3. Update documentation
4. Ensure all tests pass
5. Submit PR with detailed description
6. Address review feedback
7. Merge to `develop`

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

### Documentation
- [Frontend Specification](../../docs/distributed-llama-chat-frontend-spec.md)
- [Technical Specifications](../../TECHNICAL_SPECIFICATIONS.md)
- [Testing Summary](../../TESTING_SUMMARY.md)

### Issues
- [Bug Reports](../../issues)
- [Feature Requests](../../issues)
- [Technical Questions](../../discussions)

### Community
- [Discussions](../../discussions)
- [Contributing Guide](../../CONTRIBUTING.md)
- [Code of Conduct](../../CODE_OF_CONDUCT.md)

---

**OllamaMax Web Interface** - Modern, accessible, production-ready frontend for distributed AI inference management.