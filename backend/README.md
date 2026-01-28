# Clean Architecture Golang REST API

A production-ready REST API built with Golang, Gin framework, and Clean Architecture principles.

## 📋 Table of Contents

- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Development](#development)
- [Deployment](#deployment)
- [Best Practices](#best-practices)

## 🏗️ Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```
┌─────────────────────────────────────────┐
│     Frameworks & Drivers Layer          │
│  (Gin, GORM, PostgreSQL, HTTP)          │
└─────────────────────────────────────────┘
                  ▲
┌─────────────────────────────────────────┐
│     Interface Adapters Layer            │
│  (HTTP Handlers, Repository Impl)       │
└─────────────────────────────────────────┘
                  ▲
┌─────────────────────────────────────────┐
│        Use Cases Layer                  │
│     (Business Logic)                    │
└─────────────────────────────────────────┘
                  ▲
┌─────────────────────────────────────────┐
│         Domain Layer                    │
│    (Entities, Interfaces)               │
└─────────────────────────────────────────┘
```

**Key Principles:**

- Dependencies point inward
- Domain layer has no external dependencies
- Use cases orchestrate business logic
- Interface adapters convert data formats
- Frameworks are plugins

## 📁 Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── domain/                     # Enterprise business rules
│   │   ├── entity/                 # Domain entities
│   │   ├── repository/             # Repository interfaces
│   │   └── errors/                 # Domain errors
│   ├── usecase/                    # Application business rules
│   │   ├── user/                   # User use cases
│   │   └── auth/                   # Authentication logic
│   ├── repository/                 # Data access implementations
│   │   └── postgres/               # PostgreSQL implementations
│   ├── delivery/                   # Interface adapters
│   │   └── http/
│   │       ├── dto/                # Request/Response DTOs
│   │       ├── handler/            # HTTP handlers
│   │       ├── middleware/         # HTTP middleware
│   │       └── router/             # Route definitions
│   └── infrastructure/             # External concerns
│       ├── config/                 # Configuration
│       ├── database/               # Database setup
│       └── logger/                 # Logging setup
├── pkg/                            # Public libraries
│   └── utils/                      # Utilities
├── migrations/                     # Database migrations
├── config/                         # Configuration files
│   └── config.yaml
├── .env.example                    # Environment template
├── Makefile                        # Build commands
└── README.md
```

### Layer Responsibilities

| Layer              | Responsibility                        | Dependencies   |
| ------------------ | ------------------------------------- | -------------- |
| **Domain**         | Pure business entities and interfaces | None           |
| **Use Case**       | Business logic orchestration          | Domain only    |
| **Repository**     | Data access implementation            | Domain + GORM  |
| **Delivery**       | HTTP transport layer                  | Use Case + Gin |
| **Infrastructure** | External services setup               | External libs  |

## 🛠️ Tech Stack

- **Framework**: [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- **ORM**: [GORM](https://gorm.io/) - Database ORM
- **Database**: PostgreSQL
- **Config**: [Viper](https://github.com/spf13/viper) - Configuration management
- **Logger**: [Zap](https://github.com/uber-go/zap) - Structured logging
- **Validation**: [validator](https://github.com/go-playground/validator) - Request validation
- **JWT**: [jwt-go](https://github.com/golang-jwt/jwt) - JWT authentication
- **Testing**: [testify](https://github.com/stretchr/testify) - Testing framework

## 🚀 Getting Started

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 12 or higher
- Make (optional)

### Installation

1. **Clone the repository**

   ```bash
   cd backend
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Configure database**

   Update `config/config.yaml` or set environment variables:

   ```yaml
   database:
     host: localhost
     port: 5432
     user: postgres
     password: postgres
     dbname: tkhanchat
     sslmode: disable
   ```

5. **Run the application**

   ```bash
   # Using Go
   go run cmd/api/main.go

   # Using Make
   make run
   ```

The server will start on `http://localhost:8080`

### Using Docker

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down
```

## 📚 API Documentation

### Base URL

```
http://localhost:8080/api/v1
```

### Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Endpoints

#### Authentication

**Register User**

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

**Login**

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:

```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### User Management (Protected)

**Get Profile**

```http
GET /api/v1/users/me
Authorization: Bearer <token>
```

**Update Profile**

```http
PUT /api/v1/users/me
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Jane Doe"
}
```

**Get User by ID**

```http
GET /api/v1/users/:id
Authorization: Bearer <token>
```

**List Users**

```http
GET /api/v1/users?limit=10&offset=0
Authorization: Bearer <token>
```

**Delete User**

```http
DELETE /api/v1/users/:id
Authorization: Bearer <token>
```

### Response Format

All responses follow this structure:

**Success Response:**

```json
{
  "success": true,
  "message": "operation successful",
  "data": { ... }
}
```

**Error Response:**

```json
{
  "success": false,
  "message": "error message",
  "error": {
    "code": "ERROR_CODE",
    "details": "detailed error information"
  }
}
```

## 🧪 Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/usecase/user/...
```

### Test Structure

Tests are located alongside the code they test:

- `internal/usecase/user/user_usecase_test.go` - Use case tests with mocks
- Use `testify` for assertions
- Use `mockery` to generate mocks from interfaces

### Example Test

```go
func TestRegister_Success(t *testing.T) {
    mockRepo := new(MockUserRepository)
    uc := user.NewUserUseCase(mockRepo)

    mockRepo.On("GetByEmail", mock.Anything, "test@example.com").
        Return(nil, errors.ErrUserNotFound)
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).
        Return(nil)

    result, err := uc.Register(context.Background(),
        "test@example.com", "password123", "Test User")

    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
```

## 💻 Development

### Hot Reload

Install Air for hot reloading:

```bash
make install-tools
make dev
```

### Code Quality

```bash
# Run linter
make lint

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### Adding a New Module

Follow these steps to add a new module (e.g., Product):

1. **Create Entity** (`internal/domain/entity/product.go`)

   ```go
   type Product struct {
       ID    string
       Name  string
       Price float64
   }
   ```

2. **Create Repository Interface** (`internal/domain/repository/product_repository.go`)

   ```go
   type ProductRepository interface {
       Create(ctx context.Context, product *entity.Product) error
       GetByID(ctx context.Context, id string) (*entity.Product, error)
   }
   ```

3. **Implement Repository** (`internal/repository/postgres/product_repository.go`)

4. **Create Use Case** (`internal/usecase/product/product_usecase.go`)

5. **Create DTOs** (`internal/delivery/http/dto/product_dto.go`)

6. **Create Handler** (`internal/delivery/http/handler/product_handler.go`)

7. **Update Router** (`internal/delivery/http/router/router.go`)

8. **Wire Dependencies** (`cmd/api/main.go`)

### Adding Middleware

Create middleware in `internal/delivery/http/middleware/`:

```go
func RateLimiter() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Rate limiting logic
        c.Next()
    }
}
```

Add to router:

```go
router.Use(middleware.RateLimiter())
```

### API Versioning

```go
// v1
v1 := router.Group("/api/v1")
{
    v1.GET("/users", userHandlerV1.List)
}

// v2 - breaking changes
v2 := router.Group("/api/v2")
{
    v2.GET("/users", userHandlerV2.List)
}
```

## 🚢 Deployment

### Build for Production

```bash
# Build binary
make build

# Run binary
./bin/api
```

### Environment Variables

Set these in production:

```bash
APP_SERVER_MODE=release
APP_JWT_SECRET=<strong-random-secret>
APP_DATABASE_HOST=<production-db-host>
APP_DATABASE_PASSWORD=<secure-password>
```

### Docker Production

```bash
docker build -f Dockerfile.prod -t backend-api .
docker run -p 8080:8080 --env-file .env backend-api
```

## ✅ Best Practices

### DO:

- ✅ Keep domain layer pure (no external dependencies)
- ✅ Use interfaces for all dependencies
- ✅ Pass `context.Context` as first parameter
- ✅ Return errors, don't panic
- ✅ Use custom error types for domain errors
- ✅ Separate DTOs from entities
- ✅ Use dependency injection
- ✅ Write tests for use cases
- ✅ Use structured logging
- ✅ Validate all inputs

### DON'T:

- ❌ Import Gin or GORM in domain layer
- ❌ Use `panic` in business logic
- ❌ Expose entities directly via HTTP
- ❌ Hardcode configuration values
- ❌ Ignore errors
- ❌ Use global variables for dependencies
- ❌ Mix business logic with HTTP handlers
- ❌ Skip input validation

## 🔧 Common Pitfalls

1. **Circular Dependencies**: Keep dependency flow unidirectional (inward)
2. **GORM in Domain**: Use mapper functions to convert between models and entities
3. **Missing Context**: Always pass context for cancellation and timeouts
4. **Poor Error Handling**: Use custom error types and handle at appropriate layer
5. **No Graceful Shutdown**: Always implement graceful shutdown for production

## 📝 License

This project is licensed under the MIT License.

## 🤝 Contributing

Contributions are welcome! Please follow the existing architecture patterns.

## 📧 Contact

For questions or support, please open an issue.
