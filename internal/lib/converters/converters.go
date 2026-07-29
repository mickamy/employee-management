// Package converters registers shared type converters for automapper and
// exposes them for manual use in handlers.
package converters

import (
	"time"

	"github.com/google/uuid"
	"github.com/mickamy/mapgen/runtime/mapper"
	"google.golang.org/genproto/googleapis/type/date"

	assignmentv1 "github.com/mickamy/employee-management/gen/assignment/v1"
	amodel "github.com/mickamy/employee-management/internal/feature/assignment/model"
)

func init() {
	mapper.Register(ToDate)
	mapper.Register(UUIDToString)
	mapper.Register(ToTime)
	mapper.RegisterE(uuid.Parse)
	mapper.Register(ToAssignmentv1Position)
}

// UUIDToString renders a UUID in its canonical string form.
func UUIDToString(id uuid.UUID) string {
	return id.String()
}

// ToTime converts a calendar date into a UTC midnight time.
func ToTime(d *date.Date) time.Time {
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, time.UTC)
}

// ToDate converts a time into its calendar date.
func ToDate(t time.Time) *date.Date {
	year, month, day := t.Date()
	//nolint:gosec // Calendar year, month, and day always fit in int32.
	return &date.Date{Year: int32(year), Month: int32(month), Day: int32(day)}
}

func ToAssignmentv1Position(p amodel.Position) assignmentv1.Position {
	switch p {
	case amodel.PositionMember:
		return assignmentv1.Position_POSITION_MEMBER
	case amodel.PositionManager:
		return assignmentv1.Position_POSITION_MANAGER
	}
	return assignmentv1.Position_POSITION_UNSPECIFIED
}

func ToAssignmentPosition(p assignmentv1.Position) amodel.Position {
	switch p {
	case assignmentv1.Position_POSITION_MEMBER:
		return amodel.PositionMember
	case assignmentv1.Position_POSITION_MANAGER:
		return amodel.PositionManager
	}
	return ""
}
