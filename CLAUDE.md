# EzModel - Visual Database Schema Designer

## Project Overview

EzModel is a visual database schema design tool that empowers developers to create, design, and manage database schemas through an intuitive visual interface. Think of it as "Figma for databases" - providing real-time collaboration capabilities for database design teams.

### Key Features (Implemented)

- **Visual Schema Design**: Drag-and-drop interface using @xyflow/svelte for designing database schemas
- **Real-time Collaboration**: Fully implemented WebSocket-powered live collaboration with cursors, presence, and activities
- **Multi-Database Support**: Support for PostgreSQL, MySQL, SQLite, and SQL Server project types
- **Team Management**: Project sharing and collaboration with owner/collaborator roles
- **Auto-save**: Automatic saving of canvas data and schema changes
- **JWT Authentication**: Secure authentication with access and refresh tokens

## Tech Stack

### Backend (Go 1.24.1)

**Core Dependencies:**
- **Framework**: Chi router (v5.2.1) for HTTP handling
- **Database**: PostgreSQL with GORM (v1.26.0) ORM
- **WebSocket**: Gorilla WebSocket (v1.5.3) for real-time collaboration
- **Authentication**: JWT tokens using golang-jwt/jwt/v5 (v5.2.2)
- **Validation**: go-playground/validator (v10.26.0) for input validation
- **Configuration**: godotenv (v1.5.1) for environment management
- **Security**: golang.org/x/crypto for password hashing

### Frontend (SvelteKit + TypeScript)

**Core Dependencies:**
- **Framework**: SvelteKit (v2.22.0) with Svelte 5.0
- **UI Components**: ShadCN-Svelte (v1.0.7) + Tailwind CSS (v3.4.0)
- **Visual Canvas**: @xyflow/svelte (v1.3.1) for flow-based schema visualization
- **HTTP Client**: Axios (v1.12.2) with automatic token refresh
- **Icons**: Lucide Svelte (v0.544.0)
- **Styling**: Tailwind CSS with forms and typography plugins

## Project Architecture

### Backend Structure

```
backend/
├── cmd/api/                        # Application entry point
│   └── main.go                     # Server startup with env loading
├── internal/                       # Private application code
│   ├── api/                       # HTTP layer
│   │   ├── handlers/              # HTTP request handlers
│   │   ├── middleware/            # Authentication middleware
│   │   ├── routes/                # Route definitions
│   │   ├── dto/                   # Data transfer objects
│   │   └── responses/             # Response utilities
│   ├── services/                  # Business logic layer
│   ├── repository/                # Data access layer
│   ├── models/                    # GORM domain entities
│   ├── websocket/                 # WebSocket hub and messaging
│   ├── config/                    # Configuration management
│   ├── db/                        # Database connection
│   └── validation/                # Input validation utilities
```

### Frontend Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── components/            # Reusable UI components
│   │   │   ├── ui/                # ShadCN-Svelte base components
│   │   │   ├── layout/            # Layout components (Header)
│   │   │   ├── flow/              # Database canvas components
│   │   │   ├── collaboration/     # Real-time collaboration UI
│   │   │   └── project/           # Project management components
│   │   ├── services/              # API services and HTTP client
│   │   ├── stores/                # Svelte stores for state management
│   │   ├── types/                 # TypeScript type definitions
│   │   └── utils/                 # Utility functions
│   └── routes/                    # SvelteKit routes
│       ├── +layout.svelte         # Root layout
│       ├── +page.svelte           # Home page
│       ├── login/                 # Authentication pages
│       ├── register/
│       └── projects/              # Project management
│           └── [id]/              # Project details and editor
│               └── edit/          # Visual schema editor
```

## Development Commands

### Backend Commands

```bash
# Start development server
cd backend && go run cmd/api/main.go

# Build application
cd backend && go build -o bin/ezmodel cmd/api/main.go

# Run tests
cd backend && go test ./...

# Install dependencies
cd backend && go mod tidy

# Code quality
go fmt ./...
go vet ./...
```

### Frontend Commands

```bash
# Start development server
cd frontend && pnpm dev

# Build application
cd frontend && pnpm build

# Preview build
cd frontend && pnpm preview

# Type checking
cd frontend && pnpm check

