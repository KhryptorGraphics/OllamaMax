/**
 * Log Viewer Component
 * Displays searchable and filterable log entries with real-time updates
 */

import React, { useState, useRef, useEffect, useMemo } from 'react'
import {
  Search,
  Filter,
  Download,
  RefreshCw,
  AlertCircle,
  Info,
  AlertTriangle,
  X,
  Clock,
  Tag,
  ChevronDown,
  ChevronRight,
  Copy,
  ExternalLink
} from 'lucide-react'
import { format } from 'date-fns'
import { LogEntry, LogLevel } from '../../types/monitoring'

interface LogViewerProps {
  logs: LogEntry[]
  onRefresh?: () => void
  onExport?: (logs: LogEntry[]) => void
  maxHeight?: number
  autoScroll?: boolean
  realTime?: boolean
  className?: string
}

interface LogFilters {
  level?: LogLevel
  source?: string
  category?: string
  search?: string
  startTime?: string
  endTime?: string
  tags?: string[]
}

const LOG_LEVELS: LogLevel[] = ['debug', 'info', 'warn', 'error', 'fatal']

const LOG_LEVEL_COLORS = {
  debug: {
    bg: 'bg-gray-50',
    text: 'text-gray-700',
    icon: 'text-gray-500',
    border: 'border-gray-200'
  },
  info: {
    bg: 'bg-blue-50',
    text: 'text-blue-800',
    icon: 'text-blue-500',
    border: 'border-blue-200'
  },
  warn: {
    bg: 'bg-yellow-50',
    text: 'text-yellow-800',
    icon: 'text-yellow-500',
    border: 'border-yellow-200'
  },
  error: {
    bg: 'bg-red-50',
    text: 'text-red-800',
    icon: 'text-red-500',
    border: 'border-red-200'
  },
  fatal: {
    bg: 'bg-red-100',
    text: 'text-red-900',
    icon: 'text-red-600',
    border: 'border-red-300'
  }
}

const getLevelIcon = (level: LogLevel) => {
  const iconClass = `w-4 h-4 ${LOG_LEVEL_COLORS[level].icon}`
  
  switch (level) {
    case 'fatal':
    case 'error':
      return <AlertCircle className={iconClass} />
    case 'warn':
      return <AlertTriangle className={iconClass} />
    default:
      return <Info className={iconClass} />
  }
}

interface LogEntryProps {
  log: LogEntry
  searchTerm?: string
  expanded?: boolean
  onToggle?: () => void
}

