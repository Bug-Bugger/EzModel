package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/db"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// Define command line flags
	clearOnly := flag.Bool("clear-only", false, "Only clear the database without seeding")
	seedOnly := flag.Bool("seed-only", false, "Only seed the database without clearing")
	force := flag.Bool("force", false, "Skip confirmation prompt")
	flag.Parse()

	// Load environment file
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
	if err != nil {
		log.Printf("Warning: No %s file found or error loading it. Using default values or environment variables.", envFile)
	}

	// Load configuration
	cfg := config.New()

	// Connect to database
	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}
	defer sqlDB.Close()

	// Initialize seeder
	seeder := NewSeeder(database)

	// Get confirmation if not forced
	if !*force {
		if !*seedOnly {
			fmt.Print("This will clear all data from the database. Are you sure? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Operation cancelled.")
				return
			}
		}
	}

	// Execute operations based on flags
	if *seedOnly {
		log.Println("Seeding database...")
		if err := seeder.SeedData(); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Database seeded successfully!")
	} else if *clearOnly {
		log.Println("Clearing database...")
		if err := seeder.ClearData(); err != nil {
			log.Fatalf("Failed to clear database: %v", err)
		}
		log.Println("Database cleared successfully!")
	} else {
		// Default: clear then seed
		log.Println("Clearing database...")
		if err := seeder.ClearData(); err != nil {
			log.Fatalf("Failed to clear database: %v", err)
		}
		log.Println("Database cleared successfully!")

		log.Println("Seeding database...")
		if err := seeder.SeedData(); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Database seeded successfully!")
	}
}

// Seeder handles database clearing and seeding operations
type Seeder struct {
	db                *gorm.DB
	userRepo          repository.UserRepositoryInterface
	projectRepo       repository.ProjectRepositoryInterface
	tableRepo         repository.TableRepositoryInterface
	fieldRepo         repository.FieldRepositoryInterface
	relationshipRepo  repository.RelationshipRepositoryInterface
	collaborationRepo repository.CollaborationSessionRepositoryInterface
}

// NewSeeder creates a new seeder instance
func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{
		db:                db,
		userRepo:          repository.NewUserRepository(db),
		projectRepo:       repository.NewProjectRepository(db),
		tableRepo:         repository.NewTableRepository(db),
		fieldRepo:         repository.NewFieldRepository(db),
		relationshipRepo:  repository.NewRelationshipRepository(db),
		collaborationRepo: repository.NewCollaborationSessionRepository(db),
	}
}

// ClearData removes all data from the database in the correct order
func (s *Seeder) ClearData() error {
	// Use a transaction for consistency
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Clear in dependency order to avoid foreign key conflicts

		// 1. Clear collaboration sessions
		if err := tx.Exec("DELETE FROM collaboration_sessions").Error; err != nil {
			return fmt.Errorf("failed to clear collaboration sessions: %w", err)
		}
		log.Println("✓ Cleared collaboration sessions")

		// 2. Clear relationships
		if err := tx.Exec("DELETE FROM relationships").Error; err != nil {
			return fmt.Errorf("failed to clear relationships: %w", err)
		}
		log.Println("✓ Cleared relationships")

		// 3. Clear fields
		if err := tx.Exec("DELETE FROM fields").Error; err != nil {
			return fmt.Errorf("failed to clear fields: %w", err)
		}
		log.Println("✓ Cleared fields")

		// 4. Clear tables
		if err := tx.Exec("DELETE FROM tables").Error; err != nil {
			return fmt.Errorf("failed to clear tables: %w", err)
		}
		log.Println("✓ Cleared tables")

		// 5. Clear project collaborators (many-to-many relationship)
		if err := tx.Exec("DELETE FROM project_collaborators").Error; err != nil {
			return fmt.Errorf("failed to clear project collaborators: %w", err)
		}
		log.Println("✓ Cleared project collaborators")

		// 6. Clear projects
		if err := tx.Exec("DELETE FROM projects").Error; err != nil {
			return fmt.Errorf("failed to clear projects: %w", err)
		}
		log.Println("✓ Cleared projects")

		// 7. Clear users
		if err := tx.Exec("DELETE FROM users").Error; err != nil {
			return fmt.Errorf("failed to clear users: %w", err)
		}
		log.Println("✓ Cleared users")

		return nil
	})
}

