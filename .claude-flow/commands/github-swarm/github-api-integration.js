#!/usr/bin/env node

/**
 * GitHub API Integration
 * Handles GitHub API operations for the swarm
 */

const { Octokit } = require('@octokit/rest');

class GitHubAPIIntegration {
  constructor(token) {
    this.octokit = new Octokit({
      auth: token || process.env.GITHUB_TOKEN,
      userAgent: 'claude-flow-github-swarm v1.0.0'
    });
    this.rateLimitRemaining = 5000;
    this.rateLimitReset = Date.now();
  }

  async validateRepository(owner, repo) {
    try {
      const { data } = await this.octokit.rest.repos.get({
        owner,
        repo
      });
      return {
        valid: true,
        repository: data,
        permissions: {
          admin: data.permissions?.admin || false,
          push: data.permissions?.push || false,
          pull: data.permissions?.pull || false
        }
      };
    } catch (error) {
      return {
        valid: false,
        error: error.message
      };
    }
  }

  async getRepositoryInfo(owner, repo) {
    const { data } = await this.octokit.rest.repos.get({
      owner,
      repo
    });

    const [issues, pullRequests, branches, releases] = await Promise.all([
      this.getIssues(owner, repo, { state: 'open', per_page: 100 }),
      this.getPullRequests(owner, repo, { state: 'open', per_page: 100 }),
      this.getBranches(owner, repo),
      this.getReleases(owner, repo, { per_page: 10 })
    ]);

    return {
      repository: data,
      statistics: {
        openIssues: issues.length,
        openPullRequests: pullRequests.length,
        branches: branches.length,
        releases: releases.length,
        stars: data.stargazers_count,
        forks: data.forks_count,
        watchers: data.watchers_count
      },
      health: {
        hasReadme: data.has_readme || false,
        hasLicense: data.license !== null,
        hasWiki: data.has_wiki || false,
        hasPages: data.has_pages || false,
        hasIssues: data.has_issues || false,
        hasProjects: data.has_projects || false
      }
    };
  }

  async getIssues(owner, repo, options = {}) {
    const { data } = await this.octokit.rest.issues.listForRepo({
      owner,
      repo,
      state: options.state || 'open',
      per_page: options.per_page || 30,
      sort: options.sort || 'created',
      direction: options.direction || 'desc'
    });

    return data.filter(issue => !issue.pull_request);
  }

  async getPullRequests(owner, repo, options = {}) {
    const { data } = await this.octokit.rest.pulls.list({
      owner,
      repo,
      state: options.state || 'open',
      per_page: options.per_page || 30,
      sort: options.sort || 'created',
      direction: options.direction || 'desc'
    });

    return data;
  }

  async getBranches(owner, repo) {
    const { data } = await this.octokit.rest.repos.listBranches({
      owner,
      repo,
      per_page: 100
    });

    return data;
  }

  async getReleases(owner, repo, options = {}) {
    const { data } = await this.octokit.rest.repos.listReleases({
      owner,
      repo,
      per_page: options.per_page || 10
    });

    return data;
  }

  async analyzeIssues(owner, repo) {
    const issues = await this.getIssues(owner, repo, { per_page: 100 });
    
    const analysis = {
      total: issues.length,
      byLabel: {},
      byAssignee: {},
      byState: {},
      oldestIssue: null,
      newestIssue: null,
      averageAge: 0,
      needsLabels: 0,
      needsAssignee: 0
    };

    let totalAge = 0;
    let oldestDate = new Date();
    let newestDate = new Date(0);

    issues.forEach(issue => {
      const createdAt = new Date(issue.created_at);
      const age = Date.now() - createdAt.getTime();
      totalAge += age;

      if (createdAt < oldestDate) {
        oldestDate = createdAt;
        analysis.oldestIssue = issue;
      }

      if (createdAt > newestDate) {
        newestDate = createdAt;
        analysis.newestIssue = issue;
      }

      // Analyze labels
      if (issue.labels.length === 0) {
        analysis.needsLabels++;
      } else {
        issue.labels.forEach(label => {
          analysis.byLabel[label.name] = (analysis.byLabel[label.name] || 0) + 1;
        });
      }

      // Analyze assignees
      if (!issue.assignee) {
        analysis.needsAssignee++;
      } else {
        const assignee = issue.assignee.login;
        analysis.byAssignee[assignee] = (analysis.byAssignee[assignee] || 0) + 1;
      }

      // Analyze state
      analysis.byState[issue.state] = (analysis.byState[issue.state] || 0) + 1;
    });

    if (issues.length > 0) {
      analysis.averageAge = Math.round(totalAge / issues.length / (1000 * 60 * 60 * 24)); // days
    }

    return analysis;
  }

