package routes

import (
	"github.com/Bug-Bugger/ezmodel/internal/api/handlers"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r *chi.Mux, userService services.UserServiceInterface) {
	// Basic routes
	r.Get("/", handlers.HomeHandler())
	r.Get("/api", handlers.APIHandler())

	// User routes
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService)
	r.Post("/login", authHandler.Login())

	r.Route("/users", func(r chi.Router) {
		r.Get("/", userHandler.GetAll())
		r.Post("/", userHandler.Create())

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", userHandler.GetByID())
			r.Put("/", userHandler.Update())
			r.Delete("/", userHandler.Delete())
			r.Post("/verify-email", userHandler.VerifyEmail())
			r.Put("/password", userHandler.UpdatePassword())
		})
	})
}
