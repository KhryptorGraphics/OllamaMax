import React, { useState } from 'react';
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
  Divider
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Add as AddIcon,
  Settings as SettingsIcon,
  Warning as WarningIcon,
  Error as ErrorIcon,
  CheckCircle as CheckCircleIcon,
  CloudSync as CloudSyncIcon,
  NetworkCheck as NetworkCheckIcon,
  Security as SecurityIcon,
  Speed as SpeedIcon,
  Storage as StorageIcon,
  Memory as MemoryIcon
} from '@mui/icons-material';
import { useFederation } from '../../hooks/federation/useFederation';
import { FederationCluster, ClusterHealth } from '../../types/federation';
import { formatBytes, formatDuration, formatNumber } from '../../utils/formatting';
import ClusterList from './ClusterList';
import PolicyManager from './PolicyManager';
import ReplicationStatus from './ReplicationStatus';
import ServiceDiscovery from './ServiceDiscovery';
import FederationEvents from './FederationEvents';
import ClusterDialog from './ClusterDialog';

const FederationDashboard: React.FC = () => {
  const {
    clusters,
    policies,
    events,
    replication,
    discovery,
    loading,
    error,
    connected,
    refreshDiscovery
  } = useFederation();

  const [activeTab, setActiveTab] = useState<'overview' | 'clusters' | 'policies' | 'replication' | 'discovery' | 'events'>('overview');
  const [clusterDialogOpen, setClusterDialogOpen] = useState(false);

  // Calculate overall federation health
  const federationHealth = React.useMemo(() => {
    if (clusters.length === 0) return { status: 'unknown', score: 0 };

    const healthScores = clusters.map(cluster => {
      switch (cluster.health.overall) {
        case 'healthy': return 100;
        case 'warning': return 70;
        case 'critical': return 30;
        default: return 0;
      }
    });

    const averageScore = healthScores.reduce((sum, score) => sum + score, 0) / healthScores.length;
    
    let status: 'healthy' | 'warning' | 'critical';
    if (averageScore >= 80) status = 'healthy';
    else if (averageScore >= 50) status = 'warning';
    else status = 'critical';

    return { status, score: averageScore };
  }, [clusters]);

  // Calculate federation statistics
  const stats = React.useMemo(() => {
    const totalNodes = clusters.reduce((sum, cluster) => sum + cluster.nodes, 0);
    const totalModels = clusters.reduce((sum, cluster) => sum + cluster.activeModels, 0);
    const onlineClusters = clusters.filter(cluster => cluster.status === 'online').length;
    const avgResponseTime = clusters.length > 0 
      ? clusters.reduce((sum, cluster) => sum + cluster.metrics.responseTime, 0) / clusters.length 
      : 0;
    const totalThroughput = clusters.reduce((sum, cluster) => sum + cluster.metrics.throughput, 0);
    const totalRequests = clusters.reduce((sum, cluster) => sum + cluster.metrics.requestsPerSecond, 0);

    return {
      totalClusters: clusters.length,
      onlineClusters,
      totalNodes,
      totalModels,
      avgResponseTime,
      totalThroughput,
      totalRequests
    };
  }, [clusters]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'online':
        return 'success';
      case 'warning':
      case 'degraded':
        return 'warning';
      case 'critical':
      case 'offline':
        return 'error';
      default:
        return 'default';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'online':
        return <CheckCircleIcon />;
      case 'warning':
      case 'degraded':
        return <WarningIcon />;
      case 'critical':
      case 'offline':
        return <ErrorIcon />;
      default:
        return <NetworkCheckIcon />;
    }
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="h4" gutterBottom>
          Federation Management
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
          Federation Management
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
            onClick={refreshDiscovery}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setClusterDialogOpen(true)}
          >
            Add Cluster
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
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Box sx={{ display: 'flex', gap: 2, p: 2 }}>
            {[
              { key: 'overview', label: 'Overview' },
              { key: 'clusters', label: 'Clusters' },
              { key: 'policies', label: 'Policies' },
              { key: 'replication', label: 'Replication' },
              { key: 'discovery', label: 'Discovery' },
              { key: 'events', label: 'Events' }
            ].map(tab => (
              <Button
                key={tab.key}
                variant={activeTab === tab.key ? 'contained' : 'text'}
                onClick={() => setActiveTab(tab.key as any)}
              >
                {tab.label}
                {tab.key === 'events' && events.length > 0 && (
                  <Badge
                    badgeContent={events.filter(e => e.severity === 'error').length}
                    color="error"
                    sx={{ ml: 1 }}
                  />
                )}
              </Button>
            ))}
          </Box>
        </Box>
      </Paper>

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <>
          {/* Federation Health Card */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Federation Health
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                <Box sx={{ color: getStatusColor(federationHealth.status) + '.main' }}>
                  {getStatusIcon(federationHealth.status)}
                </Box>
                <Typography variant="h5" sx={{ textTransform: 'capitalize' }}>
                  {federationHealth.status}
                </Typography>
                <Chip
                  label={`${Math.round(federationHealth.score)}% Health Score`}
                  color={getStatusColor(federationHealth.status) as any}
                  variant="outlined"
                />
              </Box>
              <LinearProgress
                variant="determinate"
                value={federationHealth.score}
                color={getStatusColor(federationHealth.status) as any}
                sx={{ height: 8, borderRadius: 4 }}
              />
            </CardContent>
          </Card>

          {/* Statistics Grid */}
          <Grid container spacing={3} sx={{ mb: 3 }}>
            <Grid item xs={12} sm={6} md={3}>
              <Card>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <CloudSyncIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.onlineClusters}/{stats.totalClusters}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Online Clusters
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
                    <NetworkCheckIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{formatNumber(stats.totalNodes)}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Total Nodes
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
                    <MemoryIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{formatNumber(stats.totalModels)}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Active Models
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
                      <Typography variant="h4">{Math.round(stats.avgResponseTime)}ms</Typography>
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
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Performance Metrics
              </Typography>
              <Grid container spacing={3}>
                <Grid item xs={12} md={4}>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Requests/Second
                    </Typography>
                    <Typography variant="h6">
                      {formatNumber(stats.totalRequests)}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} md={4}>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Throughput
                    </Typography>
                    <Typography variant="h6">
                      {formatBytes(stats.totalThroughput)}/s
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} md={4}>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Active Policies
                    </Typography>
                    <Typography variant="h6">
                      {policies.filter(p => p.enabled).length}/{policies.length}
                    </Typography>
                  </Box>
                </Grid>
              </Grid>
            </CardContent>
          </Card>

          {/* Recent Events */}
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Recent Events
              </Typography>
              {events.slice(0, 5).map((event, index) => (
                <Box key={event.id}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, py: 1 }}>
                    <Box sx={{ color: getStatusColor(event.severity) + '.main' }}>
                      {getStatusIcon(event.severity)}
                    </Box>
                    <Box sx={{ flex: 1 }}>
                      <Typography variant="body2">
                        {event.message}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {event.cluster} • {new Date(event.timestamp).toLocaleString()}
                      </Typography>
                    </Box>
                  </Box>
                  {index < 4 && <Divider />}
                </Box>
              ))}
              {events.length === 0 && (
                <Typography variant="body2" color="text.secondary">
                  No recent events
                </Typography>
              )}
            </CardContent>
          </Card>
        </>
      )}

      {/* Other Tabs */}
      {activeTab === 'clusters' && <ClusterList />}
      {activeTab === 'policies' && <PolicyManager />}
      {activeTab === 'replication' && <ReplicationStatus />}
      {activeTab === 'discovery' && <ServiceDiscovery />}
      {activeTab === 'events' && <FederationEvents />}

      {/* Add Cluster Dialog */}
      <ClusterDialog
        open={clusterDialogOpen}
        onClose={() => setClusterDialogOpen(false)}
      />
    </Box>
  );
};

export default FederationDashboard;