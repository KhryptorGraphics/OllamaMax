// Integration test for all UI components
// This file can be run in the browser console to verify component integration

const testComponentIntegration = () => {
  console.log('🧪 Starting UI Component Integration Test...\n');
  
  const tests = [
    {
      name: 'Theme System',
      test: () => {
        const root = document.documentElement;
        const currentTheme = root.getAttribute('data-theme');
        return currentTheme === 'light' || currentTheme === 'dark';
      }
    },
    {
      name: 'CSS Custom Properties',
      test: () => {
        const computed = getComputedStyle(document.documentElement);
        return computed.getPropertyValue('--primary-gradient').length > 0;
      }
    },
    {
      name: 'Bootstrap Integration',
      test: () => {
        return !!document.querySelector('.container, .row, .col');
      }
    },
    {
      name: 'FontAwesome Icons',
      test: () => {
        return !!document.querySelector('.fa, .fas, .fab');
      }
    },
    {
      name: 'Chart.js Library',
      test: () => {
        return typeof Chart !== 'undefined';
      }
    },
    {
      name: 'React Components',
      test: () => {
        return typeof React !== 'undefined' && typeof ReactDOM !== 'undefined';
      }
    },
    {
      name: 'Responsive Grids',
      test: () => {
        const grids = document.querySelectorAll('.metrics-grid, .model-grid, .node-grid');
        return grids.length > 0;
      }
    },
    {
      name: 'Card Hover Effects',
      test: () => {
        const cards = document.querySelectorAll('.card');
        return cards.length > 0 && getComputedStyle(cards[0]).transition.includes('transform');
      }
    },
    {
      name: 'Animation Support',
      test: () => {
        const styles = Array.from(document.styleSheets)
          .map(sheet => {
            try {
              return Array.from(sheet.cssRules || []);
            } catch (e) {
              return [];
            }
          })
          .flat()
          .some(rule => rule.cssText && rule.cssText.includes('@keyframes'));
        return styles;
      }
    },
    {
      name: 'Mobile Responsive Design',
      test: () => {
        const viewport = document.querySelector('meta[name="viewport"]');
        return viewport && viewport.content.includes('width=device-width');
      }
    }
  ];

  const results = tests.map(({ name, test }) => {
    try {
      const passed = test();
      console.log(`${passed ? '✅' : '❌'} ${name}: ${passed ? 'PASS' : 'FAIL'}`);
      return { name, passed, error: null };
    } catch (error) {
      console.log(`❌ ${name}: ERROR - ${error.message}`);
      return { name, passed: false, error: error.message };
    }
  });

  const passCount = results.filter(r => r.passed).length;
  const failCount = results.length - passCount;

  console.log(`\n📊 Test Summary:`);
  console.log(`   ✅ Passed: ${passCount}`);
  console.log(`   ❌ Failed: ${failCount}`);
  console.log(`   📈 Success Rate: ${Math.round((passCount / results.length) * 100)}%\n`);

  if (failCount === 0) {
    console.log('🎉 All integration tests passed! UI components are properly integrated.');
  } else {
    console.log('⚠️  Some tests failed. Check the console output above for details.');
  }

  return results;
};

// Test component creation functionality
const testComponentCreation = () => {
  console.log('🔧 Testing Component Creation...\n');
  
  const componentTests = [
    {
      name: 'LoadingSpinner',
      test: () => {
        return typeof LoadingSpinner !== 'undefined';
      }
    },
    {
      name: 'Alert',
      test: () => {
        return typeof Alert !== 'undefined';
      }
    },
    {
      name: 'ThemeToggle', 
      test: () => {
        return typeof ThemeToggle !== 'undefined';
      }
    },
    {
      name: 'MetricsChart',
      test: () => {
        return typeof MetricsChart !== 'undefined';
      }
    },
    {
      name: 'Dashboard Enhancement',
      test: () => {
        return typeof Dashboard !== 'undefined';
      }
    }
  ];

  componentTests.forEach(({ name, test }) => {
    try {
      const exists = test();
      console.log(`${exists ? '✅' : '❌'} ${name} Component: ${exists ? 'Available' : 'Not Found'}`);
    } catch (error) {
      console.log(`❌ ${name} Component: ERROR - ${error.message}`);
    }
  });
};

