import React, { useState, useRef, useEffect } from 'react'
import { Card, CardHeader, CardContent, CardTitle } from '@/design-system/components/Card/Card'
import type { Node } from '../NodesPage'

interface NetworkTopologyProps {
  nodes: Node[]
  onNodeSelect: (node: Node) => void
}

interface NodePosition {
  x: number
  y: number
  node: Node
}

export const NetworkTopology: React.FC<NetworkTopologyProps> = ({ nodes, onNodeSelect }) => {
  const svgRef = useRef<SVGSVGElement>(null)
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'geographic' | 'logical' | 'performance'>('logical')
  const [showConnections, setShowConnections] = useState(true)
  const [showLabels, setShowLabels] = useState(true)

  // Calculate node positions based on view mode
  const getNodePositions = (): NodePosition[] => {
    const width = 800
    const height = 600
    const padding = 80

    if (viewMode === 'geographic') {
      // Group by region/datacenter
      const regions = new Map<string, Node[]>()
      nodes.forEach(node => {
        const region = node.location?.region || 'unknown'
        if (!regions.has(region)) regions.set(region, [])
        regions.get(region)!.push(node)
      })

      const positions: NodePosition[] = []
      const regionCount = regions.size
      let regionIndex = 0

      regions.forEach((regionNodes, region) => {
        const angle = (regionIndex / regionCount) * 2 * Math.PI
        const regionX = width / 2 + Math.cos(angle) * (width / 4)
        const regionY = height / 2 + Math.sin(angle) * (height / 4)

        regionNodes.forEach((node, nodeIndex) => {
          const nodeAngle = (nodeIndex / regionNodes.length) * 2 * Math.PI
          const x = regionX + Math.cos(nodeAngle) * 60
          const y = regionY + Math.sin(nodeAngle) * 60
          
          positions.push({
            x: Math.max(padding, Math.min(width - padding, x)),
            y: Math.max(padding, Math.min(height - padding, y)),
            node
          })
        })
        regionIndex++
      })

      return positions
    }

    if (viewMode === 'performance') {
      // Position based on performance metrics
      return nodes.map((node, index) => {
        const healthFactor = node.health_score / 100
        const performanceFactor = Math.min(1, node.performance.requests_per_second / 20)
        
        // High performance nodes toward center, lower performance toward edges
        const distance = (width / 3) * (1 - healthFactor)
        const angle = (index / nodes.length) * 2 * Math.PI + performanceFactor * 0.5
        
        return {
          x: Math.max(padding, Math.min(width - padding, width / 2 + Math.cos(angle) * distance)),
          y: Math.max(padding, Math.min(height - padding, height / 2 + Math.sin(angle) * distance)),
          node
        }
      })
    }

    // Logical view - circular layout with status grouping
    const statusGroups = {
      online: nodes.filter(n => n.status === 'online'),
      offline: nodes.filter(n => n.status === 'offline'),
      draining: nodes.filter(n => n.status === 'draining'),
      maintenance: nodes.filter(n => n.status === 'maintenance'),
      error: nodes.filter(n => n.status === 'error')
    }

    const positions: NodePosition[] = []
    let currentAngle = 0
    const centerX = width / 2
    const centerY = height / 2

    Object.entries(statusGroups).forEach(([status, statusNodes]) => {
      if (statusNodes.length === 0) return

      const radius = status === 'online' ? 120 : status === 'offline' ? 200 : 160
      const statusAngleSpan = (statusNodes.length / nodes.length) * 2 * Math.PI

      statusNodes.forEach((node, index) => {
        const angle = currentAngle + (index / statusNodes.length) * statusAngleSpan
        const x = centerX + Math.cos(angle) * radius
        const y = centerY + Math.sin(angle) * radius

        positions.push({
          x: Math.max(padding, Math.min(width - padding, x)),
          y: Math.max(padding, Math.min(height - padding, y)),
          node
        })
      })

      currentAngle += statusAngleSpan + 0.3 // Add gap between status groups
    })

    return positions
  }

  const nodePositions = getNodePositions()

  const getNodeColor = (node: Node) => {
    switch (node.status) {
      case 'online': return '#10b981' // success
      case 'offline': return '#ef4444' // error
      case 'draining': return '#f59e0b' // warning
      case 'maintenance': return '#3b82f6' // info
      case 'error': return '#dc2626' // error (darker)
      default: return '#6b7280' // muted
    }
  }

  const getNodeSize = (node: Node) => {
    if (viewMode === 'performance') {
      // Size based on performance metrics
      const healthFactor = node.health_score / 100
      const performanceFactor = Math.min(1, node.performance.requests_per_second / 20)
      return 8 + ((healthFactor + performanceFactor) / 2) * 12
    }
    
    // Size based on capabilities
    const baseSizeFactor = node.capabilities.gpu_enabled ? 1.5 : 1
    const capacityFactor = node.capabilities.max_concurrent_requests / 8
    return 8 + baseSizeFactor * capacityFactor * 3
  }

  const handleNodeClick = (node: Node) => {
    setSelectedNodeId(node.id)
    onNodeSelect(node)
  }

  const getConnectionStrength = (node1: Node, node2: Node) => {
    // Simulate connection strength based on region proximity and performance
    const sameRegion = node1.location?.region === node2.location?.region
    const sameDC = node1.location?.datacenter === node2.location?.datacenter
    
    if (sameDC) return 0.9
    if (sameRegion) return 0.6
    return 0.3
  }

  const shouldShowConnection = (node1: Node, node2: Node) => {
    if (!showConnections) return false
    
    // Show connections between nodes in same region or high-performance nodes
    const sameRegion = node1.location?.region === node2.location?.region
    const bothHighPerf = node1.health_score > 80 && node2.health_score > 80
    
    return sameRegion || bothHighPerf
  }

  const formatUptime = (uptime: number) => {
    return `${uptime.toFixed(1)}%`
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            Network Topology
          </CardTitle>
          
          <div className="flex items-center gap-2">
            <select
              value={viewMode}
              onChange={(e) => setViewMode(e.target.value as any)}
              className="text-xs px-2 py-1 border border-border rounded bg-background text-foreground"
            >
              <option value="logical">Logical</option>
              <option value="geographic">Geographic</option>
              <option value="performance">Performance</option>
            </select>
            
            <button
              onClick={() => setShowConnections(!showConnections)}
              className={`text-xs px-2 py-1 rounded ${showConnections ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'}`}
            >
              Connections
            </button>
            
            <button
              onClick={() => setShowLabels(!showLabels)}
              className={`text-xs px-2 py-1 rounded ${showLabels ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'}`}
            >
              Labels
            </button>
          </div>
        </div>
      </CardHeader>

      <CardContent>
        <div className="relative">
          <svg
            ref={svgRef}
            width="100%"
            height="600"
            viewBox="0 0 800 600"
            className="border border-border rounded-lg bg-card"
          >
            {/* Grid background */}
            <defs>
              <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
                <path d="M 40 0 L 0 0 0 40" fill="none" stroke="currentColor" strokeWidth="0.5" opacity="0.1"/>
              </pattern>
            </defs>
            <rect width="100%" height="100%" fill="url(#grid)" />

            {/* Connections */}
            {showConnections && nodePositions.map((pos1, i) =>
              nodePositions.slice(i + 1).map((pos2, j) => {
                if (!shouldShowConnection(pos1.node, pos2.node)) return null
                
                const strength = getConnectionStrength(pos1.node, pos2.node)
                const opacity = strength * 0.6
                const strokeWidth = strength * 2
                
                return (
                  <line
                    key={`${pos1.node.id}-${pos2.node.id}`}
                    x1={pos1.x}
                    y1={pos1.y}
                    x2={pos2.x}
                    y2={pos2.y}
                    stroke="currentColor"
                    strokeWidth={strokeWidth}
                    opacity={opacity}
                    className="text-muted-foreground"
                  />
                )
              })
            )}

            {/* Region/Datacenter labels (in geographic mode) */}
            {viewMode === 'geographic' && showLabels && (
              <>
                {Array.from(new Set(nodes.map(n => n.location?.region).filter(Boolean))).map((region, index) => {
                  const regionNodes = nodes.filter(n => n.location?.region === region)
                  if (regionNodes.length === 0) return null
                  
                  const regionPositions = nodePositions.filter(p => p.node.location?.region === region)
                  const avgX = regionPositions.reduce((sum, p) => sum + p.x, 0) / regionPositions.length
                  const avgY = regionPositions.reduce((sum, p) => sum + p.y, 0) / regionPositions.length
                  
                  return (
                    <text
                      key={region}
                      x={avgX}
                      y={avgY - 40}
                      textAnchor="middle"
                      className="text-xs font-medium fill-muted-foreground"
                    >
                      {region}
                    </text>
                  )
                })}
              </>
            )}

            {/* Nodes */}
            {nodePositions.map(({ x, y, node }) => {
              const isSelected = selectedNodeId === node.id
              const isHovered = hoveredNodeId === node.id
              const nodeSize = getNodeSize(node)
              const nodeColor = getNodeColor(node)
              
              return (
                <g key={node.id}>
                  {/* Node glow effect for selected/hovered */}
                  {(isSelected || isHovered) && (
                    <circle
                      cx={x}
                      cy={y}
                      r={nodeSize + 8}
                      fill={nodeColor}
                      opacity="0.3"
                      className="animate-pulse"
                    />
                  )}
                  
                  {/* Main node circle */}
                  <circle
                    cx={x}
                    cy={y}
                    r={nodeSize}
                    fill={nodeColor}
                    stroke={isSelected ? 'currentColor' : 'transparent'}
                    strokeWidth="2"
                    className={`cursor-pointer transition-all ${isSelected ? 'text-primary' : ''}`}
                    onClick={() => handleNodeClick(node)}
                    onMouseEnter={() => setHoveredNodeId(node.id)}
                    onMouseLeave={() => setHoveredNodeId(null)}
                  />
                  
                  {/* GPU indicator */}
                  {node.capabilities.gpu_enabled && (
                    <circle
                      cx={x + nodeSize - 3}
                      cy={y - nodeSize + 3}
                      r="3"
                      fill="#8b5cf6"
                      stroke="white"
                      strokeWidth="1"
                    />
                  )}
                  
                  {/* Status indicator */}
                  <circle
                    cx={x + nodeSize - 2}
                    cy={y + nodeSize - 2}
                    r="2"
                    fill={node.status === 'online' ? '#10b981' : node.status === 'offline' ? '#ef4444' : '#f59e0b'}
                  />
                  
                  {/* Node label */}
                  {showLabels && (
                    <text
                      x={x}
                      y={y + nodeSize + 15}
                      textAnchor="middle"
                      className="text-xs font-medium fill-foreground"
                    >
                      {node.name}
                    </text>
                  )}
                  
                  {/* Performance metrics (in performance mode) */}
                  {viewMode === 'performance' && showLabels && (
                    <text
                      x={x}
                      y={y + nodeSize + 28}
                      textAnchor="middle"
                      className="text-xs fill-muted-foreground"
                    >
                      {node.performance.requests_per_second.toFixed(1)} RPS
                    </text>
                  )}
                  
                  {/* Tooltip on hover */}
                  {isHovered && (
                    <g>
                      <rect
                        x={x + nodeSize + 10}
                        y={y - 30}
                        width="120"
                        height="60"
                        rx="4"
                        fill="currentColor"
                        className="text-background opacity-90"
                        stroke="currentColor"
                        strokeWidth="1"
                        className="text-border"
                      />
                      <text x={x + nodeSize + 16} y={y - 15} className="text-xs font-medium fill-foreground">
                        {node.name}
                      </text>
                      <text x={x + nodeSize + 16} y={y - 2} className="text-xs fill-muted-foreground">
                        Health: {node.health_score}%
                      </text>
                      <text x={x + nodeSize + 16} y={y + 11} className="text-xs fill-muted-foreground">
                        CPU: {node.resources.cpu_usage}%
                      </text>
                      <text x={x + nodeSize + 16} y={y + 24} className="text-xs fill-muted-foreground">
                        {node.resources.active_requests} active
                      </text>
                    </g>
                  )}
                </g>
              )
            })}
          </svg>

          {/* Legend */}
          <div className="absolute top-4 right-4 bg-card border border-border rounded-lg p-3 text-xs">
            <h4 className="font-medium text-foreground mb-2">Legend</h4>
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <circle cx="6" cy="6" r="4" fill="#10b981" className="w-3 h-3" />
                <span className="text-muted-foreground">Online</span>
              </div>
              <div className="flex items-center gap-2">
                <circle cx="6" cy="6" r="4" fill="#ef4444" className="w-3 h-3" />
                <span className="text-muted-foreground">Offline</span>
              </div>
              <div className="flex items-center gap-2">
                <circle cx="6" cy="6" r="4" fill="#f59e0b" className="w-3 h-3" />
                <span className="text-muted-foreground">Draining</span>
              </div>
              <div className="flex items-center gap-2">
                <circle cx="6" cy="6" r="4" fill="#3b82f6" className="w-3 h-3" />
                <span className="text-muted-foreground">Maintenance</span>
              </div>
              {nodes.some(n => n.capabilities.gpu_enabled) && (
                <div className="flex items-center gap-2 pt-1 border-t border-border">
                  <circle cx="6" cy="6" r="2" fill="#8b5cf6" className="w-3 h-3" />
                  <span className="text-muted-foreground">GPU Enabled</span>
                </div>
              )}
            </div>
          </div>

          {/* Statistics */}
          <div className="absolute bottom-4 left-4 bg-card border border-border rounded-lg p-3 text-xs">
            <h4 className="font-medium text-foreground mb-2">Statistics</h4>
            <div className="space-y-1">
              <div className="flex justify-between gap-4">
                <span className="text-muted-foreground">Total Nodes:</span>
                <span className="text-foreground font-medium">{nodes.length}</span>
              </div>
              <div className="flex justify-between gap-4">
                <span className="text-muted-foreground">Online:</span>
                <span className="text-success font-medium">{nodes.filter(n => n.status === 'online').length}</span>
              </div>
              <div className="flex justify-between gap-4">
                <span className="text-muted-foreground">GPU Nodes:</span>
                <span className="text-info font-medium">{nodes.filter(n => n.capabilities.gpu_enabled).length}</span>
              </div>
              <div className="flex justify-between gap-4">
                <span className="text-muted-foreground">Avg Health:</span>
                <span className="text-foreground font-medium">
                  {(nodes.reduce((sum, n) => sum + n.health_score, 0) / nodes.length).toFixed(1)}%
                </span>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export default NetworkTopology