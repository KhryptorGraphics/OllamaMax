// Model management types

export interface ModelState {
  models: ModelInfo[]
  loading: boolean
  error: string | null
  syncStatus: ModelSyncMap
}

export interface ModelInfo {
  name: string
  tag: string
  size: number
  format: string
  family: string
  digest: string
  createdAt: string
  modifiedAt: string
  syncStatus: ModelSyncStatus
}

export interface ModelSyncStatus {
  status: 'synchronized' | 'syncing' | 'failed' | 'pending'
  progress: number
  nodesTotal: number
  nodesSynced: number
  error?: string
}

export type ModelSyncMap = Record<string, ModelSyncStatus>