// SeedData populates the database with sample data
func (s *Seeder) SeedData() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create temporary seeder with transaction DB
		txSeeder := &Seeder{
			db:                tx,
			userRepo:          repository.NewUserRepository(tx),
			projectRepo:       repository.NewProjectRepository(tx),
			tableRepo:         repository.NewTableRepository(tx),
			fieldRepo:         repository.NewFieldRepository(tx),
			relationshipRepo:  repository.NewRelationshipRepository(tx),
			collaborationRepo: repository.NewCollaborationSessionRepository(tx),
		}

		// Seed users
		users, err := txSeeder.seedUsers()
		if err != nil {
			return fmt.Errorf("failed to seed users: %w", err)
		}
		log.Printf("✓ Created %d users", len(users))

		// Seed projects
		projects, err := txSeeder.seedProjects(users)
		if err != nil {
			return fmt.Errorf("failed to seed projects: %w", err)
		}
		log.Printf("✓ Created %d projects", len(projects))

		// Seed tables, fields, and relationships for each project
		for _, project := range projects {
			if err := txSeeder.seedProjectSchema(project); err != nil {
				return fmt.Errorf("failed to seed schema for project %s: %w", project.Name, err)
			}
		}
		log.Println("✓ Created tables, fields, and relationships")

		// Add collaborators to projects
		if err := txSeeder.seedCollaborators(projects, users); err != nil {
			return fmt.Errorf("failed to seed collaborators: %w", err)
		}
		log.Println("✓ Added project collaborators")

		return nil
	})
}

