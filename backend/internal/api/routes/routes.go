package routes

import (
	"github.com/Bug-Bugger/ezmodel/internal/api/handlers"
	"github.com/Bug-Bugger/ezmodel/internal/api/middleware"
	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	websocketPkg "github.com/Bug-Bugger/ezmodel/internal/websocket"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(
	r *chi.Mux,
	cfg *config.Config,
	userService services.UserServiceInterface,
	projectService services.ProjectServiceInterface,
	tableService services.TableServiceInterface,
	fieldService services.FieldServiceInterface,
	relationshipService services.RelationshipServiceInterface,
	collaborationService services.CollaborationSessionServiceInterface,
	jwtService *services.JWTService,
	authMiddleware *middleware.AuthMiddleware,
	websocketHub *websocketPkg.Hub,
) {
	// Basic routes
	r.Get("/", handlers.HomeHandler())

	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, jwtService, cfg)
	projectHandler := handlers.NewProjectHandler(projectService)
	tableHandler := handlers.NewTableHandler(tableService)
	fieldHandler := handlers.NewFieldHandler(fieldService)
	relationshipHandler := handlers.NewRelationshipHandler(relationshipService)
	collaborationHandler := handlers.NewCollaborationHandler(collaborationService)
	websocketHandler := handlers.NewWebSocketHandler(cfg, websocketHub, jwtService, userService, projectService, tableService)

	// Mount all API routes under /api prefix
	r.Route("/api", func(r chi.Router) {
		// API info route
		r.Get("/", handlers.APIHandler())

		// Public auth routes
		r.Post("/login", authHandler.Login())
		r.Post("/refresh-token", authHandler.RefreshToken())
		r.Post("/register", userHandler.Create())
		r.Post("/logout", authHandler.Logout())

		// WebSocket routes (handle authentication internally)
		r.Get("/projects/{project_id}/collaborate", websocketHandler.HandleWebSocket) // WebSocket endpoint for real-time collaboration

		// Protected routes
		r.Group(func(r chi.Router) {
			// Apply JWT authentication middleware
			r.Use(authMiddleware.Authenticate)

			// Current user route
			r.Get("/me", userHandler.GetMe())

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Get("/", userHandler.GetAll())

				r.Route("/{user_id}", func(r chi.Router) {
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

				r.Route("/{project_id}", func(r chi.Router) {
					r.Get("/", projectHandler.GetByID())
					r.Put("/", projectHandler.Update())
					r.Delete("/", projectHandler.Delete())
					r.Post("/collaborators", projectHandler.AddCollaborator())
					r.Delete("/collaborators/{user_id}", projectHandler.RemoveCollaborator())

					// Table routes within projects
					r.Route("/tables", func(r chi.Router) {
						r.Post("/", tableHandler.Create())        // Create table in project
						r.Get("/", tableHandler.GetByProjectID()) // Get all tables in project

						r.Route("/{table_id}", func(r chi.Router) {
							r.Get("/", tableHandler.GetByID())                // Get specific table
							r.Put("/", tableHandler.Update())                 // Update table
							r.Delete("/", tableHandler.Delete())              // Delete table
							r.Put("/position", tableHandler.UpdatePosition()) // Update table position

							// Field routes within tables
							r.Route("/fields", func(r chi.Router) {
								r.Post("/", fieldHandler.Create())        // Create field in table
								r.Get("/", fieldHandler.GetByTableID())   // Get all fields in table
								r.Put("/reorder", fieldHandler.Reorder()) // Reorder fields

								r.Route("/{field_id}", func(r chi.Router) {
									r.Get("/", fieldHandler.GetByID())   // Get specific field
									r.Put("/", fieldHandler.Update())    // Update field
									r.Delete("/", fieldHandler.Delete()) // Delete field
								})
							})
						})
					})

					// Relationship routes within projects
					r.Route("/relationships", func(r chi.Router) {
						r.Post("/", relationshipHandler.Create())        // Create relationship in project
						r.Get("/", relationshipHandler.GetByProjectID()) // Get all relationships in project

						r.Route("/{relationship_id}", func(r chi.Router) {
							r.Get("/", relationshipHandler.GetByID())   // Get specific relationship
							r.Put("/", relationshipHandler.Update())    // Update relationship
							r.Delete("/", relationshipHandler.Delete()) // Delete relationship
						})
					})

					// Collaboration session routes within projects
					r.Route("/sessions", func(r chi.Router) {
						r.Post("/", collaborationHandler.Create())                    // Create collaboration session
						r.Get("/", collaborationHandler.GetByProjectID())             // Get all sessions in project
						r.Get("/active", collaborationHandler.GetActiveByProjectID()) // Get active sessions

						r.Route("/{session_id}", func(r chi.Router) {
							r.Get("/", collaborationHandler.GetByID())             // Get specific session
							r.Put("/", collaborationHandler.Update())              // Update session
							r.Delete("/", collaborationHandler.Delete())           // Delete session
							r.Put("/cursor", collaborationHandler.UpdateCursor())  // Update cursor position
							r.Put("/inactive", collaborationHandler.SetInactive()) // Set session inactive
						})
					})
				})
			})

		})
	})
}
