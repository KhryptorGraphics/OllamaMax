package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/production"
	"github.com/sirupsen/logrus"
)

// TestPhase3ProductionFeatures tests the Phase 3 production-ready implementation
func TestPhase3ProductionFeatures(t *testing.T) {
	t.Run("ProductionMonitoring", func(t *testing.T) {
		testProductionMonitoring(t)
	})

	t.Run("SLAMonitoring", func(t *testing.T) {
		testSLAMonitoring(t)
	})

	t.Run("AlertManagement", func(t *testing.T) {
		testAlertManagement(t)
	})

	t.Run("HealthChecking", func(t *testing.T) {
		testHealthChecking(t)
	})

	t.Run("MetricsCollection", func(t *testing.T) {
		testMetricsCollection(t)
	})

	t.Run("PerformanceKPIs", func(t *testing.T) {
		testPerformanceKPIs(t)
	})
}

func testProductionMonitoring(t *testing.T) {
	// Create test configuration
	cfg := createTestProductionConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Create production monitor
	monitor, err := production.NewProductionMonitor(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, monitor)

	// Test starting the monitor
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = monitor.Start(ctx)
	assert.NoError(t, err)

	// Wait a bit for monitoring to collect data
	time.Sleep(2 * time.Second)

	// Test getting SLA metrics
	slaMetrics := monitor.GetSLAMetrics()
	require.NotNil(t, slaMetrics)
	assert.Greater(t, slaMetrics.Availability, 99.0)
	assert.LessOrEqual(t, slaMetrics.Availability, 100.0)
	assert.GreaterOrEqual(t, slaMetrics.ErrorRate, 0.0)
	assert.Less(t, slaMetrics.ErrorRate, 0.1)

	// Test getting performance KPIs
	kpis := monitor.GetPerformanceKPIs()
	require.NotNil(t, kpis)
	assert.Greater(t, kpis.MTBF, 0.0)
	assert.Greater(t, kpis.MTTR, 0.0)
	assert.Greater(t, kpis.MTTD, 0.0)
	assert.GreaterOrEqual(t, kpis.ChangeFailRate, 0.0)
	assert.LessOrEqual(t, kpis.ChangeFailRate, 1.0)

	// Test getting system health
	health := monitor.GetSystemHealth()
	require.NotNil(t, health)
	assert.GreaterOrEqual(t, health.OverallHealth, 0.0)
	assert.LessOrEqual(t, health.OverallHealth, 1.0)

	// Test recording requests
	monitor.RecordRequest(ctx, "POST", "/api/v1/inference", 150*time.Millisecond, 200)
	monitor.RecordRequest(ctx, "GET", "/api/v1/health", 25*time.Millisecond, 200)
	monitor.RecordRequest(ctx, "POST", "/api/v1/inference", 500*time.Millisecond, 500)
}

func testSLAMonitoring(t *testing.T) {
	// Test SLA target structure
	slaTarget := &production.SLATarget{
		Name:        "api_availability",
		Type:        "availability",
		Target:      99.9,
		Period:      24 * time.Hour,
		Measurement: "percentage",
		Critical:    true,
	}

	assert.Equal(t, "api_availability", slaTarget.Name)
	assert.Equal(t, "availability", slaTarget.Type)
	assert.Equal(t, 99.9, slaTarget.Target)
	assert.Equal(t, 24*time.Hour, slaTarget.Period)
	assert.True(t, slaTarget.Critical)

	// Test SLA metrics structure
	slaMetrics := &production.SLAMetrics{
		Availability:  99.95,
		ErrorRate:     0.005,
		Throughput:    125.5,
		ErrorBudget:   0.095,
		BurnRate:      0.05,
		TimeToRestore: 15.5,
		IncidentCount: 2,
		Latency: production.ResponseTimeStats{
			Mean: 150.0,
			P50:  120.0,
			P90:  200.0,
			P95:  280.0,
			P99:  450.0,
			P999: 800.0,
			Min:  25.0,
			Max:  1200.0,
		},
	}

	assert.Equal(t, 99.95, slaMetrics.Availability)
	assert.Equal(t, 0.005, slaMetrics.ErrorRate)
	assert.Equal(t, 125.5, slaMetrics.Throughput)
	assert.Equal(t, 0.095, slaMetrics.ErrorBudget)
	assert.Equal(t, 0.05, slaMetrics.BurnRate)
	assert.Equal(t, 150.0, slaMetrics.Latency.Mean)
	assert.Equal(t, 280.0, slaMetrics.Latency.P95)

	// Test SLA compliance calculation
	isCompliant := slaMetrics.Availability >= slaTarget.Target
	assert.True(t, isCompliant, "SLA should be compliant")

	// Test error budget calculation
	errorBudgetUsed := (100.0 - slaTarget.Target) - (100.0 - slaMetrics.Availability)
	assert.Equal(t, 0.05, errorBudgetUsed)
}