const LogEntryComponent: React.FC<LogEntryProps> = ({
  log,
  searchTerm,
  expanded = false,
  onToggle
}) => {
  const colors = LOG_LEVEL_COLORS[log.level]
  
  const highlightText = (text: string, search?: string) => {
    if (!search) return text
    
    const regex = new RegExp(`(${search})`, 'gi')
    const parts = text.split(regex)
    
    return parts.map((part, index) => 
      regex.test(part) ? (
        <mark key={index} className="bg-yellow-200 px-1 rounded">
          {part}
        </mark>
      ) : (
        part
      )
    )
  }
  
  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      // Could add a toast notification here
    })
  }
  
  const formatTimestamp = (timestamp: string) => {
    return format(new Date(timestamp), 'MMM dd, HH:mm:ss.SSS')
  }
  
  return (
    <div className={`
      border-l-4 p-3 mb-2 rounded-r-md transition-all duration-200
      ${colors.bg} ${colors.border} hover:shadow-sm
    `}>
      <div className="flex items-start justify-between">
        <div className="flex items-start space-x-3 flex-1 min-w-0">
          {getLevelIcon(log.level)}
          
          <div className="flex-1 min-w-0">
            <div className="flex items-center space-x-2 mb-1">
              <span className={`text-xs font-medium uppercase ${colors.text}`}>
                {log.level}
              </span>
              <span className="text-xs text-gray-500">
                {formatTimestamp(log.timestamp)}
              </span>
              <span className="text-xs text-gray-500">
                {log.source}
              </span>
              {log.category && (
                <span className="text-xs text-gray-500">
                  / {log.category}
                </span>
              )}
              {log.correlation_id && (
                <span className="text-xs font-mono text-gray-400">
                  {log.correlation_id.slice(0, 8)}...
                </span>
              )}
            </div>
            
            <div className={`text-sm ${colors.text} break-words`}>
              {highlightText(log.message, searchTerm)}
            </div>
            
            {log.tags.length > 0 && (
              <div className="flex flex-wrap gap-1 mt-2">
                {log.tags.map((tag, index) => (
                  <span
                    key={index}
                    className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-gray-100 text-gray-800"
                  >
                    <Tag className="w-3 h-3 mr-1" />
                    {tag}
                  </span>
                ))}
              </div>
            )}
            
            {expanded && log.metadata && (
              <div className="mt-3 p-2 bg-gray-100 rounded-md">
                <div className="text-xs font-medium text-gray-700 mb-1">Metadata:</div>
                <pre className="text-xs text-gray-600 whitespace-pre-wrap overflow-x-auto">
                  {JSON.stringify(log.metadata, null, 2)}
                </pre>
              </div>
            )}
          </div>
        </div>
        
        <div className="flex items-center space-x-1 ml-2">
          <button
            onClick={() => copyToClipboard(log.message)}
            className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
            title="Copy message"
          >
            <Copy className="w-3 h-3" />
          </button>
          
          {log.metadata && onToggle && (
            <button
              onClick={onToggle}
              className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
              title={expanded ? "Hide metadata" : "Show metadata"}
            >
              {expanded ? (
                <ChevronDown className="w-3 h-3" />
              ) : (
                <ChevronRight className="w-3 h-3" />
              )}
            </button>
          )}
          
          {log.correlation_id && (
            <button
              onClick={() => {
                // Could navigate to correlation view
                console.log('View correlation:', log.correlation_id)
              }}
              className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
              title="View correlation"
            >
              <ExternalLink className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

const FilterPanel: React.FC<{
  filters: LogFilters
  onFiltersChange: (filters: LogFilters) => void
  availableSources: string[]
  availableCategories: string[]
  availableTags: string[]
}> = ({
  filters,
  onFiltersChange,
  availableSources,
  availableCategories,
  availableTags
}) => {
  const [showAdvanced, setShowAdvanced] = useState(false)
  
  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4 mb-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-medium text-gray-900">Log Filters</h3>
        <button
          onClick={() => setShowAdvanced(!showAdvanced)}
          className="text-sm text-blue-600 hover:text-blue-800"
        >
          {showAdvanced ? 'Simple' : 'Advanced'} Filters
        </button>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-4">
        {/* Search */}
        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Search
          </label>
          <div className="relative">
            <Search className="absolute left-3 top-2.5 w-4 h-4 text-gray-400" />
            <input
              type="text"
              value={filters.search || ''}
              onChange={(e) => onFiltersChange({ ...filters, search: e.target.value })}
              placeholder="Search logs..."
              className="w-full pl-9 pr-3 py-2 border border-gray-300 rounded-md text-sm"
            />
          </div>
        </div>
        
        {/* Level */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Level
          </label>
          <select
            value={filters.level || ''}
            onChange={(e) => onFiltersChange({
              ...filters,
              level: e.target.value as LogLevel || undefined
            })}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            <option value="">All Levels</option>
            {LOG_LEVELS.map(level => (
              <option key={level} value={level} className="capitalize">
                {level}
              </option>
            ))}
          </select>
        </div>
        
        {/* Source */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Source
          </label>
          <select
            value={filters.source || ''}
            onChange={(e) => onFiltersChange({ ...filters, source: e.target.value || undefined })}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            <option value="">All Sources</option>
            {availableSources.map(source => (
              <option key={source} value={source}>
                {source}
              </option>
            ))}
          </select>
        </div>
        
        {/* Category */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Category
          </label>
          <select
            value={filters.category || ''}
            onChange={(e) => onFiltersChange({ ...filters, category: e.target.value || undefined })}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            <option value="">All Categories</option>
            {availableCategories.map(category => (
              <option key={category} value={category}>
                {category}
              </option>
            ))}
          </select>
        </div>
        
        {/* Clear Filters */}
        <div className="flex items-end">
          <button
            onClick={() => onFiltersChange({})}
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            Clear All
          </button>
        </div>
      </div>
      
      {showAdvanced && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4 pt-4 border-t border-gray-200">
          {/* Time Range */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Start Time
            </label>
            <input
              type="datetime-local"
              value={filters.startTime || ''}
              onChange={(e) => onFiltersChange({ ...filters, startTime: e.target.value || undefined })}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              End Time
            </label>
            <input
              type="datetime-local"
              value={filters.endTime || ''}
              onChange={(e) => onFiltersChange({ ...filters, endTime: e.target.value || undefined })}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
            />
          </div>
          
          {/* Tags */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Tags
            </label>
            <select
              multiple
              value={filters.tags || []}
              onChange={(e) => {
                const tags = Array.from(e.target.selectedOptions, option => option.value)
                onFiltersChange({ ...filters, tags: tags.length > 0 ? tags : undefined })
              }}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
              size={3}
            >
              {availableTags.map(tag => (
                <option key={tag} value={tag}>
                  {tag}
                </option>
              ))}
            </select>
          </div>
        </div>
      )}
    </div>
  )
}

export const LogViewer: React.FC<LogViewerProps> = ({
  logs,
  onRefresh,
  onExport,
  maxHeight = 600,
  autoScroll = true,
  realTime = false,
  className = ''
}) => {
  const [filters, setFilters] = useState<LogFilters>({})
  const [expandedLogs, setExpandedLogs] = useState<Set<string>>(new Set())
  const containerRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  
  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [logs, autoScroll])
  
  // Extract unique values for filter options
  const filterOptions = useMemo(() => {
    const sources = new Set<string>()
    const categories = new Set<string>()
    const tags = new Set<string>()
    
    logs.forEach(log => {
      sources.add(log.source)
      if (log.category) categories.add(log.category)
      log.tags.forEach(tag => tags.add(tag))
    })
    
    return {
      sources: Array.from(sources).sort(),
      categories: Array.from(categories).sort(),
      tags: Array.from(tags).sort()
    }
  }, [logs])
  
  // Filter logs based on current filters
  const filteredLogs = useMemo(() => {
    return logs.filter(log => {
      if (filters.level && log.level !== filters.level) return false
      if (filters.source && log.source !== filters.source) return false
      if (filters.category && log.category !== filters.category) return false
      if (filters.startTime && new Date(log.timestamp) < new Date(filters.startTime)) return false
      if (filters.endTime && new Date(log.timestamp) > new Date(filters.endTime)) return false
      if (filters.tags && !filters.tags.some(tag => log.tags.includes(tag))) return false
      if (filters.search) {
        const search = filters.search.toLowerCase()
        return (
          log.message.toLowerCase().includes(search) ||
          log.source.toLowerCase().includes(search) ||
          log.category?.toLowerCase().includes(search) ||
          log.tags.some(tag => tag.toLowerCase().includes(search))
        )
      }
      return true
    })
  }, [logs, filters])
  
  const toggleLogExpansion = (logId: string) => {
    const newExpanded = new Set(expandedLogs)
    if (newExpanded.has(logId)) {
      newExpanded.delete(logId)
    } else {
      newExpanded.add(logId)
    }
    setExpandedLogs(newExpanded)
  }
  
  const handleExport = () => {
    if (onExport) {
      onExport(filteredLogs)
    }
  }
  
  return (
    <div className={`space-y-4 ${className}`}>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-gray-900">
          System Logs ({filteredLogs.length})
        </h2>
        
        <div className="flex items-center space-x-2">
          {realTime && (
            <div className="flex items-center space-x-2 text-sm text-green-600">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              Live
            </div>
          )}
          
          {onExport && (
            <button
              onClick={handleExport}
              className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
            >
              <Download className="w-4 h-4 mr-2" />
              Export
            </button>
          )}
          
          {onRefresh && (
            <button
              onClick={onRefresh}
              className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
            >
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </button>
          )}
        </div>
      </div>
      
      <FilterPanel
        filters={filters}
        onFiltersChange={setFilters}
        availableSources={filterOptions.sources}
        availableCategories={filterOptions.categories}
        availableTags={filterOptions.tags}
      />
      
      <div
        ref={containerRef}
        className="bg-gray-50 border border-gray-200 rounded-lg p-4 overflow-y-auto"
        style={{ maxHeight: `${maxHeight}px` }}
      >
        {filteredLogs.length === 0 ? (
          <div className="text-center py-8">
            <AlertCircle className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <div className="text-gray-500">
              {logs.length === 0 ? 'No logs found' : 'No logs match your filters'}
            </div>
          </div>
        ) : (
          <div className="space-y-1">
            {filteredLogs.map(log => (
              <LogEntryComponent
                key={log.id}
                log={log}
                searchTerm={filters.search}
                expanded={expandedLogs.has(log.id)}
                onToggle={() => toggleLogExpansion(log.id)}
              />
            ))}
            <div ref={bottomRef} />
          </div>
        )}
      </div>
    </div>
  )
}