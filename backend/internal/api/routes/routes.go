package routes

import (
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/handlers"
	"github.com/Bug-Bugger/ezmodel/internal/api/middleware"
)

func SetupRoutes(mux *http.ServeMux) {
	mux.Handle("/", middleware.Logger(handlers.HomeHandler()))
	mux.Handle("/api", middleware.Logger(handlers.APIHandler()))
}
