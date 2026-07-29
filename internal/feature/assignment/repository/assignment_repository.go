package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/assignment/model"
	"github.com/mickamy/employee-management/internal/feature/assignment/queries"
	"github.com/mickamy/employee-management/internal/infra/es"
	"github.com/mickamy/employee-management/internal/infra/storage/db"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

// EventsTable is the assignment context's event table.
const EventsTable = "assignment_events"

var store = es.NewStore(EventsTable)

// Stream is the command side's port: the event log, nothing else. Append
// takes the loaded stream so the stream identity and the revision expectation
// cannot be supplied separately (or wrongly).
type Stream interface {
	Load(ctx context.Context, employeeID uuid.UUID) (model.Stream, error)
	Append(ctx context.Context, stream model.Stream, events ...model.Event) (int64, error)
	FindByAssignment(ctx context.Context, assignmentID uuid.UUID) (uuid.UUID, error)
	Bind(tx tx.Tx) Stream
}

// AssignmentView maintains the assignment view. Commands apply it in the same
// transaction as the event append; nothing else writes the view.
type AssignmentView interface {
	Insert(ctx context.Context, a model.Assignment) error
	Release(ctx context.Context, id uuid.UUID, releasedOn time.Time) error
	ClearRelease(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	Bind(tx tx.Tx) AssignmentView
}

// Assignment is the query side's port: it only reads the view.
type Assignment interface {
	Current(ctx context.Context, employeeID uuid.UUID, today time.Time) (model.Assignment, error)
	History(ctx context.Context, employeeID uuid.UUID) ([]model.Assignment, error)
	ManagerTenures(ctx context.Context, employeeID uuid.UUID) ([]model.ManagerTenure, error)
	Bind(tx tx.Tx) Assignment
}

type stream struct {
	q  *queries.Queries
	es es.Bound
}

var _ Stream = stream{}

func NewStream(reader db.Reader) Stream {
	return stream{q: queries.New(reader), es: store.With(reader)}
}

func (r stream) Bind(tx tx.Tx) Stream {
	return stream{q: queries.New(tx.DBTX()), es: store.With(tx.DBTX())}
}

func (r stream) Load(ctx context.Context, employeeID uuid.UUID) (model.Stream, error) {
	records, err := r.es.Load(ctx, employeeID)
	if err != nil {
		return model.Stream{}, fmt.Errorf("load assignment stream: %w", err)
	}

	events := make([]model.Event, len(records))
	for i, rec := range records {
		event, err := decodeEvent(rec)
		if err != nil {
			return model.Stream{}, err
		}
		events[i] = event
	}

	var lastRevision int64
	if len(records) > 0 {
		lastRevision = records[len(records)-1].StreamRevision
	}
	return model.Replay(employeeID, lastRevision, events), nil
}

func (r stream) Append(ctx context.Context, s model.Stream, events ...model.Event) (int64, error) {
	unsaved := make([]es.Unsaved, len(events))
	for i, e := range events {
		payload, err := json.Marshal(e)
		if err != nil {
			return 0, fmt.Errorf("encode %s: %w", e.EventType(), err)
		}
		unsaved[i] = es.Unsaved{Type: e.EventType(), Payload: payload}
	}

	position, err := r.es.Append(ctx, s.EmployeeID, s.LastRevision, unsaved...)
	if errors.Is(err, es.ErrRevisionConflict) {
		return 0, aerrors.Conflict("assignment stream")
	}
	if err != nil {
		return 0, fmt.Errorf("append assignment events: %w", err)
	}
	return position, nil
}

func (r stream) FindByAssignment(ctx context.Context, assignmentID uuid.UUID) (uuid.UUID, error) {
	streamID, err := r.q.FindAssignmentStream(ctx, assignmentID.String())
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, aerrors.NotFound("assignment")
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("find assignment stream: %w", err)
	}
	return streamID, nil
}

type assignmentView struct {
	q *queries.Queries
}

var _ AssignmentView = assignmentView{}

func NewAssignmentView(reader db.Reader) AssignmentView {
	return assignmentView{q: queries.New(reader)}
}

func (r assignmentView) Bind(tx tx.Tx) AssignmentView {
	return assignmentView{q: queries.New(tx.DBTX())}
}

func (r assignmentView) Insert(ctx context.Context, a model.Assignment) error {
	err := r.q.InsertAssignmentView(ctx, queries.InsertAssignmentViewParams{
		ID:           a.ID,
		EmployeeID:   a.EmployeeID,
		DepartmentID: a.DepartmentID,
		Position:     a.Position.String(),
		AssignedOn:   a.AssignedOn,
	})
	if err != nil {
		return fmt.Errorf("insert assignment view: %w", err)
	}
	return nil
}

func (r assignmentView) Release(ctx context.Context, id uuid.UUID, releasedOn time.Time) error {
	err := r.q.ReleaseAssignmentView(ctx, queries.ReleaseAssignmentViewParams{
		ID:         id,
		ReleasedOn: &releasedOn,
	})
	if err != nil {
		return fmt.Errorf("release assignment view: %w", err)
	}
	return nil
}

