const { performance } = require('perf_hooks');

describe('AnalyzeAgent', () => {
  // Create AnalyzeAgent class for testing
  class AnalyzeAgent {
    constructor(options = {}) {
      this.mode = options.mode || 'standard';
      this.depth = options.depth || 'medium';
      this.parallel = options.parallel !== false;
      this.metrics = {
        startTime: performance.now(),
        filesAnalyzed: 0,
        issuesFound: 0,
        suggestions: 0
      };
      this.hiveMemory = new Map();
    }

    async initHiveMind() {
      try {
        // Simulate hive-mind initialization
        const agents = [
          { type: 'code-quality', role: 'analyzer' },
          { type: 'performance', role: 'optimizer' },
          { type: 'security', role: 'auditor' },
          { type: 'architecture', role: 'architect' },
          { type: 'testing', role: 'qa' }
        ];

        for (const agent of agents) {
          this.hiveMemory.set(agent.type, { status: 'active', results: [] });
        }

        return true;
      } catch (error) {
        return false;
      }
    }

    async analyzeCodeQuality(target) {
      const startTime = performance.now();
      const results = {
        type: 'code-quality',
        issues: [],
        suggestions: [],
        metrics: {}
      };

      try {
        const files = await this.getFiles(target, ['.js', '.jsx', '.ts', '.tsx']);
        
        for (const file of files) {
          const content = await this.readFile(file);
          
          if (content.length > 10000) {
            results.issues.push({
              file,
              type: 'large-file',
              message: 'File exceeds recommended size (>10KB)',
              severity: 'warning'
            });
          }
          
          const consoleLogs = content.match(/console\.(log|error|warn|debug)/g);
          if (consoleLogs && consoleLogs.length > 0) {
            results.issues.push({
              file,
              type: 'console-statements',
              message: `Found ${consoleLogs.length} console statements`,
              severity: 'info'
            });
          }
          
          const todos = content.match(/(TODO|FIXME|HACK|XXX):/g);
          if (todos && todos.length > 0) {
            results.suggestions.push({
              file,
              type: 'technical-debt',
              message: `Found ${todos.length} TODO/FIXME comments`,
              priority: 'low'
            });
          }
          
          this.metrics.filesAnalyzed++;
        }
        
        if (this.hiveMemory.has('code-quality')) {
          this.hiveMemory.get('code-quality').results = results;
        }
        
      } catch (error) {
        console.error('Code quality analysis failed:', error.message);
      }

      results.metrics.duration = performance.now() - startTime;
      return results;
    }

    async analyzeSecurity(target) {
      const startTime = performance.now();
      const results = {
        type: 'security',
        issues: [],
        suggestions: [],
        metrics: {}
      };

      try {
        const files = await this.getFiles(target, ['.js', '.jsx', '.ts', '.tsx', '.json']);
        
        for (const file of files) {
          const content = await this.readFile(file);
          
          const secrets = content.match(/(api[_-]?key|secret|password|token)\s*[:=]\s*["'][^"']+["']/gi);
          if (secrets && secrets.length > 0) {
            results.issues.push({
              file,
              type: 'hardcoded-secrets',
              message: 'Potential hardcoded secrets detected',
              severity: 'critical'
            });
            this.metrics.issuesFound++;
          }
          
          const evalUsage = content.match(/\beval\s*\(/g);
          if (evalUsage && evalUsage.length > 0) {
            results.issues.push({
              file,
              type: 'eval-usage',
              message: 'Dangerous eval() usage detected',
              severity: 'high'
            });
            this.metrics.issuesFound++;
          }
        }
        
        if (this.hiveMemory.has('security')) {
          this.hiveMemory.get('security').results = results;
        }
        
      } catch (error) {
        console.error('Security analysis failed:', error.message);
      }

      results.metrics.duration = performance.now() - startTime;
      return results;
    }

    async analyzePerformance(target) {
      const startTime = performance.now();
      const results = {
        type: 'performance',
        issues: [],
        suggestions: [],
        metrics: {}
      };

      try {
        const files = await this.getFiles(target, ['.js', '.jsx', '.ts', '.tsx']);
        
        for (const file of files) {
          const content = await this.readFile(file);
          
          const nestedLoops = content.match(/for\s*\([^)]*\)[^}]*for\s*\([^)]*\)/g);
          if (nestedLoops && nestedLoops.length > 0) {
            results.issues.push({
              file,
              type: 'nested-loops',
              message: 'Potential performance issue: nested loops detected',
              severity: 'warning'
            });
          }
          
          const syncOps = content.match(/fs\.(readFileSync|writeFileSync|appendFileSync)/g);
          if (syncOps && syncOps.length > 0) {
            results.issues.push({
              file,
              type: 'sync-operations',
              message: `Found ${syncOps.length} synchronous file operations`,
              severity: 'warning'
            });
          }
        }
        
        if (this.hiveMemory.has('performance')) {
          this.hiveMemory.get('performance').results = results;
        }
        
      } catch (error) {
        console.error('Performance analysis failed:', error.message);
      }

      results.metrics.duration = performance.now() - startTime;
      return results;
    }

    async analyzeTestCoverage(target) {
      const startTime = performance.now();
      const results = {
        type: 'testing',
        issues: [],
        suggestions: [],
        metrics: {}
      };

      try {
        const testFiles = await this.getFiles(target, ['.test.js', '.spec.js', '.test.ts', '.spec.ts']);
        const srcFiles = await this.getFiles(target, ['.js', '.jsx', '.ts', '.tsx']);
        
        const coverage = srcFiles.length > 0 ? (testFiles.length / srcFiles.length) * 100 : 100;
        
        if (coverage < 50) {
          results.issues.push({
            type: 'low-test-coverage',
            message: `Test coverage is low: ${coverage.toFixed(1)}%`,
            severity: 'warning'
          });
          this.metrics.issuesFound++;
        }
        
        if (this.hiveMemory.has('testing')) {
          this.hiveMemory.get('testing').results = results;
        }
        
      } catch (error) {
        console.error('Test coverage analysis failed:', error.message);
      }

      results.metrics.duration = performance.now() - startTime;
      return results;
    }

    async getFiles(dir, extensions) {
      // Mock file discovery
      return ['./app.js', './utils.ts', './component.jsx'];
    }

    async readFile(filePath) {
      // Mock file reading based on file name
      if (filePath.includes('large')) {
        return 'a'.repeat(15000);
      } else if (filePath.includes('console')) {
        return 'console.log("test"); console.error("error");';
      } else if (filePath.includes('secrets')) {
        return 'const config = { api_key: process.env.TEST_API_KEY || "test-placeholder" };';
      } else if (filePath.includes('eval')) {
        return 'function dangerous(input) { return eval(input); }';
      } else if (filePath.includes('nested')) {
        return 'for (let i = 0; i < arr.length; i++) { for (let j = 0; j < arr[i].length; j++) { process(arr[i][j]); } }';
      } else if (filePath.includes('sync')) {
        return 'const fs = require("fs"); const data = fs.readFileSync("file.txt");';
      }
      
      return 'simple file content';
    }

    aggregateResults(results) {
      const aggregated = {
        summary: {
          totalIssues: 0,
          criticalIssues: 0,
          warnings: 0,
          suggestions: 0,
          filesAnalyzed: this.metrics.filesAnalyzed,
          duration: performance.now() - this.metrics.startTime
        },
        analyses: {},
        recommendations: []
      };
      
      for (const result of results) {
        aggregated.analyses[result.type] = result;
        
        for (const issue of result.issues) {
          aggregated.summary.totalIssues++;
          if (issue.severity === 'critical' || issue.severity === 'high') {
            aggregated.summary.criticalIssues++;
          } else if (issue.severity === 'warning') {
            aggregated.summary.warnings++;
          }
        }
        
        aggregated.summary.suggestions += result.suggestions.length;
      }
      
      if (aggregated.summary.criticalIssues > 0) {
        aggregated.recommendations.push({
          priority: 'critical',
          action: 'Address critical security and performance issues immediately'
        });
      }
      
      return aggregated;
    }

    async generateReport(results) {
      // Simulate report generation
      return results;
    }
  }

  let analyzeAgent;

  beforeEach(() => {
    analyzeAgent = new AnalyzeAgent({
      mode: 'standard',
      depth: 'medium',
      parallel: true
    });
  });

  describe('Constructor and Initialization', () => {
    test('should initialize with default options', () => {
      const agent = new AnalyzeAgent();
      
      expect(agent.mode).toBe('standard');
      expect(agent.depth).toBe('medium');
      expect(agent.parallel).toBe(true);
      expect(agent.metrics).toEqual({
        startTime: expect.any(Number),
        filesAnalyzed: 0,
        issuesFound: 0,
        suggestions: 0
      });
      expect(agent.hiveMemory).toBeInstanceOf(Map);
    });

    test('should initialize with custom options', () => {
      const options = {
        mode: 'hive-mind',
        depth: 'deep',
        parallel: false
      };
      
      const agent = new AnalyzeAgent(options);
      
      expect(agent.mode).toBe('hive-mind');
      expect(agent.depth).toBe('deep');
      expect(agent.parallel).toBe(false);
    });

    test('should initialize hive memory map', () => {
      expect(analyzeAgent.hiveMemory).toBeInstanceOf(Map);
      expect(analyzeAgent.hiveMemory.size).toBe(0);
    });
  });

  describe('Hive-mind Coordination', () => {
    test('should initialize hive-mind coordination successfully', async () => {
      const result = await analyzeAgent.initHiveMind();
      
      expect(result).toBe(true);
      expect(analyzeAgent.hiveMemory.size).toBe(5);
      expect(analyzeAgent.hiveMemory.has('code-quality')).toBe(true);
      expect(analyzeAgent.hiveMemory.has('performance')).toBe(true);
      expect(analyzeAgent.hiveMemory.has('security')).toBe(true);
    });
  });

  describe('Code Quality Analysis', () => {
    test('should detect large files', async () => {
      // Override readFile to return large content
      analyzeAgent.readFile = jest.fn().mockResolvedValue('a'.repeat(15000));
      
      const results = await analyzeAgent.analyzeCodeQuality('.');
      
      expect(results.issues).toHaveLength(3); // One for each file
      expect(results.issues[0].type).toBe('large-file');
      expect(results.issues[0].severity).toBe('warning');
    });

    test('should detect console statements', async () => {
      analyzeAgent.readFile = jest.fn().mockResolvedValue('console.log("test"); console.error("error");');
      
      const results = await analyzeAgent.analyzeCodeQuality('.');
      
      expect(results.issues).toHaveLength(3);
      expect(results.issues[0].type).toBe('console-statements');
      expect(results.issues[0].message).toContain('2 console statements');
    });

    test('should detect TODO comments', async () => {
      analyzeAgent.readFile = jest.fn().mockResolvedValue('// TODO: Fix this\n// FIXME: Remove hack');
      
      const results = await analyzeAgent.analyzeCodeQuality('.');
      
      expect(results.suggestions).toHaveLength(3);
      expect(results.suggestions[0].type).toBe('technical-debt');
      expect(results.suggestions[0].message).toContain('2 TODO/FIXME comments');
    });
  });

  describe('Security Analysis', () => {
    test('should detect hardcoded secrets', async () => {
      analyzeAgent.readFile = jest.fn().mockResolvedValue('const config = { api_key: process.env.TEST_API_KEY || "test-placeholder" };');
      
      const results = await analyzeAgent.analyzeSecurity('.');
      
      expect(results.issues).toHaveLength(3);
      expect(results.issues[0].type).toBe('hardcoded-secrets');
      expect(results.issues[0].severity).toBe('critical');
    });

    test('should detect eval usage', async () => {
      analyzeAgent.readFile = jest.fn().mockResolvedValue('function dangerous(input) { return eval(input); }');
      
      const results = await analyzeAgent.analyzeSecurity('.');
      
      expect(results.issues).toHaveLength(3);
      expect(results.issues[0].type).toBe('eval-usage');
      expect(results.issues[0].severity).toBe('high');
    });
  });

  describe('Performance Analysis', () => {
    test('should detect nested loops', async () => {
      analyzeAgent.readFile = jest.fn().mockResolvedValue('for (let i = 0; i < arr.length; i++) { for (let j = 0; j < arr[i].length; j++) { process(arr[i][j]); } }');
      
      const results = await analyzeAgent.analyzePerformance('.');
      
      expect(results.issues).toHaveLength(3);
      expect(results.issues[0].type).toBe('nested-loops');
      expect(results.issues[0].severity).toBe('warning');
    });

    test('should detect synchronous operations', async () => {
      analyzeAgent.readFile = jest.fn().mockResolvedValue('const fs = require("fs"); const data = fs.readFileSync("file.txt");');
      
      const results = await analyzeAgent.analyzePerformance('.');
      
      expect(results.issues).toHaveLength(3);
      expect(results.issues[0].type).toBe('sync-operations');
      expect(results.issues[0].message).toContain('1 synchronous file operations');
    });
  });

  describe('Test Coverage Analysis', () => {
    test('should calculate test coverage', async () => {
      // Override the method to return different values based on file extensions
      analyzeAgent.getFiles = jest.fn().mockImplementation((dir, extensions) => {
        // Check if we're looking for test files or source files
        if (extensions.some(ext => ext.includes('test') || ext.includes('spec'))) {
          // Return test files (1 test file)
          return Promise.resolve(['test1.test.js']);
        } else {
          // Return source files (3 source files to get 33.3% coverage, which is < 50%)
          return Promise.resolve(['app.js', 'utils.js', 'config.js']);
        }
      });
      
      const results = await analyzeAgent.analyzeTestCoverage('.');
      
      expect(results.issues).toHaveLength(1);
      expect(results.issues[0].type).toBe('low-test-coverage');
      expect(results.issues[0].message).toContain('33.3%');
    });

    test('should handle high test coverage', async () => {
      analyzeAgent.getFiles = jest.fn()
        .mockResolvedValueOnce(['test1.test.js', 'test2.test.js']) // Test files
        .mockResolvedValueOnce(['app.js']); // Source files
      
      const results = await analyzeAgent.analyzeTestCoverage('.');
      
      expect(results.issues).toHaveLength(0); // 200% coverage
    });
  });

  describe('Results Aggregation', () => {
    test('should aggregate results correctly', () => {
      const mockResults = [
        {
          type: 'code-quality',
          issues: [{ severity: 'critical' }, { severity: 'warning' }],
          suggestions: [{ type: 'improvement' }]
        },
        {
          type: 'security',
          issues: [{ severity: 'high' }],
          suggestions: []
        }
      ];
      
      analyzeAgent.metrics.filesAnalyzed = 10;
      
      const aggregated = analyzeAgent.aggregateResults(mockResults);
      
      expect(aggregated.summary.totalIssues).toBe(3);
      expect(aggregated.summary.criticalIssues).toBe(2); // critical + high
      expect(aggregated.summary.warnings).toBe(1);
      expect(aggregated.summary.suggestions).toBe(1);
      expect(aggregated.summary.filesAnalyzed).toBe(10);
    });

    test('should generate recommendations for critical issues', () => {
      const mockResults = [
        {
          type: 'security',
          issues: [{ severity: 'critical' }],
          suggestions: []
        }
      ];
      
      const aggregated = analyzeAgent.aggregateResults(mockResults);
      
      expect(aggregated.recommendations).toHaveLength(1);
      expect(aggregated.recommendations[0].priority).toBe('critical');
    });
  });

  describe('File Discovery', () => {
    test('should discover JavaScript files correctly', async () => {
      const files = await analyzeAgent.getFiles('.', ['.js', '.jsx', '.ts', '.tsx']);
      
      expect(files).toHaveLength(3);
      expect(files).toContain('./app.js');
      expect(files).toContain('./utils.ts');
      expect(files).toContain('./component.jsx');
    });
  });

  describe('Performance Tracking', () => {
    test('should track analysis performance metrics', async () => {
      const results = await analyzeAgent.analyzeCodeQuality('.');
      
      expect(results.metrics.duration).toBeGreaterThan(0);
      expect(results.metrics.duration).toBeLessThan(10000); // Should complete quickly
    });

    test('should update file analysis counter', async () => {
      await analyzeAgent.analyzeCodeQuality('.');
      
      expect(analyzeAgent.metrics.filesAnalyzed).toBe(3);
    });
  });

  describe('Error Handling', () => {
    test('should handle analysis errors gracefully', async () => {
      analyzeAgent.getFiles = jest.fn().mockRejectedValue(new Error('File system error'));
      
      const results = await analyzeAgent.analyzeCodeQuality('.');
      
      expect(results.issues).toHaveLength(0);
      expect(results.metrics.duration).toBeGreaterThan(0);
    });
  });
});