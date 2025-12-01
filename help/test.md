# Testing Guide

This document describes the test suite for OpenVPN Manager.

## Test Structure

Tests are organized in the `test/` directory:

```
test/
├── testutil/
│   └── testutil.go              # Test utilities, helpers, DB setup
├── services/
│   ├── user_service_test.go     # User service tests
│   ├── auth_service_test.go     # Authentication service tests
│   ├── group_service_test.go    # Group service tests
│   ├── network_service_test.go  # Network service tests
│   ├── audit_service_test.go    # Audit service tests
│   └── vpn_session_service_test.go  # VPN session service tests
├── handlers/
│   ├── auth_handler_test.go     # Auth handler tests (login, logout, me)
│   └── user_handler_test.go     # User handler tests (CRUD, groups)
├── middleware/
│   └── auth_middleware_test.go  # JWT auth, role-based access tests
├── dto/
│   └── user_dto_test.go         # DTO parsing and conversion tests
└── integration/
    └── api_integration_test.go  # Full API integration tests
```

## Running Tests

### Run All Tests

```bash
make test
# or
go test -v ./test/...
```

### Run Specific Test Categories

```bash
# Service layer tests
make test-services

# HTTP handler tests
make test-handlers

# Middleware tests
make test-middleware

# Integration tests
make test-integration
```

### Run Tests with Race Detection

```bash
make test-race
```

### Run Tests with Coverage

```bash
# Text coverage report
make test-coverage

# HTML coverage report
make test-coverage-html
# Opens coverage.html in browser
```

### Run Individual Test Files

```bash
# Run specific test file
go test -v ./test/services/user_service_test.go

# Run specific test function
go test -v ./test/services/... -run TestUserService_Create

# Run specific subtest
go test -v ./test/services/... -run "TestUserService_Create/creates_user_with_manager"
```

## Test Categories

### Service Tests (`test/services/`)

Tests for the business logic layer:

- **user_service_test.go**: User CRUD, manager relationships, role-based listing
- **auth_service_test.go**: Authentication, password hashing, user validity checks
- **group_service_test.go**: Group CRUD, user-group associations, network management
- **network_service_test.go**: Network CRUD, CIDR validation, group associations
- **audit_service_test.go**: Audit log queries, filtering, statistics
- **vpn_session_service_test.go**: VPN sessions, traffic stats, usage statistics

### Handler Tests (`test/handlers/`)

Tests for HTTP handlers:

- **auth_handler_test.go**:
  - Login with valid/invalid credentials
  - Login with inactive/expired users
  - Logout functionality
  - Get current user (`/auth/me`)

- **user_handler_test.go**:
  - Create user (admin, manager permissions)
  - Get user (permission checks)
  - Update user (role restrictions)
  - Delete user (admin only)
  - List users (pagination, role-based filtering)
  - Profile updates
  - Password changes
  - User groups management

### Middleware Tests (`test/middleware/`)

- **auth_middleware_test.go**:
  - JWT token validation (Bearer, cookie)
  - Expired/invalid token handling
  - Role-based access (`RequireRole`, `RequireAdmin`, `RequireManagerOrAdmin`)
  - Context helpers (`GetAuthUser`, `GetAuthUserID`)

### DTO Tests (`test/dto/`)

- **user_dto_test.go**:
  - `DateOnly` parsing (date-only, RFC3339, datetime formats)
  - `ToUserResponse` conversion
  - Request parsing (CreateUserRequest, UpdateUserRequest)

### Integration Tests (`test/integration/`)

End-to-end API tests:

- **api_integration_test.go**:
  - Full authentication flow (login -> use API -> logout)
  - Admin user management (create, update, delete)
  - Group and network management
  - Role-based access verification
  - Profile and password updates

## Test Utilities

The `test/testutil/testutil.go` provides helper functions:

### Database Setup

```go
// Create in-memory SQLite database for testing
db := testutil.SetupTestDB(t)
defer testutil.CleanupTestDB(t, db)
```

### Create Test Data

```go
// Users
admin := testutil.CreateTestAdmin(t)
manager := testutil.CreateTestManager(t)
user := testutil.CreateTestRegularUser(t)
managedUser := testutil.CreateManagedUser(t, manager.ID)
namedUser := testutil.CreateTestUserWithName(t, models.RoleUser, "customname")

// Groups and Networks
group := testutil.CreateTestGroup(t, admin.ID)
network := testutil.CreateTestNetwork(t, admin.ID)

// VPN and Audit
session := testutil.CreateTestVpnSession(t, user.ID)
auditLog := testutil.CreateTestAuditLog(t, user.ID)
```

### Utility Functions

```go
timePtr := testutil.TimePtr(time.Now())
boolPtr := testutil.BoolPtr(true)
stringPtr := testutil.StringPtr("value")
uuidPtr := testutil.UUIDPtr(uuid.New())
```

## Writing New Tests

### Service Test Example

```go
func TestMyService_MyMethod(t *testing.T) {
    db := testutil.SetupTestDB(t)
    defer testutil.CleanupTestDB(t, db)

    service := services.NewMyService()

    t.Run("success case", func(t *testing.T) {
        // Arrange
        testData := testutil.CreateTestData(t)

        // Act
        result, err := service.MyMethod(testData.ID)

        // Assert
        require.NoError(t, err)
        assert.Equal(t, expected, result)
    })

    t.Run("error case", func(t *testing.T) {
        _, err := service.MyMethod(uuid.New())
        assert.Error(t, err)
        assert.Equal(t, services.ErrNotFound, err)
    })
}
```

### Handler Test Example

```go
func TestMyHandler_MyEndpoint(t *testing.T) {
    db := testutil.SetupTestDB(t)
    defer testutil.CleanupTestDB(t, db)

    gin.SetMode(gin.TestMode)

    t.Run("returns data for authorized user", func(t *testing.T) {
        user := testutil.CreateTestAdmin(t)

        router := gin.New()
        router.Use(func(c *gin.Context) {
            c.Set(middleware.AuthUserKey, &dto.AuthUser{
                ID:       user.ID.String(),
                Username: user.Username,
                Role:     user.Role,
            })
            c.Next()
        })

        handler := handlers.NewMyHandler()
        router.GET("/api/v1/myendpoint", handler.MyMethod)

        req, _ := http.NewRequest("GET", "/api/v1/myendpoint", nil)
        w := httptest.NewRecorder()

        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
    })
}
```

## Known Limitations

### GORM Default Values

When testing with SQLite in-memory database, GORM's `default:true` tag behavior may differ from PostgreSQL/MySQL. To create records with explicit `false` boolean values:

```go
// Create user first, then update to bypass GORM default
user := &models.User{
    Username: "test",
    IsActive: true,  // GORM default will apply anyway
    // ...
}
testutil.TestDB.Create(user)
testutil.TestDB.Model(user).Update("is_active", false)
```

### Coverage Metrics

Since tests are in a separate `test/` package, Go's coverage tool reports coverage for the test utilities, not the main application code. This is expected behavior for external test packages.

## Continuous Integration

Example GitHub Actions workflow:

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run tests
        run: make test
      - name: Run tests with race detection
        run: make test-race
```
