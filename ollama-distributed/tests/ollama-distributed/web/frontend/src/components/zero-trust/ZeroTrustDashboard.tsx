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
  CircularProgress
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Add as AddIcon,
  Settings as SettingsIcon,
  Warning as WarningIcon,
  Error as ErrorIcon,
  CheckCircle as CheckCircleIcon,
  Security as SecurityIcon,
  Shield as ShieldIcon,
  VpnKey as VpnKeyIcon,
  Certificate as CertificateIcon,
  NetworkCheck as NetworkCheckIcon,
  Person as PersonIcon,
  Policy as PolicyIcon,
  TrendingUp as TrendingUpIcon,
  TrendingDown as TrendingDownIcon,
  Remove as RemoveIcon
} from '@mui/icons-material';
import { useZeroTrust } from '../../hooks/zero-trust/useZeroTrust';
import { Identity, NetworkPolicy, TrustScore, SecurityEvent } from '../../types/zero-trust';
import { formatNumber, formatDate } from '../../utils/formatting';
import IdentityManager from './IdentityManager';
import PolicyManager from './PolicyManager';
import CertificateManager from './CertificateManager';
import ChannelManager from './ChannelManager';
import TrustScoreView from './TrustScoreView';
import SecurityEvents from './SecurityEvents';
import IdentityDialog from './IdentityDialog';
import PolicyDialog from './PolicyDialog';

