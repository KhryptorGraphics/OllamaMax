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
  Tabs,
  CircularProgress
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Add as AddIcon,
  Settings as SettingsIcon,
  Warning as WarningIcon,
  Error as ErrorIcon,
  CheckCircle as CheckCircleIcon,
  Hub as HubIcon,
  Security as SecurityIcon,
  Timeline as TimelineIcon,
  Visibility as VisibilityIcon,
  Traffic as TrafficIcon,
  CloudSync as CloudSyncIcon,
  NetworkCheck as NetworkCheckIcon,
  Shield as ShieldIcon,
  Speed as SpeedIcon,
  Memory as MemoryIcon,
  Storage as StorageIcon,
  PlayCircleFilled as PlayCircleFilledIcon,
  Router as RouterIcon
} from '@mui/icons-material';
import { useServiceMesh } from '../../hooks/service-mesh/useServiceMesh';
import { ServiceMeshConfiguration, ServiceMeshService, ServiceMeshWorkload, CanaryDeployment } from '../../types/service-mesh';
import { formatNumber, formatDate, formatDuration } from '../../utils/formatting';
import ServiceTopology from './ServiceTopology';
import ServiceList from './ServiceList';
import WorkloadManager from './WorkloadManager';
import SecurityPolicyManager from './SecurityPolicyManager';
import TrafficManagement from './TrafficManagement';
import CanaryDeployments from './CanaryDeployments';
import ObservabilityDashboard from './ObservabilityDashboard';
import ConfigurationManager from './ConfigurationManager';

