package task

import (
	"errors"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
)

func TestNormalizeDetails(t *testing.T) {
	date := "2026-11-20"
	in, err := NormalizeDetails(DetailsInput{
		Title: "  Prepare slides  ", Description: "  Review presentation  ",
		Priority: " HIGH ", DueDate: &date,
	})
	if err != nil { t.Fatal(err) }
	if in.Title != "Prepare slides" || in.Description != "Review presentation" ||
		in.Priority != "high" || *in.DueDate != date {
		t.Fatalf("unexpected normalized details: %+v", in)
	}

	defaults, err := NormalizeDetails(DetailsInput{Title: "Test"})
	if err != nil || defaults.Priority != "medium" || defaults.AssigneeID != nil || defaults.DueDate != nil {
		t.Fatalf("invalid defaults: %+v %v", defaults, err)
	}
}

func TestNormalizeDetailsRejectsInvalidFields(t *testing.T) {
	cases := []struct {
		name string
		input DetailsInput
		want error
	}{
		{"empty title", DetailsInput{Title: "   "}, ErrInvalidTitle},
		{"long title", DetailsInput{Title: strings.Repeat("x", 201)}, ErrInvalidTitle},
		{"unicode title", DetailsInput{Title: strings.Repeat("界", 201)}, ErrInvalidTitle},
		{"long description", DetailsInput{Title: "Task", Description: strings.Repeat("a", 5001)}, ErrInvalidDescription},
		{"bad priority", DetailsInput{Title: "Task", Priority: "critical"}, ErrInvalidPriority},
		{"nonexistent day", DetailsInput{Title: "Task", DueDate: ptr("2026-02-30")}, ErrInvalidDueDate},
		{"timestamp not date", DetailsInput{Title: "Task", DueDate: ptr("2026-11-20T13:00:00Z")}, ErrInvalidDueDate},
		{"empty date", DetailsInput{Title: "Task", DueDate: ptr("")}, ErrInvalidDueDate},
		{"nil uuid", DetailsInput{Title: "Task", AssigneeID: uuidPtr(uuid.Nil)}, ErrInvalidAssignee},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NormalizeDetails(tc.input); !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestNormalizeStatus(t *testing.T) {
	for _, value := range []string{"todo", "in_progress", "done"} {
		got, err := NormalizeStatus(" " + strings.ToUpper(value) + " ")
		if err != nil || got != value { t.Fatalf("%q: %q %v", value, got, err) }
	}
	for _, value := range []string{"", "in progress", "archived", "pending"} {
		if _, err := NormalizeStatus(value); !errors.Is(err, ErrInvalidStatus) {
			t.Fatalf("expected invalid status for %q; got %v", value, err)
		}
	}
}

func ptr(value string) *string { return &value }
func uuidPtr(value uuid.UUID) *uuid.UUID { return &value }