  async analyzePullRequests(owner, repo) {
    const pullRequests = await this.getPullRequests(owner, repo, { per_page: 100 });
    
    const analysis = {
      total: pullRequests.length,
      byState: {},
      byAuthor: {},
      needsReview: 0,
      hasConflicts: 0,
      averageAge: 0,
      oldestPR: null,
      newestPR: null
    };

    let totalAge = 0;
    let oldestDate = new Date();
    let newestDate = new Date(0);

    for (const pr of pullRequests) {
      const createdAt = new Date(pr.created_at);
      const age = Date.now() - createdAt.getTime();
      totalAge += age;

      if (createdAt < oldestDate) {
        oldestDate = createdAt;
        analysis.oldestPR = pr;
      }

      if (createdAt > newestDate) {
        newestDate = createdAt;
        analysis.newestPR = pr;
      }

      // Analyze state
      analysis.byState[pr.state] = (analysis.byState[pr.state] || 0) + 1;

      // Analyze author
      const author = pr.user.login;
      analysis.byAuthor[author] = (analysis.byAuthor[author] || 0) + 1;

      // Check if needs review
      if (pr.requested_reviewers.length === 0 && pr.state === 'open') {
        analysis.needsReview++;
      }

      // Check for conflicts (would need additional API call)
      if (pr.mergeable === false) {
        analysis.hasConflicts++;
      }
    }

    if (pullRequests.length > 0) {
      analysis.averageAge = Math.round(totalAge / pullRequests.length / (1000 * 60 * 60 * 24)); // days
    }

    return analysis;
  }

  async createIssueComment(owner, repo, issueNumber, body) {
    const { data } = await this.octokit.rest.issues.createComment({
      owner,
      repo,
      issue_number: issueNumber,
      body
    });

    return data;
  }

  async addLabelsToIssue(owner, repo, issueNumber, labels) {
    const { data } = await this.octokit.rest.issues.addLabels({
      owner,
      repo,
      issue_number: issueNumber,
      labels
    });

    return data;
  }

  async createPullRequestReview(owner, repo, pullNumber, event, body, comments = []) {
    const { data } = await this.octokit.rest.pulls.createReview({
      owner,
      repo,
      pull_number: pullNumber,
      event, // 'APPROVE', 'REQUEST_CHANGES', 'COMMENT'
      body,
      comments
    });

    return data;
  }

  async getRateLimitStatus() {
    const { data } = await this.octokit.rest.rateLimit.get();
    
    this.rateLimitRemaining = data.rate.remaining;
    this.rateLimitReset = data.rate.reset * 1000;

    return {
      remaining: data.rate.remaining,
      limit: data.rate.limit,
      reset: new Date(data.rate.reset * 1000),
      used: data.rate.used
    };
  }

  async waitForRateLimit() {
    if (this.rateLimitRemaining < 100) {
      const waitTime = this.rateLimitReset - Date.now();
      if (waitTime > 0) {
        console.log(`⏳ Rate limit low, waiting ${Math.round(waitTime / 1000)}s...`);
        await new Promise(resolve => setTimeout(resolve, waitTime));
      }
    }
  }
}

module.exports = GitHubAPIIntegration;
