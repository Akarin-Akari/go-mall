# TDD Examples and Use Cases

This document provides practical examples of using the TDD commands and workflows in Claude Code.

## Quick Start

### Basic TDD Cycle

```bash
# Complete TDD cycle for a new feature
/workflows:tdd-cycle "Add user authentication with JWT tokens"

# Individual phase commands
/tools:tdd-red "Create tests for user registration endpoint"
/tools:tdd-green "Implement minimal user registration"
/tools:tdd-refactor "Optimize user registration code"
```

## Example 1: API Endpoint Development

### Scenario: Building a REST API endpoint for user management

```bash
# Step 1: Start with the RED phase - Write failing tests
/tools:tdd-red "User CRUD operations with validation"
# This will generate:
# - GET /users/:id tests
# - POST /users tests with validation
# - PUT /users/:id tests
# - DELETE /users/:id tests
# All tests will fail initially (RED)

# Step 2: GREEN phase - Minimal implementation
/tools:tdd-green "Make user CRUD tests pass"
# Implements just enough code to pass tests:
# - Basic route handlers
# - Minimal validation
# - Simple database operations

# Step 3: REFACTOR phase - Improve code quality
/tools:tdd-refactor "Optimize user CRUD implementation"
# Refactors while keeping tests green:
# - Extract validation middleware
# - Add proper error handling
# - Implement repository pattern
# - Add logging and monitoring
```

## Example 2: Frontend Component Development

### Scenario: React component with TDD

```bash
# Use the complete workflow for component development
/workflows:tdd-cycle --strict-tdd --framework=react "
Build a SearchBar component with:
- Real-time search as you type
- Debouncing for 300ms
- Loading state during search
- Error handling
- Keyboard navigation for results
"

# The workflow will:
# 1. Generate comprehensive component tests
# 2. Verify tests fail appropriately
# 3. Implement minimal component
# 4. Add features incrementally
# 5. Refactor for performance and readability
```

## Example 3: Algorithm Implementation

### Scenario: Implementing a complex sorting algorithm

```bash
# Phase 1: Define behavior through tests
/tools:tdd-red "
Quick sort algorithm with:
- Handle empty arrays
- Handle single element
- Handle already sorted
- Handle reverse sorted
- Handle duplicates
- Handle large arrays (10000+ elements)
"

# Phase 2: Progressive implementation
/tools:tdd-green --incremental "
Implement quick sort progressively:
1. Handle base cases first
2. Basic partitioning
3. Recursive sorting
4. Optimization for duplicates
"

# Phase 3: Performance optimization
/tools:tdd-refactor --focus=performance "
Optimize quick sort:
- Three-way partitioning for duplicates
- Insertion sort for small subarrays
- Tail call optimization
"
```

## Example 4: Database Migration with TDD

### Scenario: Database schema changes with zero downtime

```bash
# TDD approach to database migrations
/workflows:tdd-cycle --domain=database "
Migrate user table to support multiple email addresses:
- Maintain backward compatibility
- Zero downtime migration
- Data integrity preservation
- Rollback capability
"

# This generates:
# 1. Migration tests including rollback scenarios
# 2. Staged migration implementation
# 3. Compatibility layer for existing code
# 4. Performance tests for new schema
```

## Example 5: Microservice Development

### Scenario: Building a microservice from scratch

```bash
# Complete microservice with TDD
/workflows:tdd-cycle --type=microservice --framework=fastapi "
Payment processing microservice:
- Process credit card payments
- Handle webhooks from payment provider
- Implement retry logic
- Store transaction history
- Generate audit logs
"

# Alternative: Step-by-step approach
/tools:tdd-red "Payment processing service contract tests"
/tools:tdd-red "Payment validation unit tests"
/tools:tdd-red "Webhook handling integration tests"
/tools:tdd-green "Implement payment service core"
/tools:tdd-green "Add webhook handlers"
/tools:tdd-refactor "Apply hexagonal architecture"
```

## Example 6: Legacy Code Refactoring

### Scenario: Adding tests to existing code before refactoring

```bash
# TDD approach to legacy code
/workflows:tdd-cycle --legacy-mode "
Refactor legacy OrderProcessor class:
- Currently 500+ lines
- No existing tests
- Mixed responsibilities
- Direct database access
"

# The workflow will:
# 1. Generate characterization tests for current behavior
# 2. Add tests incrementally for each method
# 3. Refactor once tests are in place
# 4. Extract smaller, focused classes
# 5. Introduce dependency injection
```

## Example 7: Performance-Critical Code

### Scenario: TDD for performance requirements

```bash
# Performance TDD
/workflows:tdd-cycle --performance-tdd "
Image processing pipeline:
- Process 1000 images per second
- Memory usage under 500MB
- Support JPEG, PNG, WebP
- Maintain quality above 95%
"

# Generates performance tests first:
# - Throughput benchmarks
# - Memory profiling tests
# - Quality measurement tests
# Then implements with performance in mind
```

