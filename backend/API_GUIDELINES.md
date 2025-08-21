# API Development Guidelines

## Overview

This document outlines the standards and best practices for developing RESTful APIs in this backend system.

## RESTful API Design Principles

### Resource-Based URLs
Design URLs around resources, not actions:

```
✅ Good:
GET    /users              # Get all users
GET    /users/{id}         # Get specific user
POST   /users              # Create user
PUT    /users/{id}         # Update user
DELETE /users/{id}         # Delete user

❌ Bad:
GET    /getUsers           # Action-based
POST   /createUser         # Action-based
PUT    /updateUser/{id}    # Action-based
```

### HTTP Methods
Use appropriate HTTP methods for different operations:

- **GET**: Retrieve resources (safe, idempotent)
- **POST**: Create new resources (not idempotent)
- **PUT**: Update/replace entire resource (idempotent)
- **PATCH**: Partial updates (not typically used in our system)
- **DELETE**: Remove resources (idempotent)

### Status Codes
Use appropriate HTTP status codes:

#### Success Codes
- **200 OK**: Successful GET, PUT, DELETE
- **201 Created**: Successful POST (resource creation)
- **204 No Content**: Successful operation with no response body

#### Client Error Codes
- **400 Bad Request**: Invalid request data/validation errors
- **401 Unauthorized**: Authentication required
- **403 Forbidden**: Valid auth but insufficient permissions
- **404 Not Found**: Resource doesn't exist
- **409 Conflict**: Resource conflict (duplicate email, etc.)

#### Server Error Codes
- **500 Internal Server Error**: Unexpected server errors

## Request/Response Patterns

### Standard Response Format
All API responses follow this structure:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response payload
  }
}
```

Error responses:
```json
{
  "success": false,
  "message": "Error description",
  "errors": {
    "field_name": "Specific error message"
  }
}
```

### Request Validation
All requests should be validated at multiple levels:

1. **JSON Structure**: Valid JSON format
2. **Required Fields**: All required fields present
3. **Field Validation**: Data types, formats, constraints
4. **Business Rules**: Domain-specific validation

Example validation in handlers:
```go
func (h *UserHandler) Create() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req dto.CreateUserRequest
        
        // 1. JSON structure validation
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            responses.RespondWithError(w, http.StatusBadRequest, "Invalid JSON format")
            return
        }
        
        // 2. Field validation
        if err := validation.Validate(req); err != nil {
            errors := validation.ValidationErrors(err)
            responses.RespondWithValidationErrors(w, errors)
            return
        }
        
        // 3. Business logic continues in service layer...
    }
}
```

## Authentication and Authorization

### JWT Token Structure
```go
type CustomClaims struct {
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
    jwt.RegisteredClaims
}
```

### Authorization Header
```
Authorization: Bearer <jwt_token>
```

### Protected Route Implementation
```go
// In routes.go
r.Group(func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    
    r.Route("/users", func(r chi.Router) {
        r.Get("/", userHandler.GetAll())
        r.Get("/{id}", userHandler.GetByID())
    })
})
```

### Context-Based User Access
```go
func (h *UserHandler) GetProfile() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, ok := middleware.GetUserIDFromContext(r.Context())
        if !ok {
            responses.RespondWithError(w, http.StatusUnauthorized, "User context not found")
            return
        }
        
        // Use userID for operations...
    }
}
```

## Route Organization

### Route Grouping
Organize routes by resource and protection level:

```go
func SetupRoutes(r *chi.Mux, handlers...) {
    // Public routes
    r.Post("/register", userHandler.Create())
    r.Post("/login", authHandler.Login())
    r.Post("/refresh-token", authHandler.RefreshToken())
    
    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)
        
        // User management
        r.Route("/users", func(r chi.Router) {
            r.Get("/", userHandler.GetAll())
            r.Route("/{id}", func(r chi.Router) {
                r.Get("/", userHandler.GetByID())
                r.Put("/", userHandler.Update())
                r.Delete("/", userHandler.Delete())
            })
        })
        
        // Other resources...
    })
}
```

### URL Parameters
Use consistent parameter naming:

```go
// Route parameter extraction
func (h *UserHandler) GetByID() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idStr := chi.URLParam(r, "id")
        id, err := uuid.Parse(idStr)
        if err != nil {
            responses.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
            return
        }
        
        // Continue with business logic...
    }
}
```

## Error Handling Standards

### Error Response Consistency
Always use the standardized response functions:

```go
// For general errors
responses.RespondWithError(w, http.StatusBadRequest, "Error message")

// For validation errors
responses.RespondWithValidationErrors(w, validationErrors)