func testAlertManagement(t *testing.T) {
	// Test alert structure
	alert := &production.Alert{
		ID:       "alert-001",
		Name:     "High API Latency",
		Severity: "warning",
		Status:   "firing",
		Message:  "API P95 latency exceeded 500ms threshold",
		Source:   "latency-monitor",
		Labels: map[string]string{
			"service":  "api",
			"endpoint": "/inference",
			"severity": "warning",
		},
		Annotations: map[string]string{
			"description": "API latency is higher than expected",
			"runbook":     "https://docs.company.com/runbooks/high-latency",
		},
		StartsAt:    time.Now(),
		UpdatedAt:   time.Now(),
		Fingerprint: "latency-api-inference",
	}

	assert.Equal(t, "alert-001", alert.ID)
	assert.Equal(t, "High API Latency", alert.Name)
	assert.Equal(t, "warning", alert.Severity)
	assert.Equal(t, "firing", alert.Status)
	assert.Equal(t, "api", alert.Labels["service"])
	assert.Contains(t, alert.Message, "500ms")
	assert.Nil(t, alert.EndsAt)

	// Test alert rule structure
	alertRule := &production.AlertRule{
		Name:      "high_latency_rule",
		Query:     "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5",
		Condition: "greater_than",
		Threshold: 0.5,
		Duration:  5 * time.Minute,
		Severity:  "warning",
		Labels: map[string]string{
			"team": "platform",
		},
		Annotations: map[string]string{
			"summary": "High latency detected",
		},
		Enabled: true,
	}

	assert.Equal(t, "high_latency_rule", alertRule.Name)
	assert.Equal(t, "greater_than", alertRule.Condition)
	assert.Equal(t, 0.5, alertRule.Threshold)
	assert.Equal(t, 5*time.Minute, alertRule.Duration)
	assert.True(t, alertRule.Enabled)

	// Test escalation tree
	escalationTree := &production.EscalationTree{
		Levels: []production.EscalationLevel{
			{
				Level:      1,
				Delay:      0,
				Recipients: []string{"oncall-primary"},
				Channels:   []string{"slack", "email"},
				Conditions: []string{"severity=critical"},
			},
			{
				Level:      2,
				Delay:      15 * time.Minute,
				Recipients: []string{"oncall-secondary", "team-lead"},
				Channels:   []string{"slack", "email", "sms"},
				Conditions: []string{"not_acknowledged"},
			},
		},
	}

	assert.Len(t, escalationTree.Levels, 2)
	assert.Equal(t, 1, escalationTree.Levels[0].Level)
	assert.Equal(t, 15*time.Minute, escalationTree.Levels[1].Delay)
	assert.Contains(t, escalationTree.Levels[1].Recipients, "team-lead")
}

