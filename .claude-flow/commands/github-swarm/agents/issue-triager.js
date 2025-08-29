#!/usr/bin/env node

/**
 * Issue Triager Agent
 * Specialized agent for analyzing and categorizing GitHub issues
 */

const GitHubAPIIntegration = require('../github-api-integration');

class IssueTriagerAgent {
  constructor(githubToken) {
    this.github = new GitHubAPIIntegration(githubToken);
    this.name = 'Issue Triager';
    this.capabilities = ['issue-analysis', 'labeling', 'prioritization', 'duplicate-detection'];
  }

  async execute(config) {
    const { repository, features } = config;
    const [owner, repo] = repository.split('/');

    console.log(`🔍 ${this.name} analyzing repository: ${repository}`);

    const results = {
      agent: this.name,
      repository,
      tasksCompleted: 0,
      summary: '',
      actions: [],
      recommendations: []
    };

    try {
      // Analyze existing issues
      const issueAnalysis = await this.github.analyzeIssues(owner, repo);
      results.tasksCompleted++;

      // Categorize issues that need labels
      if (features.issueLabels && issueAnalysis.needsLabels > 0) {
        const labelingResults = await this.categorizeUnlabeledIssues(owner, repo);
        results.actions.push(...labelingResults.actions);
        results.tasksCompleted++;
      }

      // Detect potential duplicates
      const duplicates = await this.detectDuplicateIssues(owner, repo);
      if (duplicates.length > 0) {
        results.actions.push({
          type: 'duplicate-detection',
          count: duplicates.length,
          duplicates: duplicates.slice(0, 5) // Show first 5
        });
        results.tasksCompleted++;
      }

      // Generate priority recommendations
      const priorityRecommendations = await this.generatePriorityRecommendations(issueAnalysis);
      results.recommendations.push(...priorityRecommendations);
      results.tasksCompleted++;

      // Generate summary
      results.summary = this.generateSummary(issueAnalysis, results.actions);

      console.log(`✅ ${this.name} completed analysis`);
      return results;

    } catch (error) {
      console.error(`❌ ${this.name} failed:`, error.message);
      throw error;
    }
  }

  async categorizeUnlabeledIssues(owner, repo) {
    const issues = await this.github.getIssues(owner, repo, { per_page: 50 });
    const unlabeledIssues = issues.filter(issue => issue.labels.length === 0);

    const actions = [];
    const labelSuggestions = [];

    for (const issue of unlabeledIssues.slice(0, 10)) { // Process first 10
      const suggestedLabels = this.suggestLabelsForIssue(issue);
      
      if (suggestedLabels.length > 0) {
        labelSuggestions.push({
          issueNumber: issue.number,
          title: issue.title,
          suggestedLabels,
          confidence: this.calculateLabelConfidence(issue, suggestedLabels)
        });
      }
    }

    if (labelSuggestions.length > 0) {
      actions.push({
        type: 'label-suggestions',
        count: labelSuggestions.length,
        suggestions: labelSuggestions
      });
    }

    return { actions };
  }