const ZeroTrustDashboard: React.FC = () => {
  const {
    identities,
    policies,
    certificates,
    channels,
    trustScores,
    events,
    loading,
    error,
    connected,
    calculateTrustScore
  } = useZeroTrust();

  const [activeTab, setActiveTab] = useState<'overview' | 'identities' | 'policies' | 'certificates' | 'channels' | 'trust' | 'events'>('overview');
  const [identityDialogOpen, setIdentityDialogOpen] = useState(false);
  const [policyDialogOpen, setPolicyDialogOpen] = useState(false);

  // Calculate security posture
  const securityPosture = useMemo(() => {
    if (identities.length === 0) return { score: 0, level: 'unknown' };

    const activeIdentities = identities.filter(i => i.status === 'active').length;
    const activePolicies = policies.filter(p => p.enabled).length;
    const validCertificates = certificates.filter(c => c.status === 'valid').length;
    const activeChannels = channels.filter(c => c.status === 'active').length;

    // Calculate weighted score
    const identityScore = (activeIdentities / Math.max(identities.length, 1)) * 25;
    const policyScore = (activePolicies / Math.max(policies.length, 1)) * 25;
    const certificateScore = (validCertificates / Math.max(certificates.length, 1)) * 25;
    const channelScore = (activeChannels / Math.max(channels.length, 1)) * 25;

    const totalScore = identityScore + policyScore + certificateScore + channelScore;

    let level: 'critical' | 'warning' | 'good' | 'excellent';
    if (totalScore >= 90) level = 'excellent';
    else if (totalScore >= 70) level = 'good';
    else if (totalScore >= 50) level = 'warning';
    else level = 'critical';

    return { score: Math.round(totalScore), level };
  }, [identities, policies, certificates, channels]);

  // Calculate statistics
  const stats = useMemo(() => {
    const activeIdentities = identities.filter(i => i.status === 'active').length;
    const activePolicies = policies.filter(p => p.enabled).length;
    const validCertificates = certificates.filter(c => c.status === 'valid').length;
    const expiringCertificates = certificates.filter(c => {
      const expiryDate = new Date(c.notAfter);
      const thirtyDaysFromNow = new Date();
      thirtyDaysFromNow.setDate(thirtyDaysFromNow.getDate() + 30);
      return expiryDate <= thirtyDaysFromNow && c.status === 'valid';
    }).length;
    
    const activeChannels = channels.filter(c => c.status === 'active').length;
    const avgTrustScore = trustScores.length > 0 
      ? trustScores.reduce((sum, score) => sum + score.score, 0) / trustScores.length
      : 0;
    
    const criticalEvents = events.filter(e => e.severity === 'critical' && !e.resolved).length;
    const unresolvedEvents = events.filter(e => !e.resolved).length;

    return {
      activeIdentities,
      totalIdentities: identities.length,
      activePolicies,
      totalPolicies: policies.length,
      validCertificates,
      totalCertificates: certificates.length,
      expiringCertificates,
      activeChannels,
      totalChannels: channels.length,
      avgTrustScore: Math.round(avgTrustScore),
      criticalEvents,
      unresolvedEvents
    };
  }, [identities, policies, certificates, channels, trustScores, events]);

  const getSecurityLevelColor = (level: string) => {
    switch (level) {
      case 'excellent': return 'success';
      case 'good': return 'info';
      case 'warning': return 'warning';
      case 'critical': return 'error';
      default: return 'default';
    }
  };

  const getSecurityLevelIcon = (level: string) => {
    switch (level) {
      case 'excellent': return <CheckCircleIcon />;
      case 'good': return <ShieldIcon />;
      case 'warning': return <WarningIcon />;
      case 'critical': return <ErrorIcon />;
      default: return <SecurityIcon />;
    }
  };

  const getTrustTrendIcon = (trend: string) => {
    switch (trend) {
      case 'increasing': return <TrendingUpIcon color="success" />;
      case 'decreasing': return <TrendingDownIcon color="error" />;
      default: return <RemoveIcon color="disabled" />;
    }
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="h4" gutterBottom>
          Zero Trust Security
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
          Zero Trust Security
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
            onClick={() => setIdentityDialogOpen(true)}
          >
            Add Identity
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
              { key: 'identities', label: 'Identities' },
              { key: 'policies', label: 'Policies' },
              { key: 'certificates', label: 'Certificates' },
              { key: 'channels', label: 'Channels' },
              { key: 'trust', label: 'Trust Scores' },
              { key: 'events', label: 'Events' }
            ].map(tab => (
              <Button
                key={tab.key}
                variant={activeTab === tab.key ? 'contained' : 'text'}
                onClick={() => setActiveTab(tab.key as any)}
              >
                {tab.label}
                {tab.key === 'events' && stats.unresolvedEvents > 0 && (
                  <Badge
                    badgeContent={stats.unresolvedEvents}
                    color="error"
                    sx={{ ml: 1 }}
                  />
                )}
                {tab.key === 'certificates' && stats.expiringCertificates > 0 && (
                  <Badge
                    badgeContent={stats.expiringCertificates}
                    color="warning"
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
          {/* Security Posture Card */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Security Posture
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                <Box sx={{ position: 'relative', display: 'inline-flex' }}>
                  <CircularProgress
                    variant="determinate"
                    value={securityPosture.score}
                    size={80}
                    color={getSecurityLevelColor(securityPosture.level) as any}
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
                      {securityPosture.score}
                    </Typography>
                  </Box>
                </Box>
                <Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Box sx={{ color: getSecurityLevelColor(securityPosture.level) + '.main' }}>
                      {getSecurityLevelIcon(securityPosture.level)}
                    </Box>
                    <Typography variant="h5" sx={{ textTransform: 'capitalize' }}>
                      {securityPosture.level}
                    </Typography>
                  </Box>
                  <Typography variant="body2" color="text.secondary">
                    Overall security posture based on active policies, valid certificates, and trust scores
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
                    <PersonIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.activeIdentities}/{stats.totalIdentities}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Active Identities
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
                    <PolicyIcon color="primary" />
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
                    <CertificateIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.validCertificates}/{stats.totalCertificates}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Valid Certificates
                      </Typography>
                      {stats.expiringCertificates > 0 && (
                        <Chip
                          label={`${stats.expiringCertificates} expiring`}
                          color="warning"
                          size="small"
                          sx={{ mt: 0.5 }}
                        />
                      )}
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} sm={6} md={3}>
              <Card>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <VpnKeyIcon color="primary" />
                    <Box>
                      <Typography variant="h4">{stats.activeChannels}/{stats.totalChannels}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        Secure Channels
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          </Grid>

          {/* Trust Scores and Events */}
          <Grid container spacing={3} sx={{ mb: 3 }}>
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Trust Score Overview
                  </Typography>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                    <Typography variant="h4">
                      {stats.avgTrustScore}/100
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Average Trust Score
                    </Typography>
                  </Box>
                  {trustScores.slice(0, 5).map((score, index) => (
                    <Box key={score.identity}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, py: 1 }}>
                        {getTrustTrendIcon(score.trend)}
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="body2">
                            {identities.find(i => i.id === score.identity)?.name || score.identity}
                          </Typography>
                          <LinearProgress
                            variant="determinate"
                            value={score.score}
                            color={score.score >= 80 ? 'success' : score.score >= 60 ? 'warning' : 'error'}
                            sx={{ height: 4, borderRadius: 2 }}
                          />
                        </Box>
                        <Typography variant="body2" sx={{ minWidth: 40, textAlign: 'right' }}>
                          {score.score}
                        </Typography>
                      </Box>
                      {index < 4 && <Divider />}
                    </Box>
                  ))}
                  {trustScores.length === 0 && (
                    <Typography variant="body2" color="text.secondary">
                      No trust scores available
                    </Typography>
                  )}
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Recent Security Events
                  </Typography>
                  {stats.criticalEvents > 0 && (
                    <Alert severity="error" sx={{ mb: 2 }}>
                      {stats.criticalEvents} critical security events require attention
                    </Alert>
                  )}
                  {events.slice(0, 5).map((event, index) => (
                    <Box key={event.id}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, py: 1 }}>
                        <Box sx={{ color: event.severity === 'critical' ? 'error.main' : event.severity === 'warning' ? 'warning.main' : 'info.main' }}>
                          {event.severity === 'critical' ? <ErrorIcon /> : event.severity === 'warning' ? <WarningIcon /> : <CheckCircleIcon />}
                        </Box>
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="body2">
                            {event.message}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {event.identity} • {formatDate(event.timestamp)}
                          </Typography>
                        </Box>
                        <Chip
                          label={event.severity}
                          color={event.severity === 'critical' ? 'error' : event.severity === 'warning' ? 'warning' : 'info'}
                          size="small"
                        />
                      </Box>
                      {index < 4 && <Divider />}
                    </Box>
                  ))}
                  {events.length === 0 && (
                    <Typography variant="body2" color="text.secondary">
                      No recent security events
                    </Typography>
                  )}
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        </>
      )}

      {/* Other Tabs */}
      {activeTab === 'identities' && <IdentityManager />}
      {activeTab === 'policies' && <PolicyManager />}
      {activeTab === 'certificates' && <CertificateManager />}
      {activeTab === 'channels' && <ChannelManager />}
      {activeTab === 'trust' && <TrustScoreView />}
      {activeTab === 'events' && <SecurityEvents />}

      {/* Dialogs */}
      <IdentityDialog
        open={identityDialogOpen}
        onClose={() => setIdentityDialogOpen(false)}
      />
      <PolicyDialog
        open={policyDialogOpen}
        onClose={() => setPolicyDialogOpen(false)}
      />
    </Box>
  );
};

export default ZeroTrustDashboard;