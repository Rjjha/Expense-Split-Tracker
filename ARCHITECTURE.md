# Expense Split Tracker - Architecture Overview

## Summary

This is a comprehensive expense splitting application built in Go following clean architecture principles. The application provides robust expense management with multiple split types, debt tracking, and advanced features like idempotency and debt simplification.

## Key Features Implemented

### ✅ Core Business Logic
- **User Management**: Complete CRUD operations for users
- **Group Management**: Create and manage expense groups with members
- **Multiple Split Types**:
  - Equal split (divide equally among all members)
  - Exact amount split (assign specific amounts to users)
  - Percentage split (divide by percentages)
- **Balance Tracking**: Real-time balance calculations
- **Debt Settlement**: Record payments between users
- **Debt Simplification**: Minimize transaction count (architecture ready)

### ✅ Technical Excellence
- **Clean Architecture**: Proper separation of concerns across layers
- **Database Transactions**: ACID compliance with proper rollback handling
- **Selective Idempotency**: Prevents duplicate financial operations using UUID keys
- **Concurrency Safety**: Thread-safe operations with proper locking
- **Currency Support**: Multi-currency validation and handling
- **Comprehensive Error Handling**: Custom error types with proper HTTP status codes
- **Structured Logging**: Detailed logging with Zap logger
- **Input Validation**: Comprehensive validation at all layers

## API Documentation

- A ready-to-use Postman collection is available at `docs/postman/expense-split-tracker.postman_collection.json`.
  - Import it into Postman to explore and run all endpoints.

## Architecture Layers

### 1. **Database Layer** (`internal/database/`)
- Connection management with pooling
- Transaction support with automatic rollback
- Health checks and monitoring
- Migration support

### 2. **Repository Layer** (`internal/repository/`)
- Data access abstraction
- CRUD operations for all entities
- Complex queries with joins
- Idempotency key management

### 3. **Service Layer** (`internal/service/`)
- Business logic implementation
- Transaction coordination
- Validation and error handling
- Cross-cutting concerns

### 4. **Controller Layer** (`internal/controller/`)
- HTTP request handling
- Request/response mapping
- API documentation via Postman collection 

### 5. **Middleware** (`internal/middleware/`)
- **Idempotency**: Duplicate request prevention
- **Transaction**: Automatic transaction management
- **CORS**: Cross-origin resource sharing
- **Logging**: Structured request/response logging

## Database Schema

### Core Tables
```sql
users              - User information
groups             - Expense groups
group_members      - Group membership (many-to-many)
expenses           - Expense records
expense_splits     - How expenses are split among users
settlements        - Debt payment records
user_balances      - Cached balance information (performance)
idempotency_keys   - Request deduplication
```

### Key Relationships
- Users ↔ Groups (many-to-many via group_members)
- Groups → Expenses (one-to-many)
- Expenses → ExpenseSplits (one-to-many)
- Users → Settlements (many-to-many via from_user/to_user)

## API Design

### RESTful Endpoints (highlight)
- Users: create, list, get by UUID/email
- Groups: create, list, get, add/remove members, list members, user's groups
- Expenses: create; list with filters; group/user scoped lists
- Settlements: create; list; get by UUID; group/user scoped lists; simplify debts (GET suggestions)
- Balances: group balance sheet; user balance in group

### Idempotency
- Financial operations (expenses, settlements) require `Idempotency-Key`
- User and group operations do not use idempotency
- UUID format validation
- Request fingerprinting with SHA-256
- TTL-based cleanup (configurable, default 24h)

### Error Handling
- Standardized error responses
- Proper HTTP status codes
- Structured error information
- No sensitive data leakage

## Configuration Management

### Environment-based Configuration
```env
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
SERVER_PORT, SERVER_HOST
ENV (development/production)
LOG_LEVEL
IDEMPOTENCY_TTL_HOURS
```

### Database Setup
- MySQL 8.0+ with proper configuration
- Connection pooling and health checks
- Migration scripts for schema management

## Testing Strategy

### Unit Tests
- Service layer testing with mocks
- Repository testing with test database
- Comprehensive test coverage
- Mock implementations for external dependencies

### Integration Tests
- End-to-end API testing
- Database integration testing
- Transaction rollback testing
- Idempotency verification

## Performance Optimizations

### Database
- Proper indexing on frequently queried columns
- Connection pooling (25 max, 10 idle)
- Balance caching for quick lookups
- Optimized queries with joins

### Application
- Pagination for all list endpoints
- Efficient memory usage
- Concurrent request handling
- Request/response streaming

## Security Considerations

### Input Security
- SQL injection prevention (parameterized queries)
- Input validation at all layers
- Email format validation
- UUID format validation

### Operational Security
- Secure error messages
- Request logging without sensitive data
- CORS configuration
- Health check endpoints

## Development Tools

### Running Locally
- Run server: `go run cmd/server/main.go`
- Apply migrations: use MySQL CLI with files in `internal/database/migrations/*.sql`

### Code Quality
- Go fmt for formatting
- Linting with golangci-lint
- Security scanning with gosec
- Dependency management with go mod

## Deployment

### Production Deployment
- Binary compilation for target platform
- Environment-based configuration
- Secrets management ready
- Logging configuration
- Performance tuning parameters
- MySQL database setup and migration

## What's Implemented vs. Planned

### ✅ Fully Implemented
- User management (complete CRUD)
- Database schema and migrations
- Repository layer (users, groups, expenses, settlements, balances, idempotency)
- Service layer (users, groups, expenses with all split types, settlements, balances)
- Controller layer (users, groups, expenses, settlements, balances)
- Debt simplification (greedy suggestions algorithm)
- Middleware (idempotency, transactions, CORS, logging)
- Configuration management
- Testing framework and unit tests (splits, settlements, simplification, error handling)
- API documentation (Postman collection)



## Test Cases Covered

The application has unit tests covering the required scenarios:

1. **Equal Split**: $90 ÷ 3 users = $30 each
2. **Exact Amount**: Validation that sums equal the expense amount
3. **Percentage Split**: Percentages sum to 100; correct amounts
4. **Debt Settlement**: Success path and insufficient debt errors
5. **Debt Simplification**: Suggestions and transaction savings

## Next Steps

1. **Documentation**: Add Swagger UI endpoints and CI step to generate docs
2. **Integration Tests**: Add API-level tests with seeded database
3. **Observability**: Add request metrics and tracing spans
4. **Hardening**: Add rate limiting and input size constraints
