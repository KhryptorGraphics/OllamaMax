import React, { useState, useEffect } from 'react'
import { 
  Play, 
  Square, 
  Save, 
  Download, 
  Upload, 
  Copy,
  Settings,
  Book,
  Database,
  Activity,
  Zap,
  Clock,
  BarChart3,
  Code,
  Search,
  Filter,
  RefreshCw,
  Share,
  History,
  Eye,
  EyeOff
} from 'lucide-react'
import { Button, Card, Input, Alert, Badge, Switch } from '@/design-system'
import { useAPI } from '@/hooks/useAPI'
import { formatBytes, formatNumber, formatDuration } from '@/utils/format'

interface GraphQLSchema {
  types: GraphQLType[]
  queries: GraphQLField[]
  mutations: GraphQLField[]
  subscriptions: GraphQLField[]
}

interface GraphQLType {
  name: string
  kind: 'OBJECT' | 'SCALAR' | 'ENUM' | 'INPUT_OBJECT' | 'INTERFACE' | 'UNION'
  description?: string
  fields?: GraphQLField[]
  enumValues?: GraphQLEnumValue[]
  inputFields?: GraphQLInputValue[]
}

interface GraphQLField {
  name: string
  description?: string
  type: GraphQLTypeRef
  args?: GraphQLInputValue[]
  isDeprecated: boolean
  deprecationReason?: string
}

interface GraphQLInputValue {
  name: string
  description?: string
  type: GraphQLTypeRef
  defaultValue?: any
}

interface GraphQLEnumValue {
  name: string
  description?: string
  isDeprecated: boolean
  deprecationReason?: string
}

interface GraphQLTypeRef {
  kind: string
  name?: string
  ofType?: GraphQLTypeRef
}

interface QueryExecution {
  id: string
  query: string
  variables?: Record<string, any>
  operationName?: string
  timestamp: string
  duration: number
  status: 'success' | 'error'
  data?: any
  errors?: any[]
  complexity: number
  depth: number
}

interface QueryMetrics {
  totalQueries: number
  successRate: number
  avgDuration: number
  complexityDistribution: { [key: string]: number }
  popularFields: { field: string; count: number }[]
  errors: { type: string; count: number }[]
}

interface SavedQuery {
  id: string
  name: string
  description: string
  query: string
  variables?: Record<string, any>
  tags: string[]
  createdAt: string
  updatedAt: string
  isPublic: boolean
}

