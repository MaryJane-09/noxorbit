package workspace

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid"
)

var (
	ErrInvalidName = errors.New("workspace name must contain 1 to 100 characters")
	ErrInvalidPurpose = errors.New("workspace purpose must be general, education, development, or business")
	ErrNotFound = errors.New("workspace not found")
	ErrForbidden = errors.New("only the workspace owner can edit its settings")
)

type Workspace struct {
	ID uuid.UUID `json:"id"`
	Name string `json:"name"`
	Purpose string `json:"purpose"`
	Role string `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRequest struct {
	Name string `json:"name"`
	Purpose string `json:"purpose"`
}

// NormalizeCreate validates a workspace before any database work.
// A purpose is a preference; it never changes membership permissions.
func NormalizeCreate(req CreateRequest) (CreateRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	if n := utf8.RuneCountInString(req.Name); n < 1 || n > 100 {
		return CreateRequest{}, ErrInvalidName
	}

	req.Purpose = strings.ToLower(strings.TrimSpace(req.Purpose))
	if req.Purpose == "" {
		req.Purpose = "general"
	}
	switch req.Purpose {
	case "general", "education", "development", "business":
		return req, nil
	default:
		return CreateRequest{}, ErrInvalidPurpose
	}
}