// seedUsers creates sample users
func (s *Seeder) seedUsers() ([]*models.User, error) {
	users := []*models.User{
		{
			Email:    "test1@example.com",
			Username: "test1",
		},
		{
			Email:    "test2@example.com",
			Username: "test2",
		},
		{
			Email:    "test3@example.com",
			Username: "test3",
		},
	}

	var createdUsers []*models.User
	for _, user := range users {
		hashedPassword, err := hashPassword("123321")
		if err != nil {
			return nil, fmt.Errorf("failed to hash password for user %s: %w", user.Username, err)
		}
		user.PasswordHash = hashedPassword

		userID, err := s.userRepo.Create(user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
		user.ID = userID
		createdUsers = append(createdUsers, user)
	}

	return createdUsers, nil
}

// seedProjects creates sample projects
func (s *Seeder) seedProjects(users []*models.User) ([]*models.Project, error) {
	projects := []*models.Project{
		{
			Name:         "Collaborative E-commerce Platform",
			Description:  "Complete online store with user management, product catalog, and order processing - collaborative project",
			OwnerID:      users[0].ID, // test1
			DatabaseType: "postgresql",
			CanvasData:   `{"zoom": 1, "position": {"x": 0, "y": 0}}`,
		},
	}

	var createdProjects []*models.Project
	for _, project := range projects {
		projectID, err := s.projectRepo.Create(project)
		if err != nil {
			return nil, fmt.Errorf("failed to create project %s: %w", project.Name, err)
		}
		project.ID = projectID
		createdProjects = append(createdProjects, project)
	}

	return createdProjects, nil
}

// seedProjectSchema creates tables, fields, and relationships for a project
func (s *Seeder) seedProjectSchema(project *models.Project) error {
	switch project.Name {
	case "Collaborative E-commerce Platform":
		return s.seedEcommerceSchema(project)
	default:
		return nil
	}
}

// seedEcommerceSchema creates e-commerce database schema
func (s *Seeder) seedEcommerceSchema(project *models.Project) error {
	// Create Users table
	usersTable := &models.Table{
		ProjectID: project.ID,
		Name:      "users",
		PosX:      100,
		PosY:      100,
	}
	usersTableID, err := s.tableRepo.Create(usersTable)
	if err != nil {
		return err
	}
	usersTable.ID = usersTableID

	// Create Users fields
	userFields := []*models.Field{
		{TableID: usersTableID, Name: "id", DataType: "SERIAL", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: usersTableID, Name: "email", DataType: "VARCHAR(255)", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: usersTableID, Name: "password_hash", DataType: "VARCHAR(255)", IsPrimaryKey: false, IsNullable: false, Position: 3},
		{TableID: usersTableID, Name: "first_name", DataType: "VARCHAR(100)", IsPrimaryKey: false, IsNullable: true, Position: 4},
		{TableID: usersTableID, Name: "last_name", DataType: "VARCHAR(100)", IsPrimaryKey: false, IsNullable: true, Position: 5},
		{TableID: usersTableID, Name: "created_at", DataType: "TIMESTAMP", IsPrimaryKey: false, IsNullable: false, DefaultValue: "CURRENT_TIMESTAMP", Position: 6},
	}

	var userIDField *models.Field
	for _, field := range userFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "id" {
			userIDField = field
		}
	}

	// Create Products table
	productsTable := &models.Table{
		ProjectID: project.ID,
		Name:      "products",
		PosX:      400,
		PosY:      100,
	}
	productsTableID, err := s.tableRepo.Create(productsTable)
	if err != nil {
		return err
	}
	productsTable.ID = productsTableID

	// Create Products fields
	productFields := []*models.Field{
		{TableID: productsTableID, Name: "id", DataType: "SERIAL", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: productsTableID, Name: "name", DataType: "VARCHAR(255)", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: productsTableID, Name: "description", DataType: "TEXT", IsPrimaryKey: false, IsNullable: true, Position: 3},
		{TableID: productsTableID, Name: "price", DataType: "DECIMAL(10,2)", IsPrimaryKey: false, IsNullable: false, Position: 4},
		{TableID: productsTableID, Name: "stock_quantity", DataType: "INTEGER", IsPrimaryKey: false, IsNullable: false, DefaultValue: "0", Position: 5},
		{TableID: productsTableID, Name: "created_at", DataType: "TIMESTAMP", IsPrimaryKey: false, IsNullable: false, DefaultValue: "CURRENT_TIMESTAMP", Position: 6},
	}

	var productIDField *models.Field
	for _, field := range productFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "id" {
			productIDField = field
		}
	}

	// Create Orders table
	ordersTable := &models.Table{
		ProjectID: project.ID,
		Name:      "orders",
		PosX:      250,
		PosY:      350,
	}
	ordersTableID, err := s.tableRepo.Create(ordersTable)
	if err != nil {
		return err
	}
	ordersTable.ID = ordersTableID

	// Create Orders fields
	orderFields := []*models.Field{
		{TableID: ordersTableID, Name: "id", DataType: "SERIAL", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: ordersTableID, Name: "user_id", DataType: "INTEGER", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: ordersTableID, Name: "total_amount", DataType: "DECIMAL(10,2)", IsPrimaryKey: false, IsNullable: false, Position: 3},
		{TableID: ordersTableID, Name: "status", DataType: "VARCHAR(50)", IsPrimaryKey: false, IsNullable: false, DefaultValue: "'pending'", Position: 4},
		{TableID: ordersTableID, Name: "created_at", DataType: "TIMESTAMP", IsPrimaryKey: false, IsNullable: false, DefaultValue: "CURRENT_TIMESTAMP", Position: 5},
	}

	var orderIDField, orderUserIDField *models.Field
	for _, field := range orderFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "id" {
			orderIDField = field
		}
		if field.Name == "user_id" {
			orderUserIDField = field
		}
	}

	// Create Order Items table
	orderItemsTable := &models.Table{
		ProjectID: project.ID,
		Name:      "order_items",
		PosX:      500,
		PosY:      350,
	}
	orderItemsTableID, err := s.tableRepo.Create(orderItemsTable)
	if err != nil {
		return err
	}
	orderItemsTable.ID = orderItemsTableID

	// Create Order Items fields
	orderItemFields := []*models.Field{
		{TableID: orderItemsTableID, Name: "id", DataType: "SERIAL", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: orderItemsTableID, Name: "order_id", DataType: "INTEGER", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: orderItemsTableID, Name: "product_id", DataType: "INTEGER", IsPrimaryKey: false, IsNullable: false, Position: 3},
		{TableID: orderItemsTableID, Name: "quantity", DataType: "INTEGER", IsPrimaryKey: false, IsNullable: false, Position: 4},
		{TableID: orderItemsTableID, Name: "unit_price", DataType: "DECIMAL(10,2)", IsPrimaryKey: false, IsNullable: false, Position: 5},
	}

	var orderItemOrderIDField, orderItemProductIDField *models.Field
	for _, field := range orderItemFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "order_id" {
			orderItemOrderIDField = field
		}
		if field.Name == "product_id" {
			orderItemProductIDField = field
		}
	}

	// Create relationships
	relationships := []*models.Relationship{
		{
			ProjectID:     project.ID,
			SourceTableID: ordersTable.ID,
			SourceFieldID: orderUserIDField.ID,
			TargetTableID: usersTable.ID,
			TargetFieldID: userIDField.ID,
			RelationType:  "many_to_one",
		},
		{
			ProjectID:     project.ID,
			SourceTableID: orderItemsTable.ID,
			SourceFieldID: orderItemOrderIDField.ID,
			TargetTableID: ordersTable.ID,
			TargetFieldID: orderIDField.ID,
			RelationType:  "many_to_one",
		},
		{
			ProjectID:     project.ID,
			SourceTableID: orderItemsTable.ID,
			SourceFieldID: orderItemProductIDField.ID,
			TargetTableID: productsTable.ID,
			TargetFieldID: productIDField.ID,
			RelationType:  "many_to_one",
		},
	}

	for _, rel := range relationships {
		_, err := s.relationshipRepo.Create(rel)
		if err != nil {
			return err
		}
	}

	return nil
}

