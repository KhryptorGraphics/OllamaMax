import React, { useState, useMemo } from 'react';
import {
  Grid,
  Paper,
  Typography,
  Box,
  Chip,
  IconButton,
  Button,
  Card,
  CardContent,
  LinearProgress,
  Alert,
  Tooltip,
  Badge,
  Divider,
  Tab,
  Tabs
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Add as AddIcon,
  Settings as SettingsIcon,
  Warning as WarningIcon,
  Error as ErrorIcon,
  CheckCircle as CheckCircleIcon,
  GraphicEq as GraphicEqIcon,
  Schema as SchemaIcon,
  Cloud as CloudIcon,
  QueryStats as QueryStatsIcon,
  Subscriptions as SubscriptionsIcon,
  Compare as CompareIcon,
  Code as CodeIcon,
  PlayArrow as PlayArrowIcon,
  Speed as SpeedIcon,
  Storage as StorageIcon,
  NetworkCheck as NetworkCheckIcon
} from '@mui/icons-material';
import { useGraphQL } from '../../hooks/graphql/useGraphQL';
import { GraphQLSchema, GraphQLEndpoint, GraphQLQuery, QueryExecution } from '../../types/graphql';
import { formatNumber, formatDate, formatDuration } from '../../utils/formatting';
import SchemaManager from './SchemaManager';
import EndpointManager from './EndpointManager';
import QueryExplorer from './QueryExplorer';
import GraphQLPlayground from './GraphQLPlayground';
import SubscriptionManager from './SubscriptionManager';
import SchemaComparison from './SchemaComparison';
import QueryMetrics from './QueryMetrics';
import EndpointDialog from './EndpointDialog';
import SchemaDialog from './SchemaDialog';

