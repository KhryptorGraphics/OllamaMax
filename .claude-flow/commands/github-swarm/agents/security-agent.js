#!/usr/bin/env node

/**
 * Security Agent
 * Specialized agent for security analysis and vulnerability detection
 */

const GitHubAPIIntegration = require('../github-api-integration');

class SecurityAgent {
  constructor(githubToken) {
    this.github = new GitHubAPIIntegration(githubToken);
    this.name = 'Security Agent';
    this.capabilities = ['vulnerability-scan', 'dependency-check', 'security-audit', 'compliance'];
  }

  async execute(config) {
    const { repository, features } = config;
    const [owner, repo] = repository.split('/');

    console.log(`🔒 ${this.name} analyzing repository: ${repository}`);

    const results = {
      agent: this.name,
      repository,
      tasksCompleted: 0,
      summary: '',
      actions: [],
      recommendations: []
    };

    try {
      // Analyze security policies
      const securityPolicies = await this.analyzeSecurityPolicies(owner, repo);
      if (securityPolicies.recommendations.length > 0) {
        results.recommendations.push(...securityPolicies.recommendations);
        results.tasksCompleted++;
      }

      // Check dependency security
      const dependencyAnalysis = await this.analyzeDependencySecurity(owner, repo);
      if (dependencyAnalysis.vulnerabilities.length > 0) {
        results.actions.push({
          type: 'dependency-vulnerabilities',
          vulnerabilities: dependencyAnalysis.vulnerabilities
        });
        results.tasksCompleted++;
      }

      // Analyze secrets and sensitive data
      const secretsAnalysis = await this.analyzeSecretsExposure(owner, repo);
      if (secretsAnalysis.risks.length > 0) {
        results.actions.push({
          type: 'secrets-exposure-risks',
          risks: secretsAnalysis.risks
        });
        results.tasksCompleted++;
      }

      // Check security configurations
      const configAnalysis = await this.analyzeSecurityConfigurations(owner, repo);
      if (configAnalysis.recommendations.length > 0) {
        results.recommendations.push(...configAnalysis.recommendations);
        results.tasksCompleted++;
      }

      // Analyze branch protection
      const branchProtection = await this.analyzeBranchProtection(owner, repo);
      if (branchProtection.recommendations.length > 0) {
        results.recommendations.push(...branchProtection.recommendations);
        results.tasksCompleted++;
      }

      // Generate summary
      results.summary = this.generateSummary(results.actions, results.recommendations);

      console.log(`✅ ${this.name} completed analysis`);
      return results;

    } catch (error) {
      console.error(`❌ ${this.name} failed:`, error.message);
      throw error;
    }
  }

  async analyzeSecurityPolicies(owner, repo) {
    const recommendations = [];

    // Check for SECURITY.md
    try {
      await this.github.octokit.rest.repos.getContent({
        owner,
        repo,
        path: 'SECURITY.md'
      });
    } catch (error) {
      if (error.status === 404) {
        recommendations.push({
          type: 'missing-security-policy',
          priority: 'high',
          message: 'No SECURITY.md file found - add security policy and vulnerability reporting guidelines'
        });
      }
    }

    // Check for security advisories
    try {
      const { data } = await this.github.octokit.rest.securityAdvisories.listRepositoryAdvisories({
        owner,
        repo
      });

      if (data.length > 0) {
        const openAdvisories = data.filter(advisory => advisory.state === 'published');
        if (openAdvisories.length > 0) {
          recommendations.push({
            type: 'open-security-advisories',
            priority: 'critical',
            message: `${openAdvisories.length} published security advisories found - review and address`
          });
        }
      }
    } catch (error) {
      // Security advisories API might not be available
    }

    return { recommendations };
  }

