package task

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid"
)

var (
	ErrNotFound = errors.New("task or project not found")
	ErrForbidden = errors.New("not permitted to change this task")
	ErrInvalidAssignee = errors.New("assignee must be a member of this workspace")
	ErrArchived = errors.New("archived projects cannot be changed")
	ErrInvalidTitle = errors.New("task title must contain 1 to 200 characters")
	ErrInvalidDescription = errors.New("task description cannot exceed 5000 characters")
	ErrInvalidStatus = errors.New("task status must be todo, in_progress, or done")
	ErrInvalidPriority = errors.New("task priority must be low, medium, high, or urgent")
	ErrInvalidDueDate = errors.New("due_date must be a valid YYYY-MM-DD date")
	ErrPastDueDate = errors.New("due date cannot be in the past")
)

type Task struct {
	ID uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Title string `json:"title"`
	Description string `json:"description"`
	Status string `json:"status"`
	Priority string `json:"priority"`
	DueDate *string `json:"due_date"`
	CreatedBy *uuid.UUID `json:"created_by"`
	AssigneeID *uuid.UUID `json:"assignee_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DetailsInput is used for creation and owner/creator edits.
// Assignees are intentionally unable to edit these fields themselves.
type DetailsInput struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Priority string `json:"priority"`
	DueDate *string `json:"due_date"`
	AssigneeID *uuid.UUID `json:"assignee_id"`
}

type StatusInput struct {
	Status string `json:"status"`
}

func NormalizeDetails(in DetailsInput) (DetailsInput, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.Priority = strings.ToLower(strings.TrimSpace(in.Priority))
	if in.Priority == "" {
		in.Priority = "medium"
	}
	if n := utf8.RuneCountInString(in.Title); n < 1 || n > 200 {
		return DetailsInput{}, ErrInvalidTitle
	}
	if utf8.RuneCountInString(in.Description) > 5000 {
		return DetailsInput{}, ErrInvalidDescription
	}
	switch in.Priority {
	case "low", "medium", "high", "urgent":
	default:
		return DetailsInput{}, ErrInvalidPriority
	}
	if in.DueDate != nil {
		// Strict date-only representation avoids timezone-driven deadline shifts.
		value := *in.DueDate
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil || parsed.Format("2006-01-02") != value {
			return DetailsInput{}, ErrInvalidDueDate
		}
	}
	if in.AssigneeID != nil && *in.AssigneeID == uuid.Nil {
		return DetailsInput{}, ErrInvalidAssignee
	}
	return in, nil
}

func NormalizeStatus(value string) (string, error) {
	status := strings.ToLower(strings.TrimSpace(value))
	switch status {
	case "todo", "in_progress", "done":
		return status, nil
	default:
		return "", ErrInvalidStatus
	}
}

	
// ValidateNewDueDate rejects past dates. Existing overdue dates can be
// preserved when editing other fields; moving one earlier is not allowed.
func ValidateNewDueDate(date *string, previous *string, now time.Time) error {
	if date == nil || (previous != nil && *date == *previous) {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", *date)
	if err != nil { return ErrInvalidDueDate }
	// Workspace deadlines are calendar dates. Use the app's Nigerian
	// business timezone, independently of the deployment server timezone.
	location, err := time.LoadLocation("Africa/Lagos")
	if err != nil { return err }
	today := now.In(location).Format("2006-01-02")
	if parsed.Format("2006-01-02") < today { return ErrPastDueDate }
	return nil
}
