# Expense Split Tracker

A robust expense splitting application built with Go, featuring advanced debt management, multiple split types, and comprehensive API endpoints.

## Postman Documentation

- Link: [https://www.loom.com/share/00000000000000000000000000000000
](https://galactic-shadow-957901.postman.co/workspace/New-Team-Workspace~763b5af6-1599-4d05-a40e-ac0ac2457cdc/collection/24302902-6922c2b5-b44b-46dc-a849-a05efa93113f?action=share&creator=24302902)
## Features

### Core Features
- **Group Management**: Create and manage expense groups
- **Multiple Split Types**: 
  - Equal split (divide equally among members)
  - Exact amount split (assign specific amounts)
  - Percentage split (divide by percentages)
- **Balance Tracking**: Real-time balance calculations and debt tracking
- **Debt Settlement**: Record payments and settle debts between users
- **Debt Simplification**: Automatically minimize the number of transactions needed

### Technical Features
- **Idempotency**: Prevent duplicate operations with idempotency keys
- **Transactions**: ACID compliance with database transactions
- **Concurrency**: Safe concurrent operations with proper locking
- **Currency Support**: Multi-currency support with validation
- **API Documentation**: Comprehensive REST API (Postman collection )
- **Structured Logging**: Detailed logging with Zap
- **Health Checks**: Built-in health monitoring endpoints

## Architecture

The application follows Clean Architecture principles:

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and transactions
│   ├── models/          # Domain models and DTOs
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic layer
│   ├── controller/      # HTTP handlers
│   ├── middleware/      # HTTP middleware (CORS, logging, etc.)
│   ├── utils/           # Utility functions
│   └── routes/          # Route definitions
├── pkg/
│   ├── errors/          # Custom error types
│   └── response/        # Standardized API responses
└── tests/               # Test files
```

## Database Schema

### Key Tables
- **users**: User information
- **groups**: Expense groups
- **group_members**: Group membership
- **expenses**: Expense records
- **expense_splits**: How expenses are split
- **settlements**: Debt payments
- **user_balances**: Cached balance information
- **idempotency_keys**: Idempotency tracking

## Getting Started

### Prerequisites
- Go 1.21 or higher
- MySQL 8.0 or higher
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd expense-split-tracker
   ```

2. **Set up the database**
   ```sql
   CREATE DATABASE expense_split_tracker;
   ```

3. **Configure environment**
   ```bash
   # Edit config.env with your database credentials
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Run database migrations**
   ```bash
   # Apply migrations
   mysql -u root -p expense_split_tracker < internal/database/migrations/001_initial_schema.up.sql
   ```

6. **Start the server**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080`

### Configuration

Edit `config.env` with your settings:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=expense_split_tracker

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost

# Environment
ENV=development

# Logging
LOG_LEVEL=info

# Idempotency
IDEMPOTENCY_TTL_HOURS=24
```

## API Documentation

Use the Postman collection to explore and test endpoints.

- Path: `docs/postman/expense-split-tracker.postman_collection.json`
- Import in Postman and set variables: `base_url`, `group_uuid`, `user1_uuid`, `user2_uuid`, `user3_uuid`, `idempotency_key`.
- Includes example bodies, filtered list queries (transaction history), and test scenarios.
 
### Base URL
```
http://localhost:8080/api/v1
```

### API Endpoints

#### Users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users` - List users (paginated)
- `GET /api/v1/users/{uuid}` - Get user by UUID
- `GET /api/v1/users/by-email?email=...` - Get user by email

#### Groups
- `POST /api/v1/groups` - Create group
- `GET /api/v1/groups` - List groups
- `GET /api/v1/groups/{uuid}` - Get group details
- `POST /api/v1/groups/{uuid}/members` - Add member
- `DELETE /api/v1/groups/{uuid}/members/{userUuid}` - Remove member
- `GET /api/v1/groups/{uuid}/members` - List members
- `GET /api/v1/users/{uuid}/groups` - Get user's groups

#### Expenses
- `POST /api/v1/expenses` - Create expense
- `GET /api/v1/expenses` - List expenses (with filters)
- Filters: `group_uuid`, `user_uuid`, `split_type` (equal|exact|percentage), `currency`, `from_date` (YYYY-MM-DD), `to_date` (YYYY-MM-DD), `page`, `limit`
- `GET /api/v1/groups/{uuid}/expenses` - Get group expenses
- `GET /api/v1/users/{uuid}/expenses` - Get user expenses

#### Settlements
- `POST /api/v1/settlements` - Record settlement
- `GET /api/v1/settlements` - List settlements
- Filters: `group_uuid`, `user_uuid`, `from_date` (YYYY-MM-DD), `to_date` (YYYY-MM-DD), `page`, `limit`
- `GET /api/v1/settlements/{uuid}` - Get settlement details
- `GET /api/v1/groups/{uuid}/settlements` - Get group settlements
- `GET /api/v1/groups/{uuid}/simplify-debts` - Get debt simplification suggestions

#### Balances
- `GET /api/v1/groups/{uuid}/balance-sheet` - Get group balance sheet
- `GET /api/v1/groups/{groupUuid}/users/{userUuid}/balance` - Get user balance

### Health Check
- `GET /health` - Service health status

## Testing

### What’s covered (unit)
- Expense splits: equal, exact (with sum validation), percentage (sum to 100)
- Settlements: success path, amount exceeds debt, same payer/receiver validation
- Debt simplification: suggestions and savings
- Error handling: invalid UUIDs across services

1. **Equal Split**: Expense divided equally among users
2. **Exact Amount Split**: Specific amounts assigned to users
3. **Percentage Split**: Expense divided by percentages
4. **Debt Settlement**: Recording and tracking payments
5. **Debt Simplification**: Minimizing transaction count

### Running Tests

```bash
# Unit tests only
go test ./tests/unit/... -v

# All tests with coverage
go test -cover ./...
```

## Problem Statement & Approach

- **Problem**: Build an expense split tracker that supports multiple split types, maintains running balances per user and group, records settlements, and suggests minimal transactions to settle debts.
- **Approach**:
  - Clean architecture with repositories/services/controllers for testability.
  - Strong validation, idempotency for financial mutations, transaction middleware.
  - Deterministic rounding for splits using `shopspring/decimal`.
  - Greedy debt simplification to reduce the number of transactions.

## Explanation of Complex Logic / Algorithms

- **Split Calculations**
  - Equal: `amount / N` rounded to 2 decimals; last split receives remainder to ensure sum equals total.
  - Exact: Validates sum of split amounts equals the expense amount.
  - Percentage: Validates percentages sum to 100; amount computed per user and rounded to 2 decimals.
- **Balance Updates**
  - Each split increases the debtor’s balance; payer’s balance decreased by total amount.
- **Settlements**
  - Validates members and sufficient debt before allowing settlement; updates both sides’ balances.
- **Debt Simplification**
  - Greedy matching largest debtor with largest creditor until all balances reach zero; tracks suggested transactions and savings.

## Areas Requiring Special Consideration

- **Idempotency**: Required for expenses and settlements to prevent duplicates; include `Idempotency-Key`.
- **Rounding**: Deterministic handling of cents in equal/percentage splits.
- **Transactions**: All financial operations run in DB transactions with rollback on errors.
- **Validation**: UUIDs, currencies, amounts, and membership checks at each step.
- **Pagination & Limits**: Defensive defaults for list endpoints.


## Performance Considerations

- **Database Indexes**: Proper indexing on frequently queried columns
- **Connection Pooling**: Configured database connection pool
- **Balance Caching**: User balances cached for quick lookups
- **Pagination**: All list endpoints support pagination
- **Concurrent Safety**: Proper locking for balance calculations

## Security Features

- **SQL Injection Protection**: Parameterized queries
- **Input Validation**: Comprehensive input validation
- **Error Handling**: Secure error messages without data leakage
- **CORS Configuration**: Configurable CORS settings
- **Request Logging**: Detailed request/response logging

## Monitoring and Logging

- **Structured Logging**: JSON formatted logs with Zap
- **Request Tracing**: Request ID tracking
- **Error Tracking**: Detailed error logging
- **Performance Metrics**: Request duration and status tracking

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
