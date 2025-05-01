package server

import (
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/routes"
	"github.com/Bug-Bugger/ezmodel/internal/config"
)

type Server struct {
	config *config.Config
	router *http.ServeMux
}

func New(cfg *config.Config) *Server {
	s := &Server{
		config: cfg,
		router: http.NewServeMux(),
	}

	routes.SetupRoutes(s.router)

	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.config.Port, s.router)
}
