# Database Design and Model Guide

## Overview

This guide covers database design patterns, model definitions, and data access strategies used in this backend system.

## Database Architecture

### Technology Stack
- **Database**: PostgreSQL
- **ORM**: GORM (Go Object-Relational Mapping)
- **Migration**: GORM Auto-Migration for development
- **Connection**: Connection pooling via GORM

### Connection Configuration
```go
// Database connection with proper configuration
func Connect(cfg *config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
        cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
        cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode)

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        // Add configuration options as needed
    })
    if err != nil {
        return nil, err
    }

    // Auto-migrate schemas
    err = db.AutoMigrate(&models.User{})
    if err != nil {
        return nil, err
    }

    return db, nil
}
```

## Model Design Patterns

### Base Model Structure
All models should follow this consistent structure:

```go
type BaseModel struct {
    ID        uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime;type:timestamp with time zone"`
    UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime;type:timestamp with time zone"`
    DeletedAt *time.Time `json:"-" gorm:"index"` // Soft delete support
}
```

### Model Definition Standards
Each model should include proper tags and validation:

```go
type User struct {
    ID            uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Email         string     `json:"email" gorm:"type:varchar(255);unique;not null" validate:"required,email"`
    PasswordHash  string     `json:"-" gorm:"type:varchar(255);not null" validate:"required"`
    Username      string     `json:"username" gorm:"type:varchar(100);not null" validate:"required,min=3,max=100"`
    AvatarURL     string     `json:"avatar_url,omitempty" gorm:"type:varchar(500)"`
    EmailVerified bool       `json:"email_verified" gorm:"default:false"`
    LastLoginAt   *time.Time `json:"last_login_at,omitempty" gorm:"type:timestamp with time zone"`
    CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime;type:timestamp with time zone"`
    UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime;type:timestamp with time zone"`
}
```

### Tag Conventions

#### GORM Tags
- `primaryKey`: Designates primary key
- `type:uuid`: Explicit PostgreSQL UUID type
- `default:gen_random_uuid()`: PostgreSQL UUID generation
- `autoCreateTime`/`autoUpdateTime`: Automatic timestamp management
- `unique`: Unique constraint
- `not null`: NOT NULL constraint
- `index`: Creates database index
- `foreignKey`: Specifies foreign key relationship

#### JSON Tags
- Use snake_case for JSON field names
- Use `omitempty` for optional fields
- Use `"-"` to exclude sensitive fields (passwords, internal IDs)

#### Validation Tags
- `required`: Field must be present
- `email`: Valid email format
- `min`/`max`: String length or numeric range
- Custom validation tags as needed

### Relationship Patterns

#### One-to-Many Relationships
```go
type User struct {
    ID    uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Posts []Post    `json:"posts,omitempty" gorm:"foreignKey:AuthorID"`
}

