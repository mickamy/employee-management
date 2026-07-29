package model

import (
	"time"

	"github.com/google/uuid"
)

// Event is a domain event in an employee's assignment stream.
type Event interface {
	EventType() string
}

// AssignmentDecided records the decision to assign; assigned_on is the date
// the decision takes effect, not when it occurred (docs/data-modeling.md).
type AssignmentDecided struct {
	AssignmentID uuid.UUID `json:"assignment_id"`
	DepartmentID uuid.UUID `json:"department_id"`
	Position     Position  `json:"position"`
	AssignedOn   time.Time `json:"assigned_on"`
}

func (AssignmentDecided) EventType() string { return "AssignmentDecided" }

// ReleaseDecided records the decision to release; released_on is the date
// the assignment ends, which may lie in the future.
type ReleaseDecided struct {
	AssignmentID uuid.UUID `json:"assignment_id"`
	ReleasedOn   time.Time `json:"released_on"`
}

func (ReleaseDecided) EventType() string { return "ReleaseDecided" }

// AssignmentRevoked cancels a decision that has not yet taken effect. The
// decided event stays untouched; projections stop counting it.
type AssignmentRevoked struct {
	AssignmentID uuid.UUID `json:"assignment_id"`
	Reason       string    `json:"reason"`
}

func (AssignmentRevoked) EventType() string { return "AssignmentRevoked" }

// ReleaseRevoked cancels a release decision that has not yet taken effect;
// the assignment goes back to being open.
type ReleaseRevoked struct {
	AssignmentID uuid.UUID `json:"assignment_id"`
	Reason       string    `json:"reason"`
}

func (ReleaseRevoked) EventType() string { return "ReleaseRevoked" }
