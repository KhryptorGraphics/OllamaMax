# GitHub Swarm - Specialized Repository Management

A specialized swarm system for intelligent GitHub repository management using AI agents.

## Overview

GitHub Swarm deploys specialized AI agents to analyze, maintain, and improve GitHub repositories. Each agent focuses on specific aspects of repository health and development workflow optimization.

## Features

- **Intelligent Issue Triage** - Automatically categorize and label issues
- **PR Review Automation** - AI-powered code review and suggestions
- **Documentation Management** - Keep docs up-to-date and comprehensive
- **Test Coverage Analysis** - Identify gaps and suggest improvements
- **Security Auditing** - Scan for vulnerabilities and compliance issues

## Installation

```bash
# Install dependencies
npm install

# Make executable
chmod +x index.js

# Optional: Install globally
npm install -g .
```

## Usage

### Basic Commands

```bash
# Basic GitHub swarm
./index.js --repository owner/repo

# Maintenance-focused swarm
./index.js -r owner/repo -f maintenance --issue-labels

# Development swarm with PR automation
./index.js -r owner/repo -f development --auto-pr --code-review

# Full-featured triage swarm
./index.js -r owner/repo -a 8 -f triage --issue-labels --auto-pr
```

### Command Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--repository` | `-r` | Target GitHub repository (owner/repo) | Required |
| `--agents` | `-a` | Number of specialized agents | 5 |
| `--focus` | `-f` | Focus area (maintenance, development, review, triage) | maintenance |
| `--auto-pr` | | Enable automatic pull request enhancements | false |
| `--issue-labels` | | Auto-categorize and label issues | false |
| `--code-review` | | Enable AI-powered code reviews | false |
| `--help` | `-h` | Show help message | false |

## Agent Types

### Issue Triager
- **Purpose**: Analyzes and categorizes issues
- **Capabilities**: 
  - Issue analysis and labeling
  - Priority assignment
  - Duplicate detection
  - Workflow optimization
- **Best for**: Repositories with high issue volume

### PR Reviewer
- **Purpose**: Reviews code changes and suggests improvements
- **Capabilities**:
  - Code quality analysis
  - Best practices enforcement
  - Security review
  - Performance analysis
- **Best for**: Active development repositories

### Documentation Agent
- **Purpose**: Maintains and improves documentation
- **Capabilities**:
  - README updates
  - API documentation
  - Changelog maintenance
  - Wiki management
- **Best for**: Open source projects

### Test Agent
- **Purpose**: Ensures comprehensive test coverage
- **Capabilities**:
  - Test gap identification
  - Coverage analysis
  - Test case suggestions
  - Quality assurance
- **Best for**: Production applications

### Security Agent
- **Purpose**: Identifies and addresses security issues
- **Capabilities**:
  - Vulnerability scanning
  - Dependency auditing
  - Security compliance
  - Risk assessment
- **Best for**: Enterprise applications

## Focus Strategies

### Maintenance Focus
- **Agents**: Issue Triager, Documentation Agent, Security Agent
- **Priority**: Repository health and stability
- **Tasks**: Dependency updates, documentation sync, issue cleanup, security audit
- **Best for**: Mature projects needing ongoing maintenance

### Development Focus
- **Agents**: PR Reviewer, Test Agent, Documentation Agent
- **Priority**: Code quality and development velocity
- **Tasks**: Code review, test coverage, CI/CD optimization, performance analysis
- **Best for**: Active development projects

### Review Focus
- **Agents**: PR Reviewer, Test Agent, Security Agent
- **Priority**: Pull request quality and security
- **Tasks**: PR analysis, code quality checks, security review, test validation
- **Best for**: Projects with frequent contributions

### Triage Focus
- **Agents**: Issue Triager, PR Reviewer, Documentation Agent
- **Priority**: Issue and workflow management
- **Tasks**: Issue categorization, priority assignment, duplicate detection, workflow optimization
- **Best for**: High-traffic repositories

## Configuration

### Environment Variables

```bash
# Required for GitHub API access
export GITHUB_TOKEN="your_github_token"

# Optional: Custom API endpoint
export GITHUB_API_URL="https://api.github.com"
```

### GitHub Token Setup

1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate new token with these scopes:
   - `repo` (for private repositories)
   - `public_repo` (for public repositories)
   - `read:org` (for organization repositories)
   - `write:repo_hook` (for webhook management)

## Examples

### Repository Health Check
```bash
./index.js -r myorg/myrepo -f maintenance -a 5 --issue-labels
```

### Code Quality Review
```bash
./index.js -r myorg/myrepo -f development --code-review --auto-pr
```

### Issue Management
```bash
./index.js -r myorg/myrepo -f triage -a 8 --issue-labels
```

### Security Audit
```bash
./index.js -r myorg/myrepo -f maintenance -a 3
```

## Output

The swarm provides detailed reports including:

- **Agent Deployment Summary** - Which agents were activated
- **Task Completion Status** - What each agent accomplished
- **Actionable Recommendations** - Specific next steps
- **Performance Metrics** - Execution time and efficiency
- **Repository Health Score** - Overall assessment

## Testing

```bash
# Run all tests
npm test

# Run specific test suite
node test-runner.js

# Validate setup
npm run validate
```

## Integration

### Claude Code Integration
```javascript
mcp__claude-flow__github_swarm { 
  repository: "owner/repo", 
  agents: 6, 
  focus: "maintenance",
  features: {
    autoPr: true,
    issueLabels: true,
    codeReview: true
  }
}
```

### CI/CD Integration
```yaml
# GitHub Actions example
- name: Run GitHub Swarm
  run: |
    npx claude-flow github swarm \
      --repository ${{ github.repository }} \
      --focus development \
      --auto-pr \
      --code-review
```

## Troubleshooting

### Common Issues

1. **Rate Limit Exceeded**
   - Solution: Wait for rate limit reset or use authenticated requests

2. **Permission Denied**
   - Solution: Ensure GitHub token has required scopes

3. **Repository Not Found**
   - Solution: Verify repository name and access permissions

4. **Agent Execution Failed**
   - Solution: Check logs for specific error details

### Debug Mode
```bash
DEBUG=github-swarm ./index.js -r owner/repo
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request

## License

MIT License - see LICENSE file for details.
