#!/usr/bin/env node

/**
 * PR Reviewer Agent
 * Specialized agent for reviewing pull requests and suggesting improvements
 */

const GitHubAPIIntegration = require('../github-api-integration');

class PRReviewerAgent {
  constructor(githubToken) {
    this.github = new GitHubAPIIntegration(githubToken);
    this.name = 'PR Reviewer';
    this.capabilities = ['code-review', 'best-practices', 'security-check', 'performance-analysis'];
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
      // Analyze existing pull requests
      const prAnalysis = await this.github.analyzePullRequests(owner, repo);
      results.tasksCompleted++;

      // Review open pull requests
      if (features.codeReview) {
        const reviewResults = await this.reviewOpenPullRequests(owner, repo);
        results.actions.push(...reviewResults.actions);
        results.tasksCompleted++;
      }

      // Check for PR best practices
      const bestPracticesCheck = await this.checkPRBestPractices(owner, repo);
      results.recommendations.push(...bestPracticesCheck);
      results.tasksCompleted++;

      // Analyze PR patterns
      const patternAnalysis = await this.analyzePRPatterns(prAnalysis);
      results.recommendations.push(...patternAnalysis);
      results.tasksCompleted++;

      // Generate summary
      results.summary = this.generateSummary(prAnalysis, results.actions);

      console.log(`✅ ${this.name} completed analysis`);
      return results;

    } catch (error) {
      console.error(`❌ ${this.name} failed:`, error.message);
      throw error;
    }
  }

  async reviewOpenPullRequests(owner, repo) {
    const pullRequests = await this.github.getPullRequests(owner, repo, { 
      state: 'open', 
      per_page: 10 
    });

    const actions = [];
    const reviewSuggestions = [];

    for (const pr of pullRequests) {
      const review = await this.analyzePullRequest(owner, repo, pr);
      
      if (review.suggestions.length > 0) {
        reviewSuggestions.push({
          prNumber: pr.number,
          title: pr.title,
          author: pr.user.login,
          suggestions: review.suggestions,
          severity: review.severity,
          score: review.score
        });
      }
    }

    if (reviewSuggestions.length > 0) {
      actions.push({
        type: 'pr-reviews',
        count: reviewSuggestions.length,
        reviews: reviewSuggestions
      });
    }

    return { actions };
  }

  async analyzePullRequest(owner, repo, pr) {
    const suggestions = [];
    let severity = 'low';
    let score = 85; // Base score

    // Check PR title and description
    const titleIssues = this.analyzePRTitle(pr.title);
    if (titleIssues.length > 0) {
      suggestions.push(...titleIssues);
      score -= 5;
    }

    const descriptionIssues = this.analyzePRDescription(pr.body);
    if (descriptionIssues.length > 0) {
      suggestions.push(...descriptionIssues);
      score -= 10;
    }

    // Check PR size and complexity
    const sizeAnalysis = this.analyzePRSize(pr);
    if (sizeAnalysis.issues.length > 0) {
      suggestions.push(...sizeAnalysis.issues);
      score -= sizeAnalysis.penalty;
      if (sizeAnalysis.severity === 'high') severity = 'high';
    }

    // Check for common issues
    const commonIssues = this.checkCommonPRIssues(pr);
    if (commonIssues.length > 0) {
      suggestions.push(...commonIssues);
      score -= 15;
      severity = 'medium';
    }

    return {
      suggestions,
      severity,
      score: Math.max(score, 0)
    };
  }

  analyzePRTitle(title) {
    const issues = [];

    if (!title || title.length < 10) {
      issues.push({
        type: 'title-too-short',
        message: 'PR title should be more descriptive (at least 10 characters)',
        suggestion: 'Add more context about what the PR changes'
      });
    }

    if (title && title.length > 100) {
      issues.push({
        type: 'title-too-long',
        message: 'PR title is too long (over 100 characters)',
        suggestion: 'Shorten the title and move details to description'
      });
    }

    const hasConventionalCommit = /^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?: .+/.test(title);
    if (!hasConventionalCommit) {
      issues.push({
        type: 'conventional-commits',
        message: 'Consider using conventional commit format',
        suggestion: 'Use prefixes like feat:, fix:, docs:, etc.'
      });
    }

    return issues;
  }

  analyzePRDescription(body) {
    const issues = [];

    if (!body || body.trim().length < 20) {
      issues.push({
        type: 'missing-description',
        message: 'PR description is missing or too brief',
        suggestion: 'Add description explaining what changes and why'
      });
    }

    if (body && !body.includes('## ') && !body.includes('### ')) {
      issues.push({
        type: 'unstructured-description',
        message: 'PR description lacks structure',
        suggestion: 'Use headers like ## Changes, ## Testing, ## Notes'
      });
    }

    const hasTestingInfo = body && (
      body.toLowerCase().includes('test') ||
      body.toLowerCase().includes('testing') ||
      body.toLowerCase().includes('verified')
    );

    if (!hasTestingInfo) {
      issues.push({
        type: 'missing-testing-info',
        message: 'No testing information provided',
        suggestion: 'Add section describing how changes were tested'
      });
    }

    return issues;
  }

  analyzePRSize(pr) {
    const issues = [];
    let penalty = 0;
    let severity = 'low';

    // Note: In real implementation, you'd get file changes via API
    const estimatedChanges = pr.additions + pr.deletions;

    if (estimatedChanges > 1000) {
      issues.push({
        type: 'large-pr',
        message: `PR is very large (${estimatedChanges} changes)`,
        suggestion: 'Consider breaking into smaller, focused PRs'
      });
      penalty = 20;
      severity = 'high';
    } else if (estimatedChanges > 500) {
      issues.push({
        type: 'medium-pr',
        message: `PR is moderately large (${estimatedChanges} changes)`,
        suggestion: 'Ensure PR has single responsibility'
      });
      penalty = 10;
      severity = 'medium';
    }

    return { issues, penalty, severity };
  }

  checkCommonPRIssues(pr) {
    const issues = [];

    // Check if PR targets main/master directly
    if (['main', 'master'].includes(pr.base.ref) && pr.head.ref.includes('feature')) {
      issues.push({
        type: 'direct-to-main',
        message: 'Feature branch targeting main directly',
        suggestion: 'Consider using develop branch or feature flags'
      });
    }

    // Check for draft PRs that might be ready
    if (pr.draft && pr.additions > 0) {
      issues.push({
        type: 'draft-status',
        message: 'PR is marked as draft but has changes',
        suggestion: 'Mark as ready for review if complete'
      });
    }

    // Check for missing reviewers
    if (pr.requested_reviewers.length === 0 && !pr.draft) {
      issues.push({
        type: 'no-reviewers',
        message: 'No reviewers assigned',
        suggestion: 'Add appropriate reviewers for code review'
      });
    }

    return issues;
  }

  async checkPRBestPractices(owner, repo) {
    const recommendations = [];
    const pullRequests = await this.github.getPullRequests(owner, repo, { 
      state: 'all', 
      per_page: 50 
    });

    // Analyze PR patterns
    const avgPRSize = pullRequests.reduce((sum, pr) => sum + (pr.additions + pr.deletions), 0) / pullRequests.length;
    const largePRs = pullRequests.filter(pr => (pr.additions + pr.deletions) > 500).length;
    const draftPRs = pullRequests.filter(pr => pr.draft).length;

    if (avgPRSize > 300) {
      recommendations.push({
        type: 'pr-size',
        priority: 'medium',
        message: `Average PR size is ${Math.round(avgPRSize)} changes - consider smaller PRs`
      });
    }

    if (largePRs > pullRequests.length * 0.3) {
      recommendations.push({
        type: 'large-prs',
        priority: 'high',
        message: `${largePRs} large PRs found - break down complex changes`
      });
    }

    if (draftPRs > 5) {
      recommendations.push({
        type: 'draft-prs',
        priority: 'low',
        message: `${draftPRs} draft PRs - review and clean up stale drafts`
      });
    }

    return recommendations;
  }

  async analyzePRPatterns(prAnalysis) {
    const recommendations = [];

    if (prAnalysis.needsReview > 3) {
      recommendations.push({
        type: 'review-bottleneck',
        priority: 'high',
        message: `${prAnalysis.needsReview} PRs waiting for review - potential bottleneck`
      });
    }

    if (prAnalysis.hasConflicts > 0) {
      recommendations.push({
        type: 'merge-conflicts',
        priority: 'medium',
        message: `${prAnalysis.hasConflicts} PRs have merge conflicts`
      });
    }

    if (prAnalysis.averageAge > 7) {
      recommendations.push({
        type: 'stale-prs',
        priority: 'medium',
        message: `Average PR age is ${prAnalysis.averageAge} days - consider faster review cycles`
      });
    }

    return recommendations;
  }

  generateSummary(analysis, actions) {
    const reviewCount = actions.reduce((sum, action) => sum + (action.count || 0), 0);
    
    return `Reviewed ${analysis.total} pull requests, provided ${reviewCount} detailed reviews. ` +
           `Found ${analysis.needsReview} PRs needing review, ${analysis.hasConflicts} with conflicts.`;
  }
}

module.exports = PRReviewerAgent;
