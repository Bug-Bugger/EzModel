package server

import (
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/routes"
	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"gorm.io/gorm"
)

type Server struct {
	config   *config.Config
	router   *http.ServeMux
	db       *gorm.DB
	userRepo *repository.UserRepository
}

func New(cfg *config.Config, db *gorm.DB) *Server {
	s := &Server{
		config: cfg,
		router: http.NewServeMux(),
		db:     db,
	}

	// Initialize repositories
	s.userRepo = repository.NewUserRepository(db)

	// Setup routes
	routes.SetupRoutes(s.router, s.userRepo)

	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.config.Port, s.router)
}
