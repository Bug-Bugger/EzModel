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
	projectService services.ProjectServiceInterface,
	jwtService *services.JWTService,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Basic routes
	r.Get("/", handlers.HomeHandler())
	r.Get("/api", handlers.APIHandler())

	// User handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, jwtService)
	projectHandler := handlers.NewProjectHandler(projectService)

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
				r.Put("/password", userHandler.UpdatePassword())
			})
		})

		// Project routes
		r.Route("/projects", func(r chi.Router) {
			r.Post("/", projectHandler.Create())
			r.Get("/", projectHandler.GetAll())
			r.Get("/my", projectHandler.GetMyProjects())

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", projectHandler.GetByID())
				r.Put("/", projectHandler.Update())
				r.Delete("/", projectHandler.Delete())
				r.Post("/collaborators", projectHandler.AddCollaborator())
				r.Delete("/collaborators/{collaborator_id}", projectHandler.RemoveCollaborator())
			})
		})
	})
}
