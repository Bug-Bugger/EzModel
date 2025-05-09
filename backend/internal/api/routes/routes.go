package routes

import (
	"github.com/Bug-Bugger/ezmodel/internal/api/handlers"
	"github.com/Bug-Bugger/ezmodel/internal/api/middleware"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(
	r *chi.Mux,
	userService services.UserServiceInterface,
	jwtService *services.JWTService,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Basic routes
	r.Get("/", handlers.HomeHandler())
	r.Get("/api", handlers.APIHandler())

	// User handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, jwtService)

	// Public auth routes
	r.Post("/login", authHandler.Login())
	r.Post("/refresh-token", authHandler.RefreshToken())
	r.Post("/register", userHandler.Create())

	// Protected routes
	r.Group(func(r chi.Router) {
		// Apply JWT authentication middleware
		r.Use(authMiddleware.Authenticate)

		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/", userHandler.GetAll())

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", userHandler.GetByID())
				r.Put("/", userHandler.Update())
				r.Delete("/", userHandler.Delete())
				r.Post("/verify-email", userHandler.VerifyEmail())
				r.Put("/password", userHandler.UpdatePassword())
			})
		})
	})
}
