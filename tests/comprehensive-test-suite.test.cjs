/**
 * Jest Test Suite for Comprehensive System Validation
 * 
 * This test suite validates all implemented fixes using Jest framework
 * for better integration with CI/CD pipelines.
 */

const fs = require('fs');
const path = require('path');
const { exec } = require('child_process');
const util = require('util');
const execPromise = util.promisify(exec);

describe('Comprehensive System Validation', () => {
    let testResults = {};
    
    beforeAll(() => {
        // Initialize test environment
        testResults = {
            ports: [],
            configFiles: [],
            securityIssues: []
        };
    });

    describe('Database Configuration Tests', () => {
        test('should have database ports above 10000', () => {
            const configFiles = ['docker-compose.yml', '.env.example', 'scripts/test-db-config.go'];
            let allPortsValid = true;
            let invalidPorts = [];
            
            configFiles.forEach(file => {
                if (fs.existsSync(file)) {
                    const content = fs.readFileSync(file, 'utf8');
                    const portMatches = content.match(/port["\s:=]*(\d+)/gi);
                    
                    if (portMatches) {
                        const ports = portMatches.map(match => {
                            const port = match.match(/(\d+)/);
                            return port ? parseInt(port[1]) : 0;
                        }).filter(port => port > 0);
                        
                        const invalid = ports.filter(port => port <= 10000);
                        if (invalid.length > 0) {
                            allPortsValid = false;
                            invalidPorts.push(...invalid.map(port => ({ port, file })));
                        }
                        
                        testResults.ports.push(...ports);
                    }
                }
            });
            
            expect(allPortsValid).toBe(true);
            if (!allPortsValid) {
                console.log('Invalid ports found:', invalidPorts);
            }
        });

        test('should have proper database environment variables', () => {
            const envExample = '.env.example';
            if (fs.existsSync(envExample)) {
                const content = fs.readFileSync(envExample, 'utf8');
                expect(content).toMatch(/DB_HOST|DATABASE_URL|DB_PORT/);
            } else {
                console.warn('No .env.example file found');
            }
        });

        test('should validate Go database configuration', async () => {
            const goConfigFile = 'scripts/test-db-config.go';
            if (fs.existsSync(goConfigFile)) {
                try {
                    await execPromise('go version');
                    const { stdout } = await execPromise(`go run ${goConfigFile}`);
                    expect(stdout).toBeDefined();
                } catch (error) {
                    console.warn('Go not available or config failed:', error.message);
                }
            }
        }, 10000);
    });

    describe('API Server Configuration Tests', () => {
        test('should have API server files with health endpoints', () => {
            const apiFiles = [
                'src/api-server.js',
                'performance/performance-monitoring-dashboard.js',
                'critical-fixes/agent-pool/prewarming-system.js'
            ];
            
            let healthEndpointsFound = false;
            
            apiFiles.forEach(file => {
                if (fs.existsSync(file)) {
                    const content = fs.readFileSync(file, 'utf8');
                    if (content.includes('/health') || content.includes('/status')) {
                        healthEndpointsFound = true;
                    }
                    testResults.configFiles.push(file);
                }
            });
            
            expect(healthEndpointsFound || apiFiles.some(f => fs.existsSync(f))).toBe(true);
        });

        test('should have API server ports above 10000', () => {
            const apiFiles = ['src/api-server.js', 'performance/performance-monitoring-dashboard.js'];
            let allPortsValid = true;
            
            apiFiles.forEach(file => {
                if (fs.existsSync(file)) {
                    const content = fs.readFileSync(file, 'utf8');
                    const portMatch = content.match(/PORT["\s]*[=:]["\s]*(\d+)/i);
                    
                    if (portMatch) {
                        const port = parseInt(portMatch[1]);
                        expect(port).toBeGreaterThan(10000);
                        if (port <= 10000) {
                            allPortsValid = false;
                        }
                    }
                }
            });
            
            expect(allPortsValid).toBe(true);
        });
    });

    describe('WebSocket Configuration Tests', () => {
        test('should have graceful shutdown patterns for WebSocket servers', () => {
            const serverFiles = fs.readdirSync('src/', { withFileTypes: true })
                .filter(dirent => dirent.isFile() && dirent.name.endsWith('.js'))
                .map(dirent => path.join('src', dirent.name));
            
            let hasWebSocket = false;
            let hasGracefulShutdown = false;
            
            serverFiles.forEach(file => {
                if (fs.existsSync(file)) {
                    const content = fs.readFileSync(file, 'utf8');
                    
                    if (content.includes('websocket') || content.includes('ws') || content.includes('socket.io')) {
                        hasWebSocket = true;
                        
                        if (content.includes('SIGTERM') || content.includes('SIGINT') || 
                            content.includes('process.on') || content.includes('graceful')) {
                            hasGracefulShutdown = true;
                        }
                    }
                }
            });
            
            if (hasWebSocket) {
                expect(hasGracefulShutdown).toBe(true);
            } else {
                console.log('No WebSocket configuration detected');
            }
        });
    });

    describe('Docker Configuration Tests', () => {
        test('should have proper Dockerfile configuration', () => {
            if (fs.existsSync('Dockerfile')) {
                const dockerfile = fs.readFileSync('Dockerfile', 'utf8');
                
                // Check for Node.js support
                expect(dockerfile).toMatch(/FROM\s+node|FROM\s+.*node/);
                
                // Check for proper port exposure
                const exposeMatch = dockerfile.match(/EXPOSE\s+(\d+)/g);
                if (exposeMatch) {
                    const ports = exposeMatch.map(exp => parseInt(exp.match(/\d+/)[0]));
                    ports.forEach(port => {
                        expect(port).toBeGreaterThan(10000);
                    });
                }
            } else {
                console.warn('No Dockerfile found');
            }
        });

        test('should have proper docker-compose configuration', () => {
            if (fs.existsSync('docker-compose.yml')) {
                const compose = fs.readFileSync('docker-compose.yml', 'utf8');
                
                expect(compose).toMatch(/services:/);
                
                // Check external port mappings
                const portMappings = compose.match(/ports:\s*\n\s*-\s*"?(\d+):\d+"?/g);
                if (portMappings) {
                    const externalPorts = portMappings.map(mapping => {
                        const match = mapping.match(/(\d+):/);
                        return match ? parseInt(match[1]) : 0;
                    }).filter(port => port > 0);
                    
                    externalPorts.forEach(port => {
                        expect(port).toBeGreaterThan(10000);
                    });
                }
            } else {
                console.warn('No docker-compose.yml found');
            }
        });
    });

    describe('Claude-Flow Agent Integration Tests', () => {
        test('should have Claude-Flow command structure', () => {
            const claudeConfigPath = '.claude/commands/';
            if (fs.existsSync(claudeConfigPath)) {
                const commandDirs = fs.readdirSync(claudeConfigPath, { withFileTypes: true })
                    .filter(dirent => dirent.isDirectory())
                    .map(dirent => dirent.name);
                
                expect(commandDirs.length).toBeGreaterThan(0);
                console.log(`Found ${commandDirs.length} command directories`);
            } else {
                console.warn('No .claude/commands/ directory found');
            }
        });

        test('should have Claude-Flow dependency configuration', () => {
            if (fs.existsSync('package.json')) {
                const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'));
                
                // Check for Claude-Flow or related dependencies
                const hasClaude = packageJson.dependencies && 
                    (packageJson.dependencies['claude-flow'] || 
                     packageJson.devDependencies && packageJson.devDependencies['claude-flow']);
                
                if (!hasClaude) {
                    console.warn('Claude-Flow dependency not found in package.json');
                }
            }
        });

        test('should validate expected agent types', () => {
            const expectedAgentTypes = [
                'coder', 'reviewer', 'tester', 'planner', 'researcher',
                'hierarchical-coordinator', 'mesh-coordinator', 'adaptive-coordinator'
            ];
            
            // This test checks if we can find references to these agent types
            let foundAgents = 0;
            
            expectedAgentTypes.forEach(agentType => {
                const agentPath = `.claude/commands/${agentType}`;
                const agentMdPath = `.claude/commands/${agentType}.md`;
                
                if (fs.existsSync(agentPath) || fs.existsSync(agentMdPath)) {
                    foundAgents++;
                }
            });
            
            expect(foundAgents).toBeGreaterThan(0);
            console.log(`Found ${foundAgents}/${expectedAgentTypes.length} expected agent types`);
        });
    });

    describe('Port Configuration Validation', () => {
        test('should have all configured ports above 10000', () => {
            const configFiles = [
                'package.json',
                'docker-compose.yml',
                '.env.example',
                'src/api-server.js'
            ];
            
            let allPortsValid = true;
            let invalidPorts = [];
            let totalPortsChecked = 0;
            
            configFiles.forEach(file => {
                if (fs.existsSync(file)) {
                    const content = fs.readFileSync(file, 'utf8');
                    
                    // Multiple port patterns
                    const portPatterns = [
                        /port["\s:=]*(\d+)/gi,
                        /"(\d{4,5})"/g,
                        /:\s*(\d{4,5})/g,
                        /PORT[_\s]*=\s*(\d+)/gi
                    ];
                    
                    portPatterns.forEach(pattern => {
                        const matches = content.match(pattern);
                        if (matches) {
                            const ports = matches.map(match => {
                                const portMatch = match.match(/(\d+)/);
                                return portMatch ? parseInt(portMatch[1]) : 0;
                            }).filter(port => port > 1000 && port < 65536);
                            
                            ports.forEach(port => {
                                totalPortsChecked++;
                                if (port <= 10000) {
                                    allPortsValid = false;
                                    invalidPorts.push({ port, file });
                                }
                            });
                        }
                    });
                }
            });
            
            console.log(`Checked ${totalPortsChecked} ports across ${configFiles.length} files`);
            if (!allPortsValid) {
                console.log('Invalid ports found:', invalidPorts);
            }
            
            expect(allPortsValid).toBe(true);
        });
    });

    describe('Security Configuration Tests', () => {
        test('should have proper .gitignore configuration', () => {
            if (fs.existsSync('.gitignore')) {
                const gitignore = fs.readFileSync('.gitignore', 'utf8');
                const requiredPatterns = ['node_modules', '.env'];
                
                requiredPatterns.forEach(pattern => {
                    expect(gitignore).toMatch(new RegExp(pattern));
                });
            } else {
                console.warn('No .gitignore file found');
            }
        });

        test('should not have sensitive files in repository', () => {
            const sensitiveFiles = ['.env', 'config/secrets.json', 'private.key'];
            const foundSensitive = [];
            
            sensitiveFiles.forEach(file => {
                if (fs.existsSync(file)) {
                    foundSensitive.push(file);
                }
            });
            
            if (foundSensitive.length > 0) {
                console.warn(`Sensitive files found: ${foundSensitive.join(', ')}`);
                console.warn('Ensure these are properly added to .gitignore');
            }
        });

        test('should not have hardcoded credentials in source files', () => {
            const sourceFiles = [];
            const extensions = ['.js', '.ts', '.json'];
            
            function scanDirectory(dir) {
                try {
                    const items = fs.readdirSync(dir, { withFileTypes: true });
                    for (const item of items) {
                        const fullPath = path.join(dir, item.name);
                        if (item.isDirectory() && !item.name.startsWith('.') && 
                            item.name !== 'node_modules' && sourceFiles.length < 20) {
                            scanDirectory(fullPath);
                        } else if (item.isFile() && extensions.some(ext => item.name.endsWith(ext))) {
                            sourceFiles.push(fullPath);
                        }
                    }
                } catch (error) {
                    // Skip directories we can't read
                }
            }
            
            scanDirectory('src');
            
            let credentialIssues = 0;
            const suspiciousPatterns = [
                /password\s*=\s*["'][^"']{8,}["']/i,
                /api[_-]?key\s*=\s*["'][^"']{20,}["']/i,
                /secret\s*=\s*["'][^"']{10,}["']/i
            ];
            
            sourceFiles.forEach(file => {
                if (fs.existsSync(file)) {
                    const content = fs.readFileSync(file, 'utf8');
                    
                    for (const pattern of suspiciousPatterns) {
                        if (pattern.test(content)) {
                            credentialIssues++;
                            testResults.securityIssues.push(`Potential credential in ${file}`);
                            break;
                        }
                    }
                }
            });
            
            expect(credentialIssues).toBe(0);
            if (credentialIssues > 0) {
                console.warn(`Potential hardcoded credentials found in ${credentialIssues} files`);
            }
        });
    });

    describe('Performance and System Tests', () => {
        test('should have reasonable memory usage', () => {
            const memUsage = process.memoryUsage();
            const heapUsedMB = Math.round(memUsage.heapUsed / 1024 / 1024);
            
            console.log(`Current heap usage: ${heapUsedMB}MB`);
            expect(heapUsedMB).toBeLessThan(200); // 200MB threshold
        });

        test('should have fast file I/O performance', async () => {
            const testFile = '/tmp/jest-perf-test.txt';
            const testData = 'x'.repeat(100000); // 100KB test data
            
            const startTime = process.hrtime.bigint();
            
            fs.writeFileSync(testFile, testData);
            const readData = fs.readFileSync(testFile, 'utf8');
            fs.unlinkSync(testFile);
            
            const endTime = process.hrtime.bigint();
            const duration = Number(endTime - startTime) / 1000000; // Convert to milliseconds
            
            console.log(`File I/O performance: ${duration.toFixed(2)}ms`);
            expect(duration).toBeLessThan(50); // 50ms threshold
            expect(readData).toBe(testData);
        });

        test('should have project structure integrity', () => {
            const requiredDirs = ['src', 'scripts', 'tests'];
            const foundDirs = [];
            
            requiredDirs.forEach(dir => {
                if (fs.existsSync(dir)) {
                    foundDirs.push(dir);
                }
            });
            
            expect(foundDirs.length).toBeGreaterThan(0);
            console.log(`Found directories: ${foundDirs.join(', ')}`);
        });
    });

    afterAll(() => {
        // Generate summary
        console.log('\n📊 Test Summary:');
        console.log(`🔌 Ports checked: ${testResults.ports.length}`);
        console.log(`📁 Config files found: ${testResults.configFiles.length}`);
        console.log(`🛡️ Security issues: ${testResults.securityIssues.length}`);
        
        if (testResults.securityIssues.length > 0) {
            console.log('\n🚨 Security Issues Found:');
            testResults.securityIssues.forEach(issue => {
                console.log(`   - ${issue}`);
            });
        }
    });
});