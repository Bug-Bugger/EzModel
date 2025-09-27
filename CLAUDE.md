# EzModel - Visual Database Schema Designer

## Project Overview

EzModel is a visual database schema design tool that empowers developers to create, design, and manage database schemas through an intuitive visual interface. Think of it as "Figma for databases" - providing real-time collaboration capabilities for database design teams.

### Key Features

- **Visual Schema Design**: Drag-and-drop interface for designing database schemas
- **Real-time Collaboration**: WebSocket-powered live collaboration for teams
- **Multi-Database Support**: Generate SQL for PostgreSQL, MySQL, SQLite, and SQL Server
- **Code Generation**: Automatic ORM model generation and SQL script output
- **Team Management**: Project sharing and collaboration with role-based access

Whether you're a beginner learning database design or a seasoned developer architecting complex systems, EzModel streamlines the database creation process.

## Tech Stack

### Backend (Golang)
**Why Golang was chosen:**
- **WebSocket Excellence**: Perfect for handling real-time collaboration with excellent concurrent processing
- **Code Generation**: Strong template system and reflection capabilities for generating SQL/ORM code
- **Performance**: Compiled language with strict typing prevents bugs in critical schema generation
- **Concurrency**: Goroutines handle multiple users collaborating simultaneously
- **Database Integration**: Excellent ORM support with GORM

**Current Stack:**
- **Framework**: Chi router for HTTP handling
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT tokens (access + refresh)
- **Validation**: go-playground/validator for input validation
- **Configuration**: Environment-based config with .env support

## Project Architecture

### Clean Architecture Pattern
```
backend/
├── cmd/api/                    # Application entry point
├── internal/                   # Private application code
│   ├── api/                   # HTTP layer (handlers, middleware, routes)
│   ├── services/              # Business logic layer
│   ├── repository/            # Data access layer
│   ├── models/               # Domain entities
│   ├── config/               # Configuration management
│   ├── db/                   # Database connection
│   └── validation/           # Input validation utilities
```

### Core Design Patterns
- **Dependency Injection**: Services receive dependencies through constructors
- **Repository Pattern**: Abstract data access behind interfaces
- **Interface Segregation**: Services depend on interfaces, not implementations
- **Clean Separation**: Each layer has single responsibility

## Development Commands

### Essential Commands
```bash
# Start development server
cd backend && go run cmd/api/main.go

# Build application
cd backend && go build -o bin/ezmodel cmd/api/main.go

# Run tests
cd backend && go test ./...

# Install dependencies
cd backend && go mod tidy

# Database migration (auto-migration via GORM)
# Runs automatically on startup
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run with race detection
go run -race cmd/api/main.go
```

## Core Domain Models

### Project Management
```go
type Project struct {
    ID           uuid.UUID     `json:"id"`
    Name         string        `json:"name"`
    Description  string        `json:"description"`
    OwnerID      uuid.UUID     `json:"owner_id"`
    DatabaseType string        `json:"database_type"` // postgresql, mysql, sqlite, sqlserver
    CanvasData   string        `json:"canvas_data"`   // Visual design stored as JSONB

    // Relationships
    Owner         User                    `json:"owner"`
    Collaborators []User                  `json:"collaborators"`
    Tables        []Table                 `json:"tables"`
    Relationships []Relationship          `json:"relationships"`
}
```

### Database Design Entities
- **Table**: Database table definitions with fields
- **Field**: Column definitions with types, constraints, and validation
- **Relationship**: Foreign key relationships between tables
- **CollaborationSession**: Real-time collaboration tracking

## API Architecture

### Authentication System
- **JWT-based**: Access tokens (15min) + Refresh tokens (7 days)
- **Middleware Protection**: Route-level authentication
- **Context-based**: User info passed through request context

### Core API Endpoints

#### Authentication
```
POST /register          # User registration
POST /login             # User authentication
POST /refresh-token     # Token refresh
```

#### Project Management
```
GET    /projects        # List user's projects
POST   /projects        # Create new project
GET    /projects/{id}   # Get project details
PUT    /projects/{id}   # Update project
DELETE /projects/{id}   # Delete project

POST   /projects/{id}/collaborators/{user_id}    # Add collaborator
DELETE /projects/{id}/collaborators/{user_id}    # Remove collaborator
```

#### User Management
```
GET    /users          # List users (admin)
GET    /users/{id}     # Get user profile
PUT    /users/{id}     # Update user
DELETE /users/{id}     # Delete user
```

