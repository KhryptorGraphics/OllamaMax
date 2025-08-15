#!/usr/bin/env node

/**
 * Security Audit Script
 * 
 * Comprehensive security audit for the OllamaMax web application.
 * Checks for vulnerabilities, security best practices, and compliance.
 */

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const { execSync } = require('child_process');

class SecurityAuditor {
  constructor() {
    this.findings = [];
    this.severity = {
      CRITICAL: 'CRITICAL',
      HIGH: 'HIGH',
      MEDIUM: 'MEDIUM',
      LOW: 'LOW',
      INFO: 'INFO',
    };
  }

  // Add finding
  addFinding(severity, category, title, description, file = null, line = null) {
    this.findings.push({
      severity,
      category,
      title,
      description,
      file,
      line,
      timestamp: new Date().toISOString(),
    });
  }

  // Audit dependencies
  auditDependencies() {
    console.log('🔍 Auditing dependencies...');
    
    try {
      // Run npm audit
      const auditResult = execSync('npm audit --json', { encoding: 'utf8' });
      const audit = JSON.parse(auditResult);
      
      if (audit.vulnerabilities) {
        Object.entries(audit.vulnerabilities).forEach(([pkg, vuln]) => {
          const severity = vuln.severity.toUpperCase();
          this.addFinding(
            this.severity[severity] || this.severity.MEDIUM,
            'Dependencies',
            `Vulnerable dependency: ${pkg}`,
            `${vuln.title} - ${vuln.url}`,
            'package.json'
          );
        });
      }
    } catch (error) {
      this.addFinding(
        this.severity.HIGH,
        'Dependencies',
        'Failed to run dependency audit',
        'Could not execute npm audit. Ensure npm is installed and package.json exists.'
      );
    }
  }

  // Audit source code
  auditSourceCode() {
    console.log('🔍 Auditing source code...');
    
    const srcDir = path.join(process.cwd(), 'src');
    this.scanDirectory(srcDir);
  }

  // Scan directory recursively
  scanDirectory(dir) {
    if (!fs.existsSync(dir)) return;
    
    const files = fs.readdirSync(dir);
    
    files.forEach(file => {
      const filePath = path.join(dir, file);
      const stat = fs.statSync(filePath);
      
      if (stat.isDirectory()) {
        this.scanDirectory(filePath);
      } else if (file.endsWith('.js') || file.endsWith('.jsx') || file.endsWith('.ts') || file.endsWith('.tsx')) {
        this.scanFile(filePath);
      }
    });
  }

  // Scan individual file
  scanFile(filePath) {
    const content = fs.readFileSync(filePath, 'utf8');
    const lines = content.split('\n');
    
    lines.forEach((line, index) => {
      const lineNumber = index + 1;
      
      // Check for dangerous patterns
      this.checkDangerousPatterns(line, filePath, lineNumber);
      
      // Check for hardcoded secrets
      this.checkHardcodedSecrets(line, filePath, lineNumber);
      
      // Check for insecure practices
      this.checkInsecurePractices(line, filePath, lineNumber);
      
      // Check for XSS vulnerabilities
      this.checkXSSVulnerabilities(line, filePath, lineNumber);
    });
  }