# Install dependencies
cd frontend && pnpm install
```

## Database Models (GORM)

### Core Entities

```go
// User represents a user in the system
type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Email        string    `gorm:"uniqueIndex;not null"`
    Username     string    `gorm:"uniqueIndex;not null"`
    PasswordHash string    `gorm:"not null"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// Project represents a database schema design project
type Project struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Name         string    `gorm:"not null"`
    Description  string
    OwnerID      uuid.UUID `gorm:"type:uuid;not null"`
    DatabaseType string    `gorm:"default:'postgresql'"` // postgresql, mysql, sqlite, sqlserver
    CanvasData   string    `gorm:"type:jsonb"`           // Visual layout/positioning data
    CreatedAt    time.Time
    UpdatedAt    time.Time

    // Relationships
    Owner         User           `gorm:"foreignKey:OwnerID"`
    Collaborators []User         `gorm:"many2many:project_collaborators;"`
    Tables        []Table        `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
    Relationships []Relationship `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// Table represents a database table in the schema
type Table struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    ProjectID uuid.UUID `gorm:"type:uuid;not null"`
    Name      string    `gorm:"not null"`
    PosX      float64   // Canvas position
    PosY      float64   // Canvas position
    CreatedAt time.Time
    UpdatedAt time.Time

    Fields []Field `gorm:"foreignKey:TableID;constraint:OnDelete:CASCADE"`
}

// Field represents a column in a database table
type Field struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    TableID      uuid.UUID `gorm:"type:uuid;not null"`
    Name         string    `gorm:"not null"`
    DataType     string    `gorm:"not null"` // VARCHAR, INT, TEXT, etc.
    IsPrimaryKey bool      `gorm:"default:false"`
    IsNullable   bool      `gorm:"default:true"`
    DefaultValue string
    Position     int       // Field order in table
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// Relationship represents a foreign key relationship between tables
type Relationship struct {
    ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    ProjectID     uuid.UUID `gorm:"type:uuid;not null"`
    SourceTableID uuid.UUID `gorm:"type:uuid;not null"`
    SourceFieldID uuid.UUID `gorm:"type:uuid;not null"`
    TargetTableID uuid.UUID `gorm:"type:uuid;not null"`
    TargetFieldID uuid.UUID `gorm:"type:uuid;not null"`
    RelationType  string    `gorm:"default:'one_to_many'"` // one_to_one, one_to_many, many_to_many
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

// CollaborationSession tracks active collaboration sessions
type CollaborationSession struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    UserID    uuid.UUID `gorm:"type:uuid;not null"`
    ProjectID uuid.UUID `gorm:"type:uuid;not null"`
    IsActive  bool      `gorm:"default:true"`
    CursorX   float64
    CursorY   float64
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## API Endpoints (Implemented)

### Authentication Endpoints

```
POST /api/login              # User login
POST /api/refresh-token      # JWT token refresh
POST /api/register           # User registration
```

### Protected Endpoints (Require JWT)

#### User Management
```
GET    /api/me                      # Get current user profile
GET    /api/users                   # List all users
GET    /api/users/{user_id}         # Get user by ID
PUT    /api/users/{user_id}         # Update user
DELETE /api/users/{user_id}         # Delete user
PUT    /api/users/{user_id}/password # Update user password
```

#### Project Management
```
GET    /api/projects                # List all projects
POST   /api/projects                # Create new project
GET    /api/projects/my             # Get current user's projects
GET    /api/projects/{project_id}   # Get project details
PUT    /api/projects/{project_id}   # Update project
DELETE /api/projects/{project_id}   # Delete project

