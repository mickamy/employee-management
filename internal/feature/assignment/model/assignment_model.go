package model

import (
	"time"

	"github.com/google/uuid"
)

type Position string

const (
	PositionMember  Position = "MEMBER"
	PositionManager Position = "MANAGER"
)

func (p Position) String() string {
	return string(p)
}

// Assignment is a read shape composed from the decision and release events;
// it is not a stored entity (docs/data-modeling.md).
type Assignment struct {
	ID           uuid.UUID
	EmployeeID   uuid.UUID
	DepartmentID uuid.UUID
	Position     Position
	AssignedOn   time.Time
	// ReleasedOn is nil while the assignment is unreleased.
	ReleasedOn *time.Time
}

// Active reports whether the assignment is in effect on the given date.
func (a Assignment) Active(on time.Time) bool {
	if a.AssignedOn.After(on) {
		return false
	}
	return a.ReleasedOn == nil || a.ReleasedOn.After(on)
}

// ManagerTenure is a derived read model entry: who managed the employee, for
// which department, over which interval.
type ManagerTenure struct {
	ManagerEmployeeID uuid.UUID
	ManagerName       string
	DepartmentID      uuid.UUID
	StartedOn         time.Time
	// EndedOn is nil while the tenure is ongoing.
	EndedOn *time.Time
}
