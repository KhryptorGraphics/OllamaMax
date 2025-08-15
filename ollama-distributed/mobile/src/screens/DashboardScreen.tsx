/**
 * Dashboard Screen - React Native
 * 
 * Main dashboard screen with real-time monitoring and native mobile interactions.
 */

import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  RefreshControl,
  Alert,
  Dimensions,
  Platform,
  StatusBar,
} from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useNavigation, useFocusEffect } from '@react-navigation/native';
import LinearGradient from 'react-native-linear-gradient';
import HapticFeedback from 'react-native-haptic-feedback';

// Components
import { MetricCard } from '../components/MetricCard';
import { NodeCard } from '../components/NodeCard';
import { StatusIndicator } from '../components/StatusIndicator';
import { FloatingActionButton } from '../components/FloatingActionButton';
import { PullToRefreshIndicator } from '../components/PullToRefreshIndicator';

// Hooks and Services
import { useTheme } from '../contexts/ThemeContext';
import { useAuth } from '../contexts/AuthContext';
import { useNetwork } from '../contexts/NetworkContext';
import { dashboardService } from '../services/DashboardService';
import { notificationService } from '../services/NotificationService';
import { analyticsService } from '../services/AnalyticsService';

// Types
interface DashboardData {
  clusterStatus: 'healthy' | 'warning' | 'error';
  nodeCount: number;
  activeModels: number;
  totalRequests: number;
  avgResponseTime: number;
  errorRate: number;
  uptime: string;
  nodes: Array<{
    id: string;
    status: 'healthy' | 'warning' | 'error';
    cpu: number;
    memory: number;
    requests: number;
  }>;
}

const { width: screenWidth } = Dimensions.get('window');

