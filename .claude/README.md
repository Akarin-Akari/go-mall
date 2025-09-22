# Claude Code Slash Commands

A comprehensive collection of production-ready slash commands for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) that provides intelligent automation and multi-agent orchestration capabilities for modern software development.

## Overview

This repository provides **56 production-ready slash commands** (15 workflows, 41 tools) that extend Claude Code's capabilities through:

- **Workflows**: Multi-agent orchestration systems that coordinate complex, multi-step operations across different domains
- **Tools**: Specialized single-purpose utilities for focused development tasks

## System Requirements

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) installed and configured
- [Claude Code Subagents](https://github.com/wshobson/agents) collection for workflow orchestration capabilities
- Git for repository management

## Installation

```bash
# Navigate to Claude configuration directory
cd ~/.claude

# Clone the commands repository
git clone https://github.com/wshobson/commands.git

# Clone the agents repository (required for workflow execution)
git clone https://github.com/wshobson/agents.git
```

## Command Invocation

Commands are organized in `tools/` and `workflows/` directories and invoked using directory prefixes:

```bash
# Workflow invocation
/workflows:feature-development implement OAuth2 authentication

# Tool invocation
/tools:security-scan perform vulnerability assessment

# Multiple argument example
/tools:api-scaffold create user management endpoints with RBAC
```

### Alternative Setup (No Prefixes)

To invoke commands without directory prefixes, copy files to the root directory:

```bash
cp tools/*.md .
cp workflows/*.md .

# Then invoke directly
/api-scaffold create REST endpoints
/feature-development implement payment system
```

## Command Architecture

### Workflows (15 commands)

Workflows implement multi-agent orchestration patterns for complex, cross-domain tasks. Each workflow analyzes requirements, delegates to specialized agents, and coordinates execution across multiple subsystems.

#### Core Development Workflows

| Command | Purpose | Agent Coordination |
|---------|---------|-------------------|
| `feature-development` | End-to-end feature implementation | Backend, frontend, testing, deployment |
| `full-review` | Multi-perspective code analysis | Architecture, security, performance, quality |
| `smart-fix` | Intelligent problem resolution | Dynamic agent selection based on issue type |
| `tdd-cycle` | Test-driven development orchestration | Test writer, implementer, refactoring specialist |

#### Process Automation Workflows

| Command | Purpose | Scope |
|---------|---------|-------|
| `git-workflow` | Version control process automation | Branching strategies, commit standards, PR templates |
| `improve-agent` | Agent optimization | Prompt engineering, performance tuning |
| `legacy-modernize` | Codebase modernization | Architecture migration, dependency updates, pattern refactoring |
| `ml-pipeline` | Machine learning pipeline construction | Data engineering, model training, deployment |
| `multi-platform` | Cross-platform development | Web, mobile, desktop coordination |
| `workflow-automate` | CI/CD pipeline automation | Build, test, deploy, monitor |

#### Advanced Orchestration Workflows

| Command | Primary Focus | Specialized Agents |
|---------|---------------|-------------------|
| `full-stack-feature` | Multi-tier implementation | Backend API, frontend UI, mobile, database |
| `security-hardening` | Security-first development | Threat modeling, vulnerability assessment, remediation |
| `data-driven-feature` | ML-powered functionality | Data science, feature engineering, model deployment |
| `performance-optimization` | System-wide optimization | Profiling, caching, query optimization, load testing |
| `incident-response` | Production issue resolution | Diagnostics, root cause analysis, hotfix deployment |

### Tools (41 commands)

Tools provide focused, single-purpose utilities for specific development operations. Each tool is optimized for its domain with production-ready implementations.

#### AI and Machine Learning (5 tools)

| Command | Functionality | Key Features |
|---------|--------------|--------------|
| `ai-assistant` | AI assistant implementation | LLM integration, conversation management, context handling |
| `ai-review` | ML code review | Model architecture validation, training pipeline review |
| `langchain-agent` | LangChain agent creation | RAG patterns, tool integration, memory management |
| `ml-pipeline` | ML pipeline construction | Data processing, training, evaluation, deployment |
| `prompt-optimize` | Prompt engineering | Performance testing, cost optimization, quality metrics |

#### Architecture and Code Quality (4 tools)

| Command | Purpose | Capabilities |
|---------|---------|--------------|
| `code-explain` | Code documentation | AST analysis, complexity metrics, flow diagrams |
| `code-migrate` | Migration automation | Framework upgrades, language porting, API migrations |
| `refactor-clean` | Code improvement | Pattern detection, dead code removal, structure optimization |
| `tech-debt` | Debt assessment | Complexity analysis, risk scoring, remediation planning |

#### Data and Database (3 tools)

| Command | Focus Area | Technologies |
|---------|------------|--------------|
| `data-pipeline` | ETL/ELT architecture | Apache Spark, Airflow, dbt, streaming platforms |
| `data-validation` | Data quality | Schema validation, anomaly detection, constraint checking |
| `db-migrate` | Database migrations | Schema versioning, zero-downtime strategies, rollback plans |

#### DevOps and Infrastructure (6 tools)

| Command | Domain | Implementation |
|---------|--------|----------------|
| `deploy-checklist` | Deployment preparation | Pre-flight checks, rollback procedures, monitoring setup |
| `docker-optimize` | Container optimization | Multi-stage builds, layer caching, size reduction |
| `k8s-manifest` | Kubernetes configuration | Deployments, services, ingress, autoscaling, security policies |
| `monitor-setup` | Observability | Metrics, logging, tracing, alerting rules |
| `slo-implement` | SLO/SLI definition | Error budgets, monitoring, automated responses |
| `workflow-automate` | Pipeline automation | CI/CD, GitOps, infrastructure as code |

#### Testing and Development (6 tools)

| Command | Testing Focus | Framework Support |
|---------|---------------|-------------------|
| `api-mock` | Mock generation | REST, GraphQL, gRPC, WebSocket |
| `api-scaffold` | Endpoint creation | CRUD operations, authentication, validation |
| `test-harness` | Test suite generation | Unit, integration, e2e, performance |
| `tdd-red` | Test-first development | Failing test creation, edge case coverage |
| `tdd-green` | Implementation | Minimal code to pass tests |
| `tdd-refactor` | Code improvement | Optimization while maintaining green tests |

#### Security and Compliance (3 tools)

| Command | Security Domain | Standards |
|---------|-----------------|-----------|
| `accessibility-audit` | WCAG compliance | ARIA, keyboard navigation, screen reader support |
| `compliance-check` | Regulatory compliance | GDPR, HIPAA, SOC2, PCI-DSS |
| `security-scan` | Vulnerability assessment | OWASP, CVE scanning, dependency audits |

#### Debugging and Analysis (4 tools)

| Command | Analysis Type | Output |
|---------|---------------|--------|
| `debug-trace` | Runtime analysis | Stack traces, memory profiles, execution paths |
| `error-analysis` | Error patterns | Root cause analysis, frequency analysis, impact assessment |
| `error-trace` | Production debugging | Log correlation, distributed tracing, error reproduction |
| `issue` | Issue tracking | Standardized templates, reproduction steps, acceptance criteria |

#### Dependency and Configuration Management (3 tools)

| Command | Management Area | Features |
|---------|-----------------|----------|
| `config-validate` | Configuration management | Schema validation, environment variables, secrets handling |
| `deps-audit` | Dependency analysis | Security vulnerabilities, license compliance, version conflicts |
| `deps-upgrade` | Version management | Breaking change detection, compatibility testing, rollback support |

#### Documentation and Collaboration (3 tools)

| Command | Documentation Type | Format |
|---------|-------------------|--------|
| `doc-generate` | API documentation | OpenAPI, JSDoc, TypeDoc, Sphinx |
| `pr-enhance` | Pull request optimization | Description generation, checklist creation, review suggestions |
| `standup-notes` | Status reporting | Progress tracking, blocker identification, next steps |

#### Operations and Context (4 tools)

| Command | Operational Focus | Use Case |
|---------|------------------|----------|
| `cost-optimize` | Resource optimization | Cloud spend analysis, right-sizing, reserved capacity |
| `onboard` | Environment setup | Development tools, access configuration, documentation |
| `context-save` | State persistence | Architecture decisions, configuration snapshots |
| `context-restore` | State recovery | Context reload, decision history, configuration restore |

## Usage Patterns

### Common Development Scenarios

#### Feature Implementation
```bash
# Complete feature with multi-agent orchestration
/workflows:feature-development OAuth2 authentication with JWT tokens

# API-first development
/tools:api-scaffold REST endpoints for user management with RBAC

# Test-driven approach
/workflows:tdd-cycle shopping cart with discount calculation logic
```

#### Debugging and Performance
```bash
# Intelligent issue resolution
/workflows:smart-fix high memory consumption in production workers

# Targeted error analysis
/tools:error-trace investigate Redis connection timeouts

# Performance optimization
/workflows:performance-optimization optimize database query performance
```

#### Security and Compliance
```bash
# Security assessment
/tools:security-scan OWASP Top 10 vulnerability scan

# Compliance verification
/tools:compliance-check GDPR data handling requirements

# Security hardening workflow
/workflows:security-hardening implement zero-trust architecture
```

### Test-Driven Development

#### Standard TDD Flow
```bash
# Complete TDD cycle with orchestration
/workflows:tdd-cycle payment processing with Stripe integration

# Manual TDD phases for granular control
/tools:tdd-red create failing tests for order validation
/tools:tdd-green implement minimal order validation logic
/tools:tdd-refactor optimize validation performance
```

### Command Composition Strategies

#### Sequential Execution Pattern
```bash
# Feature implementation pipeline
/workflows:feature-development real-time notifications with WebSockets
/tools:security-scan WebSocket implementation vulnerabilities
/workflows:performance-optimization WebSocket connection handling
/tools:deploy-checklist notification service deployment requirements
/tools:k8s-manifest WebSocket service with session affinity
```

#### Modernization Pipeline
```bash
# Legacy system upgrade
/workflows:legacy-modernize migrate monolith to microservices
/tools:deps-audit check dependency vulnerabilities
/tools:deps-upgrade update to latest stable versions
/tools:refactor-clean remove deprecated patterns
/tools:test-harness generate comprehensive test coverage
/tools:docker-optimize create optimized container images
/tools:k8s-manifest deploy with rolling update strategy
```

## Command Selection Guidelines

### Workflow vs Tool Decision Matrix

| Criteria | Use Workflows | Use Tools |
|----------|--------------|-----------|
| **Problem Complexity** | Multi-domain, cross-cutting concerns | Single domain, focused scope |
| **Solution Clarity** | Exploratory, undefined approach | Clear implementation path |
| **Agent Coordination** | Multiple specialists required | Single expertise sufficient |
| **Implementation Scope** | End-to-end features | Specific components |
| **Control Level** | Automated orchestration preferred | Manual control required |

### Workflow Selection Examples

| Requirement | Recommended Workflow | Rationale |
|-------------|---------------------|-----------|
| "Build complete authentication system" | `/workflows:feature-development` | Multi-tier implementation required |
| "Debug production performance issues" | `/workflows:smart-fix` | Unknown root cause, needs analysis |
| "Modernize legacy application" | `/workflows:legacy-modernize` | Complex refactoring across stack |
| "Implement ML-powered feature" | `/workflows:data-driven-feature` | Requires data science expertise |

### Tool Selection Examples

| Task | Recommended Tool | Output |
|------|-----------------|--------|
| "Generate Kubernetes configs" | `/tools:k8s-manifest` | YAML manifests with best practices |
| "Audit security vulnerabilities" | `/tools:security-scan` | Vulnerability report with fixes |
| "Create API documentation" | `/tools:doc-generate` | OpenAPI/Swagger specifications |
| "Optimize Docker images" | `/tools:docker-optimize` | Multi-stage Dockerfile |

## Execution Best Practices

### Context Optimization

1. **Technology Stack Specification**: Include framework versions, database systems, deployment targets
2. **Constraint Definition**: Specify performance requirements, security standards, compliance needs
3. **Integration Requirements**: Define external services, APIs, authentication methods
4. **Output Preferences**: Indicate coding standards, testing frameworks, documentation formats

### Command Chaining Strategies

1. **Progressive Enhancement**: Start with workflows for foundation, refine with tools
2. **Pipeline Construction**: Chain commands in logical sequence for complete solutions
3. **Iterative Refinement**: Use tool outputs as inputs for subsequent commands
4. **Parallel Execution**: Run independent tools simultaneously when possible

### Performance Considerations

- Workflows typically require 30-90 seconds for complete orchestration
- Tools execute in 5-30 seconds for focused operations
- Provide detailed requirements upfront to minimize iteration cycles
- Use saved context (`context-save`/`context-restore`) for multi-session projects

## Technical Architecture

### Command Structure

Each slash command is a markdown file with the following characteristics:

| Component | Description | Example |
|-----------|-------------|---------|
| **Filename** | Determines command name | `api-scaffold.md` → `/tools:api-scaffold` |
| **Content** | Execution instructions | Agent prompts and orchestration logic |
| **Variables** | `$ARGUMENTS` placeholder | Captures and processes user input |
| **Directory** | Command category | `tools/` for utilities, `workflows/` for orchestration |

### File Organization

```
~/.claude/commands/
├── workflows/          # Multi-agent orchestration commands
│   ├── feature-development.md
│   ├── smart-fix.md
│   └── ...
├── tools/             # Single-purpose utility commands
│   ├── api-scaffold.md
│   ├── security-scan.md
│   └── ...
└── README.md          # This documentation
```

## Development Guidelines

### Creating New Commands

#### Workflow Development

1. **File Creation**: Place in `workflows/` directory with descriptive naming
2. **Agent Orchestration**: Define delegation logic for multiple specialists
3. **Error Handling**: Include fallback strategies and error recovery
4. **Output Coordination**: Specify how agent outputs should be combined

#### Tool Development

1. **File Creation**: Place in `tools/` directory with single-purpose naming
2. **Implementation**: Provide complete, production-ready code generation
3. **Framework Detection**: Auto-detect and adapt to project stack
4. **Best Practices**: Include security, performance, and scalability considerations

### Naming Conventions

- Use lowercase with hyphens: `feature-name.md`
- Be descriptive but concise: `security-scan` not `scan`
- Indicate action clearly: `deps-upgrade` not `dependencies`
- Maintain consistency with existing commands

## Troubleshooting Guide

### Diagnostic Steps

| Issue | Cause | Resolution |
|-------|-------|------------|
| Command not recognized | File missing or misnamed | Verify file exists in correct directory |
| Slow execution | Normal workflow behavior | Workflows coordinate multiple agents (30-90s typical) |
| Incomplete output | Insufficient context | Provide technology stack and requirements |
| Integration failures | Path or configuration issues | Check file paths and dependencies |

### Performance Optimization

1. **Context Caching**: Use `context-save` for multi-session projects
2. **Batch Operations**: Combine related tasks in single workflow
3. **Tool Selection**: Use tools for known problems, workflows for exploration
4. **Requirement Clarity**: Detailed specifications reduce iteration cycles

## Featured Command Implementations

### Test-Driven Development Suite

| Command | Type | Capabilities |
|---------|------|--------------|
| `tdd-cycle` | Workflow | Complete red-green-refactor orchestration with test coverage analysis |
| `tdd-red` | Tool | Failing test generation with edge case coverage and mocking |
| `tdd-green` | Tool | Minimal implementation to achieve test passage |
| `tdd-refactor` | Tool | Code optimization while maintaining test integrity |

**Framework Support**: Jest, Mocha, PyTest, RSpec, JUnit, Go testing, Rust tests

### Security and Infrastructure

| Command | Specialization | Key Features |
|---------|---------------|--------------|
| `security-scan` | Vulnerability detection | SAST/DAST analysis, dependency scanning, secret detection |
| `docker-optimize` | Container optimization | Multi-stage builds, layer caching, size reduction (50-90% typical) |
| `k8s-manifest` | Kubernetes deployment | HPA, NetworkPolicy, PodSecurityPolicy, service mesh ready |
| `monitor-setup` | Observability | Prometheus metrics, Grafana dashboards, alert rules |

**Security Tools Integration**: Bandit, Safety, Trivy, Semgrep, Snyk, GitGuardian

### Data and Database Operations

| Command | Database Support | Migration Strategies |
|---------|-----------------|---------------------|
| `db-migrate` | PostgreSQL, MySQL, MongoDB, DynamoDB | Blue-green, expand-contract, versioned schemas |
| `data-pipeline` | Batch and streaming | Apache Spark, Kafka, Airflow, dbt integration |
| `data-validation` | Schema and quality | Great Expectations, Pandera, custom validators |

**Zero-Downtime Patterns**: Rolling migrations, feature flags, dual writes, backfill strategies

### Performance and Optimization

| Command | Analysis Type | Optimization Techniques |
|---------|--------------|------------------------|
| `performance-optimization` | Full-stack profiling | Query optimization, caching strategies, CDN configuration |
| `cost-optimize` | Cloud resource analysis | Right-sizing, spot instances, reserved capacity planning |
| `docker-optimize` | Container performance | Build cache optimization, minimal base images, layer reduction |

## Additional Resources

### Documentation

- [Claude Code Official Documentation](https://docs.anthropic.com/en/docs/claude-code)
- [Slash Commands Reference](https://docs.anthropic.com/en/docs/claude-code/slash-commands)
- [Subagents Architecture](https://docs.anthropic.com/en/docs/claude-code/sub-agents)

### Source Repositories

- [Command Collection](https://github.com/wshobson/commands)
- [Agent Collection](https://github.com/wshobson/agents)
- [Claude Code Repository](https://github.com/anthropics/claude-code)

### Integration Examples

```bash
# Complete feature development pipeline
/workflows:feature-development user authentication system
/tools:security-scan authentication implementation
/tools:test-harness authentication test suite
/tools:docker-optimize authentication service
/tools:k8s-manifest authentication deployment
/tools:monitor-setup authentication metrics
```

## License

MIT License - See LICENSE file for complete terms.

## Support and Contribution

- **Issues**: [GitHub Issues](https://github.com/wshobson/commands/issues)
- **Contributions**: Pull requests welcome following the development guidelines
- **Questions**: Open a discussion in the repository