  // Check for dangerous patterns
  checkDangerousPatterns(line, file, lineNumber) {
    const dangerousPatterns = [
      {
        pattern: /eval\s*\(/,
        title: 'Use of eval()',
        description: 'eval() can execute arbitrary code and is a security risk',
        severity: this.severity.HIGH,
      },
      {
        pattern: /innerHTML\s*=/,
        title: 'Direct innerHTML assignment',
        description: 'Direct innerHTML assignment can lead to XSS vulnerabilities',
        severity: this.severity.MEDIUM,
      },
      {
        pattern: /document\.write\s*\(/,
        title: 'Use of document.write()',
        description: 'document.write() can be exploited for XSS attacks',
        severity: this.severity.MEDIUM,
      },
      {
        pattern: /window\.location\s*=\s*[^"']*[+]/,
        title: 'Dynamic location assignment',
        description: 'Dynamic window.location assignment can lead to open redirects',
        severity: this.severity.MEDIUM,
      },
      {
        pattern: /setTimeout\s*\(\s*["'][^"']*["']/,
        title: 'setTimeout with string',
        description: 'Using setTimeout with strings can execute arbitrary code',
        severity: this.severity.HIGH,
      },
    ];

    dangerousPatterns.forEach(({ pattern, title, description, severity }) => {
      if (pattern.test(line)) {
        this.addFinding(severity, 'Code Security', title, description, file, lineNumber);
      }
    });
  }

  // Check for hardcoded secrets
  checkHardcodedSecrets(line, file, lineNumber) {
    const secretPatterns = [
      {
        pattern: /(?:password|pwd|pass)\s*[:=]\s*["'][^"']{8,}["']/i,
        title: 'Hardcoded password',
        description: 'Password appears to be hardcoded in source code',
      },
      {
        pattern: /(?:api[_-]?key|apikey)\s*[:=]\s*["'][^"']{16,}["']/i,
        title: 'Hardcoded API key',
        description: 'API key appears to be hardcoded in source code',
      },
      {
        pattern: /(?:secret|token)\s*[:=]\s*["'][^"']{16,}["']/i,
        title: 'Hardcoded secret/token',
        description: 'Secret or token appears to be hardcoded in source code',
      },
      {
        pattern: /(?:private[_-]?key|privatekey)\s*[:=]\s*["'][^"']{32,}["']/i,
        title: 'Hardcoded private key',
        description: 'Private key appears to be hardcoded in source code',
      },
    ];

    secretPatterns.forEach(({ pattern, title, description }) => {
      if (pattern.test(line)) {
        this.addFinding(this.severity.CRITICAL, 'Secrets', title, description, file, lineNumber);
      }
    });
  }

  // Check for insecure practices
  checkInsecurePractices(line, file, lineNumber) {
    const insecurePatterns = [
      {
        pattern: /http:\/\/(?!localhost|127\.0\.0\.1)/,
        title: 'Insecure HTTP URL',
        description: 'HTTP URLs should be replaced with HTTPS for security',
        severity: this.severity.MEDIUM,
      },
      {
        pattern: /localStorage\.setItem\s*\(\s*["'][^"']*(?:token|password|secret)/i,
        title: 'Sensitive data in localStorage',
        description: 'Sensitive data should not be stored in localStorage',
        severity: this.severity.HIGH,
      },
      {
        pattern: /console\.log\s*\([^)]*(?:password|token|secret|key)/i,
        title: 'Sensitive data in console.log',
        description: 'Sensitive data should not be logged to console',
        severity: this.severity.MEDIUM,
      },
      {
        pattern: /Math\.random\s*\(\s*\)/,
        title: 'Use of Math.random() for security',
        description: 'Math.random() is not cryptographically secure',
        severity: this.severity.LOW,
      },
    ];

    insecurePatterns.forEach(({ pattern, title, description, severity }) => {
      if (pattern.test(line)) {
        this.addFinding(severity, 'Insecure Practices', title, description, file, lineNumber);
      }
    });
  }

  // Check for XSS vulnerabilities
  checkXSSVulnerabilities(line, file, lineNumber) {
    const xssPatterns = [
      {
        pattern: /dangerouslySetInnerHTML/,
        title: 'Use of dangerouslySetInnerHTML',
        description: 'Ensure content is properly sanitized before using dangerouslySetInnerHTML',
        severity: this.severity.HIGH,
      },
      {
        pattern: /\.innerHTML\s*=\s*[^"']*\+/,
        title: 'Dynamic innerHTML assignment',
        description: 'Dynamic innerHTML assignment can lead to XSS if not properly sanitized',
        severity: this.severity.HIGH,
      },
      {
        pattern: /href\s*=\s*["']javascript:/,
        title: 'JavaScript URL in href',
        description: 'JavaScript URLs in href attributes can be exploited for XSS',
        severity: this.severity.HIGH,
      },
    ];

    xssPatterns.forEach(({ pattern, title, description, severity }) => {
      if (pattern.test(line)) {
        this.addFinding(severity, 'XSS', title, description, file, lineNumber);
      }
    });
  }

  // Audit configuration files
  auditConfiguration() {
    console.log('🔍 Auditing configuration...');
    
    // Check package.json
    this.auditPackageJson();
    
    // Check environment files
    this.auditEnvironmentFiles();
    
    // Check build configuration
    this.auditBuildConfig();
  }

  // Audit package.json
  auditPackageJson() {
    const packagePath = path.join(process.cwd(), 'package.json');
    
    if (!fs.existsSync(packagePath)) {
      this.addFinding(
        this.severity.HIGH,
        'Configuration',
        'Missing package.json',
        'package.json file not found'
      );
      return;
    }

    const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
    
    // Check for security-related scripts
    if (!packageJson.scripts || !packageJson.scripts.audit) {
      this.addFinding(
        this.severity.LOW,
        'Configuration',
        'Missing audit script',
        'Consider adding npm audit script to package.json',
        'package.json'
      );
    }

    // Check for outdated dependencies
    if (packageJson.dependencies) {
      Object.entries(packageJson.dependencies).forEach(([pkg, version]) => {
        if (version.includes('^') || version.includes('~')) {
          // This is good - allows patch updates
        } else if (version === '*' || version === 'latest') {
          this.addFinding(
            this.severity.MEDIUM,
            'Configuration',
            'Unpinned dependency version',
            `Dependency ${pkg} uses unpinned version: ${version}`,
            'package.json'
          );
        }
      });
    }
  }

  // Audit environment files
  auditEnvironmentFiles() {
    const envFiles = ['.env', '.env.local', '.env.development', '.env.production'];
    
    envFiles.forEach(envFile => {
      const envPath = path.join(process.cwd(), envFile);
      
      if (fs.existsSync(envPath)) {
        const content = fs.readFileSync(envPath, 'utf8');
        
        // Check if .env files are in .gitignore
        const gitignorePath = path.join(process.cwd(), '.gitignore');
        if (fs.existsSync(gitignorePath)) {
          const gitignore = fs.readFileSync(gitignorePath, 'utf8');
          if (!gitignore.includes('.env')) {
            this.addFinding(
              this.severity.HIGH,
              'Configuration',
              'Environment files not in .gitignore',
              'Environment files should be added to .gitignore to prevent committing secrets',
              '.gitignore'
            );
          }
        }

        // Check for weak secrets
        const lines = content.split('\n');
        lines.forEach((line, index) => {
          if (line.includes('=')) {
            const [key, value] = line.split('=');
            if (key.toLowerCase().includes('secret') || key.toLowerCase().includes('key')) {
              if (value.length < 16) {
                this.addFinding(
                  this.severity.MEDIUM,
                  'Configuration',
                  'Weak secret in environment file',
                  `${key} appears to have a weak value (less than 16 characters)`,
                  envFile,
                  index + 1
                );
              }
            }
          }
        });
      }
    });
  }

  // Audit build configuration
  auditBuildConfig() {
    // Check for source maps in production
    const buildFiles = ['webpack.config.js', 'vite.config.js', 'next.config.js'];
    
    buildFiles.forEach(configFile => {
      const configPath = path.join(process.cwd(), configFile);
      
      if (fs.existsSync(configPath)) {
        const content = fs.readFileSync(configPath, 'utf8');
        
        if (content.includes('devtool') && content.includes('source-map')) {
          this.addFinding(
            this.severity.MEDIUM,
            'Configuration',
            'Source maps in production',
            'Ensure source maps are disabled in production builds',
            configFile
          );
        }
      }
    });
  }

  // Generate report
  generateReport() {
    console.log('\n📊 Security Audit Report');
    console.log('========================\n');

    const severityCounts = {
      [this.severity.CRITICAL]: 0,
      [this.severity.HIGH]: 0,
      [this.severity.MEDIUM]: 0,
      [this.severity.LOW]: 0,
      [this.severity.INFO]: 0,
    };

    this.findings.forEach(finding => {
      severityCounts[finding.severity]++;
    });

    // Summary
    console.log('Summary:');
    Object.entries(severityCounts).forEach(([severity, count]) => {
      if (count > 0) {
        const emoji = {
          [this.severity.CRITICAL]: '🔴',
          [this.severity.HIGH]: '🟠',
          [this.severity.MEDIUM]: '🟡',
          [this.severity.LOW]: '🔵',
          [this.severity.INFO]: '⚪',
        }[severity];
        
        console.log(`${emoji} ${severity}: ${count}`);
      }
    });

    console.log(`\nTotal findings: ${this.findings.length}\n`);

    // Detailed findings
    if (this.findings.length > 0) {
      console.log('Detailed Findings:');
      console.log('==================\n');

      this.findings.forEach((finding, index) => {
        const emoji = {
          [this.severity.CRITICAL]: '🔴',
          [this.severity.HIGH]: '🟠',
          [this.severity.MEDIUM]: '🟡',
          [this.severity.LOW]: '🔵',
          [this.severity.INFO]: '⚪',
        }[finding.severity];

        console.log(`${index + 1}. ${emoji} [${finding.severity}] ${finding.title}`);
        console.log(`   Category: ${finding.category}`);
        console.log(`   Description: ${finding.description}`);
        
        if (finding.file) {
          console.log(`   File: ${finding.file}${finding.line ? `:${finding.line}` : ''}`);
        }
        
        console.log('');
      });
    } else {
      console.log('✅ No security issues found!');
    }

    // Recommendations
    this.generateRecommendations();

    return {
      summary: severityCounts,
      findings: this.findings,
      totalFindings: this.findings.length,
    };
  }

  // Generate recommendations
  generateRecommendations() {
    console.log('Recommendations:');
    console.log('================\n');

    const recommendations = [
      '1. 🔒 Implement Content Security Policy (CSP) headers',
      '2. 🛡️ Use HTTPS for all external resources',
      '3. 🔐 Store sensitive data in secure storage (not localStorage)',
      '4. 🧹 Regularly update dependencies to patch vulnerabilities',
      '5. 🔍 Use static analysis tools in CI/CD pipeline',
      '6. 🚫 Never commit secrets to version control',
      '7. 🔑 Use environment variables for configuration',
      '8. 🧪 Implement security testing in your test suite',
      '9. 📝 Regular security audits and penetration testing',
      '10. 🎯 Follow OWASP security guidelines',
    ];

    recommendations.forEach(rec => console.log(rec));
    console.log('');
  }

  // Run complete audit
  async run() {
    console.log('🔐 Starting Security Audit for OllamaMax\n');

    this.auditDependencies();
    this.auditSourceCode();
    this.auditConfiguration();

    const report = this.generateReport();

    // Save report to file
    const reportPath = path.join(process.cwd(), 'security-audit-report.json');
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    console.log(`📄 Detailed report saved to: ${reportPath}\n`);

    // Exit with error code if critical or high severity issues found
    const criticalCount = report.summary[this.severity.CRITICAL];
    const highCount = report.summary[this.severity.HIGH];

    if (criticalCount > 0 || highCount > 0) {
      console.log('❌ Security audit failed due to critical or high severity issues.');
      process.exit(1);
    } else {
      console.log('✅ Security audit passed!');
      process.exit(0);
    }
  }
}

// Run audit if called directly
if (require.main === module) {
  const auditor = new SecurityAuditor();
  auditor.run().catch(error => {
    console.error('Security audit failed:', error);
    process.exit(1);
  });
}

module.exports = SecurityAuditor;