// seedBlogSchema creates blog management schema
func (s *Seeder) seedBlogSchema(project *models.Project) error {
	// Users table
	usersTable := &models.Table{
		ProjectID: project.ID,
		Name:      "users",
		PosX:      150,
		PosY:      80,
	}
	usersTableID, err := s.tableRepo.Create(usersTable)
	if err != nil {
		return err
	}

	userFields := []*models.Field{
		{TableID: usersTableID, Name: "id", DataType: "INT AUTO_INCREMENT", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: usersTableID, Name: "username", DataType: "VARCHAR(50)", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: usersTableID, Name: "email", DataType: "VARCHAR(255)", IsPrimaryKey: false, IsNullable: false, Position: 3},
		{TableID: usersTableID, Name: "bio", DataType: "TEXT", IsPrimaryKey: false, IsNullable: true, Position: 4},
	}

	var userIDField *models.Field
	for _, field := range userFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "id" {
			userIDField = field
		}
	}

	// Posts table
	postsTable := &models.Table{
		ProjectID: project.ID,
		Name:      "posts",
		PosX:      400,
		PosY:      80,
	}
	postsTableID, err := s.tableRepo.Create(postsTable)
	if err != nil {
		return err
	}

	postFields := []*models.Field{
		{TableID: postsTableID, Name: "id", DataType: "INT AUTO_INCREMENT", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: postsTableID, Name: "title", DataType: "VARCHAR(255)", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: postsTableID, Name: "content", DataType: "LONGTEXT", IsPrimaryKey: false, IsNullable: false, Position: 3},
		{TableID: postsTableID, Name: "author_id", DataType: "INT", IsPrimaryKey: false, IsNullable: false, Position: 4},
		{TableID: postsTableID, Name: "published_at", DataType: "DATETIME", IsPrimaryKey: false, IsNullable: true, Position: 5},
	}

	var postIDField, postAuthorIDField *models.Field
	for _, field := range postFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "id" {
			postIDField = field
		}
		if field.Name == "author_id" {
			postAuthorIDField = field
		}
	}

	// Comments table
	commentsTable := &models.Table{
		ProjectID: project.ID,
		Name:      "comments",
		PosX:      275,
		PosY:      300,
	}
	commentsTableID, err := s.tableRepo.Create(commentsTable)
	if err != nil {
		return err
	}

	commentFields := []*models.Field{
		{TableID: commentsTableID, Name: "id", DataType: "INT AUTO_INCREMENT", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: commentsTableID, Name: "post_id", DataType: "INT", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: commentsTableID, Name: "author_id", DataType: "INT", IsPrimaryKey: false, IsNullable: false, Position: 3},
		{TableID: commentsTableID, Name: "content", DataType: "TEXT", IsPrimaryKey: false, IsNullable: false, Position: 4},
		{TableID: commentsTableID, Name: "created_at", DataType: "DATETIME", IsPrimaryKey: false, IsNullable: false, DefaultValue: "CURRENT_TIMESTAMP", Position: 5},
	}

	var commentPostIDField, commentAuthorIDField *models.Field
	for _, field := range commentFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "post_id" {
			commentPostIDField = field
		}
		if field.Name == "author_id" {
			commentAuthorIDField = field
		}
	}

	// Create relationships
	relationships := []*models.Relationship{
		{
			ProjectID:     project.ID,
			SourceTableID: postsTable.ID,
			SourceFieldID: postAuthorIDField.ID,
			TargetTableID: usersTable.ID,
			TargetFieldID: userIDField.ID,
			RelationType:  "many_to_one",
		},
		{
			ProjectID:     project.ID,
			SourceTableID: commentsTable.ID,
			SourceFieldID: commentPostIDField.ID,
			TargetTableID: postsTable.ID,
			TargetFieldID: postIDField.ID,
			RelationType:  "many_to_one",
		},
		{
			ProjectID:     project.ID,
			SourceTableID: commentsTable.ID,
			SourceFieldID: commentAuthorIDField.ID,
			TargetTableID: usersTable.ID,
			TargetFieldID: userIDField.ID,
			RelationType:  "many_to_one",
		},
	}

	for _, rel := range relationships {
		_, err := s.relationshipRepo.Create(rel)
		if err != nil {
			return err
		}
	}

	return nil
}