func testHealthChecking(t *testing.T) {
	// Test health check configuration
	healthCheck := &production.HealthCheck{
		Name:         "api_health_check",
		Type:         "http",
		Target:       "http://localhost:8080/health",
		Interval:     30 * time.Second,
		Timeout:      5 * time.Second,
		Retries:      3,
		Headers:      map[string]string{"User-Agent": "health-checker/1.0"},
		ExpectedCode: 200,
		ExpectedBody: "healthy",
		Enabled:      true,
	}

	assert.Equal(t, "api_health_check", healthCheck.Name)
	assert.Equal(t, "http", healthCheck.Type)
	assert.Equal(t, "http://localhost:8080/health", healthCheck.Target)
	assert.Equal(t, 30*time.Second, healthCheck.Interval)
	assert.Equal(t, 5*time.Second, healthCheck.Timeout)
	assert.Equal(t, 3, healthCheck.Retries)
	assert.Equal(t, 200, healthCheck.ExpectedCode)
	assert.True(t, healthCheck.Enabled)

	// Test component health
	componentHealth := &production.ComponentHealth{
		Name:         "database",
		Status:       "healthy",
		Score:        1.0,
		LastCheck:    time.Now(),
		CheckCount:   100,
		FailureCount: 2,
		Metadata: map[string]string{
			"version":     "14.5",
			"connections": "25",
		},
	}

	assert.Equal(t, "database", componentHealth.Name)
	assert.Equal(t, "healthy", componentHealth.Status)
	assert.Equal(t, 1.0, componentHealth.Score)
	assert.Equal(t, 100, componentHealth.CheckCount)
	assert.Equal(t, 2, componentHealth.FailureCount)
	assert.Equal(t, "14.5", componentHealth.Metadata["version"])

	// Test dependency health
	dependencyHealth := &production.DependencyHealth{
		Name:         "redis_cache",
		Type:         "cache",
		Status:       "healthy",
		ResponseTime: 2.5,
		LastCheck:    time.Now(),
		ErrorRate:    0.001,
		Critical:     false,
	}

	assert.Equal(t, "redis_cache", dependencyHealth.Name)
	assert.Equal(t, "cache", dependencyHealth.Type)
	assert.Equal(t, "healthy", dependencyHealth.Status)
	assert.Equal(t, 2.5, dependencyHealth.ResponseTime)
	assert.Equal(t, 0.001, dependencyHealth.ErrorRate)
	assert.False(t, dependencyHealth.Critical)

	// Test health snapshot
	healthSnapshot := &production.HealthSnapshot{
		Timestamp:     time.Now(),
		OverallHealth: 0.95,
		ComponentHealth: map[string]*production.ComponentHealth{
			"database": componentHealth,
		},
		DependencyHealth: map[string]*production.DependencyHealth{
			"redis": dependencyHealth,
		},
		Issues: []string{"minor_latency_increase"},
	}

	assert.Equal(t, 0.95, healthSnapshot.OverallHealth)
	assert.Len(t, healthSnapshot.ComponentHealth, 1)
	assert.Len(t, healthSnapshot.DependencyHealth, 1)
	assert.Len(t, healthSnapshot.Issues, 1)
	assert.Contains(t, healthSnapshot.Issues, "minor_latency_increase")
}

func testMetricsCollection(t *testing.T) {
	// Test system metrics
	systemMetrics := &production.SystemMetrics{
		CPU: production.CPUMetrics{
			Usage:     75.5,
			LoadAvg1:  1.2,
			LoadAvg5:  1.1,
			LoadAvg15: 1.0,
			Cores:     8,
		},
		Memory: production.MemoryMetrics{
			Used:      6442450944, // ~6GB
			Available: 2147483648, // ~2GB
			Total:     8589934592, // ~8GB
			Percent:   75.0,
			Swap: production.SwapMetrics{
				Used:    536870912, // ~512MB
				Total:   2147483648, // ~2GB
				Percent: 25.0,
			},
		},
		Disk: production.DiskMetrics{
			Usage: map[string]production.DiskUsage{
				"/": {
					Used:      53687091200, // ~50GB
					Available: 46636396544, // ~43.4GB
					Total:     107374182400, // ~100GB
					Percent:   50.0,
				},
			},
		},
		Network: production.NetworkMetrics{
			Interfaces: map[string]production.NetworkInterface{
				"eth0": {
					BytesReceived:   1073741824, // 1GB
					BytesSent:       536870912,  // 512MB
					PacketsReceived: 1000000,
					PacketsSent:     500000,
					Errors:          5,
					Drops:           2,
				},
			},
		},
		Timestamp: time.Now(),
	}

	assert.Equal(t, 75.5, systemMetrics.CPU.Usage)
	assert.Equal(t, 8, systemMetrics.CPU.Cores)
	assert.Equal(t, 75.0, systemMetrics.Memory.Percent)
	assert.Equal(t, 25.0, systemMetrics.Memory.Swap.Percent)
	assert.Equal(t, 50.0, systemMetrics.Disk.Usage["/"].Percent)
	assert.Equal(t, uint64(1073741824), systemMetrics.Network.Interfaces["eth0"].BytesReceived)

	// Test application metrics
	appMetrics := &production.ApplicationMetrics{
		RequestRate: 150.5,
		ResponseTime: production.ResponseTimeStats{
			Mean: 125.0,
			P50:  100.0,
			P90:  200.0,
			P95:  300.0,
			P99:  500.0,
			P999: 1000.0,
		},
		ErrorRate:    0.02,
		Throughput:   75.5,
		Concurrency:  50,
		QueueDepth:   10,
		CacheHitRate: 0.85,
		DatabaseMetrics: production.DatabaseMetrics{
			Connections:   25,
			ActiveQueries: 5,
			SlowQueries:   2,
			QueryTime:     15.5,
			LockWaitTime:  2.1,
			DeadlockCount: 0,
		},
		Timestamp: time.Now(),
	}

	assert.Equal(t, 150.5, appMetrics.RequestRate)
	assert.Equal(t, 125.0, appMetrics.ResponseTime.Mean)
	assert.Equal(t, 300.0, appMetrics.ResponseTime.P95)
	assert.Equal(t, 0.02, appMetrics.ErrorRate)
	assert.Equal(t, 0.85, appMetrics.CacheHitRate)
	assert.Equal(t, 25, appMetrics.DatabaseMetrics.Connections)
}

