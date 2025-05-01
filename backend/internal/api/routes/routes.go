package routes

import (
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/handlers"
	"github.com/Bug-Bugger/ezmodel/internal/api/middleware"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
)

func SetupRoutes(mux *http.ServeMux, userRepo *repository.UserRepository) {
	// Basic routes
	mux.Handle("/", middleware.Logger(handlers.HomeHandler()))
	mux.Handle("/api", middleware.Logger(handlers.APIHandler()))

	// User routes
	userHandler := handlers.NewUserHandler(userRepo)
	mux.Handle("/users", middleware.Logger(userHandler.GetAll()))
	mux.Handle("/users/", middleware.Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetByID().ServeHTTP(w, r)
		case http.MethodPut:
			userHandler.Update().ServeHTTP(w, r)
		case http.MethodDelete:
			userHandler.Delete().ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/users/create", middleware.Logger(userHandler.Create()))
}
