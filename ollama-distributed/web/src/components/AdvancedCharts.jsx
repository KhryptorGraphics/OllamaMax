import React, { useState, useEffect } from 'react';
import { Card, Form, Button, Badge } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faChartLine,
  faChartBar,
  faChartPie,
  faChartArea,
  faDownload,
  faExpand,
  faCompress,
  faCog
} from '@fortawesome/free-solid-svg-icons';

const AdvancedCharts = ({ 
  data = {}, 
  title = "Advanced Metrics", 
  chartType = 'line',
  realTime = false,
  className = "",
  height = 300,
  onExport,
  interactive = true
}) => {
  const [currentChartType, setCurrentChartType] = useState(chartType);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [chartData, setChartData] = useState([]);
  const [animationEnabled, setAnimationEnabled] = useState(true);
  const [gridEnabled, setGridEnabled] = useState(true);
  const [labelsEnabled, setLabelsEnabled] = useState(true);

  const chartTypes = [
    { value: 'line', label: 'Line Chart', icon: faChartLine },
    { value: 'bar', label: 'Bar Chart', icon: faChartBar },
    { value: 'area', label: 'Area Chart', icon: faChartArea },
    { value: 'pie', label: 'Pie Chart', icon: faChartPie },
    { value: 'scatter', label: 'Scatter Plot', icon: faChartLine },
    { value: 'heatmap', label: 'Heatmap', icon: faChartArea }
  ];

  useEffect(() => {
    if (data && Object.keys(data).length > 0) {
      processChartData(data);
    }
  }, [data, currentChartType]);

  useEffect(() => {
    if (realTime) {
      const interval = setInterval(() => {
        generateRealtimeData();
      }, 1000);

      return () => clearInterval(interval);
    }
  }, [realTime]);

  const processChartData = (rawData) => {
    const processedData = Array.isArray(rawData) ? rawData : generateMockData();
    setChartData(processedData);
  };

  const generateMockData = () => {
    const dataPoints = 50;
    const now = Date.now();
    
    return Array.from({ length: dataPoints }, (_, i) => {
      const timestamp = now - (dataPoints - i) * 60000;
      return {
        x: new Date(timestamp).toLocaleTimeString(),
        timestamp: timestamp,
        value: Math.sin(i * 0.1) * 50 + 50 + Math.random() * 20,
        category: i % 3 === 0 ? 'High' : i % 3 === 1 ? 'Medium' : 'Low'
      };
    });
  };

  const generateRealtimeData = () => {
    setChartData(prevData => {
      const newPoint = {
        x: new Date().toLocaleTimeString(),
        timestamp: Date.now(),
        value: Math.sin(Date.now() * 0.001) * 50 + 50 + Math.random() * 20,
        category: ['High', 'Medium', 'Low'][Math.floor(Math.random() * 3)]
      };
      
      return [...prevData.slice(-49), newPoint];
    });
  };

  const renderLineChart = () => {
    if (!chartData.length) return null;

    const maxValue = Math.max(...chartData.map(d => d.value));
    const minValue = Math.min(...chartData.map(d => d.value));
    const valueRange = maxValue - minValue || 1;

    const svgHeight = height - 60;
    const svgWidth = 800;
    const padding = 40;

    const pathData = chartData.map((point, index) => {
      const x = padding + (index / (chartData.length - 1)) * (svgWidth - 2 * padding);
      const y = padding + (1 - (point.value - minValue) / valueRange) * (svgHeight - 2 * padding);
      return `${index === 0 ? 'M' : 'L'} ${x} ${y}`;
    }).join(' ');

    return (
      <div className="chart-container position-relative">
        <svg width="100%" height={height} viewBox={`0 0 ${svgWidth} ${height}`} className="advanced-chart">
          {/* Grid */}
          {gridEnabled && (
            <g className="grid">
              {Array.from({ length: 5 }, (_, i) => {
                const y = padding + (i / 4) * (svgHeight - 2 * padding);
                return (
                  <g key={i}>
                    <line 
                      x1={padding} 
                      y1={y} 
                      x2={svgWidth - padding} 
                      y2={y} 
                      stroke="var(--border-color)" 
                      strokeWidth="1" 
                      strokeDasharray="3,3"
                      opacity="0.3"
                    />
                    {labelsEnabled && (
                      <text 
                        x={padding - 5} 
                        y={y + 4} 
                        textAnchor="end" 
                        fontSize="12" 
                        fill="var(--text-secondary)"
                      >
                        {(maxValue - (i / 4) * valueRange).toFixed(0)}
                      </text>
                    )}
                  </g>
                );
              })}
            </g>
          )}

          {/* Line Path */}
          <path
            d={pathData}
            fill="none"
            stroke="var(--primary-color)"
            strokeWidth="2"
            className={animationEnabled ? 'animate-draw' : ''}
          />

          {/* Area Fill */}
          {currentChartType === 'area' && (
            <path
              d={`${pathData} L ${svgWidth - padding} ${svgHeight - padding} L ${padding} ${svgHeight - padding} Z`}
              fill="var(--primary-color)"
              fillOpacity="0.2"
            />
          )}

          {/* Data Points */}
          {chartData.map((point, index) => {
            const x = padding + (index / (chartData.length - 1)) * (svgWidth - 2 * padding);
            const y = padding + (1 - (point.value - minValue) / valueRange) * (svgHeight - 2 * padding);
            
            return (
              <g key={index}>
                <circle
                  cx={x}
                  cy={y}
                  r="4"
                  fill="var(--primary-color)"
                  className={interactive ? 'chart-point-interactive' : ''}
                  onMouseEnter={(e) => {
                    if (interactive) {
                      const tooltip = e.target.closest('.chart-container').querySelector('.tooltip');
                      if (tooltip) {
                        tooltip.style.display = 'block';
                        tooltip.style.left = e.pageX + 'px';
                        tooltip.style.top = e.pageY - 30 + 'px';
                        tooltip.textContent = `${point.x}: ${point.value.toFixed(2)}`;
                      }
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (interactive) {
                      const tooltip = e.target.closest('.chart-container').querySelector('.tooltip');
                      if (tooltip) {
                        tooltip.style.display = 'none';
                      }
                    }
                  }}
                />
              </g>
            );
          })}
        </svg>
        
        {/* Tooltip */}
        {interactive && (
          <div className="tooltip position-absolute bg-dark text-white p-2 rounded" style={{ display: 'none', zIndex: 1000, pointerEvents: 'none' }}></div>
        )}
      </div>
    );
  };

  const renderBarChart = () => {
    if (!chartData.length) return null;

    const maxValue = Math.max(...chartData.map(d => d.value));
    const svgHeight = height - 60;
    const svgWidth = 800;
    const padding = 40;
    const barWidth = (svgWidth - 2 * padding) / chartData.length * 0.8;
    const barSpacing = (svgWidth - 2 * padding) / chartData.length * 0.2;

    return (
      <div className="chart-container position-relative">
        <svg width="100%" height={height} viewBox={`0 0 ${svgWidth} ${height}`} className="advanced-chart">
          {/* Grid */}
          {gridEnabled && (
            <g className="grid">
              {Array.from({ length: 5 }, (_, i) => {
                const y = padding + (i / 4) * (svgHeight - 2 * padding);
                return (
                  <line 
                    key={i}
                    x1={padding} 
                    y1={y} 
                    x2={svgWidth - padding} 
                    y2={y} 
                    stroke="var(--border-color)" 
                    strokeWidth="1" 
                    strokeDasharray="3,3"
                    opacity="0.3"
                  />
                );
              })}
            </g>
          )}

          {/* Bars */}
          {chartData.map((point, index) => {
            const x = padding + index * ((svgWidth - 2 * padding) / chartData.length) + barSpacing / 2;
            const barHeight = (point.value / maxValue) * (svgHeight - 2 * padding);
            const y = svgHeight - padding - barHeight;
            
            return (
              <g key={index}>
                <rect
                  x={x}
                  y={y}
                  width={barWidth}
                  height={barHeight}
                  fill="var(--primary-color)"
                  className={interactive ? 'chart-bar-interactive' : ''}
                  rx="2"
                  onMouseEnter={(e) => {
                    if (interactive) {
                      e.target.setAttribute('fill', 'var(--primary-dark)');
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (interactive) {
                      e.target.setAttribute('fill', 'var(--primary-color)');
                    }
                  }}
                />
                {labelsEnabled && (
                  <text 
                    x={x + barWidth / 2} 
                    y={svgHeight - padding + 15} 
                    textAnchor="middle" 
                    fontSize="12" 
                    fill="var(--text-secondary)"
                  >
                    {point.x}
                  </text>
                )}
              </g>
            );
          })}
        </svg>
      </div>
    );
  };

  const renderPieChart = () => {
    if (!chartData.length) return null;

    const total = chartData.reduce((sum, point) => sum + point.value, 0);
    const centerX = 200;
    const centerY = 150;
    const radius = 100;
    let currentAngle = -90;

    const colors = [
      'var(--primary-color)',
      'var(--secondary-color)',
      'var(--accent-color)',
      'var(--info-color)',
      'var(--success-color)',
      'var(--warning-color)'
    ];

    return (
      <div className="chart-container position-relative">
        <svg width="400" height="300" className="advanced-chart">
          {chartData.slice(0, 6).map((point, index) => {
            const percentage = point.value / total;
            const angleSize = percentage * 360;
            const endAngle = currentAngle + angleSize;
            
            const x1 = centerX + radius * Math.cos((currentAngle * Math.PI) / 180);
            const y1 = centerY + radius * Math.sin((currentAngle * Math.PI) / 180);
            const x2 = centerX + radius * Math.cos((endAngle * Math.PI) / 180);
            const y2 = centerY + radius * Math.sin((endAngle * Math.PI) / 180);
            
            const largeArcFlag = angleSize > 180 ? 1 : 0;
            
            const pathData = [
              `M ${centerX} ${centerY}`,
              `L ${x1} ${y1}`,
              `A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2}`,
              'Z'
            ].join(' ');
            
            const result = (
              <g key={index}>
                <path
                  d={pathData}
                  fill={colors[index % colors.length]}
                  className={interactive ? 'chart-pie-interactive' : ''}
                  onMouseEnter={(e) => {
                    if (interactive) {
                      e.target.style.opacity = '0.8';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (interactive) {
                      e.target.style.opacity = '1';
                    }
                  }}
                />
                <text
                  x={centerX + (radius * 0.7) * Math.cos(((currentAngle + angleSize / 2) * Math.PI) / 180)}
                  y={centerY + (radius * 0.7) * Math.sin(((currentAngle + angleSize / 2) * Math.PI) / 180)}
                  textAnchor="middle"
                  fontSize="12"
                  fill="white"
                  fontWeight="bold"
                >
                  {(percentage * 100).toFixed(1)}%
                </text>
              </g>
            );
            
            currentAngle = endAngle;
            return result;
          })}
        </svg>
        
        {/* Legend */}
        <div className="chart-legend mt-3">
          {chartData.slice(0, 6).map((point, index) => (
            <div key={index} className="d-flex align-items-center me-3 mb-2">
              <div 
                className="legend-color me-2" 
                style={{ 
                  width: '12px', 
                  height: '12px', 
                  backgroundColor: colors[index % colors.length],
                  borderRadius: '2px'
                }}
              ></div>
              <small>{point.category || point.x}: {point.value.toFixed(1)}</small>
            </div>
          ))}
        </div>
      </div>
    );
  };

  const renderChart = () => {
    switch (currentChartType) {
      case 'line':
      case 'area':
        return renderLineChart();
      case 'bar':
        return renderBarChart();
      case 'pie':
        return renderPieChart();
      default:
        return renderLineChart();
    }
  };

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
  };

  const handleExport = () => {
    if (onExport) {
      onExport({
        chartType: currentChartType,
        data: chartData,
        title,
        timestamp: new Date().toISOString()
      });
    } else {
      // Default export behavior
      const exportData = {
        chartType: currentChartType,
        data: chartData,
        title,
        timestamp: new Date().toISOString()
      };
      
      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: 'application/json'
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `chart-${title.toLowerCase().replace(/\s+/g, '-')}-${Date.now()}.json`;
      a.click();
      URL.revokeObjectURL(url);
    }
  };

  return (
    <Card className={`advanced-charts ${className} ${isFullscreen ? 'position-fixed' : ''}`} 
          style={isFullscreen ? { top: 0, left: 0, right: 0, bottom: 0, zIndex: 1050 } : {}}>
      <Card.Header className="d-flex justify-content-between align-items-center">
        <h6 className="mb-0">{title}</h6>
        <div className="d-flex align-items-center gap-2">
          {/* Chart Type Selector */}
          <Form.Select 
            value={currentChartType} 
            onChange={(e) => setCurrentChartType(e.target.value)}
            size="sm"
            style={{ width: '150px' }}
          >
            {chartTypes.map(type => (
              <option key={type.value} value={type.value}>
                {type.label}
              </option>
            ))}
          </Form.Select>
          
          {/* Chart Options */}
          <div className="btn-group" role="group">
            <Button 
              variant="outline-secondary" 
              size="sm"
              onClick={() => setGridEnabled(!gridEnabled)}
              className={gridEnabled ? 'active' : ''}
              title="Toggle Grid"
            >
              Grid
            </Button>
            <Button 
              variant="outline-secondary" 
              size="sm"
              onClick={() => setAnimationEnabled(!animationEnabled)}
              className={animationEnabled ? 'active' : ''}
              title="Toggle Animation"
            >
              Anim
            </Button>
            <Button 
              variant="outline-secondary" 
              size="sm"
              onClick={() => setLabelsEnabled(!labelsEnabled)}
              className={labelsEnabled ? 'active' : ''}
              title="Toggle Labels"
            >
              Labels
            </Button>
          </div>
          
          {/* Action Buttons */}
          <Button variant="outline-primary" size="sm" onClick={handleExport}>
            <FontAwesomeIcon icon={faDownload} />
          </Button>
          <Button variant="outline-primary" size="sm" onClick={toggleFullscreen}>
            <FontAwesomeIcon icon={isFullscreen ? faCompress : faExpand} />
          </Button>
        </div>
      </Card.Header>
      <Card.Body className={`p-0 ${isFullscreen ? 'h-100 d-flex flex-column' : ''}`}>
        <div className={`chart-wrapper ${isFullscreen ? 'flex-grow-1 p-4' : 'p-3'}`}>
          {renderChart()}
        </div>
        
        {/* Real-time indicator */}
        {realTime && (
          <div className="position-absolute top-0 end-0 m-3">
            <Badge bg="success" className="pulse">
              <div className="status-indicator status-online me-1"></div>
              LIVE
            </Badge>
          </div>
        )}
      </Card.Body>
    </Card>
  );
};

export default AdvancedCharts;