  suggestLabelsForIssue(issue) {
    const title = issue.title.toLowerCase();
    const body = (issue.body || '').toLowerCase();
    const text = `${title} ${body}`;

    const labelRules = [
      { keywords: ['bug', 'error', 'broken', 'crash', 'fail'], label: 'bug', priority: 'high' },
      { keywords: ['feature', 'enhancement', 'improve', 'add'], label: 'enhancement', priority: 'medium' },
      { keywords: ['documentation', 'docs', 'readme', 'wiki'], label: 'documentation', priority: 'low' },
      { keywords: ['question', 'help', 'how to', 'support'], label: 'question', priority: 'low' },
      { keywords: ['security', 'vulnerability', 'exploit'], label: 'security', priority: 'critical' },
      { keywords: ['performance', 'slow', 'optimization', 'speed'], label: 'performance', priority: 'medium' },
      { keywords: ['test', 'testing', 'spec', 'coverage'], label: 'testing', priority: 'medium' },
      { keywords: ['ui', 'ux', 'interface', 'design'], label: 'ui/ux', priority: 'medium' },
      { keywords: ['api', 'endpoint', 'rest', 'graphql'], label: 'api', priority: 'medium' },
      { keywords: ['database', 'db', 'sql', 'migration'], label: 'database', priority: 'medium' }
    ];

    const suggestedLabels = [];

    labelRules.forEach(rule => {
      const matches = rule.keywords.some(keyword => text.includes(keyword));
      if (matches) {
        suggestedLabels.push({
          name: rule.label,
          priority: rule.priority,
          confidence: this.calculateKeywordConfidence(text, rule.keywords)
        });
      }
    });

    // Sort by confidence and return top 3
    return suggestedLabels
      .sort((a, b) => b.confidence - a.confidence)
      .slice(0, 3)
      .map(label => label.name);
  }

  calculateKeywordConfidence(text, keywords) {
    const matches = keywords.filter(keyword => text.includes(keyword)).length;
    return Math.min(matches / keywords.length * 100, 95);
  }

  calculateLabelConfidence(issue, suggestedLabels) {
    // Base confidence on title/body length and keyword density
    const textLength = (issue.title + (issue.body || '')).length;
    const baseConfidence = Math.min(textLength / 100 * 10, 80);
    
    return Math.round(baseConfidence + (suggestedLabels.length * 5));
  }

  async detectDuplicateIssues(owner, repo) {
    const issues = await this.github.getIssues(owner, repo, { per_page: 100 });
    const duplicates = [];

    for (let i = 0; i < issues.length; i++) {
      for (let j = i + 1; j < issues.length; j++) {
        const similarity = this.calculateIssueSimilarity(issues[i], issues[j]);
        
        if (similarity > 0.8) {
          duplicates.push({
            issue1: {
              number: issues[i].number,
              title: issues[i].title,
              created: issues[i].created_at
            },
            issue2: {
              number: issues[j].number,
              title: issues[j].title,
              created: issues[j].created_at
            },
            similarity: Math.round(similarity * 100)
          });
        }
      }
    }

    return duplicates;
  }

  calculateIssueSimilarity(issue1, issue2) {
    const title1 = issue1.title.toLowerCase();
    const title2 = issue2.title.toLowerCase();
    
    // Simple similarity based on common words
    const words1 = new Set(title1.split(/\s+/));
    const words2 = new Set(title2.split(/\s+/));
    
    const intersection = new Set([...words1].filter(x => words2.has(x)));
    const union = new Set([...words1, ...words2]);
    
    return intersection.size / union.size;
  }

  async generatePriorityRecommendations(issueAnalysis) {
    const recommendations = [];

    if (issueAnalysis.needsLabels > 5) {
      recommendations.push({
        type: 'labeling',
        priority: 'high',
        message: `${issueAnalysis.needsLabels} issues need labels for better organization`
      });
    }

    if (issueAnalysis.needsAssignee > 10) {
      recommendations.push({
        type: 'assignment',
        priority: 'medium',
        message: `${issueAnalysis.needsAssignee} issues need assignees to track ownership`
      });
    }

    if (issueAnalysis.averageAge > 30) {
      recommendations.push({
        type: 'stale-issues',
        priority: 'medium',
        message: `Average issue age is ${issueAnalysis.averageAge} days - consider reviewing older issues`
      });
    }

    return recommendations;
  }

  generateSummary(analysis, actions) {
    const actionCount = actions.reduce((sum, action) => sum + (action.count || 1), 0);
    
    return `Analyzed ${analysis.total} issues, suggested ${actionCount} improvements. ` +
           `Found ${analysis.needsLabels} unlabeled issues, ${analysis.needsAssignee} unassigned issues.`;
  }
}

module.exports = IssueTriagerAgent;
