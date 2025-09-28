package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/api/middleware"
	"github.com/Bug-Bugger/ezmodel/internal/api/responses"
	"github.com/Bug-Bugger/ezmodel/internal/api/utils"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectService services.ProjectServiceInterface
}

func NewProjectHandler(projectService services.ProjectServiceInterface) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

func (h *ProjectHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context
		userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			responses.RespondWithError(w, http.StatusUnauthorized, "User context not found")
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID")
			return
		}

		// Parse and validate request body
		var req dto.CreateProjectRequest
		if !utils.DecodeAndValidate(w, r, &req) {
			return
		}

		// Create project through service
		project, err := h.projectService.CreateProject(req.Name, req.Description, userID)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserNotFound):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid owner")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to create project")
			}
			return
		}

		// Create project response
		projectResponse := dto.ProjectSummaryResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			OwnerID:     project.OwnerID,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusCreated, "Project created successfully", projectResponse)
	}
}

func (h *ProjectHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := utils.ParseUUIDParamWithError(w, r, "project_id", "Invalid project ID")
		if !ok {
			return
		}

		project, err := h.projectService.GetProjectByID(id)
		if err != nil {
			if errors.Is(err, services.ErrProjectNotFound) {
				responses.RespondWithError(w, http.StatusNotFound, "Project not found")
			} else {
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve project")
			}
			return
		}

		// Convert to response format with full details
		var collaboratorResponses []dto.UserResponse
		for _, collaborator := range project.Collaborators {
			collaboratorResponses = append(collaboratorResponses, dto.UserResponse{
				ID:       collaborator.ID,
				Email:    collaborator.Email,
				Username: collaborator.Username,
			})
		}

		projectResponse := dto.ProjectResponse{
			ID:           project.ID,
			Name:         project.Name,
			Description:  project.Description,
			OwnerID:      project.OwnerID,
			DatabaseType: project.DatabaseType,
			CanvasData:   project.CanvasData,
			Owner: dto.UserResponse{
				ID:       project.Owner.ID,
				Email:    project.Owner.Email,
				Username: project.Owner.Username,
			},
			Collaborators: collaboratorResponses,
			CreatedAt:     project.CreatedAt,
			UpdatedAt:     project.UpdatedAt,
		}

		// Debug logging for project retrieval
		log.Printf("CANVAS DEBUG: Returning project %s with canvas data length: %d",
			project.ID.String(), len(project.CanvasData))

		responses.RespondWithSuccess(w, http.StatusOK, "Project retrieved successfully", projectResponse)
	}
}

func (h *ProjectHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := utils.ParseUUIDParamWithError(w, r, "project_id", "Invalid project ID")
		if !ok {
			return
		}

		var req dto.UpdateProjectRequest
		if !utils.DecodeAndValidate(w, r, &req) {
			return
		}

		// Empty update request
		if req.Name == nil && req.Description == nil && req.CanvasData == nil {
			responses.RespondWithError(w, http.StatusBadRequest, "No fields to update provided")
			return
		}

		project, err := h.projectService.UpdateProject(id, &req)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrProjectNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Project not found")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to update project")
			}
			return
		}

		projectResponse := dto.ProjectSummaryResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			OwnerID:     project.OwnerID,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Project updated successfully", projectResponse)
	}
}

func (h *ProjectHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := utils.ParseUUIDParamWithError(w, r, "project_id", "Invalid project ID")
		if !ok {
			return
		}

		if err := h.projectService.DeleteProject(id); err != nil {
			if errors.Is(err, services.ErrProjectNotFound) {
				responses.RespondWithError(w, http.StatusNotFound, "Project not found")
			} else {
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to delete project")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Project deleted successfully", nil)
	}
}

func (h *ProjectHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := h.projectService.GetAllProjects()
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve projects")
			return
		}

		var projectResponses []dto.ProjectSummaryResponse
		for _, project := range projects {
			projectResponses = append(projectResponses, dto.ProjectSummaryResponse{
				ID:          project.ID,
				Name:        project.Name,
				Description: project.Description,
				OwnerID:     project.OwnerID,
				CreatedAt:   project.CreatedAt,
				UpdatedAt:   project.UpdatedAt,
			})
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Projects retrieved successfully", projectResponses)
	}
}

func (h *ProjectHandler) GetMyProjects() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context
		userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			responses.RespondWithError(w, http.StatusUnauthorized, "User context not found")
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID")
			return
		}

		// Get projects owned by user
		ownedProjects, err := h.projectService.GetProjectsByOwnerID(userID)
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve owned projects")
			return
		}

		// Get projects where user is collaborator
		collaboratedProjects, err := h.projectService.GetProjectsByCollaboratorID(userID)
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve collaborated projects")
			return
		}

		// Combine and deduplicate
		projectMap := make(map[uuid.UUID]*dto.ProjectSummaryResponse)

		for _, project := range ownedProjects {
			projectMap[project.ID] = &dto.ProjectSummaryResponse{
				ID:          project.ID,
				Name:        project.Name,
				Description: project.Description,
				OwnerID:     project.OwnerID,
				CreatedAt:   project.CreatedAt,
				UpdatedAt:   project.UpdatedAt,
			}
		}

		for _, project := range collaboratedProjects {
			if _, exists := projectMap[project.ID]; !exists {
				projectMap[project.ID] = &dto.ProjectSummaryResponse{
					ID:          project.ID,
					Name:        project.Name,
					Description: project.Description,
					OwnerID:     project.OwnerID,
					CreatedAt:   project.CreatedAt,
					UpdatedAt:   project.UpdatedAt,
				}
			}
		}

		// Convert map to slice
		var projectResponses []dto.ProjectSummaryResponse
		for _, project := range projectMap {
			projectResponses = append(projectResponses, *project)
		}

		responses.RespondWithSuccess(w, http.StatusOK, "My projects retrieved successfully", projectResponses)
	}
}

func (h *ProjectHandler) AddCollaborator() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := utils.ParseUUIDParamWithError(w, r, "project_id", "Invalid project ID")
		if !ok {
			return
		}

		var req dto.AddCollaboratorRequest
		if !utils.DecodeAndValidate(w, r, &req) {
			return
		}

		err := h.projectService.AddCollaborator(projectID, req.CollaboratorID)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrProjectNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Project not found")
			case errors.Is(err, services.ErrCollaboratorNotFound):
				responses.RespondWithError(w, http.StatusBadRequest, "Collaborator not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to add collaborator")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Collaborator added successfully", nil)
	}
}

func (h *ProjectHandler) RemoveCollaborator() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := utils.ParseUUIDParamWithError(w, r, "project_id", "Invalid project ID")
		if !ok {
			return
		}

		collaboratorID, ok := utils.ParseUUIDParam(w, r, "user_id")
		if !ok {
			return
		}

		err := h.projectService.RemoveCollaborator(projectID, collaboratorID)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrProjectNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Project not found")
			case errors.Is(err, services.ErrCollaboratorNotFound):
				responses.RespondWithError(w, http.StatusBadRequest, "Collaborator not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to remove collaborator")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Collaborator removed successfully", nil)
	}
}
