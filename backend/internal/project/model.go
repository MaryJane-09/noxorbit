package project

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrNotFound = errors.New("project not found")
	ErrForbidden = errors.New("only workspace owners can change project status")
	ErrInvalidName = errors.New("project name must contain 1 to 100 characters")
	ErrInvalidDescription = errors.New("project description cannot exceed 2000 characters")
	ErrInvalidStatus = errors.New("project status must be active or archived")
)

type Project struct {
	ID uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name string `json:"name"`
	Description string `json:"description"`
	Status string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Input struct {
	Name string `json:"name"`
	Description string `json:"description"`
}

type StatusInput struct {
	Status string `json:"status"`
}

func Normalize(input Input) (Input, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if n := utf8.RuneCountInString(input.Name); n < 1 || n > 100 {
		return Input{}, ErrInvalidName
	}
	if utf8.RuneCountInString(input.Description) > 2000 {
		return Input{}, ErrInvalidDescription
	}
	return input, nil
}

func NormalizeStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "active" && status != "archived" {
		return "", ErrInvalidStatus
	}
	return status, nil
}
