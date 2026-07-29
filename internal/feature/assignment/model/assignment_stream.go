package model

import (
	"time"

	"github.com/google/uuid"
)

// Stream is the fold of one employee's assignment events: the state commands
// decide against. Revoked assignments are removed from it.
type Stream struct {
	EmployeeID   uuid.UUID
	LastRevision int64
	assignments  []Assignment
}

// Replay folds events into a Stream. lastRevision is the stream revision of
// the last event, used as the optimistic-concurrency expectation on append.
func Replay(employeeID uuid.UUID, lastRevision int64, events []Event) Stream {
	s := Stream{EmployeeID: employeeID, LastRevision: lastRevision}
	for _, e := range events {
		s.apply(e)
	}
	return s
}

// Latest returns the most recently decided assignment still in the stream.
func (s Stream) Latest() (Assignment, bool) {
	if len(s.assignments) == 0 {
		return Assignment{}, false
	}
	return s.assignments[len(s.assignments)-1], true
}

// Open returns the unreleased assignment, if any. v1 allows at most one.
func (s Stream) Open() (Assignment, bool) {
	for _, a := range s.assignments {
		if a.ReleasedOn == nil {
			return a, true
		}
	}
	return Assignment{}, false
}

// Assignment returns the assignment with the given id. Revoked assignments
// are gone from the stream state.
func (s Stream) Assignment(id uuid.UUID) (Assignment, bool) {
	for _, a := range s.assignments {
		if a.ID == id {
			return a, true
		}
	}
	return Assignment{}, false
}

// LastReleasedOn returns the latest release date in the stream, if any. A new
// assignment must not start before it.
func (s Stream) LastReleasedOn() (time.Time, bool) {
	var last time.Time
	var found bool
	for _, a := range s.assignments {
		if a.ReleasedOn != nil && a.ReleasedOn.After(last) {
			last = *a.ReleasedOn
			found = true
		}
	}
	return last, found
}

func (s *Stream) apply(e Event) {
	switch e := e.(type) {
	case AssignmentDecided:
		s.assignments = append(s.assignments, Assignment{
			ID:           e.AssignmentID,
			EmployeeID:   s.EmployeeID,
			DepartmentID: e.DepartmentID,
			Position:     e.Position,
			AssignedOn:   e.AssignedOn,
		})
	case ReleaseDecided:
		for i := range s.assignments {
			if s.assignments[i].ID == e.AssignmentID {
				s.assignments[i].ReleasedOn = new(e.ReleasedOn)
				return
			}
		}
	case AssignmentRevoked:
		for i := range s.assignments {
			if s.assignments[i].ID == e.AssignmentID {
				s.assignments = append(s.assignments[:i], s.assignments[i+1:]...)
				return
			}
		}
	case ReleaseRevoked:
		for i := range s.assignments {
			if s.assignments[i].ID == e.AssignmentID {
				s.assignments[i].ReleasedOn = nil
				return
			}
		}
	}
}