const GraphQLDashboard: React.FC = () => {
  const {
    schemas,
    endpoints,
    queries,
    executions,
    subscriptions,
    loading,
    error,
    connected,
    testEndpoint
  } = useGraphQL();

  const [activeTab, setActiveTab] = useState<'overview' | 'playground' | 'schemas' | 'endpoints' | 'queries' | 'subscriptions' | 'comparisons' | 'metrics'>('overview');
  const [endpointDialogOpen, setEndpointDialogOpen] = useState(false);
  const [schemaDialogOpen, setSchemaDialogOpen] = useState(false);

  // Calculate API health and statistics
  const apiHealth = useMemo(() => {
    if (endpoints.length === 0) return { score: 0, level: 'unknown' };

    const healthyEndpoints = endpoints.filter(e => e.status === 'healthy').length;
    const score = (healthyEndpoints / endpoints.length) * 100;

    let level: 'excellent' | 'good' | 'warning' | 'critical';
    if (score >= 95) level = 'excellent';
    else if (score >= 80) level = 'good';
    else if (score >= 60) level = 'warning';
    else level = 'critical';

    return { score: Math.round(score), level };
  }, [endpoints]);

  // Calculate statistics
  const stats = useMemo(() => {
    const healthyEndpoints = endpoints.filter(e => e.status === 'healthy').length;
    const publishedSchemas = schemas.filter(s => s.status === 'published').length;
    const activeSubscriptions = subscriptions.filter(s => s.status === 'connected').length;
    const recentExecutions = executions.filter(e => {
      const oneHourAgo = new Date();
      oneHourAgo.setHours(oneHourAgo.getHours() - 1);
      return new Date(e.startTime) > oneHourAgo;
    }).length;

    const avgResponseTime = executions.length > 0 
      ? executions.reduce((sum, exec) => sum + (exec.duration || 0), 0) / executions.length
      : 0;

    const successRate = executions.length > 0
      ? (executions.filter(e => e.status === 'success').length / executions.length) * 100
      : 0;

    const totalRequests = endpoints.reduce((sum, endpoint) => sum + endpoint.metrics.totalRequests, 0);
    const avgCacheHitRate = endpoints.length > 0
      ? endpoints.reduce((sum, endpoint) => sum + endpoint.metrics.cacheHitRate, 0) / endpoints.length
      : 0;

    return {
      totalEndpoints: endpoints.length,
      healthyEndpoints,
      totalSchemas: schemas.length,
      publishedSchemas,
      totalQueries: queries.length,
      activeSubscriptions,
      recentExecutions,
      avgResponseTime: Math.round(avgResponseTime),
      successRate: Math.round(successRate),
      totalRequests,
      avgCacheHitRate: Math.round(avgCacheHitRate)
    };
  }, [endpoints, schemas, queries, subscriptions, executions]);

  const getHealthLevelColor = (level: string) => {
    switch (level) {
      case 'excellent': return 'success';
      case 'good': return 'info';
      case 'warning': return 'warning';
      case 'critical': return 'error';
      default: return 'default';
    }
  };

  const getHealthLevelIcon = (level: string) => {
    switch (level) {
      case 'excellent': return <CheckCircleIcon />;
      case 'good': return <GraphicEqIcon />;
      case 'warning': return <WarningIcon />;
      case 'critical': return <ErrorIcon />;
      default: return <NetworkCheckIcon />;
    }
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="h4" gutterBottom>
          GraphQL API Management
        </Typography>
        <LinearProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      {/* Header */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">
          GraphQL API Management
        </Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Tooltip title="Connection Status">
            <Chip
              icon={<NetworkCheckIcon />}
              label={connected ? 'Connected' : 'Disconnected'}
              color={connected ? 'success' : 'error'}
              variant="outlined"
            />
          </Tooltip>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={() => window.location.reload()}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setEndpointDialogOpen(true)}
          >
            Add Endpoint
          </Button>
        </Box>
      </Box>

      {/* Error Alert */}
      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Navigation Tabs */}
      <Paper sx={{ mb: 3 }}>
        <Tabs
          value={activeTab}
          onChange={(_, newValue) => setActiveTab(newValue)}
          variant="scrollable"
          scrollButtons="auto"
        >
          <Tab label="Overview" value="overview" />
          <Tab 
            label="Playground" 
            value="playground" 
            icon={<PlayArrowIcon fontSize="small" />}
            iconPosition="start"
          />
          <Tab 
            label="Schemas" 
            value="schemas"
            icon={<Badge badgeContent={schemas.length} color="primary"><SchemaIcon fontSize="small" /></Badge>}
            iconPosition="start"
          />
          <Tab 
            label="Endpoints" 
            value="endpoints"
            icon={<Badge badgeContent={endpoints.length} color="primary"><CloudIcon fontSize="small" /></Badge>}
            iconPosition="start"
          />
          <Tab 
            label="Queries" 
            value="queries"
            icon={<Badge badgeContent={queries.length} color="primary"><CodeIcon fontSize="small" /></Badge>}
            iconPosition="start"
          />
          <Tab 
            label="Subscriptions" 
            value="subscriptions"
            icon={<Badge badgeContent={stats.activeSubscriptions} color="success"><SubscriptionsIcon fontSize="small" /></Badge>}
            iconPosition="start"
          />
          <Tab 
            label="Comparisons" 
            value="comparisons"
            icon={<CompareIcon fontSize="small" />}
            iconPosition="start"
          />
          <Tab 
            label="Metrics" 
            value="metrics"
            icon={<QueryStatsIcon fontSize="small" />}
            iconPosition="start"
          />
        </Tabs>
      </Paper>

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <>
          {/* API Health Card */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                API Health Status
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                <Box sx={{ color: getHealthLevelColor(apiHealth.level) + '.main' }}>
                  {getHealthLevelIcon(apiHealth.level)}
                </Box>
                <Typography variant="h5" sx={{ textTransform: 'capitalize' }}>
                  {apiHealth.level}
                </Typography>
                <Chip
                  label={`${apiHealth.score}% Healthy`}
                  color={getHealthLevelColor(apiHealth.level) as any}
                  variant="outlined"
                />
              </Box>
              <LinearProgress
                variant="determinate"
                value={apiHealth.score}
                color={getHealthLevelColor(apiHealth.level) as any}
                sx={{ height: 8, borderRadius: 4 }}
              />
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                {stats.healthyEndpoints} of {stats.totalEndpoints} endpoints are healthy
              </Typography>
            </CardContent>
          </Card>

          {/* Statistics Grid */}
          <Grid container spacing={3} sx={{ mb: 3 }}>
            <Grid item xs={12} sm={6} md={3}>
              <Card>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <CloudIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.healthyEndpoints}/{stats.totalEndpoints}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Healthy Endpoints
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} sm={6} md={3}>
              <Card>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <SchemaIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.publishedSchemas}/{stats.totalSchemas}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Published Schemas
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} sm={6} md={3}>
              <Card>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <SubscriptionsIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.activeSubscriptions}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Active Subscriptions
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} sm={6} md={3}>
              <Card>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <SpeedIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.avgResponseTime}ms</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Avg Response Time
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          </Grid>

          {/* Performance Metrics */}
          <Grid container spacing={3} sx={{ mb: 3 }}>
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Request Statistics
                  </Typography>
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          Total Requests
                        </Typography>
                        <Typography variant="h6">
                          {formatNumber(stats.totalRequests)}
                        </Typography>
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          Recent (1h)
                        </Typography>
                        <Typography variant="h6">
                          {formatNumber(stats.recentExecutions)}
                        </Typography>
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          Success Rate
                        </Typography>
                        <Typography variant="h6">
                          {stats.successRate}%
                        </Typography>
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          Cache Hit Rate
                        </Typography>
                        <Typography variant="h6">
                          {stats.avgCacheHitRate}%
                        </Typography>
                      </Box>
                    </Grid>
                  </Grid>
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Recent Executions
                  </Typography>
                  {executions.slice(0, 5).map((execution, index) => (
                    <Box key={execution.id}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, py: 1 }}>
                        <Box sx={{ color: execution.status === 'success' ? 'success.main' : execution.status === 'error' ? 'error.main' : 'warning.main' }}>
                          {execution.status === 'success' ? <CheckCircleIcon fontSize="small" /> : 
                           execution.status === 'error' ? <ErrorIcon fontSize="small" /> : 
                           <WarningIcon fontSize="small" />}
                        </Box>
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="body2">
                            {queries.find(q => q.id === execution.queryId)?.name || 'Unknown Query'}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {execution.duration ? `${execution.duration}ms` : 'Running'} • {formatDate(execution.startTime)}
                          </Typography>
                        </Box>
                        <Chip
                          label={execution.status}
                          color={execution.status === 'success' ? 'success' : execution.status === 'error' ? 'error' : 'warning'}
                          size="small"
                        />
                      </Box>
                      {index < 4 && <Divider />}
                    </Box>
                  ))}
                  {executions.length === 0 && (
                    <Typography variant="body2" color="text.secondary">
                      No recent executions
                    </Typography>
                  )}
                </CardContent>
              </Card>
            </Grid>
          </Grid>

          {/* Endpoint Status Overview */}
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Endpoint Status Overview
              </Typography>
              {endpoints.slice(0, 5).map((endpoint, index) => (
                <Box key={endpoint.id}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, py: 1 }}>
                    <Box sx={{ color: endpoint.status === 'healthy' ? 'success.main' : endpoint.status === 'degraded' ? 'warning.main' : 'error.main' }}>
                      {endpoint.status === 'healthy' ? <CheckCircleIcon fontSize="small" /> : 
                       endpoint.status === 'degraded' ? <WarningIcon fontSize="small" /> : 
                       <ErrorIcon fontSize="small" />}
                    </Box>
                    <Box sx={{ flex: 1 }}>
                      <Typography variant="body2">
                        {endpoint.name}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {endpoint.url} • {endpoint.environment}
                      </Typography>
                    </Box>
                    <Box sx={{ textAlign: 'right' }}>
                      <Typography variant="body2">
                        {Math.round(endpoint.metrics.averageResponseTime)}ms
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {Math.round(endpoint.metrics.requestsPerSecond)}/s
                      </Typography>
                    </Box>
                    <Chip
                      label={endpoint.status}
                      color={endpoint.status === 'healthy' ? 'success' : endpoint.status === 'degraded' ? 'warning' : 'error'}
                      size="small"
                      sx={{ textTransform: 'capitalize' }}
                    />
                  </Box>
                  {index < 4 && <Divider />}
                </Box>
              ))}
              {endpoints.length === 0 && (
                <Typography variant="body2" color="text.secondary">
                  No endpoints configured
                </Typography>
              )}
            </CardContent>
          </Card>
        </>
      )}

      {/* Other Tabs */}
      {activeTab === 'playground' && <GraphQLPlayground />}
      {activeTab === 'schemas' && <SchemaManager />}
      {activeTab === 'endpoints' && <EndpointManager />}
      {activeTab === 'queries' && <QueryExplorer />}
      {activeTab === 'subscriptions' && <SubscriptionManager />}
      {activeTab === 'comparisons' && <SchemaComparison />}
      {activeTab === 'metrics' && <QueryMetrics />}

      {/* Dialogs */}
      <EndpointDialog
        open={endpointDialogOpen}
        onClose={() => setEndpointDialogOpen(false)}
      />
      <SchemaDialog
        open={schemaDialogOpen}
        onClose={() => setSchemaDialogOpen(false)}
      />
    </Box>
  );
};

export default GraphQLDashboard;