type Post struct {
    ID       uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    AuthorID uuid.UUID `json:"author_id" gorm:"type:uuid;not null"`
    Author   User      `json:"author" gorm:"foreignKey:AuthorID"`
    Title    string    `json:"title" gorm:"type:varchar(255);not null"`
    Content  string    `json:"content" gorm:"type:text"`
}
```

#### Many-to-Many Relationships
```go
type User struct {
    ID    uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Roles []Role    `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}

type Role struct {
    ID    uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Name  string    `json:"name" gorm:"type:varchar(100);unique;not null"`
    Users []User    `json:"users,omitempty" gorm:"many2many:user_roles;"`
}
```

#### Polymorphic Relationships (When Needed)
```go
type Comment struct {
    ID            uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Content       string    `json:"content" gorm:"type:text;not null"`
    CommentableID uuid.UUID `json:"commentable_id" gorm:"type:uuid;not null"`
    CommentableType string  `json:"commentable_type" gorm:"type:varchar(50);not null"`
}
```

## Repository Pattern Implementation

### Interface Definition
Always define repository interfaces for testability:

```go
type UserRepositoryInterface interface {
    Create(user *models.User) (uuid.UUID, error)
    GetByID(id uuid.UUID) (*models.User, error)
    GetByEmail(email string) (*models.User, error)
    GetAll() ([]*models.User, error)
    Update(user *models.User) error
    Delete(id uuid.UUID) error
    
    // Specialized queries
    GetActiveUsers() ([]*models.User, error)
    GetUsersByRole(role string) ([]*models.User, error)
}
```

### Repository Implementation
```go
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) (uuid.UUID, error) {
    result := r.db.Create(user)
    if result.Error != nil {
        return uuid.Nil, result.Error
    }
    return user.ID, nil
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
    return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uuid.UUID) error {
    return r.db.Delete(&models.User{}, "id = ?", id).Error
}
```

### Query Optimization Patterns

#### Selective Field Loading
```go
func (r *UserRepository) GetUserProfile(id uuid.UUID) (*models.User, error) {
    var user models.User
    err := r.db.Select("id", "username", "email", "avatar_url", "created_at").
        First(&user, "id = ?", id).Error
    return &user, err
}
```

#### Preloading Relationships
```go
func (r *UserRepository) GetUserWithPosts(id uuid.UUID) (*models.User, error) {
    var user models.User
    err := r.db.Preload("Posts").First(&user, "id = ?", id).Error
    return &user, err
}

func (r *UserRepository) GetUsersWithRoles() ([]*models.User, error) {
    var users []*models.User
    err := r.db.Preload("Roles").Find(&users).Error
    return users, err
}
```

#### Complex Queries
```go
func (r *UserRepository) GetActiveUsersWithRecentActivity(days int) ([]*models.User, error) {
    var users []*models.User
    cutoffDate := time.Now().AddDate(0, 0, -days)
    
    err := r.db.Where("email_verified = ? AND last_login_at > ?", true, cutoffDate).
        Find(&users).Error
    return users, err
}
```

#### Pagination Support
```go
func (r *UserRepository) GetUsersPaginated(page, limit int, sortBy, order string) ([]*models.User, int64, error) {
    var users []*models.User
    var total int64
    
    offset := (page - 1) * limit
    
    // Count total records
    if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Get paginated results
    query := r.db.Offset(offset).Limit(limit)
    
    if sortBy != "" {
        orderClause := fmt.Sprintf("%s %s", sortBy, order)
        query = query.Order(orderClause)
    }
    
    err := query.Find(&users).Error
    return users, total, err
}
```

## Migration Strategies

### Development Migrations
Use GORM AutoMigrate for development:

```go
func runMigrations(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.User{},
        &models.Post{},
        &models.Role{},
        // Add new models here
    )
}
```

### Production Migrations
For production, create explicit migration files:

```go
// migrations/001_create_users_table.go
func CreateUsersTable(db *gorm.DB) error {
    return db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            email VARCHAR(255) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            username VARCHAR(100) NOT NULL,
            avatar_url VARCHAR(500),
            email_verified BOOLEAN DEFAULT FALSE,
            last_login_at TIMESTAMP WITH TIME ZONE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
    `).Error
}
```

### Index Management
Create indexes for performance:

```go
func CreateIndexes(db *gorm.DB) error {
    indexes := []string{
        "CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
        "CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
        "CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
        "CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id)",
        "CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC)",
    }
    
    for _, index := range indexes {
        if err := db.Exec(index).Error; err != nil {
            return err
        }
    }
    
    return nil
}
```

## Data Validation Strategies

### Model-Level Validation
Use GORM hooks for model validation:

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // Validate business rules before creation
    if len(u.Username) < 3 {
        return errors.New("username must be at least 3 characters")
    }
    
    if u.Email == "" {
        return errors.New("email is required")
    }
    
    return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
    // Validate business rules before updates
    return u.BeforeCreate(tx)
}
```

### Database Constraints
Use database-level constraints for data integrity:

```sql
-- Email uniqueness
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);

-- Check constraints
ALTER TABLE users ADD CONSTRAINT check_email_format 
    CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

-- Foreign key constraints
ALTER TABLE posts ADD CONSTRAINT fk_posts_author 
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;
```

## Performance Best Practices

### Query Optimization
1. **Use Indexes**: Create indexes on frequently queried columns
2. **Limit Results**: Always use pagination for large datasets
3. **Select Specific Fields**: Don't select all columns when not needed
4. **Avoid N+1 Queries**: Use Preload for related data

### Connection Management
```go
func configureDatabase(db *gorm.DB) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    
    // Connection pool settings
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    return nil
}
```

### Monitoring and Logging
```go
func enableQueryLogging(db *gorm.DB) *gorm.DB {
    return db.Session(&gorm.Session{
        Logger: logger.Default.LogMode(logger.Info),
    })
}
```

## Security Considerations

### SQL Injection Prevention
- Always use parameterized queries (GORM handles this automatically)
- Validate input data before database operations
- Use allowlists for dynamic query building

### Data Encryption
```go
type SensitiveData struct {
    ID            uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    EncryptedData string    `json:"-" gorm:"type:text;not null"` // Never expose in JSON
    Salt          string    `json:"-" gorm:"type:varchar(255);not null"`
}
```

### Audit Logging
```go
type AuditLog struct {
    ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    UserID    uuid.UUID `gorm:"type:uuid;not null"`
    Action    string    `gorm:"type:varchar(100);not null"`
    TableName string    `gorm:"type:varchar(100);not null"`
    RecordID  uuid.UUID `gorm:"type:uuid;not null"`
    Changes   string    `gorm:"type:jsonb"` // JSON field for change tracking
    CreatedAt time.Time `gorm:"autoCreateTime"`
}
```

## Testing Database Code

### Repository Testing
```go
func TestUserRepository_Create(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    repo := NewUserRepository(db)
    
    user := &models.User{
        Email:    "test@example.com",
        Username: "testuser",
        PasswordHash: "hashedpassword",
    }
    
    id, err := repo.Create(user)
    
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, id)
    
    // Verify in database
    retrieved, err := repo.GetByID(id)
    assert.NoError(t, err)
    assert.Equal(t, user.Email, retrieved.Email)
}
```

### Test Database Setup
```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    // Run migrations
    err = db.AutoMigrate(&models.User{})
    require.NoError(t, err)
    
    return db
}
```

This guide ensures consistent, secure, and performant database operations across the entire backend system.