import React, { useState, useMemo } from 'react';
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  TableSortLabel,
  Chip,
  IconButton,
  Menu,
  MenuItem,
  Box,
  Typography,
  LinearProgress,
  Tooltip,
  Avatar,
  Button,
  TextField,
  InputAdornment,
  Card,
  CardContent,
  Grid
} from '@mui/material';
import {
  MoreVert as MoreVertIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  PlayArrow as PlayArrowIcon,
  Stop as StopIcon,
  Settings as SettingsIcon,
  Warning as WarningIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Search as SearchIcon,
  CloudSync as CloudSyncIcon,
  Memory as MemoryIcon,
  Speed as SpeedIcon,
  Storage as StorageIcon
} from '@mui/icons-material';
import { useFederation } from '../../hooks/federation/useFederation';
import { FederationCluster } from '../../types/federation';
import { formatBytes, formatDuration, formatNumber } from '../../utils/formatting';
import ClusterDetailsDialog from './ClusterDetailsDialog';
import ClusterDialog from './ClusterDialog';

type SortField = keyof FederationCluster | 'health' | 'performance';
type SortDirection = 'asc' | 'desc';

const ClusterList: React.FC = () => {
  const {
    clusters,
    loading,
    removeCluster,
    updateCluster,
    initiateFailover
  } = useFederation();

  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [sortField, setSortField] = useState<SortField>('name');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [searchTerm, setSearchTerm] = useState('');
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedCluster, setSelectedCluster] = useState<FederationCluster | null>(null);
  const [detailsDialogOpen, setDetailsDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);

  // Filter and sort clusters
  const filteredAndSortedClusters = useMemo(() => {
    let filtered = clusters.filter(cluster =>
      cluster.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      cluster.region.toLowerCase().includes(searchTerm.toLowerCase()) ||
      cluster.endpoint.toLowerCase().includes(searchTerm.toLowerCase())
    );

    filtered.sort((a, b) => {
      let aValue: any;
      let bValue: any;

      switch (sortField) {
        case 'health':
          const healthScore = (cluster: FederationCluster) => {
            switch (cluster.health.overall) {
              case 'healthy': return 3;
              case 'warning': return 2;
              case 'critical': return 1;
              default: return 0;
            }
          };
          aValue = healthScore(a);
          bValue = healthScore(b);
          break;
        case 'performance':
          aValue = a.metrics.responseTime;
          bValue = b.metrics.responseTime;
          break;
        default:
          aValue = a[sortField];
          bValue = b[sortField];
      }

      if (typeof aValue === 'string') {
        aValue = aValue.toLowerCase();
        bValue = bValue.toLowerCase();
      }

      if (sortDirection === 'asc') {
        return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
      } else {
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
      }
    });

    return filtered;
  }, [clusters, searchTerm, sortField, sortDirection]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>, cluster: FederationCluster) => {
    setAnchorEl(event.currentTarget);
    setSelectedCluster(cluster);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
    setSelectedCluster(null);
  };

  const handleViewDetails = () => {
    setDetailsDialogOpen(true);
    handleMenuClose();
  };

  const handleEdit = () => {
    setEditDialogOpen(true);
    handleMenuClose();
  };

  const handleDelete = async () => {
    if (selectedCluster && confirm(`Are you sure you want to remove cluster "${selectedCluster.name}"?`)) {
      try {
        await removeCluster(selectedCluster.id);
      } catch (error) {
        console.error('Failed to remove cluster:', error);
      }
    }
    handleMenuClose();
  };

  const handleToggleStatus = async () => {
    if (selectedCluster) {
      const newStatus = selectedCluster.status === 'online' ? 'offline' : 'online';
      try {
        await updateCluster(selectedCluster.id, { status: newStatus });
      } catch (error) {
        console.error('Failed to update cluster status:', error);
      }
    }
    handleMenuClose();
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'success';
      case 'offline': return 'error';
      case 'degraded': return 'warning';
      case 'maintenance': return 'info';
      default: return 'default';
    }
  };

  const getHealthColor = (health: string) => {
    switch (health) {
      case 'healthy': return 'success';
      case 'warning': return 'warning';
      case 'critical': return 'error';
      default: return 'default';
    }
  };

  const getHealthIcon = (health: string) => {
    switch (health) {
      case 'healthy': return <CheckCircleIcon fontSize="small" />;
      case 'warning': return <WarningIcon fontSize="small" />;
      case 'critical': return <ErrorIcon fontSize="small" />;
      default: return <CheckCircleIcon fontSize="small" />;
    }
  };

  const getRegionAvatar = (region: string) => {
    const colors = [
      '#1976d2', '#dc004e', '#9c27b0', '#673ab7',
      '#3f51b5', '#2196f3', '#03a9f4', '#00bcd4',
      '#009688', '#4caf50', '#8bc34a', '#cddc39',
      '#ffeb3b', '#ffc107', '#ff9800', '#ff5722'
    ];
    const colorIndex = region.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) % colors.length;
    
    return (
      <Avatar sx={{ bgcolor: colors[colorIndex], width: 32, height: 32, fontSize: '0.875rem' }}>
        {region.substring(0, 2).toUpperCase()}
      </Avatar>
    );
  };

  // Calculate summary statistics
  const stats = useMemo(() => {
    const online = clusters.filter(c => c.status === 'online').length;
    const healthy = clusters.filter(c => c.health.overall === 'healthy').length;
    const totalNodes = clusters.reduce((sum, c) => sum + c.nodes, 0);
    const totalModels = clusters.reduce((sum, c) => sum + c.activeModels, 0);
    const avgResponseTime = clusters.length > 0 
      ? clusters.reduce((sum, c) => sum + c.metrics.responseTime, 0) / clusters.length 
      : 0;

    return { online, healthy, totalNodes, totalModels, avgResponseTime };
  }, [clusters]);

  return (
    <Box>
      {/* Summary Cards */}
      <Grid container spacing={2} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={2.4}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <CloudSyncIcon color="primary" sx={{ fontSize: 40, mb: 1 }} />
              <Typography variant="h6">{stats.online}/{clusters.length}</Typography>
              <Typography variant="caption" color="text.secondary">
                Online Clusters
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={2.4}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <CheckCircleIcon color="success" sx={{ fontSize: 40, mb: 1 }} />
              <Typography variant="h6">{stats.healthy}</Typography>
              <Typography variant="caption" color="text.secondary">
                Healthy Clusters
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={2.4}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <MemoryIcon color="primary" sx={{ fontSize: 40, mb: 1 }} />
              <Typography variant="h6">{formatNumber(stats.totalNodes)}</Typography>
              <Typography variant="caption" color="text.secondary">
                Total Nodes
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={2.4}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <StorageIcon color="primary" sx={{ fontSize: 40, mb: 1 }} />
              <Typography variant="h6">{formatNumber(stats.totalModels)}</Typography>
              <Typography variant="caption" color="text.secondary">
                Active Models
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={2.4}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <SpeedIcon color="primary" sx={{ fontSize: 40, mb: 1 }} />
              <Typography variant="h6">{Math.round(stats.avgResponseTime)}ms</Typography>
              <Typography variant="caption" color="text.secondary">
                Avg Response
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Search and Controls */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <TextField
          placeholder="Search clusters..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
          sx={{ width: 300 }}
        />
        <Typography variant="body2" color="text.secondary">
          {filteredAndSortedClusters.length} of {clusters.length} clusters
        </Typography>
      </Box>

      {/* Clusters Table */}
      <Paper>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>
                  <TableSortLabel
                    active={sortField === 'name'}
                    direction={sortField === 'name' ? sortDirection : 'asc'}
                    onClick={() => handleSort('name')}
                  >
                    Cluster
                  </TableSortLabel>
                </TableCell>
                <TableCell>
                  <TableSortLabel
                    active={sortField === 'region'}
                    direction={sortField === 'region' ? sortDirection : 'asc'}
                    onClick={() => handleSort('region')}
                  >
                    Region
                  </TableSortLabel>
                </TableCell>
                <TableCell>
                  <TableSortLabel
                    active={sortField === 'status'}
                    direction={sortField === 'status' ? sortDirection : 'asc'}
                    onClick={() => handleSort('status')}
                  >
                    Status
                  </TableSortLabel>
                </TableCell>
                <TableCell>
                  <TableSortLabel
                    active={sortField === 'health'}
                    direction={sortField === 'health' ? sortDirection : 'asc'}
                    onClick={() => handleSort('health')}
                  >
                    Health
                  </TableSortLabel>
                </TableCell>
                <TableCell align="right">Nodes</TableCell>
                <TableCell align="right">Models</TableCell>
                <TableCell align="right">
                  <TableSortLabel
                    active={sortField === 'performance'}
                    direction={sortField === 'performance' ? sortDirection : 'asc'}
                    onClick={() => handleSort('performance')}
                  >
                    Response Time
                  </TableSortLabel>
                </TableCell>
                <TableCell>Resources</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredAndSortedClusters
                .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                .map((cluster) => (
                  <TableRow key={cluster.id} hover>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                        {getRegionAvatar(cluster.region)}
                        <Box>
                          <Typography variant="subtitle2">
                            {cluster.name}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {cluster.endpoint}
                          </Typography>
                        </Box>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={cluster.region}
                        size="small"
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={cluster.status}
                        color={getStatusColor(cluster.status) as any}
                        size="small"
                        sx={{ textTransform: 'capitalize' }}
                      />
                    </TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Chip
                          icon={getHealthIcon(cluster.health.overall)}
                          label={cluster.health.overall}
                          color={getHealthColor(cluster.health.overall) as any}
                          size="small"
                          sx={{ textTransform: 'capitalize' }}
                        />
                        {cluster.health.issues.length > 0 && (
                          <Tooltip title={`${cluster.health.issues.length} issues`}>
                            <WarningIcon color="warning" fontSize="small" />
                          </Tooltip>
                        )}
                      </Box>
                    </TableCell>
                    <TableCell align="right">
                      <Typography variant="body2">
                        {formatNumber(cluster.nodes)}
                      </Typography>
                    </TableCell>
                    <TableCell align="right">
                      <Typography variant="body2">
                        {formatNumber(cluster.activeModels)}
                      </Typography>
                    </TableCell>
                    <TableCell align="right">
                      <Typography variant="body2">
                        {Math.round(cluster.metrics.responseTime)}ms
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Box sx={{ width: 80 }}>
                        <Typography variant="caption" color="text.secondary">
                          CPU: {cluster.health.cpu}%
                        </Typography>
                        <LinearProgress
                          variant="determinate"
                          value={cluster.health.cpu}
                          size="small"
                          color={cluster.health.cpu > 80 ? 'error' : cluster.health.cpu > 60 ? 'warning' : 'primary'}
                          sx={{ height: 4, mb: 0.5 }}
                        />
                        <Typography variant="caption" color="text.secondary">
                          Mem: {cluster.health.memory}%
                        </Typography>
                        <LinearProgress
                          variant="determinate"
                          value={cluster.health.memory}
                          size="small"
                          color={cluster.health.memory > 80 ? 'error' : cluster.health.memory > 60 ? 'warning' : 'primary'}
                          sx={{ height: 4 }}
                        />
                      </Box>
                    </TableCell>
                    <TableCell align="right">
                      <IconButton
                        size="small"
                        onClick={(e) => handleMenuOpen(e, cluster)}
                      >
                        <MoreVertIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        </TableContainer>

        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 50]}
          component="div"
          count={filteredAndSortedClusters.length}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={(_, newPage) => setPage(newPage)}
          onRowsPerPageChange={(e) => {
            setRowsPerPage(parseInt(e.target.value, 10));
            setPage(0);
          }}
        />
      </Paper>

      {/* Action Menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
      >
        <MenuItem onClick={handleViewDetails}>
          <SettingsIcon sx={{ mr: 1 }} />
          View Details
        </MenuItem>
        <MenuItem onClick={handleEdit}>
          <EditIcon sx={{ mr: 1 }} />
          Edit
        </MenuItem>
        <MenuItem onClick={handleToggleStatus}>
          {selectedCluster?.status === 'online' ? (
            <>
              <StopIcon sx={{ mr: 1 }} />
              Take Offline
            </>
          ) : (
            <>
              <PlayArrowIcon sx={{ mr: 1 }} />
              Bring Online
            </>
          )}
        </MenuItem>
        <MenuItem onClick={handleDelete} sx={{ color: 'error.main' }}>
          <DeleteIcon sx={{ mr: 1 }} />
          Remove
        </MenuItem>
      </Menu>

      {/* Dialogs */}
      {selectedCluster && (
        <>
          <ClusterDetailsDialog
            cluster={selectedCluster}
            open={detailsDialogOpen}
            onClose={() => {
              setDetailsDialogOpen(false);
              setSelectedCluster(null);
            }}
          />
          <ClusterDialog
            cluster={selectedCluster}
            open={editDialogOpen}
            onClose={() => {
              setEditDialogOpen(false);
              setSelectedCluster(null);
            }}
          />
        </>
      )}
    </Box>
  );
};

export default ClusterList;