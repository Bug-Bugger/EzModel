package server

import (
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/middleware"
	"github.com/Bug-Bugger/ezmodel/internal/api/routes"
	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type Server struct {
	config         *config.Config
	router         *chi.Mux
	db             *gorm.DB
	userRepo       repository.UserRepositoryInterface
	projectRepo    repository.ProjectRepositoryInterface
	userService    services.UserServiceInterface
	projectService services.ProjectServiceInterface
	jwtService     *services.JWTService
	authMiddleware *middleware.AuthMiddleware
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

	// Initialize repositories
	s.userRepo = repository.NewUserRepository(db)
	s.projectRepo = repository.NewProjectRepository(db)

	// Initialize services
	s.userService = services.NewUserService(s.userRepo)
	s.projectService = services.NewProjectService(s.projectRepo, s.userRepo)
	s.jwtService = services.NewJWTService(cfg)

	// Initialize middleware
	s.authMiddleware = middleware.NewAuthMiddleware(s.jwtService)

	// Setup routes
	routes.SetupRoutes(s.router, s.userService, s.projectService, s.jwtService, s.authMiddleware)

	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.config.Port, s.router)
}