# Collaboration
POST   /api/projects/{project_id}/collaborators      # Add collaborator
DELETE /api/projects/{project_id}/collaborators/{user_id} # Remove collaborator
```

#### Table Management
```
POST   /api/projects/{project_id}/tables             # Create table
GET    /api/projects/{project_id}/tables             # Get project tables
GET    /api/projects/{project_id}/tables/{table_id}  # Get table details
PUT    /api/projects/{project_id}/tables/{table_id}  # Update table
DELETE /api/projects/{project_id}/tables/{table_id}  # Delete table
PUT    /api/projects/{project_id}/tables/{table_id}/position # Update table position
```

#### Field Management
```
POST   /api/projects/{project_id}/tables/{table_id}/fields             # Create field
GET    /api/projects/{project_id}/tables/{table_id}/fields             # Get table fields
PUT    /api/projects/{project_id}/tables/{table_id}/fields/reorder     # Reorder fields
GET    /api/projects/{project_id}/tables/{table_id}/fields/{field_id}  # Get field details
PUT    /api/projects/{project_id}/tables/{table_id}/fields/{field_id}  # Update field
DELETE /api/projects/{project_id}/tables/{table_id}/fields/{field_id}  # Delete field
```

#### Relationship Management
```
POST   /api/projects/{project_id}/relationships                     # Create relationship
GET    /api/projects/{project_id}/relationships                     # Get project relationships
GET    /api/projects/{project_id}/relationships/{relationship_id}   # Get relationship details
PUT    /api/projects/{project_id}/relationships/{relationship_id}   # Update relationship
DELETE /api/projects/{project_id}/relationships/{relationship_id}   # Delete relationship
```

#### Collaboration Sessions
```
POST   /api/projects/{project_id}/sessions                  # Create collaboration session
GET    /api/projects/{project_id}/sessions                  # Get project sessions
GET    /api/projects/{project_id}/sessions/active           # Get active sessions
GET    /api/projects/{project_id}/sessions/{session_id}     # Get session details
PUT    /api/projects/{project_id}/sessions/{session_id}     # Update session
DELETE /api/projects/{project_id}/sessions/{session_id}     # Delete session
PUT    /api/projects/{project_id}/sessions/{session_id}/cursor    # Update cursor position
PUT    /api/projects/{project_id}/sessions/{session_id}/inactive  # Set session inactive
```

### WebSocket Endpoint
```
GET /api/projects/{project_id}/collaborate # WebSocket connection for real-time collaboration
```

### Response Format

All API responses follow this consistent structure:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response payload
  }
}
```

## Real-time Collaboration (WebSocket Implementation)

### WebSocket Architecture

The WebSocket system is fully implemented with the following components:

#### Hub Structure
- **Client Management**: Tracks connected clients per project
- **Message Broadcasting**: Distributes messages to project participants
- **Heartbeat System**: 30-second ping/pong for connection health
- **Graceful Cleanup**: Handles client disconnections and project cleanup

#### Message Types
```go
const (
    MessageTypeUserJoined    = "user_joined"
    MessageTypeUserLeft      = "user_left"
    MessageTypeUserPresence  = "user_presence"
    MessageTypeCursorMove    = "cursor_move"
    MessageTypeTableCreate   = "table_create"
    MessageTypeTableUpdate   = "table_update"
    MessageTypeTableDelete   = "table_delete"
    MessageTypeFieldCreate   = "field_create"
    MessageTypeFieldUpdate   = "field_update"
    MessageTypeFieldDelete   = "field_delete"
    MessageTypePing          = "ping"
    MessageTypePong          = "pong"
)
```

#### Features Implemented
- **Live Cursors**: Real-time cursor tracking and display
- **User Presence**: Active user list with join/leave notifications
- **Schema Updates**: Broadcast table/field changes instantly
- **Activity Feed**: Live activity log of schema changes
- **Connection Management**: Automatic reconnection and cleanup

## Frontend State Management

### Svelte Stores

#### Authentication Store (`auth.ts`)
- User authentication state
- JWT token management
- Auto-initialization from localStorage

#### Project Store (`project.ts`)
- Project list and current project state
- Auto-save functionality with debouncing
- CRUD operations for projects

#### Flow Store (`flow.ts`)
- Canvas state for @xyflow/svelte
- Table nodes and relationship edges
- Position and viewport management

#### Collaboration Store (`collaboration.ts`)
- WebSocket connection management
- Real-time user presence
- Cursor tracking and activity feed

#### Designer Store (`designer.ts`)
- Visual designer UI state
- Selected elements and properties
- Tool selection and modes

#### UI Store (`ui.ts`)
- Toast notifications
- Modal states
- Global UI preferences

### Services Layer

#### API Client (`api.ts`)
- Axios-based HTTP client
- Automatic JWT token refresh
- Request/response interceptors
- Error handling

#### Service Classes
- **AuthService**: Authentication operations
- **ProjectService**: Project CRUD operations
- **UserService**: User management

## Environment Configuration

### Environment Files Structure

The project uses centralized environment configuration:
- **`.env.dev`**: Development environment settings
- **`.env.prod`**: Production environment settings

Both files are located in the project root and loaded by both backend and frontend.

### Backend Environment Loading
```go
// cmd/api/main.go
env := os.Getenv("ENV")
if env == "" {
    env = "development"
}

var envFile string
if env == "production" {
    envFile = "../.env.prod"
} else {
    envFile = "../.env.dev"
}

err := godotenv.Load(envFile)
```

