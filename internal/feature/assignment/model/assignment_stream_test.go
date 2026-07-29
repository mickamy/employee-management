package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mickamy/employee-management/internal/lib/times"

	"github.com/mickamy/employee-management/internal/feature/assignment/model"
)

func TestReplay_DecidedOpensAssignment(t *testing.T) {
	t.Parallel()

	employeeID := uuid.New()
	assignmentID := uuid.New()
	stream := model.Replay(employeeID, 1, []model.Event{
		model.AssignmentDecided{
			AssignmentID: assignmentID,
			DepartmentID: uuid.New(),
			Position:     model.PositionMember,
			AssignedOn:   times.Date(2026, 4, 1),
		},
	})

	open, ok := stream.Open()
	require.True(t, ok)
	assert.Equal(t, assignmentID, open.ID)
	assert.Equal(t, employeeID, open.EmployeeID)
	assert.Nil(t, open.ReleasedOn)
	assert.Equal(t, int64(1), stream.LastRevision)
}

func TestReplay_ReleasedClosesAssignment(t *testing.T) {
	t.Parallel()

	assignmentID := uuid.New()
	stream := model.Replay(uuid.New(), 2, []model.Event{
		model.AssignmentDecided{
			AssignmentID: assignmentID, DepartmentID: uuid.New(),
			Position: model.PositionMember, AssignedOn: times.Date(2026, 4, 1),
		},
		model.ReleaseDecided{AssignmentID: assignmentID, ReleasedOn: times.Date(2026, 9, 30)},
	})

	_, open := stream.Open()
	assert.False(t, open)

	a, ok := stream.Assignment(assignmentID)
	require.True(t, ok)
	require.NotNil(t, a.ReleasedOn)
	assert.Equal(t, times.Date(2026, 9, 30), *a.ReleasedOn)

	last, ok := stream.LastReleasedOn()
	require.True(t, ok)
	assert.Equal(t, times.Date(2026, 9, 30), last)
}

func TestReplay_RevokedRemovesAssignment(t *testing.T) {
	t.Parallel()

	assignmentID := uuid.New()
	stream := model.Replay(uuid.New(), 2, []model.Event{
		model.AssignmentDecided{
			AssignmentID: assignmentID, DepartmentID: uuid.New(),
			Position: model.PositionMember, AssignedOn: times.Date(2026, 4, 1),
		},
		model.AssignmentRevoked{AssignmentID: assignmentID, Reason: "wrong department"},
	})

	_, open := stream.Open()
	assert.False(t, open)
	_, found := stream.Assignment(assignmentID)
	assert.False(t, found)
}

func TestReplay_ReleaseRevokedReopensAssignment(t *testing.T) {
	t.Parallel()

	assignmentID := uuid.New()
	stream := model.Replay(uuid.New(), 3, []model.Event{
		model.AssignmentDecided{
			AssignmentID: assignmentID, DepartmentID: uuid.New(),
			Position: model.PositionMember, AssignedOn: times.Date(2026, 4, 1),
		},
		model.ReleaseDecided{AssignmentID: assignmentID, ReleasedOn: times.Date(2026, 9, 30)},
		model.ReleaseRevoked{AssignmentID: assignmentID, Reason: "plans changed"},
	})

	open, ok := stream.Open()
	require.True(t, ok)
	assert.Equal(t, assignmentID, open.ID)
	assert.Nil(t, open.ReleasedOn)
	_, released := stream.LastReleasedOn()
	assert.False(t, released)
}

func TestAssignment_Active(t *testing.T) {
	t.Parallel()

	releasedOn := times.Date(2026, 9, 30)
	a := model.Assignment{AssignedOn: times.Date(2026, 4, 1), ReleasedOn: &releasedOn}

	assert.False(t, a.Active(times.Date(2026, 3, 31)), "before assigned_on")
	assert.True(t, a.Active(times.Date(2026, 4, 1)), "on assigned_on")
	assert.True(t, a.Active(times.Date(2026, 9, 29)), "before released_on")
	assert.False(t, a.Active(times.Date(2026, 9, 30)), "on released_on the employee is no longer assigned")
}
