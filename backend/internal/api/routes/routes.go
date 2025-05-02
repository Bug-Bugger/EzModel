package routes

import (
	"github.com/Bug-Bugger/ezmodel/internal/api/handlers"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r *chi.Mux, userRepo *repository.UserRepository) {
	// Basic routes
	r.Get("/", handlers.HomeHandler())
	r.Get("/api", handlers.APIHandler())

	// User routes
	userHandler := handlers.NewUserHandler(userRepo)

	r.Route("/users", func(r chi.Router) {
		r.Get("/", userHandler.GetAll())
		r.Post("/", userHandler.Create())

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", userHandler.GetByID())
			r.Put("/", userHandler.Update())
			r.Delete("/", userHandler.Delete())
		})
	})
}
