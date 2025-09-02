# Task Completion Workflow

## After Code Changes
1. **Rebuild Docker containers**: `docker-compose -f docker-compose.dev.yml up -d --build`
2. **Check service health**: `docker-compose -f docker-compose.dev.yml ps`
3. **View logs**: `docker-compose logs -f distributed-api`
4. **Test API endpoints**: 
   - Health: `curl http://localhost:13100/api/health`
   - Nodes: `curl http://localhost:13100/api/nodes/detailed`
   - Models: `curl http://localhost:13100/api/models`

## For Swarm Testing (Production)
1. **Deploy to swarm**: `docker stack deploy -c docker-swarm.yml ollamamax`
2. **Check stack status**: `docker stack services ollamamax`
3. **Scale services**: `docker service scale ollamamax_distributed-api=3`
4. **Monitor logs**: `docker service logs -f ollamamax_distributed-api`

## Quality Checks
- **No lint errors**: All code passes linting
- **API functionality**: All endpoints return expected responses
- **WebSocket connectivity**: Real-time features work
- **P2P model migration**: Model sharing between nodes functional
- **Monitoring integration**: Prometheus metrics collection active