# Code Improvement (/sc:improve) - Systematic Code Enhancement

Apply systematic improvements to code quality, performance, and maintainability using multi-persona coordination and domain-specific expertise.

## Overview

The Code Improvement system provides comprehensive code analysis and enhancement capabilities across four key areas:
- **Quality** - Code structure, readability, technical debt reduction
- **Performance** - Optimization, bottleneck resolution, efficiency improvements
- **Maintainability** - Documentation, complexity reduction, modularity enhancement
- **Security** - Vulnerability fixes, security pattern application

## Features

- **Multi-Persona Coordination** - Architect, Performance Expert, Quality Engineer, Security Specialist
- **Framework-Specific Optimization** - Context7 integration for best practices
- **Systematic Analysis** - Sequential MCP for complex multi-component improvements
- **Safe Refactoring** - Comprehensive validation and rollback capabilities
- **Interactive Mode** - Guided improvement decisions for complex scenarios
- **Preview Mode** - Review changes before application

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
# Quality enhancement with safe refactoring
./index.js src/ --type quality --safe

# Performance optimization with interactive guidance
./index.js api-endpoints --type performance --interactive

# Maintainability improvements with preview
./index.js legacy-modules --type maintainability --preview

# Security hardening with validation
./index.js auth-service --type security --validate
```

### Command Options

| Option | Description | Default |
|--------|-------------|---------|
| `target` | Target file, directory, or pattern | current directory |
| `--type <type>` | Improvement type (quality, performance, maintainability, security) | quality |
| `--safe` | Apply only safe improvements with rollback capability | false |
| `--interactive` | Interactive mode for complex improvement decisions | false |
| `--preview` | Preview changes before application | false |
| `--validate` | Comprehensive validation after improvements | false |
| `--help` | Show help message | false |

## Improvement Types

### Quality Improvements
- **Code Structure** - Extract methods, simplify conditionals, reduce duplication
- **Readability** - Improve naming conventions, add meaningful comments
- **Technical Debt** - Refactor legacy code, modernize patterns
- **Best Practices** - Apply coding standards and conventions

### Performance Optimizations
- **Algorithm Optimization** - Improve loop efficiency, reduce complexity
- **Caching** - Add intelligent caching for expensive operations
- **Database Optimization** - Query optimization, indexing suggestions
- **Memory Management** - Reduce memory usage, prevent leaks

### Maintainability Enhancements
- **Documentation** - Add comprehensive comments and documentation
- **Modularity** - Break down large functions, improve separation of concerns
- **Error Handling** - Implement robust error handling patterns
- **Test Coverage** - Identify areas needing test coverage

### Security Hardening
- **Input Validation** - Sanitize and validate all inputs
- **Authentication** - Implement secure authentication patterns
- **Data Protection** - Encrypt sensitive data, secure storage
- **Vulnerability Fixes** - Address known security issues

## Personas

### Software Architect
- **Focus**: Structure, design patterns, modularity
- **Expertise**: Architecture, design patterns, separation of concerns
- **Improvements**: Code organization, design pattern application, modularity enhancement

### Performance Expert
- **Focus**: Speed optimization, bottleneck resolution
- **Expertise**: Algorithms, caching, database optimization, memory management
- **Improvements**: Performance optimization, caching strategies, efficiency improvements

### Quality Engineer
- **Focus**: Code quality, maintainability, readability
- **Expertise**: Clean code, refactoring, testing, documentation
- **Improvements**: Code quality enhancement, refactoring, documentation

### Security Specialist
- **Focus**: Vulnerability analysis, security patterns
- **Expertise**: Security patterns, input validation, authentication, encryption
- **Improvements**: Security hardening, vulnerability fixes, secure coding practices

## Modes

### Safe Mode (`--safe`)
- Creates automatic backups before applying changes
- Provides rollback capabilities
- Conservative improvement application
- Comprehensive validation before changes

### Interactive Mode (`--interactive`)
- Prompts for approval before each improvement
- Provides detailed impact analysis
- Allows selective improvement application
- Guided decision making for complex scenarios

### Preview Mode (`--preview`)
- Shows proposed changes without applying them
- Detailed analysis of potential improvements
- Impact assessment and recommendations
- No actual file modifications

### Validation Mode (`--validate`)
- Comprehensive testing after improvements
- Quality metric verification
- Performance impact assessment
- Rollback if validation fails

## Output Examples

### Quality Improvement
```
🔧 Code Improvement Results
━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Improvement Summary
├── Files Analyzed: 15
├── Issues Found: 8
├── Improvements Applied: 6
├── Files Modified: 4
└── Execution Time: 1,245ms