const ServiceMeshDashboard: React.FC = () => {
  const {
    configuration,
    services,
    workloads,
    topology,
    canaryDeployments,
    securityPolicies,
    loading,
    error,
    connected,
    runHealthCheck,
    refreshServices,
    refreshWorkloads,
    refreshTopology
  } = useServiceMesh();

  const [activeTab, setActiveTab] = useState<'overview' | 'topology' | 'services' | 'workloads' | 'security' | 'traffic' | 'canary' | 'observability' | 'config'>('overview');

  // Calculate mesh health and statistics
  const meshHealth = useMemo(() => {
    if (!configuration) return { score: 0, level: 'unknown' };

    let score = 0;
    let factors = 0;

    // Control plane health
    if (configuration.status === 'active') {
      score += 30;
    } else if (configuration.status === 'degraded') {
      score += 15;
    }
    factors += 30;

    // Service health
    if (services.length > 0) {
      const healthyServices = services.filter(s => s.status === 'healthy').length;
      score += (healthyServices / services.length) * 25;
    }
    factors += 25;

    // Workload health
    if (workloads.length > 0) {
      const healthyWorkloads = workloads.filter(w => w.health.overallHealth === 'healthy').length;
      score += (healthyWorkloads / workloads.length) * 25;
    }
    factors += 25;

    // Security policies
    if (securityPolicies.length > 0) {
      const activesPolicies = securityPolicies.filter(p => p.enabled).length;
      score += (activesPolicies / securityPolicies.length) * 20;
    }
    factors += 20;

    const finalScore = Math.round((score / factors) * 100);

    let level: 'excellent' | 'good' | 'warning' | 'critical';
    if (finalScore >= 90) level = 'excellent';
    else if (finalScore >= 75) level = 'good';
    else if (finalScore >= 50) level = 'warning';
    else level = 'critical';

    return { score: finalScore, level };
  }, [configuration, services, workloads, securityPolicies]);

  // Calculate statistics
  const stats = useMemo(() => {
    const healthyServices = services.filter(s => s.status === 'healthy').length;
    const healthyWorkloads = workloads.filter(w => w.health.overallHealth === 'healthy').length;
    const injectedWorkloads = workloads.filter(w => w.sidecar.injected).length;
    const activePolicies = securityPolicies.filter(p => p.enabled).length;
    const activeCanaries = canaryDeployments.filter(c => c.status === 'running').length;

    const avgRequestRate = services.length > 0 
      ? services.reduce((sum, service) => sum + service.metrics.requestRate, 0) / services.length
      : 0;

    const avgErrorRate = services.length > 0 
      ? services.reduce((sum, service) => sum + service.metrics.errorRate, 0) / services.length
      : 0;

    const avgLatency = services.length > 0 
      ? services.reduce((sum, service) => sum + service.metrics.p95Latency, 0) / services.length
      : 0;

    const totalWorkloadCpu = workloads.reduce((sum, workload) => sum + workload.metrics.cpuUsage, 0);
    const totalWorkloadMemory = workloads.reduce((sum, workload) => sum + workload.metrics.memoryUsage, 0);

    return {
      totalServices: services.length,
      healthyServices,
      totalWorkloads: workloads.length,
      healthyWorkloads,
      injectedWorkloads,
      totalPolicies: securityPolicies.length,
      activePolicies,
      totalCanaries: canaryDeployments.length,
      activeCanaries,
      avgRequestRate: Math.round(avgRequestRate),
      avgErrorRate: Math.round(avgErrorRate * 100) / 100,
      avgLatency: Math.round(avgLatency),
      totalWorkloadCpu: Math.round(totalWorkloadCpu),
      totalWorkloadMemory: Math.round(totalWorkloadMemory)
    };
  }, [services, workloads, securityPolicies, canaryDeployments]);

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
      case 'good': return <HubIcon />;
      case 'warning': return <WarningIcon />;
      case 'critical': return <ErrorIcon />;
      default: return <NetworkCheckIcon />;
    }
  };

  const handleRefresh = async () => {
    try {
      await Promise.all([
        refreshServices(),
        refreshWorkloads(),
        refreshTopology()
      ]);
    } catch (error) {
      console.error('Failed to refresh data:', error);
    }
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="h4" gutterBottom>
          Service Mesh Management
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
          Service Mesh Management
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
          <Tooltip title="Mesh Type">
            <Chip
              icon={<HubIcon />}
              label={configuration?.type || 'Unknown'}
              color="primary"
              variant="outlined"
              sx={{ textTransform: 'capitalize' }}
            />
          </Tooltip>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={handleRefresh}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<PlayCircleFilledIcon />}
            onClick={runHealthCheck}
          >
            Health Check
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
            label="Topology" 
            value="topology"
            icon={<TimelineIcon fontSize="small" />}
            iconPosition="start"
          />
          <Tab 
            label="Services" 
            value="services"
            icon={<Badge badgeContent={stats.totalServices} color="primary"><CloudSyncIcon fontSize="small" /></Badge>}
            iconPosition="start"
          />
          <Tab 
            label="Workloads" 
            value="workloads"
            icon={<Badge badgeContent={stats.injectedWorkloads} color="success"><MemoryIcon fontSize="small" /></Badge>}
            iconPosition="start"
          />
          <Tab 
            label="Security" 
            value="security"
            icon={<Badge badgeContent={stats.activePolicies} color="error"><SecurityIcon fontSize="small" /></Badge>}
            iconPosition="start"
          />
          <Tab 
            label="Traffic" 
            value="traffic"
            icon={<TrafficIcon fontSize="small" />}
            iconPosition="start"
          />
          <Tab 
            label="Canary" 
            value="canary"
            icon={<Badge badgeContent={stats.activeCanaries} color="warning"><RouterIcon fontSize="small" /></Badge>}
            iconPosition="start"
          />
          <Tab 
            label="Observability" 
            value="observability"
            icon={<VisibilityIcon fontSize="small" />}
            iconPosition="start"
          />
          <Tab 
            label="Configuration" 
            value="config"
            icon={<SettingsIcon fontSize="small" />}
            iconPosition="start"
          />
        </Tabs>
      </Paper>

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <>
          {/* Service Mesh Health Card */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Service Mesh Health
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                <Box sx={{ position: 'relative', display: 'inline-flex' }}>
                  <CircularProgress
                    variant="determinate"
                    value={meshHealth.score}
                    size={80}
                    color={getHealthLevelColor(meshHealth.level) as any}
                  />
                  <Box
                    sx={{
                      top: 0,
                      left: 0,
                      bottom: 0,
                      right: 0,
                      position: 'absolute',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <Typography variant="h6" component="div">
                      {meshHealth.score}
                    </Typography>
                  </Box>
                </Box>
                <Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Box sx={{ color: getHealthLevelColor(meshHealth.level) + '.main' }}>
                      {getHealthLevelIcon(meshHealth.level)}
                    </Box>
                    <Typography variant="h5" sx={{ textTransform: 'capitalize' }}>
                      {meshHealth.level}
                    </Typography>
                  </Box>
                  <Typography variant="body2" color="text.secondary">
                    Overall mesh health based on control plane, services, workloads, and security policies
                  </Typography>
                </Box>
              </Box>
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
                      <Typography variant="h4">{stats.healthyServices}/{stats.totalServices}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Healthy Services
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
                      <Typography variant="h4">{stats.injectedWorkloads}/{stats.totalWorkloads}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Injected Workloads
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
                    <SecurityIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.activePolicies}/{stats.totalPolicies}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Active Policies
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
                    <RouterIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.activeCanaries}/{stats.totalCanaries}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Active Canaries
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
                    Traffic Metrics
                  </Typography>
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          Request Rate
                        </Typography>
                        <Typography variant="h6">
                          {formatNumber(stats.avgRequestRate)}/s
                        </Typography>
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          Error Rate
                        </Typography>
                        <Typography variant="h6" color={stats.avgErrorRate > 5 ? 'error.main' : 'inherit'}>
                          {stats.avgErrorRate}%
                        </Typography>
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          P95 Latency
                        </Typography>
                        <Typography variant="h6">
                          {stats.avgLatency}ms
                        </Typography>
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          Success Rate
                        </Typography>
                        <Typography variant="h6" color="success.main">
                          {Math.round((100 - stats.avgErrorRate) * 100) / 100}%
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
                    Resource Usage
                  </Typography>
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          Total CPU
                        </Typography>
                        <Typography variant="h6">
                          {stats.totalWorkloadCpu}%
                        </Typography>
                        <LinearProgress
                          variant="determinate"
                          value={Math.min(stats.totalWorkloadCpu, 100)}
                          color={stats.totalWorkloadCpu > 80 ? 'error' : stats.totalWorkloadCpu > 60 ? 'warning' : 'primary'}
                          sx={{ mt: 1, height: 6, borderRadius: 3 }}
                        />
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          Total Memory
                        </Typography>
                        <Typography variant="h6">
                          {stats.totalWorkloadMemory}%
                        </Typography>
                        <LinearProgress
                          variant="determinate"
                          value={Math.min(stats.totalWorkloadMemory, 100)}
                          color={stats.totalWorkloadMemory > 80 ? 'error' : stats.totalWorkloadMemory > 60 ? 'warning' : 'primary'}
                          sx={{ mt: 1, height: 6, borderRadius: 3 }}
                        />
                      </Box>
                    </Grid>
                    <Grid item xs={12}>
                      <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                        Mesh Version: {configuration?.version || 'Unknown'}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Namespace: {configuration?.namespace || 'istio-system'}
                      </Typography>
                    </Grid>
                  </Grid>
                </CardContent>
              </Card>
            </Grid>
          </Grid>

          {/* Control Plane Status */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Control Plane Status
              </Typography>
              <Grid container spacing={3}>
                <Grid item xs={12} md={6}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                    <Box sx={{ color: configuration?.status === 'active' ? 'success.main' : 'error.main' }}>
                      {configuration?.status === 'active' ? <CheckCircleIcon /> : <ErrorIcon />}
                    </Box>
                    <Box>
                      <Typography variant="body1">
                        Control Plane: {configuration?.status || 'Unknown'}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {configuration?.controlPlane.replicas || 0} replicas running
                      </Typography>
                    </Box>
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <Box sx={{ color: configuration?.dataPlane.autoInjection.enabled ? 'success.main' : 'warning.main' }}>
                      {configuration?.dataPlane.autoInjection.enabled ? <CheckCircleIcon /> : <WarningIcon />}
                    </Box>
                    <Box>
                      <Typography variant="body1">
                        Auto Injection: {configuration?.dataPlane.autoInjection.enabled ? 'Enabled' : 'Disabled'}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        Proxy: {configuration?.dataPlane.proxyType || 'envoy'}
                      </Typography>
                    </Box>
                  </Box>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                    <Box sx={{ color: configuration?.security.mtls.mode === 'strict' ? 'success.main' : 'warning.main' }}>
                      <ShieldIcon />
                    </Box>
                    <Box>
                      <Typography variant="body1">
                        mTLS Mode: {configuration?.security.mtls.mode || 'Unknown'}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        Certificate Provider: {configuration?.security.mtls.certificateProvider || 'Unknown'}
                      </Typography>
                    </Box>
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <Box sx={{ color: configuration?.observability.metrics.enabled ? 'success.main' : 'warning.main' }}>
                      {configuration?.observability.metrics.enabled ? <CheckCircleIcon /> : <WarningIcon />}
                    </Box>
                    <Box>
                      <Typography variant="body1">
                        Observability: {configuration?.observability.metrics.enabled ? 'Enabled' : 'Disabled'}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        Tracing: {configuration?.observability.tracing.enabled ? 'Enabled' : 'Disabled'}
                      </Typography>
                    </Box>
                  </Box>
                </Grid>
              </Grid>
            </CardContent>
          </Card>

          {/* Recent Activity */}
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Recent Canary Deployments
                  </Typography>
                  {canaryDeployments.slice(0, 5).map((canary, index) => (
                    <Box key={canary.id}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, py: 1 }}>
                        <Box sx={{ color: canary.status === 'running' ? 'success.main' : canary.status === 'failed' ? 'error.main' : 'warning.main' }}>
                          {canary.status === 'running' ? <PlayCircleFilledIcon fontSize="small" /> : 
                           canary.status === 'failed' ? <ErrorIcon fontSize="small" /> : 
                           <WarningIcon fontSize="small" />}
                        </Box>
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="body2">
                            {canary.name}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {canary.service} • Step {canary.currentStep}/{canary.steps.length}
                          </Typography>
                        </Box>
                        <Chip
                          label={canary.status}
                          color={canary.status === 'running' ? 'success' : canary.status === 'failed' ? 'error' : 'warning'}
                          size="small"
                          sx={{ textTransform: 'capitalize' }}
                        />
                      </Box>
                      {index < 4 && <Divider />}
                    </Box>
                  ))}
                  {canaryDeployments.length === 0 && (
                    <Typography variant="body2" color="text.secondary">
                      No canary deployments
                    </Typography>
                  )}
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Service Health Overview
                  </Typography>
                  {services.slice(0, 5).map((service, index) => (
                    <Box key={service.id}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, py: 1 }}>
                        <Box sx={{ color: service.status === 'healthy' ? 'success.main' : service.status === 'degraded' ? 'warning.main' : 'error.main' }}>
                          {service.status === 'healthy' ? <CheckCircleIcon fontSize="small" /> : 
                           service.status === 'degraded' ? <WarningIcon fontSize="small" /> : 
                           <ErrorIcon fontSize="small" />}
                        </Box>
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="body2">
                            {service.name}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {service.namespace} • {service.endpoints.length} endpoints
                          </Typography>
                        </Box>
                        <Box sx={{ textAlign: 'right' }}>
                          <Typography variant="body2">
                            {Math.round(service.metrics.requestRate)}/s
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {service.metrics.errorRate.toFixed(2)}% error
                          </Typography>
                        </Box>
                      </Box>
                      {index < 4 && <Divider />}
                    </Box>
                  ))}
                  {services.length === 0 && (
                    <Typography variant="body2" color="text.secondary">
                      No services found
                    </Typography>
                  )}
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        </>
      )}

      {/* Other Tabs */}
      {activeTab === 'topology' && <ServiceTopology />}
      {activeTab === 'services' && <ServiceList />}
      {activeTab === 'workloads' && <WorkloadManager />}
      {activeTab === 'security' && <SecurityPolicyManager />}
      {activeTab === 'traffic' && <TrafficManagement />}
      {activeTab === 'canary' && <CanaryDeployments />}
      {activeTab === 'observability' && <ObservabilityDashboard />}
      {activeTab === 'config' && <ConfigurationManager />}
    </Box>
  );
};

export default ServiceMeshDashboard;