// For successful operations
responses.RespondWithSuccess(w, http.StatusOK, "Success message", data)
```

### Error Logging
Log errors appropriately but don't expose internal details to clients:

```go
user, err := h.userService.GetUserByID(id)
if err != nil {
    if errors.Is(err, services.ErrUserNotFound) {
        responses.RespondWithError(w, http.StatusNotFound, "User not found")
    } else {
        log.Printf("Error retrieving user %s: %v", id, err) // Log internal error
        responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
    }
    return
}
```

## Pagination Standards

### Query Parameters
For endpoints returning lists, support pagination:

```
GET /users?page=1&limit=20&sort=created_at&order=desc
```

### Implementation Pattern
```go
type PaginationParams struct {
    Page  int    `json:"page" validate:"min=1"`
    Limit int    `json:"limit" validate:"min=1,max=100"`
    Sort  string `json:"sort"`
    Order string `json:"order" validate:"oneof=asc desc"`
}

type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Page       int         `json:"page"`
    Limit      int         `json:"limit"`
    Total      int64       `json:"total"`
    TotalPages int         `json:"total_pages"`
}
```

## Content Type Standards

### Request Content Type
- Always expect `application/json` for POST/PUT requests
- Validate Content-Type header when necessary

### Response Content Type
- Always return `application/json`
- Set appropriate headers in response functions

```go
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, _ := json.Marshal(payload)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}
```

## Security Best Practices

### Input Sanitization
```go
func (s *UserService) CreateUser(email, username, password string) (*models.User, error) {
    // Sanitize inputs
    email = strings.TrimSpace(strings.ToLower(email))
    username = strings.TrimSpace(username)
    
    // Validate after sanitization
    if len(email) < 5 || len(username) < 3 {
        return nil, ErrInvalidInput
    }
    
    // Continue processing...
}
```

### Password Handling
```go
// Always hash passwords before storage
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    return nil, err
}

// Never return password hashes in responses
type User struct {
    PasswordHash string `json:"-" gorm:"type:varchar(255);not null"`
}
```

### Rate Limiting (Recommended Implementation)
For production deployments, implement rate limiting:

```go
// Middleware for rate limiting (to be implemented)
func RateLimitMiddleware(requests int, window time.Duration) func(http.Handler) http.Handler {
    // Implementation depends on chosen rate limiting strategy
    // (in-memory, Redis-based, etc.)
}
```

## Documentation Standards

### Handler Documentation
Document each handler with clear descriptions:

```go
// Create handles user registration
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User registration data"
// @Success 201 {object} dto.APIResponse{data=models.User}
// @Failure 400 {object} dto.APIResponse
// @Router /register [post]
func (h *UserHandler) Create() http.HandlerFunc {
    // Implementation...
}
```

### API Endpoint Catalog
Maintain a clear catalog of all endpoints:

#### Authentication Endpoints
- `POST /register` - User registration
- `POST /login` - User login
- `POST /refresh-token` - Token refresh

#### User Management Endpoints
- `GET /users` - List all users (admin)
- `GET /users/{id}` - Get user by ID
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user
- `POST /users/{id}/verify-email` - Verify user email
- `PUT /users/{id}/password` - Update password

## Testing Standards

### Handler Testing
```go
func TestUserHandler_Create(t *testing.T) {
    tests := []struct {
        name           string
        requestBody    dto.CreateUserRequest
        setupMocks     func(*mocks.UserServiceInterface)
        expectedStatus int
        expectedBody   string
    }{
        {
            name: "successful user creation",
            requestBody: dto.CreateUserRequest{
                Email:    "test@example.com",
                Username: "testuser",
                Password: "password123",
            },
            setupMocks: func(m *mocks.UserServiceInterface) {
                m.On("CreateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
                  Return(&models.User{}, nil)
            },
            expectedStatus: http.StatusCreated,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation...
        })
    }
}
```

## Performance Considerations

### Database Queries
- Use selective field loading when possible
- Implement proper indexing strategies
- Avoid N+1 queries with proper preloading

```go
// Good: Preload related data
func (r *UserRepository) GetWithPosts(id uuid.UUID) (*models.User, error) {
    var user models.User
    err := r.db.Preload("Posts").First(&user, "id = ?", id).Error
    return &user, err
}

// Good: Select specific fields when full model not needed
func (r *UserRepository) GetEmailByID(id uuid.UUID) (string, error) {
    var email string
    err := r.db.Model(&models.User{}).Select("email").Where("id = ?", id).Scan(&email).Error
    return email, err
}
```

### Response Optimization
- Implement appropriate caching strategies
- Use compression for large responses
- Consider pagination for large datasets

These guidelines ensure consistent, secure, and maintainable API development across the entire backend system.