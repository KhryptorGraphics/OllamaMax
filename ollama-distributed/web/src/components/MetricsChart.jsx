import React, { useEffect, useRef } from 'react';
import Chart from 'chart.js/auto';

const MetricsChart = ({ data, title, color, type = 'line' }) => {
  const chartRef = useRef(null);
  const chartInstance = useRef(null);

  useEffect(() => {
    const ctx = chartRef.current.getContext('2d');
    
    // Destroy existing chart
    if (chartInstance.current) {
      chartInstance.current.destroy();
    }

    const chartData = data || [];
    const labels = chartData.map((_, index) => `${index * 5}s`);
    const values = chartData.map(item => item.value || 0);

    chartInstance.current = new Chart(ctx, {
      type: type,
      data: {
        labels: labels,
        datasets: [{
          label: title,
          data: values,
          borderColor: color,
          backgroundColor: `${color}20`,
          borderWidth: 2,
          fill: true,
          tension: 0.4,
          pointRadius: 0,
          pointHoverRadius: 6,
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            display: false
          }
        },
        scales: {
          x: {
            display: false,
            grid: {
              display: false
            }
          },
          y: {
            beginAtZero: true,
            max: 100,
            grid: {
              color: 'rgba(255,255,255,0.1)'
            },
            ticks: {
              color: 'rgba(255,255,255,0.8)',
              font: {
                size: 10
              }
            }
          }
        },
        interaction: {
          intersect: false,
          mode: 'index'
        },
        animation: {
          duration: 750,
          easing: 'easeInOutCubic'
        }
      }
    });

    return () => {
      if (chartInstance.current) {
        chartInstance.current.destroy();
      }
    };
  }, [data, title, color, type]);

  return (
    <div className="chart-wrapper" style={{ height: '200px', position: 'relative' }}>
      <canvas ref={chartRef}></canvas>
    </div>
  );
};

export default MetricsChart;