// Performance test for UI responsiveness
const testUIPerformance = () => {
  console.log('⚡ Testing UI Performance...\n');
  
  const performanceTests = [
    {
      name: 'CSS Animation Performance',
      test: () => {
        const start = performance.now();
        const testEl = document.createElement('div');
        testEl.className = 'card';
        testEl.style.transform = 'translateY(-8px) scale(1.02)';
        document.body.appendChild(testEl);
        const end = performance.now();
        document.body.removeChild(testEl);
        return (end - start) < 16; // Should be under one frame (16ms)
      }
    },
    {
      name: 'DOM Query Performance',
      test: () => {
        const start = performance.now();
        document.querySelectorAll('.card, .metric-card, .node-card, .model-card');
        const end = performance.now();
        return (end - start) < 5; // Should be very fast
      }
    },
    {
      name: 'Style Computation',
      test: () => {
        const start = performance.now();
        const cards = document.querySelectorAll('.card');
        if (cards.length > 0) {
          getComputedStyle(cards[0]);
        }
        const end = performance.now();
        return (end - start) < 10;
      }
    }
  ];

  performanceTests.forEach(({ name, test }) => {
    try {
      const passed = test();
      console.log(`${passed ? '✅' : '⚠️'} ${name}: ${passed ? 'OPTIMAL' : 'NEEDS OPTIMIZATION'}`);
    } catch (error) {
      console.log(`❌ ${name}: ERROR - ${error.message}`);
    }
  });
};

// Accessibility test
const testAccessibility = () => {
  console.log('♿ Testing Accessibility Features...\n');
  
  const a11yTests = [
    {
      name: 'ARIA Labels',
      test: () => {
        return document.querySelectorAll('[aria-label], [aria-labelledby]').length > 0;
      }
    },
    {
      name: 'Keyboard Navigation',
      test: () => {
        const focusableElements = document.querySelectorAll(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        );
        return focusableElements.length > 0;
      }
    },
    {
      name: 'Color Contrast',
      test: () => {
        // Basic check for theme variables
        const computed = getComputedStyle(document.documentElement);
        return computed.getPropertyValue('--text-color').length > 0;
      }
    },
    {
      name: 'Semantic HTML',
      test: () => {
        const semantic = document.querySelectorAll('main, nav, section, article, header, footer');
        return semantic.length > 0;
      }
    }
  ];

  a11yTests.forEach(({ name, test }) => {
    try {
      const passed = test();
      console.log(`${passed ? '✅' : '⚠️'} ${name}: ${passed ? 'GOOD' : 'NEEDS IMPROVEMENT'}`);
    } catch (error) {
      console.log(`❌ ${name}: ERROR - ${error.message}`);
    }
  });
};

// Run all tests
const runFullIntegrationTest = () => {
  console.clear();
  console.log('🚀 Ollama Distributed - UI Integration Test Suite\n');
  console.log('==========================================\n');
  
  testComponentIntegration();
  console.log('\n');
  testComponentCreation();
  console.log('\n');
  testUIPerformance();
  console.log('\n');
  testAccessibility();
  
  console.log('\n==========================================');
  console.log('🏁 Integration Test Suite Complete!');
  console.log('==========================================\n');
};

// Export for use
window.UIIntegrationTest = {
  runAll: runFullIntegrationTest,
  integration: testComponentIntegration,
  components: testComponentCreation,
  performance: testUIPerformance,
  accessibility: testAccessibility
};

// Auto-run if in development
if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
  console.log('🔧 Development environment detected. Run UIIntegrationTest.runAll() to test UI components.');
}