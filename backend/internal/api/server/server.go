package server

import (
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/middleware"
	"github.com/Bug-Bugger/ezmodel/internal/api/routes"
	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	websocketPkg "github.com/Bug-Bugger/ezmodel/internal/websocket"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"
)

type Server struct {
	config               *config.Config
	router               *chi.Mux
	db                   *gorm.DB
	userRepo             repository.UserRepositoryInterface
	projectRepo          repository.ProjectRepositoryInterface
	tableRepo            repository.TableRepositoryInterface
	fieldRepo            repository.FieldRepositoryInterface
	relationshipRepo     repository.RelationshipRepositoryInterface
	collaborationRepo    repository.CollaborationSessionRepositoryInterface
	authService          services.AuthorizationServiceInterface
	userService          services.UserServiceInterface
	projectService       services.ProjectServiceInterface
	tableService         services.TableServiceInterface
	fieldService         services.FieldServiceInterface
	relationshipService  services.RelationshipServiceInterface
	collaborationService services.CollaborationSessionServiceInterface
	jwtService           *services.JWTService
	authMiddleware       *middleware.AuthMiddleware
	websocketHub         *websocketPkg.Hub
}

func New(cfg *config.Config, db *gorm.DB) *Server {
	s := &Server{
		config: cfg,
		router: chi.NewRouter(),
		db:     db,
	}

	// Apply global middleware
	s.router.Use(chiMiddleware.Logger)
	s.router.Use(chiMiddleware.Recoverer)

	// CORS middleware
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:4173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Initialize WebSocket hub
	s.websocketHub = websocketPkg.NewHub()

	// Initialize repositories
	s.userRepo = repository.NewUserRepository(db)
	s.projectRepo = repository.NewProjectRepository(db)
	s.tableRepo = repository.NewTableRepository(db)
	s.fieldRepo = repository.NewFieldRepository(db)
	s.relationshipRepo = repository.NewRelationshipRepository(db)
	s.collaborationRepo = repository.NewCollaborationSessionRepository(db)

	// Initialize authorization service first
	s.authService = services.NewAuthorizationService(s.projectRepo, s.tableRepo, s.fieldRepo, s.relationshipRepo, s.collaborationRepo)

	// Initialize services with authorization service
	s.userService = services.NewUserService(s.userRepo)
	s.collaborationService = services.NewCollaborationSessionService(s.collaborationRepo, s.projectRepo, s.userRepo, s.tableRepo, s.authService, s.websocketHub)
	s.projectService = services.NewProjectService(s.projectRepo, s.userRepo, s.collaborationService)
	s.tableService = services.NewTableService(s.tableRepo, s.projectRepo, s.authService, s.collaborationService)
	s.fieldService = services.NewFieldService(s.fieldRepo, s.tableRepo, s.authService, s.collaborationService)
	s.relationshipService = services.NewRelationshipService(s.relationshipRepo, s.projectRepo, s.tableRepo, s.fieldRepo, s.authService, s.collaborationService)
	s.jwtService = services.NewJWTService(cfg)

	// Initialize middleware
	s.authMiddleware = middleware.NewAuthMiddleware(s.jwtService)

	// Setup routes
	routes.SetupRoutes(s.router, s.userService, s.projectService, s.tableService, s.fieldService, s.relationshipService, s.collaborationService, s.jwtService, s.authMiddleware, s.websocketHub)

	return s
}

func (s *Server) Start() error {
	// Start WebSocket hub in goroutine
	go s.websocketHub.Run()

	return http.ListenAndServe(s.config.Port, s.router)
}
