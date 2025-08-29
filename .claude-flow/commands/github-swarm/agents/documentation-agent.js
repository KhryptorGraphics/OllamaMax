#!/usr/bin/env node

/**
 * Documentation Agent
 * Specialized agent for maintaining and improving documentation
 */

const GitHubAPIIntegration = require('../github-api-integration');

class DocumentationAgent {
  constructor(githubToken) {
    this.github = new GitHubAPIIntegration(githubToken);
    this.name = 'Documentation Agent';
    this.capabilities = ['readme-updates', 'api-docs', 'changelog', 'wiki-maintenance'];
  }

  async execute(config) {
    const { repository, features } = config;
    const [owner, repo] = repository.split('/');

    console.log(`📚 ${this.name} analyzing repository: ${repository}`);

    const results = {
      agent: this.name,
      repository,
      tasksCompleted: 0,
      summary: '',
      actions: [],
      recommendations: []
    };

    try {
      // Analyze repository documentation
      const repoInfo = await this.github.getRepositoryInfo(owner, repo);
      const docAnalysis = await this.analyzeDocumentation(owner, repo, repoInfo);
      results.tasksCompleted++;

      // Check README quality
      const readmeAnalysis = await this.analyzeReadme(owner, repo);
      if (readmeAnalysis.issues.length > 0) {
        results.actions.push({
          type: 'readme-improvements',
          issues: readmeAnalysis.issues,
          suggestions: readmeAnalysis.suggestions
        });
        results.tasksCompleted++;
      }

      // Analyze API documentation
      const apiDocAnalysis = await this.analyzeAPIDocumentation(owner, repo);
      if (apiDocAnalysis.recommendations.length > 0) {
        results.recommendations.push(...apiDocAnalysis.recommendations);
        results.tasksCompleted++;
      }

      // Check for missing documentation
      const missingDocs = await this.identifyMissingDocumentation(repoInfo);
      if (missingDocs.length > 0) {
        results.actions.push({
          type: 'missing-documentation',
          missing: missingDocs
        });
        results.tasksCompleted++;
      }

      // Generate changelog recommendations
      const changelogAnalysis = await this.analyzeChangelog(owner, repo);
      if (changelogAnalysis.recommendations.length > 0) {
        results.recommendations.push(...changelogAnalysis.recommendations);
        results.tasksCompleted++;
      }

      // Generate summary
      results.summary = this.generateSummary(docAnalysis, results.actions);

      console.log(`✅ ${this.name} completed analysis`);
      return results;

    } catch (error) {
      console.error(`❌ ${this.name} failed:`, error.message);
      throw error;
    }
  }

  async analyzeDocumentation(owner, repo, repoInfo) {
    const analysis = {
      hasReadme: repoInfo.health.hasReadme,
      hasLicense: repoInfo.health.hasLicense,
      hasWiki: repoInfo.health.hasWiki,
      hasPages: repoInfo.health.hasPages,
      documentationScore: 0,
      issues: [],
      strengths: []
    };

    // Calculate documentation score
    let score = 0;
    if (analysis.hasReadme) score += 30;
    if (analysis.hasLicense) score += 20;
    if (analysis.hasWiki) score += 15;
    if (analysis.hasPages) score += 15;

    // Check for common documentation files
    const commonDocs = await this.checkCommonDocFiles(owner, repo);
    score += commonDocs.found * 4; // 4 points per doc file

    analysis.documentationScore = Math.min(score, 100);

    // Identify issues
    if (!analysis.hasReadme) {
      analysis.issues.push('Missing README file');
    }
    if (!analysis.hasLicense) {
      analysis.issues.push('Missing LICENSE file');
    }
    if (commonDocs.missing.length > 0) {
      analysis.issues.push(`Missing common docs: ${commonDocs.missing.join(', ')}`);
    }

    // Identify strengths
    if (analysis.hasReadme) analysis.strengths.push('Has README');
    if (analysis.hasLicense) analysis.strengths.push('Has LICENSE');
    if (analysis.hasWiki) analysis.strengths.push('Has Wiki');
    if (analysis.hasPages) analysis.strengths.push('Has GitHub Pages');

    return analysis;
  }

  async checkCommonDocFiles(owner, repo) {
    const commonFiles = [
      'CONTRIBUTING.md',
      'CODE_OF_CONDUCT.md',
      'SECURITY.md',
      'CHANGELOG.md',
      'INSTALL.md',
      'USAGE.md',
      'API.md',
      'TROUBLESHOOTING.md'
    ];

    const found = [];
    const missing = [];

    for (const file of commonFiles) {
      try {
        await this.github.octokit.rest.repos.getContent({
          owner,
          repo,
          path: file
        });
        found.push(file);
      } catch (error) {
        if (error.status === 404) {
          missing.push(file);
        }
      }
    }

    return { found: found.length, missing };
  }