// seedTaskManagementSchema creates task management schema
func (s *Seeder) seedTaskManagementSchema(project *models.Project) error {
	// Teams table
	teamsTable := &models.Table{
		ProjectID: project.ID,
		Name:      "teams",
		PosX:      100,
		PosY:      50,
	}
	teamsTableID, err := s.tableRepo.Create(teamsTable)
	if err != nil {
		return err
	}

	teamFields := []*models.Field{
		{TableID: teamsTableID, Name: "id", DataType: "SERIAL", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: teamsTableID, Name: "name", DataType: "VARCHAR(100)", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: teamsTableID, Name: "description", DataType: "TEXT", IsPrimaryKey: false, IsNullable: true, Position: 3},
	}

	var teamIDField *models.Field
	for _, field := range teamFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "id" {
			teamIDField = field
		}
	}

	// Projects table
	projectsTable := &models.Table{
		ProjectID: project.ID,
		Name:      "projects",
		PosX:      350,
		PosY:      50,
	}
	projectsTableID, err := s.tableRepo.Create(projectsTable)
	if err != nil {
		return err
	}

	projectFields := []*models.Field{
		{TableID: projectsTableID, Name: "id", DataType: "SERIAL", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: projectsTableID, Name: "name", DataType: "VARCHAR(255)", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: projectsTableID, Name: "team_id", DataType: "INTEGER", IsPrimaryKey: false, IsNullable: false, Position: 3},
		{TableID: projectsTableID, Name: "deadline", DataType: "DATE", IsPrimaryKey: false, IsNullable: true, Position: 4},
	}

	var projectIDField, projectTeamIDField *models.Field
	for _, field := range projectFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "id" {
			projectIDField = field
		}
		if field.Name == "team_id" {
			projectTeamIDField = field
		}
	}

	// Tasks table
	tasksTable := &models.Table{
		ProjectID: project.ID,
		Name:      "tasks",
		PosX:      225,
		PosY:      250,
	}
	tasksTableID, err := s.tableRepo.Create(tasksTable)
	if err != nil {
		return err
	}

	taskFields := []*models.Field{
		{TableID: tasksTableID, Name: "id", DataType: "SERIAL", IsPrimaryKey: true, IsNullable: false, Position: 1},
		{TableID: tasksTableID, Name: "title", DataType: "VARCHAR(255)", IsPrimaryKey: false, IsNullable: false, Position: 2},
		{TableID: tasksTableID, Name: "description", DataType: "TEXT", IsPrimaryKey: false, IsNullable: true, Position: 3},
		{TableID: tasksTableID, Name: "project_id", DataType: "INTEGER", IsPrimaryKey: false, IsNullable: false, Position: 4},
		{TableID: tasksTableID, Name: "status", DataType: "VARCHAR(20)", IsPrimaryKey: false, IsNullable: false, DefaultValue: "'todo'", Position: 5},
		{TableID: tasksTableID, Name: "priority", DataType: "VARCHAR(10)", IsPrimaryKey: false, IsNullable: false, DefaultValue: "'medium'", Position: 6},
	}

	var taskProjectIDField *models.Field
	for _, field := range taskFields {
		fieldID, err := s.fieldRepo.Create(field)
		if err != nil {
			return err
		}
		field.ID = fieldID
		if field.Name == "project_id" {
			taskProjectIDField = field
		}
	}

	// Create relationships
	relationships := []*models.Relationship{
		{
			ProjectID:     project.ID,
			SourceTableID: projectsTable.ID,
			SourceFieldID: projectTeamIDField.ID,
			TargetTableID: teamsTable.ID,
			TargetFieldID: teamIDField.ID,
			RelationType:  "many_to_one",
		},
		{
			ProjectID:     project.ID,
			SourceTableID: tasksTable.ID,
			SourceFieldID: taskProjectIDField.ID,
			TargetTableID: projectsTable.ID,
			TargetFieldID: projectIDField.ID,
			RelationType:  "many_to_one",
		},
	}

	for _, rel := range relationships {
		_, err := s.relationshipRepo.Create(rel)
		if err != nil {
			return err
		}
	}

	return nil
}

// seedCollaborators adds collaborators to projects
func (s *Seeder) seedCollaborators(projects []*models.Project, users []*models.User) error {
	// Add test2 and test3 as collaborators to test1's collaborative e-commerce project
	if len(projects) > 0 && len(users) >= 3 {
		// Add test2 as collaborator
		err := s.projectRepo.AddCollaborator(projects[0].ID, users[1].ID)
		if err != nil {
			return err
		}

		// Add test3 as collaborator
		err = s.projectRepo.AddCollaborator(projects[0].ID, users[2].ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// hashPassword creates a bcrypt hash of the password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
