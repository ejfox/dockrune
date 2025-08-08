# Contributing to dockrune

Thank you for your interest in contributing to dockrune! This guide will help you get started with development and understand our contribution process.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Style](#code-style)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)

## Code of Conduct

This project follows a simple principle: **Be excellent to each other**. We welcome contributions from everyone and are committed to providing a friendly, safe, and inclusive environment for all contributors.

## Getting Started

### Prerequisites

- Go 1.21+ 
- Node.js 20+
- Docker and Docker Compose
- Git

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/dockrune.git
   cd dockrune
   ```

2. **Install Dependencies**
   ```bash
   # Install Go dependencies
   go mod download
   
   # Install dashboard dependencies
   cd dashboard
   npm install
   cd ..
   ```

3. **Set Up Environment**
   ```bash
   cp .env.example .env.dev
   # Edit .env.dev with your development settings
   ```

4. **Run Development Servers**
   ```bash
   # Terminal 1: Run the Go backend
   make dev
   
   # Terminal 2: Run the Nuxt dashboard
   cd dashboard
   npm run dev
   ```

5. **Verify Setup**
   ```bash
   # Run tests
   make test
   
   # Check code formatting
   make fmt
   make lint
   ```

## Project Structure

```
dockrune/
├── cmd/dockrune/           # CLI entry point
│   └── main.go
├── internal/               # Go application code
│   ├── admin/             # Admin API server
│   ├── alerting/          # Notification systems
│   ├── cmd/               # CLI commands
│   ├── config/            # Configuration management
│   ├── deployer/          # Deployment orchestration
│   ├── detector/          # Project type detection
│   ├── github/            # GitHub API integration
│   ├── models/            # Data models
│   ├── storage/           # Database layer
│   └── webhook/           # Webhook server
├── dashboard/             # Nuxt 3 dashboard
│   ├── components/        # Vue components
│   ├── pages/            # Dashboard pages
│   ├── stores/           # Pinia state management
│   └── middleware/       # Auth middleware
├── docs/                 # Documentation
├── scripts/              # Build and utility scripts
└── __tests__/           # Integration tests
```

### Key Components

- **Webhook Server** (`internal/webhook/`): Receives and processes GitHub webhooks
- **Detector** (`internal/detector/`): Auto-detects project types from repository contents
- **Deployer** (`internal/deployer/`): Orchestrates deployment processes
- **Admin API** (`internal/admin/`): Provides REST API for dashboard
- **Dashboard** (`dashboard/`): Modern Nuxt 3 web interface
- **Storage** (`internal/storage/`): SQLite database operations

## Development Workflow

### 1. Choose or Create an Issue

- Check existing issues for bugs or features you'd like to work on
- Create a new issue if you have a new idea or found a bug
- Comment on the issue to let others know you're working on it

### 2. Create a Feature Branch

```bash
git checkout main
git pull origin main
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/feature-name` for new features
- `fix/bug-description` for bug fixes  
- `docs/documentation-update` for documentation
- `refactor/component-name` for refactoring

### 3. Make Your Changes

Follow these guidelines:

**Go Code:**
- Follow Go conventions and idioms
- Keep functions focused and testable
- Add appropriate error handling
- Include comprehensive logging

**Dashboard Code:**
- Use TypeScript where possible
- Follow Vue 3 Composition API patterns
- Keep components focused and reusable
- Add proper TypeScript types

**Database Changes:**
- Ensure schema changes are backwards compatible
- Include migration logic if needed
- Update model definitions

### 4. Write Tests

```bash
# Run Go tests
go test ./...

# Run specific package tests
go test ./internal/detector/

# Run tests with coverage
go test -cover ./...

# Run integration tests  
./scripts/integration-test.sh
```

### 5. Update Documentation

- Update relevant documentation in `docs/`
- Add inline comments for complex logic
- Update API documentation for endpoint changes
- Update README if needed

## Testing

### Go Tests

Write tests for new functionality:

```go
func TestDetectProjectType(t *testing.T) {
    tests := []struct {
        name     string
        files    []string
        expected string
    }{
        {
            name:     "detects Go project",
            files:    []string{"go.mod", "main.go"},
            expected: "go",
        },
        // Add more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := detectProjectType(tt.files)
            if result != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, result)
            }
        })
    }
}
```

### Integration Tests

Test the complete deployment flow:

```bash
# Create test repository
./scripts/create-test-repo.sh

# Run full integration test
./scripts/integration-test.sh

# Clean up test resources
./scripts/cleanup-test.sh
```

### Dashboard Tests

```bash
cd dashboard

# Run component tests  
npm run test

# Run e2e tests
npm run test:e2e
```

## Code Style

### Go Code Style

- Follow `gofmt` formatting
- Use `golint` and `go vet`  
- Keep line length reasonable (100-120 characters)
- Use meaningful variable and function names
- Document exported functions and types

```go
// Good
func DeployRepository(ctx context.Context, repo *Repository) (*Deployment, error) {
    if repo == nil {
        return nil, fmt.Errorf("repository cannot be nil")
    }
    
    // Implementation...
}