export const GraphQLPlayground: React.FC = () => {
  const [schema, setSchema] = useState<GraphQLSchema | null>(null)
  const [query, setQuery] = useState(`# Welcome to GraphQL Playground
# Type your query here

query {
  models {
    id
    name
    status
    size
    parameters
  }
}`)
  const [variables, setVariables] = useState('{}')
  const [headers, setHeaders] = useState('{}')
  const [operationName, setOperationName] = useState('')
  const [result, setResult] = useState<any>(null)
  const [isExecuting, setIsExecuting] = useState(false)
  const [executions, setExecutions] = useState<QueryExecution[]>([])
  const [metrics, setMetrics] = useState<QueryMetrics | null>(null)
  const [savedQueries, setSavedQueries] = useState<SavedQuery[]>([])
  const [activeTab, setActiveTab] = useState<'playground' | 'schema' | 'history' | 'saved'>('playground')
  const [showVariables, setShowVariables] = useState(false)
  const [showHeaders, setShowHeaders] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedType, setSelectedType] = useState<GraphQLType | null>(null)
  const [autoComplete, setAutoComplete] = useState(true)
  const [prettifyQuery, setPrettifyQuery] = useState(true)

  const { data: schemaData } = useAPI('/api/graphql/schema')
  const { data: metricsData } = useAPI('/api/graphql/metrics')
  const { data: savedQueriesData } = useAPI('/api/graphql/saved-queries')

  useEffect(() => {
    if (schemaData) {
      setSchema(schemaData.schema)
    }
    if (metricsData) {
      setMetrics(metricsData)
    }
    if (savedQueriesData) {
      setSavedQueries(savedQueriesData.queries || [])
    }
  }, [schemaData, metricsData, savedQueriesData])

  const executeQuery = async () => {
    setIsExecuting(true)
    const startTime = performance.now()
    
    try {
      let parsedVariables = {}
      let parsedHeaders = {}
      
      try {
        parsedVariables = variables ? JSON.parse(variables) : {}
        parsedHeaders = headers ? JSON.parse(headers) : {}
      } catch (error) {
        throw new Error('Invalid JSON in variables or headers')
      }

      const response = await fetch('/api/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...parsedHeaders
        },
        body: JSON.stringify({
          query,
          variables: parsedVariables,
          operationName: operationName || undefined
        })
      })

      const result = await response.json()
      const duration = performance.now() - startTime

      setResult(result)

      // Add to execution history
      const execution: QueryExecution = {
        id: Date.now().toString(),
        query,
        variables: parsedVariables,
        operationName,
        timestamp: new Date().toISOString(),
        duration,
        status: result.errors ? 'error' : 'success',
        data: result.data,
        errors: result.errors,
        complexity: calculateComplexity(query),
        depth: calculateDepth(query)
      }

      setExecutions(prev => [execution, ...prev.slice(0, 49)]) // Keep last 50 executions
    } catch (error) {
      const duration = performance.now() - startTime
      setResult({
        errors: [{ message: error.message }]
      })

      const execution: QueryExecution = {
        id: Date.now().toString(),
        query,
        variables: {},
        timestamp: new Date().toISOString(),
        duration,
        status: 'error',
        errors: [{ message: error.message }],
        complexity: 0,
        depth: 0
      }

      setExecutions(prev => [execution, ...prev.slice(0, 49)])
    } finally {
      setIsExecuting(false)
    }
  }

  const calculateComplexity = (query: string): number => {
    // Simple complexity calculation based on query structure
    const fieldCount = (query.match(/\w+\s*{/g) || []).length
    const nestedCount = (query.match(/{/g) || []).length
    return fieldCount + nestedCount
  }

  const calculateDepth = (query: string): number => {
    // Calculate nesting depth
    let depth = 0
    let maxDepth = 0
    for (const char of query) {
      if (char === '{') {
        depth++
        maxDepth = Math.max(maxDepth, depth)
      } else if (char === '}') {
        depth--
      }
    }
    return maxDepth
  }

  const saveQuery = async () => {
    const name = prompt('Enter query name:')
    if (!name) return

    try {
      const savedQuery: Omit<SavedQuery, 'id' | 'createdAt' | 'updatedAt'> = {
        name,
        description: '',
        query,
        variables: variables ? JSON.parse(variables) : undefined,
        tags: [],
        isPublic: false
      }

      const response = await fetch('/api/graphql/saved-queries', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(savedQuery)
      })

      if (response.ok) {
        const newQuery = await response.json()
        setSavedQueries(prev => [newQuery, ...prev])
      }
    } catch (error) {
      console.error('Failed to save query:', error)
    }
  }

  const loadQuery = (savedQuery: SavedQuery) => {
    setQuery(savedQuery.query)
    setVariables(savedQuery.variables ? JSON.stringify(savedQuery.variables, null, 2) : '{}')
    setActiveTab('playground')
  }

  const prettifyQueryText = () => {
    // Simple query formatting
    try {
      const formatted = query
        .replace(/\s+/g, ' ')
        .replace(/{\s*/g, '{\n  ')
        .replace(/\s*}/g, '\n}')
        .replace(/,\s*/g, ',\n  ')
      setQuery(formatted)
    } catch (error) {
      console.error('Failed to prettify query:', error)
    }
  }

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(query)
    } catch (error) {
      console.error('Failed to copy to clipboard:', error)
    }
  }

  const getTypeDisplayName = (typeRef: GraphQLTypeRef): string => {
    if (typeRef.kind === 'NON_NULL') {
      return `${getTypeDisplayName(typeRef.ofType!)}!`
    }
    if (typeRef.kind === 'LIST') {
      return `[${getTypeDisplayName(typeRef.ofType!)}]`
    }
    return typeRef.name || 'Unknown'
  }

  const filteredTypes = schema?.types.filter(type =>
    type.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    type.description?.toLowerCase().includes(searchQuery.toLowerCase())
  ) || []

  return (
    <div className="space-y-6">
      {/* Metrics Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Activity className="w-8 h-8 text-blue-500" />
              <div>
                <p className="text-sm text-muted-foreground">Total Queries</p>
                <p className="text-2xl font-bold">{formatNumber(metrics?.totalQueries || 0)}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Zap className="w-8 h-8 text-green-500" />
              <div>
                <p className="text-sm text-muted-foreground">Success Rate</p>
                <p className="text-2xl font-bold">{(metrics?.successRate || 0).toFixed(1)}%</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Clock className="w-8 h-8 text-orange-500" />
              <div>
                <p className="text-sm text-muted-foreground">Avg Duration</p>
                <p className="text-2xl font-bold">{(metrics?.avgDuration || 0).toFixed(0)}ms</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <BarChart3 className="w-8 h-8 text-purple-500" />
              <div>
                <p className="text-sm text-muted-foreground">Schema Types</p>
                <p className="text-2xl font-bold">{schema?.types.length || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>
      </div>

      {/* Main Interface */}
      <Card>
        <Card.Header>
          <Card.Title>GraphQL Playground</Card.Title>
          <div className="flex space-x-2">
            <Button
              onClick={executeQuery}
              disabled={isExecuting}
              loading={isExecuting}
              loadingText="Executing..."
            >
              <Play className="w-4 h-4 mr-2" />
              Execute
            </Button>
            <Button variant="outline" onClick={prettifyQueryText}>
              <Code className="w-4 h-4 mr-2" />
              Prettify
            </Button>
            <Button variant="outline" onClick={copyToClipboard}>
              <Copy className="w-4 h-4 mr-2" />
              Copy
            </Button>
            <Button variant="outline" onClick={saveQuery}>
              <Save className="w-4 h-4 mr-2" />
              Save
            </Button>
          </div>
        </Card.Header>

        <Card.Content>
          {/* Tab Navigation */}
          <div className="border-b mb-6">
            <nav className="flex space-x-4">
              {['playground', 'schema', 'history', 'saved'].map((tab) => (
                <button
                  key={tab}
                  className={`py-2 px-1 border-b-2 text-sm font-medium capitalize ${
                    activeTab === tab
                      ? 'border-primary text-primary'
                      : 'border-transparent text-muted-foreground hover:text-foreground'
                  }`}
                  onClick={() => setActiveTab(tab as any)}
                >
                  {tab}
                </button>
              ))}
            </nav>
          </div>

          {/* Playground Tab */}
          {activeTab === 'playground' && (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Query Editor */}
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <h4 className="text-sm font-medium">Query</h4>
                  <div className="flex items-center space-x-2">
                    <label className="flex items-center space-x-2 text-sm">
                      <Switch
                        checked={autoComplete}
                        onCheckedChange={setAutoComplete}
                      />
                      <span>Auto-complete</span>
                    </label>
                  </div>
                </div>
                
                <textarea
                  className="w-full h-64 p-3 border rounded-md font-mono text-sm"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Enter your GraphQL query here..."
                  style={{ fontFamily: 'Monaco, Menlo, monospace' }}
                />

                {/* Operation Name */}
                <Input
                  label="Operation Name (optional)"
                  value={operationName}
                  onChange={(e) => setOperationName(e.target.value)}
                  placeholder="MyQuery"
                />

                {/* Variables Toggle */}
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowVariables(!showVariables)}
                  >
                    {showVariables ? <EyeOff className="w-4 h-4 mr-2" /> : <Eye className="w-4 h-4 mr-2" />}
                    Variables
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowHeaders(!showHeaders)}
                  >
                    {showHeaders ? <EyeOff className="w-4 h-4 mr-2" /> : <Eye className="w-4 h-4 mr-2" />}
                    Headers
                  </Button>
                </div>

                {/* Variables Editor */}
                {showVariables && (
                  <div>
                    <label className="block text-sm font-medium mb-2">Variables (JSON)</label>
                    <textarea
                      className="w-full h-24 p-3 border rounded-md font-mono text-sm"
                      value={variables}
                      onChange={(e) => setVariables(e.target.value)}
                      placeholder='{"key": "value"}'
                    />
                  </div>
                )}

                {/* Headers Editor */}
                {showHeaders && (
                  <div>
                    <label className="block text-sm font-medium mb-2">Headers (JSON)</label>
                    <textarea
                      className="w-full h-24 p-3 border rounded-md font-mono text-sm"
                      value={headers}
                      onChange={(e) => setHeaders(e.target.value)}
                      placeholder='{"Authorization": "Bearer token"}'
                    />
                  </div>
                )}
              </div>

              {/* Results */}
              <div className="space-y-4">
                <h4 className="text-sm font-medium">Result</h4>
                
                <div className="border rounded-md h-96 overflow-auto">
                  {result ? (
                    <pre className="p-3 text-sm font-mono whitespace-pre-wrap">
                      {JSON.stringify(result, null, 2)}
                    </pre>
                  ) : (
                    <div className="flex items-center justify-center h-full text-muted-foreground">
                      <div className="text-center">
                        <Database className="w-12 h-12 mx-auto mb-2 opacity-50" />
                        <p>Execute a query to see results</p>
                      </div>
                    </div>
                  )}
                </div>

                {/* Query Info */}
                {executions.length > 0 && (
                  <div className="text-sm text-muted-foreground">
                    <div>Duration: {executions[0].duration.toFixed(2)}ms</div>
                    <div>Complexity: {executions[0].complexity}</div>
                    <div>Depth: {executions[0].depth}</div>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Schema Tab */}
          {activeTab === 'schema' && (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Schema Explorer */}
              <div className="space-y-4">
                <div className="flex space-x-2">
                  <Input
                    placeholder="Search types..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    leftIcon={<Search className="w-4 h-4" />}
                  />
                  <Button variant="outline" size="sm">
                    <Filter className="w-4 h-4" />
                  </Button>
                </div>

                <div className="space-y-2 max-h-96 overflow-y-auto">
                  {filteredTypes.map((type) => (
                    <div
                      key={type.name}
                      className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                        selectedType?.name === type.name ? 'border-primary bg-primary/5' : 'hover:bg-gray-50'
                      }`}
                      onClick={() => setSelectedType(type)}
                    >
                      <div className="flex items-center space-x-2">
                        <Badge variant="outline">{type.kind}</Badge>
                        <span className="font-medium">{type.name}</span>
                      </div>
                      {type.description && (
                        <p className="text-sm text-muted-foreground mt-1">{type.description}</p>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              {/* Type Details */}
              <div className="space-y-4">
                {selectedType ? (
                  <div>
                    <div className="flex items-center space-x-2 mb-4">
                      <Badge variant="outline">{selectedType.kind}</Badge>
                      <h3 className="text-lg font-semibold">{selectedType.name}</h3>
                    </div>

                    {selectedType.description && (
                      <p className="text-sm text-muted-foreground mb-4">{selectedType.description}</p>
                    )}

                    {/* Fields */}
                    {selectedType.fields && (
                      <div>
                        <h4 className="text-sm font-medium mb-2">Fields</h4>
                        <div className="space-y-2">
                          {selectedType.fields.map((field) => (
                            <div key={field.name} className="p-2 bg-gray-50 rounded">
                              <div className="flex items-center space-x-2">
                                <span className="font-medium">{field.name}</span>
                                <span className="text-sm text-muted-foreground">
                                  : {getTypeDisplayName(field.type)}
                                </span>
                                {field.isDeprecated && (
                                  <Badge variant="warning" size="sm">deprecated</Badge>
                                )}
                              </div>
                              {field.description && (
                                <p className="text-xs text-muted-foreground mt-1">{field.description}</p>
                              )}
                              {field.args && field.args.length > 0 && (
                                <div className="mt-2">
                                  <span className="text-xs font-medium">Arguments:</span>
                                  <div className="ml-2">
                                    {field.args.map((arg) => (
                                      <div key={arg.name} className="text-xs">
                                        {arg.name}: {getTypeDisplayName(arg.type)}
                                        {arg.defaultValue !== undefined && (
                                          <span className="text-muted-foreground"> = {JSON.stringify(arg.defaultValue)}</span>
                                        )}
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Enum Values */}
                    {selectedType.enumValues && (
                      <div>
                        <h4 className="text-sm font-medium mb-2">Values</h4>
                        <div className="space-y-1">
                          {selectedType.enumValues.map((value) => (
                            <div key={value.name} className="p-2 bg-gray-50 rounded">
                              <div className="flex items-center space-x-2">
                                <span className="font-medium">{value.name}</span>
                                {value.isDeprecated && (
                                  <Badge variant="warning" size="sm">deprecated</Badge>
                                )}
                              </div>
                              {value.description && (
                                <p className="text-xs text-muted-foreground mt-1">{value.description}</p>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="text-center text-muted-foreground py-12">
                    <Book className="w-12 h-12 mx-auto mb-2 opacity-50" />
                    <p>Select a type to view details</p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* History Tab */}
          {activeTab === 'history' && (
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <h4 className="text-sm font-medium">Query History ({executions.length})</h4>
                <Button variant="outline" size="sm" onClick={() => setExecutions([])}>
                  Clear History
                </Button>
              </div>

              <div className="space-y-3 max-h-96 overflow-y-auto">
                {executions.map((execution) => (
                  <div key={execution.id} className="p-3 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <Badge variant={execution.status === 'success' ? 'success' : 'destructive'}>
                            {execution.status}
                          </Badge>
                          <span className="text-sm text-muted-foreground">
                            {new Date(execution.timestamp).toLocaleString()}
                          </span>
                          <span className="text-sm text-muted-foreground">
                            {execution.duration.toFixed(2)}ms
                          </span>
                        </div>
                        
                        <pre className="text-xs font-mono bg-gray-50 p-2 rounded whitespace-pre-wrap overflow-hidden max-h-20">
                          {execution.query}
                        </pre>
                      </div>
                      
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          setQuery(execution.query)
                          if (execution.variables && Object.keys(execution.variables).length > 0) {
                            setVariables(JSON.stringify(execution.variables, null, 2))
                          }
                          if (execution.operationName) {
                            setOperationName(execution.operationName)
                          }
                          setActiveTab('playground')
                        }}
                      >
                        Load
                      </Button>
                    </div>
                  </div>
                ))}
                
                {executions.length === 0 && (
                  <div className="text-center text-muted-foreground py-12">
                    <History className="w-12 h-12 mx-auto mb-2 opacity-50" />
                    <p>No query history yet</p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Saved Queries Tab */}
          {activeTab === 'saved' && (
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <h4 className="text-sm font-medium">Saved Queries ({savedQueries.length})</h4>
                <div className="flex space-x-2">
                  <Button variant="outline" size="sm">
                    <Upload className="w-4 h-4 mr-2" />
                    Import
                  </Button>
                  <Button variant="outline" size="sm">
                    <Download className="w-4 h-4 mr-2" />
                    Export
                  </Button>
                </div>
              </div>

              <div className="space-y-3 max-h-96 overflow-y-auto">
                {savedQueries.map((savedQuery) => (
                  <div key={savedQuery.id} className="p-3 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <h4 className="text-sm font-medium">{savedQuery.name}</h4>
                          {savedQuery.isPublic && (
                            <Badge variant="default" size="sm">public</Badge>
                          )}
                          {savedQuery.tags.map((tag) => (
                            <Badge key={tag} variant="outline" size="sm">{tag}</Badge>
                          ))}
                        </div>
                        
                        {savedQuery.description && (
                          <p className="text-xs text-muted-foreground mb-2">{savedQuery.description}</p>
                        )}
                        
                        <div className="text-xs text-muted-foreground">
                          Created: {new Date(savedQuery.createdAt).toLocaleDateString()}
                          {savedQuery.updatedAt !== savedQuery.createdAt && (
                            <span> • Updated: {new Date(savedQuery.updatedAt).toLocaleDateString()}</span>
                          )}
                        </div>
                      </div>
                      
                      <div className="flex space-x-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => loadQuery(savedQuery)}
                        >
                          Load
                        </Button>
                        <Button size="sm" variant="outline">
                          <Share className="w-3 h-3" />
                        </Button>
                      </div>
                    </div>
                  </div>
                ))}
                
                {savedQueries.length === 0 && (
                  <div className="text-center text-muted-foreground py-12">
                    <Save className="w-12 h-12 mx-auto mb-2 opacity-50" />
                    <p>No saved queries yet</p>
                  </div>
                )}
              </div>
            </div>
          )}
        </Card.Content>
      </Card>

      {/* Popular Fields */}
      {metrics?.popularFields && metrics.popularFields.length > 0 && (
        <Card>
          <Card.Header>
            <Card.Title>Popular Fields</Card.Title>
          </Card.Header>
          <Card.Content>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {metrics.popularFields.slice(0, 6).map((field, index) => (
                <div key={field.field} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                  <span className="font-medium">{field.field}</span>
                  <Badge variant="outline">{formatNumber(field.count)}</Badge>
                </div>
              ))}
            </div>
          </Card.Content>
        </Card>
      )}
    </div>
  )
}

export default GraphQLPlayground