  async analyzeDependencySecurity(owner, repo) {
    const vulnerabilities = [];

    // Check for known vulnerable dependencies
    try {
      // Check package.json for Node.js projects
      const { data } = await this.github.octokit.rest.repos.getContent({
        owner,
        repo,
        path: 'package.json'
      });

      const packageJson = JSON.parse(Buffer.from(data.content, 'base64').toString());
      const dependencies = { ...packageJson.dependencies, ...packageJson.devDependencies };

      // Check for commonly vulnerable packages (simplified list)
      const knownVulnerable = {
        'lodash': { versions: ['<4.17.21'], severity: 'medium', issue: 'Prototype pollution' },
        'axios': { versions: ['<0.21.1'], severity: 'medium', issue: 'SSRF vulnerability' },
        'express': { versions: ['<4.17.1'], severity: 'medium', issue: 'Various security issues' },
        'moment': { versions: ['*'], severity: 'low', issue: 'Deprecated, use date-fns or dayjs' },
        'request': { versions: ['*'], severity: 'medium', issue: 'Deprecated, security unmaintained' }
      };

      for (const [dep, info] of Object.entries(knownVulnerable)) {
        if (dependencies[dep]) {
          vulnerabilities.push({
            package: dep,
            currentVersion: dependencies[dep],
            severity: info.severity,
            issue: info.issue,
            recommendation: `Update ${dep} to latest version`
          });
        }
      }

    } catch (error) {
      // Not a Node.js project or package.json not found
    }

    // Check for requirements.txt (Python)
    try {
      const { data } = await this.github.octokit.rest.repos.getContent({
        owner,
        repo,
        path: 'requirements.txt'
      });

      const requirements = Buffer.from(data.content, 'base64').toString();
      
      // Check for commonly vulnerable Python packages
      const pythonVulnerable = ['django<3.2', 'flask<2.0', 'requests<2.25'];
      
      pythonVulnerable.forEach(vuln => {
        if (requirements.includes(vuln.split('<')[0])) {
          vulnerabilities.push({
            package: vuln.split('<')[0],
            severity: 'medium',
            issue: 'Potentially outdated version',
            recommendation: `Update to latest version`
          });
        }
      });

    } catch (error) {
      // Not a Python project or requirements.txt not found
    }

    return { vulnerabilities };
  }

  async analyzeSecretsExposure(owner, repo) {
    const risks = [];

    // Check for common secret patterns in repository files
    const secretPatterns = [
      { pattern: 'api[_-]?key', name: 'API Key', severity: 'high' },
      { pattern: 'secret[_-]?key', name: 'Secret Key', severity: 'high' },
      { pattern: 'password\\s*=', name: 'Hardcoded Password', severity: 'critical' },
      { pattern: 'token\\s*=', name: 'Access Token', severity: 'high' },
      { pattern: 'aws[_-]?access[_-]?key', name: 'AWS Access Key', severity: 'critical' },
      { pattern: 'private[_-]?key', name: 'Private Key', severity: 'critical' }
    ];

    // Check common configuration files
    const configFiles = [
      '.env', '.env.local', '.env.production',
      'config.json', 'config.yaml', 'config.yml',
      'settings.py', 'application.properties'
    ];

    for (const file of configFiles) {
      try {
        const { data } = await this.github.octokit.rest.repos.getContent({
          owner,
          repo,
          path: file
        });

        const content = Buffer.from(data.content, 'base64').toString();
        
        secretPatterns.forEach(({ pattern, name, severity }) => {
          const regex = new RegExp(pattern, 'gi');
          if (regex.test(content)) {
            risks.push({
              file,
              type: name,
              severity,
              message: `Potential ${name.toLowerCase()} found in ${file}`,
              recommendation: 'Move secrets to environment variables or secure vault'
            });
          }
        });

      } catch (error) {
        // File doesn't exist
      }
    }

    // Check for .env files in repository (should be in .gitignore)
    try {
      await this.github.octokit.rest.repos.getContent({
        owner,
        repo,
        path: '.env'
      });

      risks.push({
        file: '.env',
        type: 'Environment File',
        severity: 'high',
        message: '.env file found in repository',
        recommendation: 'Add .env to .gitignore and remove from repository'
      });
    } catch (error) {
      // .env not found (good)
    }

    return { risks };
  }