### Response Format
All API responses follow consistent structure:
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { /* response payload */ }
}
```

## Database Schema

### Key Features
- **UUID Primary Keys**: Enhanced security and distribution support
- **JSONB Storage**: Canvas visual data stored efficiently in PostgreSQL
- **Audit Trails**: CreatedAt/UpdatedAt timestamps on all entities
- **Soft Deletes**: Maintain data integrity with soft deletion patterns
- **Referential Integrity**: Proper foreign key constraints

### Database Configuration
```go
// Support for multiple database types
DatabaseType: "postgresql" | "mysql" | "sqlite" | "sqlserver"
```

## Real-time Collaboration

### WebSocket Architecture (Planned)
- **Live Cursors**: See collaborators' cursors in real-time
- **Schema Updates**: Broadcast table/field changes instantly
- **Conflict Resolution**: Handle simultaneous edits gracefully
- **Session Management**: Track active collaboration sessions

### Current State
- Models and database schema ready for WebSocket integration
- CollaborationSession entity tracks active sessions
- Project-based collaboration permissions established

## Development Guidelines

### Code Conventions
- **Error Handling**: Custom error types for different scenarios
- **Validation**: Multi-layer validation (struct tags + business rules)
- **Security**: Password hashing, SQL injection prevention, input sanitization
- **Testing**: Unit tests for services, integration tests for repositories

### Service Layer Pattern
```go
type ProjectService struct {
    projectRepo repository.ProjectRepositoryInterface
    userRepo    repository.UserRepositoryInterface
}

func (s *ProjectService) CreateProject(name, description string, ownerID uuid.UUID) (*models.Project, error) {
    // Business logic here
}
```

### Repository Interface Pattern
```go
type ProjectRepositoryInterface interface {
    Create(project *models.Project) (uuid.UUID, error)
    GetByID(id uuid.UUID) (*models.Project, error)
    GetByOwnerID(ownerID uuid.UUID) ([]*models.Project, error)
    Update(project *models.Project) error
    Delete(id uuid.UUID) error
}
```

## Environment Configuration

EzModel uses centralized environment variable management with all configuration stored in the project root.

### Environment Files
- **`.env.dev`**: Development environment configuration
- **`.env.prod`**: Production environment configuration

Both backend and frontend read from these root-level files for consistent configuration across the entire application.

### Required Environment Variables
```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=ezmodel
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your_jwt_secret
JWT_ACCESS_TOKEN_EXP=15m
JWT_REFRESH_TOKEN_EXP=168h

# Server Configuration
PORT=8080
ENV=development

# Frontend Configuration (Vite prefixed)
VITE_API_URL=http://localhost:8080/api
```

### Environment Loading
- **Backend**: Uses `godotenv.Load("../.env")` to load from project root
- **Frontend**: Uses Vite's `envDir: '../'` configuration to read from project root

## Future Development

### Planned Features
1. **WebSocket Integration**: Real-time collaboration implementation
2. **SQL Generation**: Export complete database schemas
3. **ORM Code Generation**: Generate model classes for various frameworks
4. **Schema Validation**: Advanced constraint and relationship validation
5. **Version Control**: Schema versioning and migration management
6. **Team Management**: Advanced role-based permissions

### Code Generation Targets
- **SQL Scripts**: DDL statements for schema creation
- **GORM Models**: Go struct generation
- **Database Migrations**: Migration file generation
- **API Documentation**: Auto-generated API specs

## Testing Strategy

### Current Test Structure
- **Unit Tests**: Service layer business logic
- **Integration Tests**: Repository database operations
- **Handler Tests**: HTTP endpoint testing

### Test Commands
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/services/
```

## Security Considerations

- **Authentication**: JWT with secure secret rotation
- **Authorization**: Role-based access control ready
- **Input Validation**: Comprehensive validation at all layers
- **SQL Injection**: GORM ORM provides protection
- **Password Security**: bcrypt hashing with salt
- **CORS**: Configurable cross-origin resource sharing

## Documentation

### Additional Resources
- `ARCHITECTURE.md`: Detailed architecture documentation
- `API_GUIDELINES.md`: API development standards
- `DATABASE_GUIDE.md`: Database design guidelines
- `DEVELOPMENT_GUIDE.md`: Setup and development instructions

This project follows clean architecture principles with a focus on maintainability, testability, and scalability. The modular design enables easy extension and modification as features are added.