// Package comment
// Package detector provides automatic project type detection
// for various programming languages and frameworks.
package detector
```

### TypeScript/Vue Style

- Use TypeScript interfaces for data structures
- Follow Vue 3 Composition API patterns
- Use meaningful component and variable names
- Add type annotations where helpful

```typescript
// Good
interface DeploymentData {
  id: string
  status: DeploymentStatus
  createdAt: string
}

const { data: deployments, error } = await $fetch<DeploymentData[]>('/api/deployments')
```

### Running Style Checks

```bash
# Go formatting and linting
make fmt
make lint

# Dashboard formatting
cd dashboard
npm run lint
npm run lint:fix
```

## Documentation

### Code Documentation

- Document all exported functions, types, and packages
- Use Go's standard documentation format
- Include examples for complex functions

```go
// DetectProjectType analyzes repository files and determines the project type.
// It returns the detected project type or "unknown" if no type can be determined.
//
// Example:
//   projectType := DetectProjectType([]string{"package.json", "nuxt.config.js"})
//   // Returns: "nuxt"
func DetectProjectType(files []string) string {
    // Implementation...
}
```

### API Documentation

Update `docs/API.md` for any API changes:

```markdown
#### POST /api/deployments/{id}/restart

Restart a stopped deployment.

**Response:**
```json
{
  "message": "Deployment restarted",
  "id": "dep_123abc"
}
```

**Status Codes:**
- `200` - Deployment restarted successfully
- `404` - Deployment not found
- `400` - Deployment cannot be restarted
```

### User Documentation

Update user-facing documentation:
- README.md for setup changes
- docs/DEPLOYMENT.md for production changes
- Add new documentation files as needed

## Submitting Changes

### 1. Pre-submission Checklist

- [ ] Code follows project style guidelines
- [ ] Tests pass locally (`make test`)
- [ ] Code is properly formatted (`make fmt`)
- [ ] Linting passes (`make lint`) 
- [ ] Documentation is updated
- [ ] Commit messages are descriptive

### 2. Commit Messages

Use clear, descriptive commit messages:

```bash
# Good
git commit -m "feat: add support for Rust project detection

- Add Cargo.toml detection logic
- Include rust build commands in deployer
- Add tests for Rust project type
- Update documentation

Closes #123"

# Also good for simple changes
git commit -m "fix: handle empty repository URLs in webhook handler"

# Avoid
git commit -m "fix stuff"
git commit -m "wip"
```

### 3. Create Pull Request

1. Push your branch to your fork
   ```bash
   git push origin feature/your-feature-name
   ```

2. Create a pull request on GitHub
3. Fill out the pull request template
4. Link to related issues
5. Add appropriate labels

### Pull Request Template

```markdown
## Description
Brief description of changes and motivation.

## Changes Made
- [ ] Added new feature X
- [ ] Fixed bug Y
- [ ] Updated documentation

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing completed

## Documentation
- [ ] Code comments added
- [ ] API docs updated
- [ ] User docs updated

## Screenshots (if applicable)
Add screenshots for UI changes.

Closes #issue_number
```

### 4. Review Process

- Maintainers will review your PR
- Address feedback by pushing additional commits
- Once approved, your PR will be merged
- Delete your feature branch after merge

## Development Tips

### Useful Make Commands

```bash
make build          # Build binaries
make dev            # Run in development mode
make test           # Run all tests
make fmt            # Format code
make lint           # Run linters
make deps           # Install dependencies
make clean          # Clean build artifacts
make docker-build   # Build Docker image
```

### Debugging

**Go Application:**
```bash
# Run with debug logging
LOG_LEVEL=debug ./dockrune serve

# Use delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug cmd/dockrune/main.go
```

**Dashboard:**
```bash
# Vue devtools available in browser
cd dashboard
npm run dev -- --debug
```

### Adding New Project Types

1. **Update Detector:**
   ```go
   // internal/detector/detector.go
   func detectRust(files []string) bool {
       return containsFile(files, "Cargo.toml")
   }
   ```

2. **Add Deployment Logic:**
   ```go
   // internal/deployer/rust.go
   func (d *Deployer) deployRust(ctx context.Context, deployment *models.Deployment) error {
       // Implementation...
   }
   ```

3. **Add Tests:**
   ```go
   func TestDetectRust(t *testing.T) {
       // Test cases...
   }
   ```

4. **Update Documentation:**
   - Add to supported project types table
   - Update API documentation if needed

### Common Gotchas

- Always handle errors appropriately in Go
- Use context.Context for cancellation in long-running operations
- Be careful with database transactions and cleanup
- Test webhook handlers with actual GitHub payloads
- Ensure dashboard updates work with real-time WebSocket events

## Getting Help

- Join discussions in GitHub Discussions
- Ask questions in issues (use "question" label)
- Check existing documentation first
- Look at similar code for patterns

## Recognition

Contributors will be recognized in:
- GitHub contributors section
- CHANGELOG.md for significant changes
- Special thanks for major features

Thank you for contributing to dockrune! 🚀