## Example 8: Cross-Platform Library

### Scenario: Building a library that works across multiple platforms

```bash
# Multi-platform TDD
/workflows:tdd-cycle --platforms="node,browser,deno" "
Build a date formatting library:
- Support for 15+ locales
- Timezone handling
- Relative time formatting
- Custom format strings
- Tree-shakeable
"
```

## Advanced TDD Patterns

### Property-Based TDD

```bash
/tools:tdd-red --style=property-based "
String manipulation library with properties:
- Reversing twice returns original
- Concatenation is associative
- Trim removes only whitespace
- Case conversion is idempotent
"
```

### Contract-First TDD

```bash
/tools:tdd-red --style=contract-first --spec=openapi.yaml "
Generate tests from OpenAPI specification
"
```

### Acceptance Test-Driven Development (ATDD)

```bash
/workflows:tdd-cycle --style=atdd "
Feature: User can reset password
Given: User has forgotten password
When: User requests password reset
Then: Email with reset link is sent
And: Link expires after 24 hours
"
```

### Behavior-Driven Development (BDD) with TDD

```bash
/workflows:tdd-cycle --style=bdd --format=gherkin "
Feature: Shopping cart management
  Scenario: Add item to cart
    Given I have an empty cart
    When I add a product with quantity 2
    Then cart should contain 1 item
    And total quantity should be 2
"
```

## TDD Metrics and Reporting

### Generate TDD metrics for existing project

```bash
# Analyze TDD compliance
/workflows:full-review --tdd-review --tdd-metrics "
Analyze current codebase for TDD patterns
"

# Generate TDD adoption report
/tools:tdd-refactor --analyze-only --report "
Generate TDD metrics report for last 30 days
"
```

## Integration with CI/CD

### GitHub Actions Integration

```yaml
# .github/workflows/tdd.yml
name: TDD Workflow
on: [push, pull_request]

jobs:
  tdd-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run TDD validation
        run: |
          # Check if tests were written before code
          claude-code /workflows:full-review --tdd-review --strict-tdd
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Ensure tests exist for new code
claude-code /tools:tdd-red --check-only

# Verify all tests pass
claude-code /tools:tdd-green --verify-only

# Check for refactoring opportunities
claude-code /tools:tdd-refactor --suggest-only
```

## Common TDD Scenarios

### Starting a New Project

```bash
# Initialize project with TDD structure
/workflows:tdd-cycle --init-project "
Web application with:
- User authentication
- REST API
- PostgreSQL database
- Redis caching
"
```

### Adding Feature to Existing Project

```bash
# TDD for new feature
/workflows:feature-development --tdd "
Add real-time notifications using WebSockets
"
```

### Bug Fix with TDD

```bash
# Fix bug using TDD approach
/tools:tdd-red "
Reproduce bug: User cannot login with email containing '+'
Expected: Login successful
Actual: Invalid email error
"
/tools:tdd-green "Fix email validation regex"
/tools:tdd-refactor "Improve email validation logic"
```

### Performance Improvement

```bash
# TDD for performance optimization
/tools:tdd-red --performance "
Test: API response time should be < 100ms
Current: 500ms average
Target: 50ms p99
"
```

## Best Practices

1. **Start Small**: Begin with simple test cases and gradually add complexity
2. **One Test at a Time**: Focus on making one test pass before writing the next
3. **Keep Tests Fast**: Aim for test suites that run in seconds, not minutes
4. **Test Behavior, Not Implementation**: Focus on what the code does, not how
5. **Refactor Regularly**: Use the green phase as an opportunity to improve
6. **Maintain Test Quality**: Tests are code too - keep them clean and maintainable

## Troubleshooting

### Tests Pass Too Easily

```bash
# Verify tests are actually testing something
/tools:tdd-red --mutation-testing "Verify test effectiveness"
```

### Tests Are Too Slow

```bash
# Optimize test performance
/tools:tdd-refactor --focus=test-performance "Speed up test suite"
```

### Too Many Mocks

```bash
# Refactor to reduce mocking
/tools:tdd-refactor --reduce-mocks "Simplify test dependencies"
```

## Tips for Success

- Use `--strict-tdd` flag to enforce discipline
- Set coverage thresholds with `--test-coverage-min=90`
- Use `--incremental` for step-by-step development
- Enable metrics with `--tdd-metrics` to track progress
- Use `--style=chicago` or `--style=london` for specific TDD approaches

## Next Steps

- Review the [TDD workflow documentation](/workflows:tdd-cycle --help)
- Explore individual phase commands (`tdd-red`, `tdd-green`, `tdd-refactor`)
- Check TDD metrics with `/workflows:full-review --tdd-metrics`
- Join TDD practice sessions with `/tools:tdd-red --kata`