  async analyzeSecurityConfigurations(owner, repo) {
    const recommendations = [];

    // Check .gitignore for security-sensitive files
    try {
      const { data } = await this.github.octokit.rest.repos.getContent({
        owner,
        repo,
        path: '.gitignore'
      });

      const gitignore = Buffer.from(data.content, 'base64').toString();
      
      const securityPatterns = ['.env', '*.key', '*.pem', 'secrets/', 'config/secrets'];
      const missingPatterns = securityPatterns.filter(pattern => !gitignore.includes(pattern));

      if (missingPatterns.length > 0) {
        recommendations.push({
          type: 'gitignore-security',
          priority: 'medium',
          message: `Add security patterns to .gitignore: ${missingPatterns.join(', ')}`
        });
      }

    } catch (error) {
      if (error.status === 404) {
        recommendations.push({
          type: 'missing-gitignore',
          priority: 'medium',
          message: 'No .gitignore file found - create one to exclude sensitive files'
        });
      }
    }

    // Check for security linting configurations
    const securityLintFiles = [
      '.eslintrc.json', '.eslintrc.js', 'eslint.config.js',
      'bandit.yaml', '.bandit', 'pyproject.toml'
    ];

    let hasSecurityLinting = false;
    for (const file of securityLintFiles) {
      try {
        const { data } = await this.github.octokit.rest.repos.getContent({
          owner,
          repo,
          path: file
        });

        const content = Buffer.from(data.content, 'base64').toString();
        if (content.includes('security') || content.includes('bandit') || content.includes('eslint-plugin-security')) {
          hasSecurityLinting = true;
          break;
        }
      } catch (error) {
        // File doesn't exist
      }
    }

    if (!hasSecurityLinting) {
      recommendations.push({
        type: 'security-linting',
        priority: 'medium',
        message: 'Consider adding security linting tools (eslint-plugin-security, bandit, etc.)'
      });
    }

    return { recommendations };
  }

  async analyzeBranchProtection(owner, repo) {
    const recommendations = [];

    try {
      // Check main/master branch protection
      const mainBranches = ['main', 'master'];
      
      for (const branch of mainBranches) {
        try {
          const { data } = await this.github.octokit.rest.repos.getBranchProtection({
            owner,
            repo,
            branch
          });

          // Analyze protection settings
          if (!data.required_status_checks) {
            recommendations.push({
              type: 'branch-protection-status-checks',
              priority: 'high',
              message: `${branch} branch lacks required status checks`
            });
          }

          if (!data.enforce_admins) {
            recommendations.push({
              type: 'branch-protection-admin-enforcement',
              priority: 'medium',
              message: `${branch} branch protection doesn't enforce rules for admins`
            });
          }

          if (!data.required_pull_request_reviews || data.required_pull_request_reviews.required_approving_review_count < 1) {
            recommendations.push({
              type: 'branch-protection-reviews',
              priority: 'high',
              message: `${branch} branch requires pull request reviews`
            });
          }

          break; // Found protected main branch
        } catch (error) {
          if (error.status === 404) {
            // Branch doesn't exist or no protection
            continue;
          }
        }
      }

    } catch (error) {
      recommendations.push({
        type: 'branch-protection-setup',
        priority: 'high',
        message: 'No branch protection found - set up protection for main/master branch'
      });
    }

    return { recommendations };
  }

  generateSummary(actions, recommendations) {
    const vulnerabilityCount = actions.reduce((sum, action) => {
      if (action.type === 'dependency-vulnerabilities') {
        return sum + action.vulnerabilities.length;
      }
      if (action.type === 'secrets-exposure-risks') {
        return sum + action.risks.length;
      }
      return sum;
    }, 0);

    const criticalRecommendations = recommendations.filter(r => r.priority === 'critical' || r.priority === 'high').length;

    return `Security scan completed. Found ${vulnerabilityCount} potential vulnerabilities, ` +
           `${criticalRecommendations} high-priority security recommendations. ` +
           `${recommendations.length} total security improvements suggested.`;
  }
}

module.exports = SecurityAgent;
