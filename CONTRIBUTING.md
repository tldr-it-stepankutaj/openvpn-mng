# Contributing to OpenVPN Manager

Thank you for your interest in contributing to OpenVPN Manager! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Commit Messages](#commit-messages)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. Please:

- Be respectful and considerate in all interactions
- Accept constructive criticism gracefully
- Focus on what is best for the community and project
- Show empathy towards other community members

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/openvpn-mng.git
   cd openvpn-mng
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/tldr-it-stepankutaj/openvpn-mng.git
   ```
4. **Keep your fork synced**:
   ```bash
   git fetch upstream
   git checkout main
   git merge upstream/main
   ```

## Development Setup

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12+ or MySQL 8+
- swag CLI tool for Swagger documentation

### Setup Steps

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Install swag CLI**:
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

3. **Setup database**:
   ```sql
   -- PostgreSQL
   CREATE DATABASE openvpn_mng_dev;

   -- MySQL
   CREATE DATABASE openvpn_mng_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

4. **Configure the application**:
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your local settings
   ```

5. **Generate Swagger documentation**:
   ```bash
   swag init -g cmd/server/main.go -o docs
   ```

6. **Run the application**:
   ```bash
   go run ./cmd/server
   ```

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check if the issue already exists. When creating a bug report, include:

- **Clear title** describing the issue
- **Steps to reproduce** the behavior
- **Expected behavior** vs **actual behavior**
- **Environment details** (OS, Go version, database type)
- **Relevant logs** or error messages
- **Screenshots** if applicable

Use the GitHub issue tracker: [Create a new issue](https://github.com/tldr-it-stepankutaj/openvpn-mng/issues/new)

### Suggesting Features

Feature requests are welcome! Please:

- Check if the feature has already been requested
- Provide a clear description of the feature
- Explain the use case and benefits
- Consider if it aligns with the project's goals

### Contributing Code

1. **Find an issue** to work on, or create one for discussion
2. **Comment on the issue** to let others know you're working on it
3. **Create a feature branch** from `main`
4. **Make your changes** following our coding standards
5. **Write/update tests** for your changes
6. **Submit a pull request**

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run tests** (when available):
   ```bash
   go test ./...
   ```

3. **Run linter**:
   ```bash
   go vet ./...
   ```

4. **Format code**:
   ```bash
   go fmt ./...
   ```

5. **Update Swagger docs** if API changed:
   ```bash
   swag init -g cmd/server/main.go -o docs
   ```

6. **Build successfully**:
   ```bash
   go build ./cmd/server
   ```

### Creating the Pull Request

1. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create PR on GitHub** with:
   - Clear, descriptive title
   - Reference to related issue(s) using `Fixes #123` or `Relates to #123`
   - Description of changes made
   - Screenshots for UI changes
   - Any breaking changes noted

3. **PR Template**:
   ```markdown
   ## Description
   Brief description of changes

   ## Related Issue
   Fixes #(issue number)

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Self-reviewed the code
   - [ ] Added/updated comments for complex logic
   - [ ] Updated documentation if needed
   - [ ] Added tests if applicable
   - [ ] All tests pass locally
   ```

### Review Process

- PRs require at least one approval before merging
- Address all review comments
- Keep the PR focused and reasonably sized
- Be patient - reviews may take time

## Coding Standards

### Go Style Guide

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://go.dev/doc/effective_go).

### Project-Specific Guidelines

#### File Organization

```
internal/
├── config/      # Configuration structures
├── database/    # Database initialization
├── dto/         # Data Transfer Objects
├── handlers/    # HTTP handlers
├── logger/      # Logging utilities
├── middleware/  # HTTP middleware
├── models/      # Database models
├── routes/      # Route definitions
└── services/    # Business logic
```

#### Naming Conventions

- **Files**: lowercase with underscores (`user_service.go`)
- **Packages**: lowercase, single word (`handlers`, `models`)
- **Interfaces**: descriptive names, often ending in `-er` (`UserService`)
- **Structs**: PascalCase (`UserHandler`, `NetworkService`)
- **Functions/Methods**: PascalCase for exported, camelCase for unexported

#### Error Handling

```go
// Good - wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Good - use custom error types for specific cases
var ErrUserNotFound = errors.New("user not found")
```

#### API Responses

Use consistent response structures:

```go
// Success
c.JSON(http.StatusOK, gin.H{"data": result})

// Error
c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})

// List with pagination
c.JSON(http.StatusOK, gin.H{
    "data":       items,
    "total":      total,
    "page":       page,
    "page_size":  pageSize,
    "total_pages": totalPages,
})
```

#### Swagger Documentation

Document all API endpoints:

```go
// CreateUser godoc
// @Summary     Create a new user
// @Description Create a new user (MANAGER, ADMIN only)
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user body     dto.CreateUserRequest true "User data"
// @Success     201  {object} dto.UserResponse
// @Failure     400  {object} map[string]string
// @Failure     401  {object} map[string]string
// @Security    BearerAuth
// @Router      /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
```

## Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `style` | Code style (formatting, semicolons, etc.) |
| `refactor` | Code refactoring |
| `test` | Adding/updating tests |
| `chore` | Maintenance tasks |

### Examples

```bash
feat(users): add password reset functionality

fix(auth): resolve JWT token expiration issue

docs(readme): update installation instructions

refactor(services): extract common validation logic

chore(deps): update gin framework to v1.9.1
```

### Guidelines

- Use imperative mood: "add" not "added" or "adds"
- Keep subject line under 72 characters
- Reference issues in the footer: `Fixes #123`

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/services/...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests for multiple cases
- Mock external dependencies
- Test both success and error paths

```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   dto.CreateUserRequest
        wantErr bool
    }{
        {
            name:    "valid user",
            input:   dto.CreateUserRequest{Username: "test", Email: "test@example.com"},
            wantErr: false,
        },
        {
            name:    "missing username",
            input:   dto.CreateUserRequest{Email: "test@example.com"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Documentation

### Code Comments

- Comment exported functions and types
- Explain "why", not "what"
- Keep comments up to date with code changes

### API Documentation

- Update Swagger annotations for API changes
- Regenerate docs: `swag init -g cmd/server/main.go -o docs`

### README Updates

Update README.md when:
- Adding new features
- Changing configuration options
- Modifying API endpoints
- Updating dependencies

## Questions?

If you have questions about contributing:

1. Check existing [issues](https://github.com/tldr-it-stepankutaj/openvpn-mng/issues) and [discussions](https://github.com/tldr-it-stepankutaj/openvpn-mng/discussions)
2. Open a new issue with your question
3. Tag it with the `question` label

Thank you for contributing to OpenVPN Manager!
