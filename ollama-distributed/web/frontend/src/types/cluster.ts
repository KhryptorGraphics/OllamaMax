// Cluster and system state types

export interface ClusterState {
  status: ClusterStatus
  leader: string | null
  nodes: number
  healthyNodes: number
  loading: boolean
  error: string | null
  lastUpdated: number | null
}

export type ClusterStatus = 'healthy' | 'degraded' | 'unhealthy' | 'maintenance' | 'unknown'

export interface NodeState {
  id: string
  status: NodeStatus
  health: NodeHealth
  lastSeen: string
  error?: string
}

export type NodeStatus = 'online' | 'offline' | 'draining' | 'maintenance'

export interface NodeHealth {
  cpu: number
  memory: number
  disk: number
  network: number
  load: number[]
  errors: number
  warnings: number
}