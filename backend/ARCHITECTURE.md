# Backend Architecture Documentation

## Overview

This backend follows a **Clean Architecture** pattern with clear separation of concerns, organized into distinct layers that promote maintainability, testability, and scalability.

## Project Structure

```
backend/
├── cmd/api/                    # Application entry points
│   └── main.go                # Main server startup
├── internal/                  # Private application code
│   ├── api/                   # HTTP layer
│   │   ├── dto/               # Data Transfer Objects
│   │   ├── handlers/          # HTTP request handlers
│   │   ├── middleware/        # HTTP middleware
│   │   ├── responses/         # Standardized API responses
│   │   ├── routes/            # Route definitions
│   │   └── server/            # Server configuration
│   ├── config/                # Configuration management
│   ├── db/                    # Database connection
│   ├── models/                # Domain models/entities
│   ├── repository/            # Data access layer
│   ├── services/              # Business logic layer
│   └── validation/            # Input validation utilities
├── go.mod                     # Go module definition
└── go.sum                     # Go module checksums
```

## Architecture Layers

### 1. **Presentation Layer** (`internal/api/`)
- **Purpose**: Handles HTTP requests/responses and API concerns
- **Components**:
  - **Handlers**: Process HTTP requests, call services, format responses
  - **Middleware**: Cross-cutting concerns (authentication, logging, CORS)
  - **DTOs**: Request/response data structures
  - **Routes**: URL routing and endpoint definitions
  - **Responses**: Standardized API response formatting

### 2. **Business Logic Layer** (`internal/services/`)
- **Purpose**: Contains core business rules and application logic
- **Components**:
  - Services implement business operations
  - Coordinate between repositories and external services
  - Handle domain-specific validation and errors
  - Manage transactions and business workflows

### 3. **Data Access Layer** (`internal/repository/`)
- **Purpose**: Abstracts database operations and data persistence
- **Components**:
  - Repository interfaces define data operations
  - Repository implementations handle database queries
  - Database-agnostic data access patterns

### 4. **Domain Layer** (`internal/models/`)
- **Purpose**: Core business entities and domain objects
- **Components**:
  - Domain models with business rules
  - Entity definitions with validation tags
  - Database schema mappings

## Key Design Patterns

### 1. **Dependency Injection**
Dependencies are injected through constructors, promoting loose coupling:

```go
type UserService struct {
    userRepo repository.UserRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
    return &UserService{userRepo: userRepo}
}
```

### 2. **Interface Segregation**
Services depend on interfaces, not concrete implementations:

```go
type UserServiceInterface interface {
    CreateUser(email, username, password, avatarURL string) (*models.User, error)
    GetUserByID(id uuid.UUID) (*models.User, error)
    // ... other methods
}
```

### 3. **Repository Pattern**
Data access is abstracted behind repository interfaces:

```go
type UserRepositoryInterface interface {
    Create(user *models.User) (uuid.UUID, error)
    GetByID(id uuid.UUID) (*models.User, error)
    GetByEmail(email string) (*models.User, error)
    // ... other methods
}
```

### 4. **Standardized Error Handling**
Custom error types for different scenarios:

```go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserAlreadyExists = errors.New("user already exists")
)
```

## Configuration Management

Configuration is centralized and environment-aware:

```go
type Config struct {
    Port     string
    Env      string
    Database struct {
        Host, Port, User, Password, DBName, SSLMode string
    }
    JWT struct {
        Secret          string
        AccessTokenExp  time.Duration
        RefreshTokenExp time.Duration
    }
}
```

- Uses environment variables with sensible defaults
- Supports `.env` file loading for development
- Type-safe configuration access throughout the application

## Security Architecture

### Authentication & Authorization
- **JWT-based authentication** with access and refresh tokens
- **Middleware-based protection** for secured endpoints
- **Role-based access** (ready for extension)

### Data Protection
- **Password hashing** using bcrypt
- **SQL injection prevention** via GORM ORM
- **Input validation** at multiple layers
- **Sensitive data exclusion** from JSON responses (`json:"-"`)

## Database Design

### ORM Strategy
- **GORM** for database operations
- **Auto-migrations** for schema management
- **UUID primary keys** for better security and distribution

### Model Structure
```go
type User struct {
    ID            uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Email         string     `json:"email" gorm:"type:varchar(255);unique;not null" validate:"required,email"`
    PasswordHash  string     `json:"-" gorm:"type:varchar(255);not null"`
    // ... other fields with proper tags
}
```

## API Design Standards

### Response Format
All API responses follow a consistent structure:

```go
type APIResponse struct {
    Success bool                   `json:"success"`
    Message string                 `json:"message"`
    Data    interface{}           `json:"data,omitempty"`
    Errors  map[string]string     `json:"errors,omitempty"`
}
```

### HTTP Status Codes
- **200**: Successful operations
- **201**: Resource creation
- **400**: Bad request/validation errors
- **401**: Authentication required
- **404**: Resource not found
- **500**: Internal server errors

## Development Principles

### 1. **Single Responsibility**
Each component has one reason to change:
- Handlers only handle HTTP concerns
- Services only contain business logic
- Repositories only handle data access

### 2. **Interface-First Design**
Define interfaces before implementations to enable:
- Easy testing with mocks
- Flexible implementations
- Reduced coupling

### 3. **Error Propagation**
Errors flow up the stack with context:
- Repository errors are wrapped by services
- Service errors are handled by handlers
- Consistent error responses to clients

### 4. **Validation Strategy**
Multi-layer validation approach:
- **Struct tags** for basic validation
- **Service layer** for business rule validation
- **Handler layer** for request validation

## Testing Strategy

### Unit Testing
- Test business logic in services with mocked dependencies
- Test repository implementations against test database
- Test handlers with HTTP test servers

### Integration Testing
- End-to-end API testing
- Database integration testing
- Authentication flow testing

## Scalability Considerations

### Horizontal Scaling
- Stateless design enables multiple instances
- JWT tokens eliminate session state
- Database connection pooling

### Performance
- Repository pattern enables caching layers
- Structured logging for monitoring
- Graceful shutdown handling

## Migration and Deployment

### Database Migrations
- GORM auto-migration for development
- Manual migrations for production
- Backup strategies for data safety

### Environment Management
- Development, staging, and production configs
- Container-ready architecture
- Health check endpoints (to be implemented)

---

This architecture provides a solid foundation for building scalable, maintainable APIs while following Go best practices and clean architecture principles.