  async analyzeReadme(owner, repo) {
    const analysis = {
      exists: false,
      size: 0,
      sections: [],
      issues: [],
      suggestions: []
    };

    try {
      const { data } = await this.github.octokit.rest.repos.getContent({
        owner,
        repo,
        path: 'README.md'
      });

      analysis.exists = true;
      const content = Buffer.from(data.content, 'base64').toString();
      analysis.size = content.length;

      // Analyze README structure
      const sections = this.extractMarkdownSections(content);
      analysis.sections = sections;

      // Check for essential sections
      const essentialSections = [
        'installation',
        'usage',
        'getting started',
        'quick start',
        'examples',
        'api',
        'contributing',
        'license'
      ];

      const missingSections = essentialSections.filter(section => 
        !sections.some(s => s.toLowerCase().includes(section))
      );

      if (missingSections.length > 0) {
        analysis.issues.push(`Missing sections: ${missingSections.join(', ')}`);
        analysis.suggestions.push('Add missing essential sections to improve usability');
      }

      // Check README length
      if (analysis.size < 500) {
        analysis.issues.push('README is too brief (less than 500 characters)');
        analysis.suggestions.push('Expand README with more detailed information');
      } else if (analysis.size > 10000) {
        analysis.issues.push('README is very long (over 10,000 characters)');
        analysis.suggestions.push('Consider breaking into multiple documents');
      }

      // Check for code examples
      const hasCodeBlocks = content.includes('```');
      if (!hasCodeBlocks) {
        analysis.issues.push('No code examples found');
        analysis.suggestions.push('Add code examples to demonstrate usage');
      }

      // Check for badges
      const hasBadges = content.includes('![') || content.includes('[![');
      if (!hasBadges) {
        analysis.suggestions.push('Consider adding status badges (build, coverage, version)');
      }

    } catch (error) {
      if (error.status === 404) {
        analysis.issues.push('README.md file not found');
        analysis.suggestions.push('Create a comprehensive README.md file');
      }
    }

    return analysis;
  }

  extractMarkdownSections(content) {
    const sections = [];
    const lines = content.split('\n');
    
    for (const line of lines) {
      const match = line.match(/^#+\s+(.+)$/);
      if (match) {
        sections.push(match[1].trim());
      }
    }
    
    return sections;
  }

  async analyzeAPIDocumentation(owner, repo) {
    const recommendations = [];

    // Check for API documentation files
    const apiFiles = ['API.md', 'docs/api.md', 'api/README.md'];
    let hasApiDocs = false;

    for (const file of apiFiles) {
      try {
        await this.github.octokit.rest.repos.getContent({
          owner,
          repo,
          path: file
        });
        hasApiDocs = true;
        break;
      } catch (error) {
        // File doesn't exist, continue checking
      }
    }

    if (!hasApiDocs) {
      recommendations.push({
        type: 'api-documentation',
        priority: 'medium',
        message: 'No API documentation found - consider adding API.md'
      });
    }

    // Check for OpenAPI/Swagger specs
    const specFiles = ['openapi.yaml', 'swagger.yaml', 'api-spec.yaml'];
    let hasApiSpec = false;

    for (const file of specFiles) {
      try {
        await this.github.octokit.rest.repos.getContent({
          owner,
          repo,
          path: file
        });
        hasApiSpec = true;
        break;
      } catch (error) {
        // File doesn't exist, continue checking
      }
    }

    if (!hasApiSpec) {
      recommendations.push({
        type: 'api-specification',
        priority: 'low',
        message: 'Consider adding OpenAPI/Swagger specification'
      });
    }

    return { recommendations };
  }

  async identifyMissingDocumentation(repoInfo) {
    const missing = [];

    if (!repoInfo.health.hasReadme) {
      missing.push({
        file: 'README.md',
        priority: 'high',
        description: 'Main project documentation'
      });
    }

    if (!repoInfo.health.hasLicense) {
      missing.push({
        file: 'LICENSE',
        priority: 'high',
        description: 'Project license information'
      });
    }

    // Check for other important files
    const importantFiles = [
      { file: 'CONTRIBUTING.md', description: 'Contribution guidelines' },
      { file: 'CODE_OF_CONDUCT.md', description: 'Community standards' },
      { file: 'SECURITY.md', description: 'Security policy' }
    ];

    for (const doc of importantFiles) {
      missing.push({
        ...doc,
        priority: 'medium'
      });
    }

    return missing;
  }

  async analyzeChangelog(owner, repo) {
    const recommendations = [];

    try {
      const { data } = await this.github.octokit.rest.repos.getContent({
        owner,
        repo,
        path: 'CHANGELOG.md'
      });

      const content = Buffer.from(data.content, 'base64').toString();
      
      // Check changelog format
      if (!content.includes('## ') && !content.includes('# ')) {
        recommendations.push({
          type: 'changelog-format',
          priority: 'low',
          message: 'Changelog lacks proper section headers'
        });
      }

      // Check for recent updates
      const releases = await this.github.getReleases(owner, repo, { per_page: 5 });
      if (releases.length > 0) {
        const latestRelease = releases[0];
        const releaseTag = latestRelease.tag_name;
        
        if (!content.includes(releaseTag)) {
          recommendations.push({
            type: 'changelog-outdated',
            priority: 'medium',
            message: `Changelog missing entry for latest release ${releaseTag}`
          });
        }
      }

    } catch (error) {
      if (error.status === 404) {
        recommendations.push({
          type: 'missing-changelog',
          priority: 'medium',
          message: 'No CHANGELOG.md found - consider adding one'
        });
      }
    }

    return { recommendations };
  }

  generateSummary(analysis, actions) {
    const actionCount = actions.reduce((sum, action) => sum + (action.issues?.length || action.missing?.length || 1), 0);
    
    return `Documentation score: ${analysis.documentationScore}/100. ` +
           `Identified ${actionCount} improvement opportunities. ` +
           `Strengths: ${analysis.strengths.join(', ') || 'None identified'}.`;
  }
}

module.exports = DocumentationAgent;