const DashboardScreen: React.FC = () => {
  const navigation = useNavigation();
  const insets = useSafeAreaInsets();
  const { theme, isDark } = useTheme();
  const { user } = useAuth();
  const { isConnected } = useNetwork();

  const [dashboardData, setDashboardData] = useState<DashboardData>({
    clusterStatus: 'healthy',
    nodeCount: 3,
    activeModels: 5,
    totalRequests: 1247,
    avgResponseTime: 245,
    errorRate: 0.02,
    uptime: '99.9%',
    nodes: [
      { id: 'node-1', status: 'healthy', cpu: 45, memory: 67, requests: 423 },
      { id: 'node-2', status: 'healthy', cpu: 52, memory: 71, requests: 389 },
      { id: 'node-3', status: 'warning', cpu: 78, memory: 89, requests: 435 },
    ],
  });

  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date());

  // Load dashboard data
  const loadDashboardData = useCallback(async (showLoading = false) => {
    try {
      if (showLoading) {
        setIsRefreshing(true);
      }

      const data = await dashboardService.getDashboardData();
      setDashboardData(data);
      setLastUpdated(new Date());

      // Log analytics
      analyticsService.logEvent('dashboard_data_loaded', {
        nodeCount: data.nodeCount,
        clusterStatus: data.clusterStatus,
      });

      // Haptic feedback for successful refresh
      if (showLoading) {
        HapticFeedback.trigger('impactLight');
      }
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
      
      if (isConnected) {
        Alert.alert(
          'Error',
          'Failed to load dashboard data. Please try again.',
          [
            { text: 'Retry', onPress: () => loadDashboardData(true) },
            { text: 'Cancel', style: 'cancel' },
          ]
        );
      } else {
        notificationService.showLocalNotification({
          title: 'Offline Mode',
          body: 'Using cached data. Connect to internet for latest updates.',
        });
      }
    } finally {
      setIsRefreshing(false);
    }
  }, [isConnected]);

  // Initial load and focus effect
  useFocusEffect(
    useCallback(() => {
      loadDashboardData();
      
      // Set up real-time updates
      const interval = setInterval(() => {
        if (isConnected) {
          loadDashboardData();
        }
      }, 30000); // Update every 30 seconds

      return () => clearInterval(interval);
    }, [loadDashboardData, isConnected])
  );

  // Handle pull to refresh
  const onRefresh = useCallback(() => {
    HapticFeedback.trigger('impactMedium');
    loadDashboardData(true);
  }, [loadDashboardData]);

  // Handle metric card press
  const handleMetricPress = useCallback((metricType: string) => {
    HapticFeedback.trigger('impactLight');
    analyticsService.logEvent('metric_card_pressed', { metricType });
    
    navigation.navigate('MetricDetails', { 
      metricType,
      value: dashboardData[metricType as keyof DashboardData],
    });
  }, [navigation, dashboardData]);

  // Handle node press
  const handleNodePress = useCallback((node: any) => {
    HapticFeedback.trigger('impactLight');
    analyticsService.logEvent('node_card_pressed', { nodeId: node.id });
    
    navigation.navigate('NodeDetails', { node });
  }, [navigation]);

  // Handle floating action button press
  const handleFABPress = useCallback(() => {
    HapticFeedback.trigger('impactMedium');
    navigation.navigate('QuickActions');
  }, [navigation]);

  const styles = createStyles(theme, insets);

  return (
    <View style={styles.container}>
      <StatusBar
        barStyle={isDark ? 'light-content' : 'dark-content'}
        backgroundColor={theme.colors.background}
        translucent={false}
      />

      {/* Header */}
      <LinearGradient
        colors={theme.gradients.primary}
        style={styles.header}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
      >
        <View style={styles.headerContent}>
          <View>
            <Text style={styles.greeting}>
              Welcome back, {user?.firstName}! 👋
            </Text>
            <Text style={styles.subtitle}>
              Monitor your cluster on the go
            </Text>
          </View>
          
          <StatusIndicator
            status={isConnected ? 'online' : 'offline'}
            label={isConnected ? 'Live' : 'Offline'}
          />
        </View>
      </LinearGradient>

      {/* Content */}
      <ScrollView
        style={styles.content}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={onRefresh}
            tintColor={theme.colors.primary}
            colors={[theme.colors.primary]}
            progressBackgroundColor={theme.colors.surface}
          />
        }
      >
        {/* Metrics Grid */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>System Metrics</Text>
          
          <View style={styles.metricsGrid}>
            <MetricCard
              title="Total Requests"
              value={dashboardData.totalRequests.toLocaleString()}
              subtitle="Last 24 hours"
              trend={12}
              icon="📊"
              color={theme.colors.primary}
              onPress={() => handleMetricPress('totalRequests')}
              style={styles.metricCard}
            />
            
            <MetricCard
              title="Response Time"
              value={`${dashboardData.avgResponseTime}ms`}
              subtitle="95th percentile"
              trend={-5}
              icon="⚡"
              color={theme.colors.success}
              onPress={() => handleMetricPress('avgResponseTime')}
              style={styles.metricCard}
            />
            
            <MetricCard
              title="Error Rate"
              value={`${(dashboardData.errorRate * 100).toFixed(2)}%`}
              subtitle="Last hour"
              trend={-15}
              icon="🚨"
              color={dashboardData.errorRate > 0.05 ? theme.colors.error : theme.colors.warning}
              onPress={() => handleMetricPress('errorRate')}
              style={styles.metricCard}
            />
            
            <MetricCard
              title="Uptime"
              value={dashboardData.uptime}
              subtitle="This month"
              icon="🔄"
              color={theme.colors.info}
              onPress={() => handleMetricPress('uptime')}
              style={styles.metricCard}
            />
          </View>
        </View>

        {/* Cluster Status */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>
            Cluster Nodes ({dashboardData.nodeCount})
          </Text>
          
          {dashboardData.nodes.map((node) => (
            <NodeCard
              key={node.id}
              node={node}
              onPress={() => handleNodePress(node)}
              style={styles.nodeCard}
            />
          ))}
        </View>

        {/* Last Updated */}
        <View style={styles.lastUpdated}>
          <Text style={styles.lastUpdatedText}>
            Last updated: {lastUpdated.toLocaleTimeString()}
          </Text>
        </View>

        {/* Bottom spacing for FAB */}
        <View style={styles.bottomSpacing} />
      </ScrollView>

      {/* Floating Action Button */}
      <FloatingActionButton
        icon="⚡"
        onPress={handleFABPress}
        style={styles.fab}
      />
    </View>
  );
};

const createStyles = (theme: any, insets: any) => StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: theme.colors.background,
  },
  header: {
    paddingTop: insets.top,
    paddingHorizontal: 20,
    paddingBottom: 20,
  },
  headerContent: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 10,
  },
  greeting: {
    fontSize: 24,
    fontWeight: 'bold',
    color: theme.colors.neutral[0],
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 16,
    color: theme.colors.neutral[100],
    opacity: 0.9,
  },
  content: {
    flex: 1,
    paddingHorizontal: 20,
  },
  section: {
    marginTop: 24,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: theme.colors.text,
    marginBottom: 16,
  },
  metricsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
  },
  metricCard: {
    width: (screenWidth - 60) / 2,
    marginBottom: 16,
  },
  nodeCard: {
    marginBottom: 12,
  },
  lastUpdated: {
    alignItems: 'center',
    marginTop: 24,
    marginBottom: 16,
  },
  lastUpdatedText: {
    fontSize: 14,
    color: theme.colors.textSecondary,
  },
  bottomSpacing: {
    height: 80, // Space for FAB
  },
  fab: {
    position: 'absolute',
    bottom: insets.bottom + 20,
    right: 20,
  },
});

export default DashboardScreen;
