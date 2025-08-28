package services

import "errors"

var (
	// Shared errors
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	
	// User errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	
	// Project errors
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrCollaboratorNotFound = errors.New("collaborator not found")
)