### Frontend Environment Loading
```typescript
// vite.config.ts
export default defineConfig({
    envDir: '../', // Load environment variables from project root
    // ...
});
```

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

# Frontend Configuration
VITE_API_URL=http://localhost:8080/api
```

## Architecture Patterns

### Backend Patterns

#### Clean Architecture
- **Handlers**: HTTP request/response handling
- **Services**: Business logic and orchestration
- **Repositories**: Data access abstraction
- **Models**: Domain entities

#### Dependency Injection
```go
// server/server.go - Dependency setup
s.userRepo = repository.NewUserRepository(db)
s.userService = services.NewUserService(s.userRepo)
s.authService = services.NewAuthorizationService(s.projectRepo, s.tableRepo, s.fieldRepo, s.relationshipRepo, s.collaborationRepo)
```

#### Repository Pattern
```go
type ProjectRepositoryInterface interface {
    Create(project *models.Project) (uuid.UUID, error)
    GetByID(id uuid.UUID) (*models.Project, error)
    GetByOwnerID(ownerID uuid.UUID) ([]*models.Project, error)
    Update(project *models.Project) error
    Delete(id uuid.UUID) error
}
```

### Frontend Patterns

#### Component Architecture
- **UI Components**: Reusable ShadCN-Svelte components
- **Layout Components**: Page structure and navigation
- **Feature Components**: Domain-specific functionality
- **Flow Components**: Database canvas visualization

#### Store Pattern
```typescript
// Svelte store creation pattern
function createAuthStore() {
    const { subscribe, set, update } = writable(initialState);
    return {
        subscribe,
        init() { /* initialization logic */ },
        setUser(user: User) { /* state mutations */ },
        clear() { /* cleanup logic */ }
    };
}
```

## Security Implementation

### Authentication
- **JWT Tokens**: HS256 signing with configurable expiration
- **Refresh Tokens**: 7-day expiration with automatic rotation
- **Password Hashing**: bcrypt with salt

### Authorization
- **Middleware**: JWT verification on protected routes
- **Context Passing**: User ID available in request context
- **Resource Access**: Owner/collaborator checks for projects

### Input Validation
- **Backend**: go-playground/validator struct tags + custom validation
- **Frontend**: TypeScript interfaces + runtime validation

### Security Headers
- **CORS**: Configurable allowed origins
- **Request Limits**: Timeout and size restrictions

## Testing Infrastructure

### Backend Testing
```bash
# Test structure
backend/internal/
├── services/*_test.go          # Unit tests for business logic
├── handlers/*_test.go          # HTTP handler tests
├── websocket/*_test.go         # WebSocket functionality tests
└── testutil/test_helpers.go    # Test utilities and mocks
```

### Test Commands
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/services/
```

### Mock Generation
The project includes comprehensive mocks:
- Repository interfaces
- Service interfaces
- External dependencies

## Containerization

### Docker Configuration

#### Backend (`Dockerfile.prod`)
- Multi-stage build for optimized production image
- Go binary compilation and minimal runtime

#### Frontend (`Dockerfile.prod`, `Dockerfile.dev`)
- Node.js build environment
- Static file serving with nginx

#### Docker Compose
- **`docker-compose.dev.yml`**: Development environment
- **`docker-compose.prod.yml`**: Production environment

### Development vs Production

#### Development
- Hot reloading for both backend and frontend
- Database running in container
- Environment variable loading from `.env.dev`

#### Production
- Optimized builds and minimal images
- nginx serving static frontend files
- Health checks and restart policies

## Key File Locations

### Backend Entry Points
- **`backend/cmd/api/main.go`**: Application startup
- **`backend/internal/api/server/server.go`**: Server initialization
- **`backend/internal/api/routes/routes.go`**: Route definitions

### Frontend Entry Points
- **`frontend/src/routes/+layout.svelte`**: Root layout
- **`frontend/src/routes/projects/[id]/edit/+page.svelte`**: Visual schema editor
- **`frontend/vite.config.ts`**: Build configuration

### Configuration Files
- **`backend/internal/config/config.go`**: Backend configuration
- **`backend/internal/db/db.go`**: Database connection and migration
- **`frontend/package.json`**: Frontend dependencies and scripts

This documentation reflects the actual implementation of EzModel as of the current codebase state. All endpoints, models, and features described are implemented and functional.

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.

Frontend uses pnpm as the package manager. Run `pnpm build` and `pnpm check` when finish implementing.