👥 Persona Analysis
├── Quality Engineer: 5 issues, 4 improvements
├── Software Architect: 3 issues, 2 improvements

✅ Applied Improvements
1. 🟡 Extract method to reduce complexity
   └── File: src/utils.js (refactoring)
   └── Impact: Major code quality improvement
2. 🟢 Extract magic numbers to named constants
   └── File: src/config.js (readability)
   └── Impact: Code readability improved

📈 Quality Metrics
├── Code Quality Score: 92/100
├── Maintainability Index: 85/100
├── Technical Debt Reduction: 25%
└── Performance Improvement: 0%
```

### Performance Optimization
```
✅ Applied Improvements
1. 🔴 Optimize nested loops for better performance
   └── File: src/processor.js (performance)
   └── Impact: Significant performance improvement expected
2. 🟡 Add caching for database queries
   └── File: src/database.js (performance)
   └── Impact: Moderate performance improvement

📈 Quality Metrics
├── Performance Improvement: 35%
├── Memory Usage Reduction: 20%
├── Database Query Optimization: 40%
└── Cache Hit Rate: 85%
```

## Priority Levels

| Icon | Priority | Description |
|------|----------|-------------|
| 🔴 | High | Critical issues requiring immediate attention |
| 🟡 | Medium | Important improvements with moderate impact |
| 🟢 | Low | Minor improvements for code polish |

## Integration

### Claude Code Integration
```javascript
// Trigger code improvement
/sc:improve src/ --type quality --safe

// Performance optimization
/sc:improve api-endpoints --type performance --interactive

// Security hardening
/sc:improve auth-service --type security --validate
```

### Programmatic Usage
```javascript
const CodeImproveCLI = require('./index');

const cli = new CodeImproveCLI();
const result = await cli.run(['src/', '--type', 'quality', '--safe']);
```

## Best Practices

### Before Running Improvements
1. **Backup Your Code** - Always use version control or `--safe` mode
2. **Start Small** - Begin with single files or small directories
3. **Choose Appropriate Type** - Select the most relevant improvement type
4. **Use Preview Mode** - Review changes before applying them

### During Improvements
1. **Review Suggestions** - Use interactive mode for complex decisions
2. **Monitor Progress** - Watch for any unexpected issues
3. **Validate Changes** - Use `--validate` for critical code
4. **Test Thoroughly** - Run tests after improvements

### After Improvements
1. **Run Tests** - Ensure functionality is preserved
2. **Review Changes** - Manually review all modifications
3. **Monitor Performance** - Check for performance impacts
4. **Document Changes** - Update documentation as needed

## Troubleshooting

### Common Issues

1. **No Improvements Found**
   - Code may already be well-optimized
   - Try different improvement types
   - Lower quality thresholds if needed

2. **Validation Failures**
   - Review applied changes manually
   - Use rollback if available
   - Run tests to identify issues

3. **Performance Degradation**
   - Some optimizations may have trade-offs
   - Monitor metrics after changes
   - Consider reverting specific improvements

### Debug Mode
```bash
DEBUG=code-improve ./index.js src/ --type quality
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request

## License

MIT License - see LICENSE file for details.
