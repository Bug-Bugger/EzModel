# Development Guide

## Code Structure and Patterns

This guide outlines the coding standards and patterns used in this backend to ensure consistency across the codebase.

## Naming Conventions

### Files and Packages
- **Package names**: lowercase, single word when possible (`handlers`, `services`, `models`)
- **File names**: snake_case (`auth_handlers.go`, `user_service.go`)
- **Test files**: `*_test.go` suffix

### Variables and Functions
- **Variables**: camelCase (`userID`, `accessToken`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Constants**: PascalCase for exported, camelCase for unexported
- **Interfaces**: end with `Interface` (`UserServiceInterface`)

### Database and JSON Tags
- **Database columns**: snake_case (`user_id`, `created_at`)
- **JSON fields**: snake_case (`user_id`, `access_token`)
- **GORM tags**: follow PostgreSQL conventions

## File Organization Patterns

### Handler Structure
Each handler file should follow this pattern:

```go
package handlers

import (
    // Standard library imports first
    "encoding/json"
    "net/http"
    
    // Third-party imports
    "github.com/google/uuid"
    
    // Internal imports (grouped by layer)
    "github.com/Bug-Bugger/ezmodel/internal/api/dto"
    "github.com/Bug-Bugger/ezmodel/internal/api/responses"
    "github.com/Bug-Bugger/ezmodel/internal/services"
)

type EntityHandler struct {
    entityService services.EntityServiceInterface
    // other dependencies
}

func NewEntityHandler(entityService services.EntityServiceInterface) *EntityHandler {
    return &EntityHandler{
        entityService: entityService,
    }
}

// HTTP handlers follow RESTful conventions
func (h *EntityHandler) Create() http.HandlerFunc { /* ... */ }
func (h *EntityHandler) GetByID() http.HandlerFunc { /* ... */ }
func (h *EntityHandler) Update() http.HandlerFunc { /* ... */ }
func (h *EntityHandler) Delete() http.HandlerFunc { /* ... */ }
```

### Service Structure
Services contain business logic and follow this pattern:

```go
package services

import (
    // imports organized as above
)

type EntityService struct {
    entityRepo repository.EntityRepositoryInterface
    // other dependencies
}

func NewEntityService(entityRepo repository.EntityRepositoryInterface) *EntityService {
    return &EntityService{
        entityRepo: entityRepo,
    }
}

// Business operations
func (s *EntityService) CreateEntity(params...) (*models.Entity, error) {
    // 1. Input validation
    // 2. Business logic
    // 3. Repository calls
    // 4. Return result
}
```

### Repository Structure
Repositories handle data persistence:

```go
package repository

import (
    "github.com/Bug-Bugger/ezmodel/internal/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type EntityRepositoryInterface interface {
    Create(entity *models.Entity) (uuid.UUID, error)
    GetByID(id uuid.UUID) (*models.Entity, error)
    Update(entity *models.Entity) error
    Delete(id uuid.UUID) error
}

type EntityRepository struct {
    db *gorm.DB
}

func NewEntityRepository(db *gorm.DB) EntityRepositoryInterface {
    return &EntityRepository{db: db}
}
```

## Error Handling Patterns

### Custom Errors
Define domain-specific errors in `services/errors.go`:

```go
var (
    ErrEntityNotFound    = errors.New("entity not found")
    ErrInvalidInput      = errors.New("invalid input")
    ErrEntityExists      = errors.New("entity already exists")
)
```

### Error Propagation
Follow this pattern through the stack:

```go
// Repository layer - return raw errors
func (r *EntityRepository) GetByID(id uuid.UUID) (*models.Entity, error) {
    var entity models.Entity
    err := r.db.First(&entity, "id = ?", id).Error
    if err != nil {
        return nil, err // Return GORM error as-is
    }
    return &entity, nil
}

// Service layer - wrap with domain errors
func (s *EntityService) GetEntityByID(id uuid.UUID) (*models.Entity, error) {
    entity, err := s.entityRepo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrEntityNotFound
        }
        return nil, err
    }
    return entity, nil
}

// Handler layer - convert to HTTP responses
func (h *EntityHandler) GetByID() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // ... parse ID from request
        
        entity, err := h.entityService.GetEntityByID(id)
        if err != nil {
            if errors.Is(err, services.ErrEntityNotFound) {
                responses.RespondWithError(w, http.StatusNotFound, "Entity not found")
            } else {
                responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
            }
            return
        }
        
        responses.RespondWithSuccess(w, http.StatusOK, "Entity retrieved", entity)
    }
}
```

## Validation Patterns

### Model Validation
Use struct tags for basic validation:

```go
type Entity struct {
    ID       uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Name     string    `json:"name" gorm:"type:varchar(255);not null" validate:"required,min=3,max=255"`
    Email    string    `json:"email" gorm:"type:varchar(255);unique" validate:"required,email"`
}
```

### Service Layer Validation
Add business rule validation in services:

```go
func (s *EntityService) CreateEntity(name, email string) (*models.Entity, error) {
    // Input sanitization
    name = strings.TrimSpace(name)
    email = strings.TrimSpace(email)
    
    // Business validation
    if len(name) < 3 {
        return nil, ErrInvalidInput
    }
    
    // Check business rules (e.g., uniqueness)
    existing, err := s.entityRepo.GetByEmail(email)
    if err == nil && existing != nil {
        return nil, ErrEntityExists
    }
    
    // Create entity...
}
```

### Request Validation
Validate DTOs in handlers:

```go
func (h *EntityHandler) Create() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req dto.CreateEntityRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            responses.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
            return
        }
        
        if err := validation.Validate(req); err != nil {
            errors := validation.ValidationErrors(err)
            responses.RespondWithValidationErrors(w, errors)
            return
        }
        
        // Process request...
    }
}
```

## DTO Patterns

### Request DTOs
Define clear request structures:

```go
type CreateEntityRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=255"`
    Email string `json:"email" validate:"required,email"`
}