func (r assignmentView) ClearRelease(ctx context.Context, id uuid.UUID) error {
	if err := r.q.ClearAssignmentViewRelease(ctx, id); err != nil {
		return fmt.Errorf("clear assignment view release: %w", err)
	}
	return nil
}

func (r assignmentView) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteAssignmentView(ctx, id); err != nil {
		return fmt.Errorf("delete assignment view: %w", err)
	}
	return nil
}

type assignment struct {
	q  *queries.Queries
	db tx.DBTX
}

var _ Assignment = assignment{}

func NewAssignment(reader db.Reader) Assignment {
	return assignment{q: queries.New(reader), db: reader}
}

func (r assignment) Bind(tx tx.Tx) Assignment {
	return assignment{q: queries.New(tx.DBTX()), db: tx.DBTX()}
}

func (r assignment) Current(
	ctx context.Context,
	employeeID uuid.UUID,
	today time.Time,
) (model.Assignment, error) {
	row, err := r.q.GetCurrentAssignment(ctx, queries.GetCurrentAssignmentParams{
		EmployeeID: employeeID,
		Today:      today,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Assignment{}, aerrors.NotFound("assignment")
	}
	if err != nil {
		return model.Assignment{}, fmt.Errorf("get current assignment: %w", err)
	}
	return assignmentViewToModel(row), nil
}

func (r assignment) History(ctx context.Context, employeeID uuid.UUID) ([]model.Assignment, error) {
	rows, err := r.q.ListAssignmentHistory(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("list assignment history: %w", err)
	}

	assignments := make([]model.Assignment, len(rows))
	for i, row := range rows {
		assignments[i] = assignmentViewToModel(row)
	}
	return assignments, nil
}

// listManagerTenures lives outside sqlc: it needs a nullable CASE column that
// sqlc's inference types as non-nullable.
const listManagerTenures = `
SELECT m.employee_id                             AS manager_employee_id,
       m.department_id,
       GREATEST(mine.assigned_on, m.assigned_on) AS started_on,
       CASE
           WHEN mine.released_on IS NULL AND m.released_on IS NULL THEN NULL
           ELSE LEAST(COALESCE(mine.released_on, 'infinity'::date),
                      COALESCE(m.released_on, 'infinity'::date))
           END                                   AS ended_on
FROM assignment_view AS m
         INNER JOIN assignment_view AS mine ON mine.department_id = m.department_id
WHERE mine.employee_id = $1
  AND m.position = 'MANAGER'
  AND m.employee_id <> mine.employee_id
  AND GREATEST(mine.assigned_on, m.assigned_on)
    < LEAST(COALESCE(mine.released_on, 'infinity'::date),
            COALESCE(m.released_on, 'infinity'::date))
ORDER BY started_on`

func (r assignment) ManagerTenures(ctx context.Context, employeeID uuid.UUID) ([]model.ManagerTenure, error) {
	rows, err := r.db.Query(ctx, listManagerTenures, employeeID)
	if err != nil {
		return nil, fmt.Errorf("list manager tenures: %w", err)
	}
	defer rows.Close()

	var tenures []model.ManagerTenure
	for rows.Next() {
		var t model.ManagerTenure
		if err := rows.Scan(&t.ManagerEmployeeID, &t.DepartmentID, &t.StartedOn, &t.EndedOn); err != nil {
			return nil, fmt.Errorf("scan manager tenure: %w", err)
		}
		tenures = append(tenures, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate manager tenures: %w", err)
	}
	return tenures, nil
}

func assignmentViewToModel(row queries.AssignmentView) model.Assignment {
	return model.Assignment{
		ID:           row.ID,
		EmployeeID:   row.EmployeeID,
		DepartmentID: row.DepartmentID,
		Position:     model.Position(row.Position),
		AssignedOn:   row.AssignedOn,
		ReleasedOn:   row.ReleasedOn,
	}
}

func decodeEvent(rec es.Record) (model.Event, error) {
	switch rec.Type {
	case model.AssignmentDecided{}.EventType():
		var e model.AssignmentDecided
		if err := json.Unmarshal(rec.Payload, &e); err != nil {
			return nil, fmt.Errorf("decode %s: %w", rec.Type, err)
		}
		return e, nil
	case model.ReleaseDecided{}.EventType():
		var e model.ReleaseDecided
		if err := json.Unmarshal(rec.Payload, &e); err != nil {
			return nil, fmt.Errorf("decode %s: %w", rec.Type, err)
		}
		return e, nil
	case model.AssignmentRevoked{}.EventType():
		var e model.AssignmentRevoked
		if err := json.Unmarshal(rec.Payload, &e); err != nil {
			return nil, fmt.Errorf("decode %s: %w", rec.Type, err)
		}
		return e, nil
	case model.ReleaseRevoked{}.EventType():
		var e model.ReleaseRevoked
		if err := json.Unmarshal(rec.Payload, &e); err != nil {
			return nil, fmt.Errorf("decode %s: %w", rec.Type, err)
		}
		return e, nil
	default:
		return nil, fmt.Errorf("unknown event type %q at position %d", rec.Type, rec.GlobalPosition)
	}
}
