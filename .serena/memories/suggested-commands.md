# OllamaMax Development Commands

## Docker Commands
- `docker-compose -f docker-compose.dev.yml up -d` - Start development stack
- `docker-compose -f docker-compose.distributed.yml up -d` - Start production stack  
- `./deploy-docker.sh --dev` - Automated development deployment
- `./deploy-docker.sh --prod` - Automated production deployment

## Testing Commands
- `npm test` - Run Node.js backend tests
- `npm run lint` - Code linting
- `node api-server/distributed-inference.js` - Run API server directly

## Monitoring Commands
- `docker-compose logs -f distributed-api` - View API server logs
- `docker-compose ps` - Check service status
- `curl http://localhost:13100/api/health` - Test API health

## Development Workflow
1. Start services: `docker-compose -f docker-compose.dev.yml up -d`
2. View logs: `docker-compose logs -f`
3. Test API: `curl http://localhost:13100/api/nodes/detailed`
4. Access web UI: `http://localhost:13100`

## File Structure Commands
- API Server: `api-server/distributed-inference.js`
- Web Interface: `web-interface/index.html`
- Docker configs: `docker-compose*.yml`
- Deployment: `deploy-docker.sh`