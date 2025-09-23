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

	// Table errors
	ErrTableNotFound = errors.New("table not found")

	// Field errors
	ErrFieldNotFound = errors.New("field not found")

	// Relationship errors
	ErrRelationshipNotFound = errors.New("relationship not found")

	// Collaboration session errors
	ErrSessionNotFound = errors.New("collaboration session not found")
)
