import React, { useState, useEffect } from 'react';
import { Card, Form, Button, Badge, Table } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faChartLine,
  faDownload,
  faCalendar,
  faFilter,
  faClock,
  faServer,
  faBrain,
  faUsers,
  faExclamationTriangle,
  faCheckCircle
} from '@fortawesome/free-solid-svg-icons';
import MetricsChart from './MetricsChart';

const Analytics = ({ analyticsData = {}, onExportReport }) => {
  const [timeRange, setTimeRange] = useState('24h');
  const [selectedMetrics, setSelectedMetrics] = useState(['requests', 'latency', 'errors']);
  const [filteredData, setFilteredData] = useState({});

  const timeRangeOptions = [
    { value: '1h', label: 'Last Hour' },
    { value: '24h', label: 'Last 24 Hours' },
    { value: '7d', label: 'Last 7 Days' },
    { value: '30d', label: 'Last 30 Days' },
    { value: '90d', label: 'Last 3 Months' }
  ];

  const metricOptions = [
    { value: 'requests', label: 'Request Count', icon: faServer },
    { value: 'latency', label: 'Average Latency', icon: faClock },
    { value: 'errors', label: 'Error Rate', icon: faExclamationTriangle },
    { value: 'throughput', label: 'Throughput', icon: faChartLine },
    { value: 'nodes', label: 'Node Activity', icon: faServer },
    { value: 'models', label: 'Model Usage', icon: faBrain },
    { value: 'users', label: 'User Activity', icon: faUsers }
  ];

  useEffect(() => {
    // Filter and process data based on selected time range and metrics
    const processedData = processAnalyticsData(analyticsData, timeRange, selectedMetrics);
    setFilteredData(processedData);
  }, [analyticsData, timeRange, selectedMetrics]);

  const processAnalyticsData = (data, range, metrics) => {
    // Simulate data processing based on time range and selected metrics
    const now = Date.now();
    const timeRanges = {
      '1h': 60 * 60 * 1000,
      '24h': 24 * 60 * 60 * 1000,
      '7d': 7 * 24 * 60 * 60 * 1000,
      '30d': 30 * 24 * 60 * 60 * 1000,
      '90d': 90 * 24 * 60 * 60 * 1000
    };

    const rangeMs = timeRanges[range];
    const startTime = now - rangeMs;

    return {
      summary: {
        totalRequests: Math.floor(Math.random() * 10000),
        averageLatency: Math.floor(Math.random() * 200),
        errorRate: Math.random() * 5,
        activeNodes: Math.floor(Math.random() * 10) + 1,
        modelsUsed: Math.floor(Math.random() * 20) + 5,
        uptime: Math.random() * 100
      },
      chartData: generateChartData(metrics, rangeMs),
      topModels: generateTopModels(),
      errorBreakdown: generateErrorBreakdown(),
      performanceMetrics: generatePerformanceMetrics()
    };
  };

  const generateChartData = (metrics, timeRange) => {
    const dataPoints = Math.min(50, timeRange / (60 * 60 * 1000)); // Max 50 points
    const data = {};

    metrics.forEach(metric => {
      data[metric] = Array.from({ length: dataPoints }, (_, i) => ({
        timestamp: Date.now() - (timeRange * (dataPoints - i) / dataPoints),
        value: Math.random() * 100
      }));
    });

    return data;
  };

  const generateTopModels = () => [
    { name: 'llama2:7b', requests: 1234, latency: 145, usage: 85 },
    { name: 'codellama:13b', requests: 987, latency: 234, usage: 72 },
    { name: 'mistral:7b', requests: 756, latency: 123, usage: 68 },
    { name: 'neural-chat:7b', requests: 543, latency: 167, usage: 54 },
    { name: 'vicuna:13b', requests: 321, latency: 289, usage: 41 }
  ];

  const generateErrorBreakdown = () => [
    { type: 'Timeout', count: 23, percentage: 45.1 },
    { type: 'Network Error', count: 15, percentage: 29.4 },
    { type: 'Model Load Failed', count: 8, percentage: 15.7 },
    { type: 'Authentication', count: 3, percentage: 5.9 },
    { type: 'Other', count: 2, percentage: 3.9 }
  ];

  const generatePerformanceMetrics = () => [
    { node: 'node-001', cpu: 67, memory: 82, requests: 1456, latency: 123 },
    { node: 'node-002', cpu: 45, memory: 64, requests: 1234, latency: 145 },
    { node: 'node-003', cpu: 73, memory: 78, requests: 1098, latency: 167 },
    { node: 'node-004', cpu: 56, memory: 71, requests: 987, latency: 134 }
  ];

  const handleMetricToggle = (metric) => {
    setSelectedMetrics(prev => 
      prev.includes(metric) 
        ? prev.filter(m => m !== metric)
        : [...prev, metric]
    );
  };

  const handleExport = () => {
    const reportData = {
      timeRange,
      selectedMetrics,
      data: filteredData,
      generatedAt: new Date().toISOString()
    };
    
    if (onExportReport) {
      onExportReport(reportData);
    }
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Analytics Dashboard</h2>
        <div className="d-flex gap-2">
          <Form.Select 
            value={timeRange} 
            onChange={(e) => setTimeRange(e.target.value)}
            style={{ width: '200px' }}
          >
            {timeRangeOptions.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </Form.Select>
          <Button variant="outline-primary" onClick={handleExport}>
            <FontAwesomeIcon icon={faDownload} className="me-2" />
            Export Report
          </Button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="row mb-4">
        <div className="col-md-2">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faServer} size="2x" className="text-primary mb-2" />
              <h4>{filteredData.summary?.totalRequests?.toLocaleString()}</h4>
              <p className="mb-0 text-muted">Total Requests</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-2">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faClock} size="2x" className="text-info mb-2" />
              <h4>{filteredData.summary?.averageLatency}ms</h4>
              <p className="mb-0 text-muted">Avg Latency</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-2">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faExclamationTriangle} size="2x" className="text-warning mb-2" />
              <h4>{filteredData.summary?.errorRate?.toFixed(1)}%</h4>
              <p className="mb-0 text-muted">Error Rate</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-2">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faServer} size="2x" className="text-success mb-2" />
              <h4>{filteredData.summary?.activeNodes}</h4>
              <p className="mb-0 text-muted">Active Nodes</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-2">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faBrain} size="2x" className="text-purple mb-2" />
              <h4>{filteredData.summary?.modelsUsed}</h4>
              <p className="mb-0 text-muted">Models Used</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-2">
          <Card className="metric-card h-100">
            <Card.Body className="text-center">
              <FontAwesomeIcon icon={faCheckCircle} size="2x" className="text-success mb-2" />
              <h4>{filteredData.summary?.uptime?.toFixed(1)}%</h4>
              <p className="mb-0 text-muted">Uptime</p>
            </Card.Body>
          </Card>
        </div>
      </div>

      {/* Metric Selection */}
      <Card className="mb-4">
        <Card.Header>
          <h6 className="mb-0">
            <FontAwesomeIcon icon={faFilter} className="me-2" />
            Select Metrics to Display
          </h6>
        </Card.Header>
        <Card.Body>
          <div className="d-flex flex-wrap gap-2">
            {metricOptions.map(option => (
              <Badge
                key={option.value}
                bg={selectedMetrics.includes(option.value) ? 'primary' : 'outline-secondary'}
                className="metric-toggle"
                onClick={() => handleMetricToggle(option.value)}
                style={{ cursor: 'pointer' }}
              >
                <FontAwesomeIcon icon={option.icon} className="me-1" />
                {option.label}
              </Badge>
            ))}
          </div>
        </Card.Body>
      </Card>

      {/* Charts */}
      <div className="row mb-4">
        {selectedMetrics.map(metric => (
          <div key={metric} className="col-md-6 mb-4">
            <Card>
              <Card.Header>
                <h6 className="mb-0">
                  {metricOptions.find(opt => opt.value === metric)?.label}
                </h6>
              </Card.Header>
              <Card.Body>
                <MetricsChart 
                  data={filteredData.chartData?.[metric] || []}
                  title={metricOptions.find(opt => opt.value === metric)?.label}
                  color={`hsl(${Math.random() * 360}, 70%, 50%)`}
                />
              </Card.Body>
            </Card>
          </div>
        ))}
      </div>

      {/* Detailed Tables */}
      <div className="row">
        <div className="col-md-6">
          <Card className="mb-4">
            <Card.Header>
              <h6 className="mb-0">Top Models by Usage</h6>
            </Card.Header>
            <Card.Body>
              <Table responsive>
                <thead>
                  <tr>
                    <th>Model</th>
                    <th>Requests</th>
                    <th>Avg Latency</th>
                    <th>Usage</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredData.topModels?.map((model, index) => (
                    <tr key={index}>
                      <td>{model.name}</td>
                      <td>{model.requests}</td>
                      <td>{model.latency}ms</td>
                      <td>
                        <div className="progress" style={{ height: '6px' }}>
                          <div 
                            className="progress-bar bg-primary" 
                            style={{ width: `${model.usage}%` }}
                          ></div>
                        </div>
                        {model.usage}%
                      </td>
                    </tr>
                  ))}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </div>

        <div className="col-md-6">
          <Card className="mb-4">
            <Card.Header>
              <h6 className="mb-0">Error Breakdown</h6>
            </Card.Header>
            <Card.Body>
              <Table responsive>
                <thead>
                  <tr>
                    <th>Error Type</th>
                    <th>Count</th>
                    <th>Percentage</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredData.errorBreakdown?.map((error, index) => (
                    <tr key={index}>
                      <td>{error.type}</td>
                      <td>{error.count}</td>
                      <td>{error.percentage}%</td>
                    </tr>
                  ))}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </div>
      </div>

      {/* Node Performance Table */}
      <Card>
        <Card.Header>
          <h6 className="mb-0">Node Performance Summary</h6>
        </Card.Header>
        <Card.Body>
          <Table responsive>
            <thead>
              <tr>
                <th>Node ID</th>
                <th>CPU Usage</th>
                <th>Memory Usage</th>
                <th>Requests Handled</th>
                <th>Avg Latency</th>
              </tr>
            </thead>
            <tbody>
              {filteredData.performanceMetrics?.map((node, index) => (
                <tr key={index}>
                  <td>{node.node}</td>
                  <td>
                    <span className={`badge ${
                      node.cpu > 80 ? 'bg-danger' : 
                      node.cpu > 60 ? 'bg-warning' : 
                      'bg-success'
                    }`}>
                      {node.cpu}%
                    </span>
                  </td>
                  <td>
                    <span className={`badge ${
                      node.memory > 80 ? 'bg-danger' : 
                      node.memory > 60 ? 'bg-warning' : 
                      'bg-success'
                    }`}>
                      {node.memory}%
                    </span>
                  </td>
                  <td>{node.requests}</td>
                  <td>{node.latency}ms</td>
                </tr>
              ))}
            </tbody>
          </Table>
        </Card.Body>
      </Card>
    </div>
  );
};

export default Analytics;