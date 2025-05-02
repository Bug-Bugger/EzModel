package server

import (
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/routes"
	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type Server struct {
	config   *config.Config
	router   *chi.Mux
	db       *gorm.DB
	userRepo *repository.UserRepository
}

func New(cfg *config.Config, db *gorm.DB) *Server {
	s := &Server{
		config: cfg,
		router: chi.NewRouter(),
		db:     db,
	}

	// Apply global middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Initialize repositories
	s.userRepo = repository.NewUserRepository(db)

	// Setup routes
	routes.SetupRoutes(s.router, s.userRepo)

	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.config.Port, s.router)
}