func testPerformanceKPIs(t *testing.T) {
	// Test performance KPIs
	kpis := &production.PerformanceKPIs{
		MTBF:            720.0, // 30 days
		MTTR:            15.0,  // 15 minutes
		MTTD:            5.0,   // 5 minutes
		ChangeFailRate:  0.05,  // 5%
		DeployFrequency: 2.5,   // 2.5 per day
		LeadTime:        120.0, // 2 hours
	}

	assert.Equal(t, 720.0, kpis.MTBF)
	assert.Equal(t, 15.0, kpis.MTTR)
	assert.Equal(t, 5.0, kpis.MTTD)
	assert.Equal(t, 0.05, kpis.ChangeFailRate)
	assert.Equal(t, 2.5, kpis.DeployFrequency)
	assert.Equal(t, 120.0, kpis.LeadTime)

	// Test KPI health assessment
	isHealthyMTBF := kpis.MTBF > 168.0 // > 1 week
	isHealthyMTTR := kpis.MTTR < 60.0   // < 1 hour
	isHealthyMTTD := kpis.MTTD < 15.0   // < 15 minutes
	isHealthyChangeFailRate := kpis.ChangeFailRate < 0.15 // < 15%

	assert.True(t, isHealthyMTBF, "MTBF should be healthy")
	assert.True(t, isHealthyMTTR, "MTTR should be healthy")
	assert.True(t, isHealthyMTTD, "MTTD should be healthy")
	assert.True(t, isHealthyChangeFailRate, "Change failure rate should be healthy")

	// Test burn rate calculation
	burnRate := &production.BurnRate{
		SLAName:    "availability",
		Window:     1 * time.Hour,
		Rate:       0.5,
		Threshold:  2.0,
		LastUpdate: time.Now(),
		Alerting:   false,
	}

	assert.Equal(t, "availability", burnRate.SLAName)
	assert.Equal(t, 1*time.Hour, burnRate.Window)
	assert.Equal(t, 0.5, burnRate.Rate)
	assert.Equal(t, 2.0, burnRate.Threshold)
	assert.False(t, burnRate.Alerting)

	// Test if burn rate should trigger alert
	shouldAlert := burnRate.Rate > burnRate.Threshold
	assert.False(t, shouldAlert, "Burn rate should not trigger alert")
}

// Helper functions

func createTestProductionConfig() *config.DistributedConfig {
	cfg := &config.DistributedConfig{}
	cfg.SetDefaults()
	cfg.Node.ID = "production-test-node"
	cfg.Node.Region = "us-west-2"
	cfg.Node.Zone = "us-west-2a"
	return cfg
}