type UpdateEntityRequest struct {
    Name  *string `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
}
```

### Response DTOs
Create response-specific structures when needed:

```go
type EntityResponse struct {
    ID        uuid.UUID `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (e *Entity) ToResponse() *EntityResponse {
    return &EntityResponse{
        ID:        e.ID,
        Name:      e.Name,
        Email:     e.Email,
        CreatedAt: e.CreatedAt,
        UpdatedAt: e.UpdatedAt,
    }
}
```

## Security Patterns

### Authentication Middleware
Protect routes using middleware:

```go
// In routes setup
r.Group(func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    
    r.Route("/entities", func(r chi.Router) {
        r.Post("/", entityHandler.Create())
        r.Get("/{id}", entityHandler.GetByID())
    })
})
```

### Authorization Patterns
Check permissions in handlers or services:

```go
func (h *EntityHandler) Delete() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, ok := middleware.GetUserIDFromContext(r.Context())
        if !ok {
            responses.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
            return
        }
        
        // Check if user owns the entity or has admin rights
        entity, err := h.entityService.GetEntityByID(entityID)
        if err != nil {
            // handle error
        }
        
        if entity.OwnerID.String() != userID {
            responses.RespondWithError(w, http.StatusForbidden, "Access denied")
            return
        }
        
        // Process deletion...
    }
}
```

## Database Patterns

### Model Definitions
Follow consistent model patterns:

```go
type Entity struct {
    ID        uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Name      string     `json:"name" gorm:"type:varchar(255);not null"`
    OwnerID   uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
    Owner     User       `json:"owner" gorm:"foreignKey:OwnerID"`
    CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt *time.Time `json:"-" gorm:"index"` // Soft delete
}
```

### Repository Queries
Use consistent query patterns:

```go
func (r *EntityRepository) GetByOwnerID(ownerID uuid.UUID) ([]*models.Entity, error) {
    var entities []*models.Entity
    err := r.db.Where("owner_id = ?", ownerID).Find(&entities).Error
    return entities, err
}

func (r *EntityRepository) GetWithOwner(id uuid.UUID) (*models.Entity, error) {
    var entity models.Entity
    err := r.db.Preload("Owner").First(&entity, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &entity, nil
}
```

## Testing Patterns

### Handler Tests
```go
func TestEntityHandler_Create(t *testing.T) {
    // Setup
    mockService := &mocks.EntityServiceInterface{}
    handler := handlers.NewEntityHandler(mockService)
    
    // Test cases
    tests := []struct {
        name           string
        requestBody    interface{}
        setupMocks     func()
        expectedStatus int
        expectedBody   string
    }{
        // test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Run test...
        })
    }
}
```

### Service Tests
```go
func TestEntityService_CreateEntity(t *testing.T) {
    // Setup
    mockRepo := &mocks.EntityRepositoryInterface{}
    service := services.NewEntityService(mockRepo)
    
    // Test implementation...
}
```

## Import Organization

Always organize imports in this order:
1. Standard library packages
2. Third-party packages  
3. Internal packages (grouped by layer)

```go
import (
    // Standard library
    "context"
    "encoding/json"
    "net/http"
    
    // Third-party
    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "gorm.io/gorm"
    
    // Internal - by layer
    "github.com/Bug-Bugger/ezmodel/internal/models"
    "github.com/Bug-Bugger/ezmodel/internal/repository"
    "github.com/Bug-Bugger/ezmodel/internal/services"
    "github.com/Bug-Bugger/ezmodel/internal/api/dto"
)
```

This structure ensures consistency and maintainability across the entire codebase.