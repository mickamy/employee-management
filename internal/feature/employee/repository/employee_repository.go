package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/employee/model"
	"github.com/mickamy/employee-management/internal/feature/employee/queries"
	"github.com/mickamy/employee-management/internal/infra/storage/db"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

type Employee interface {
	Find(ctx context.Context, id uuid.UUID) (model.Employee, error)
	Names(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error)
	List(ctx context.Context, afterID uuid.UUID, limit int32) ([]model.Employee, error)
	Create(ctx context.Context, employee model.Employee) error
	CreateHire(ctx context.Context, hire model.EmployeeHire) error
	Bind(tx tx.Tx) Employee
}

type employee struct {
	q *queries.Queries
}

var _ Employee = employee{}

func NewEmployee(reader db.Reader) Employee {
	return employee{q: queries.New(reader)}
}

func (r employee) Bind(tx tx.Tx) Employee {
	return employee{q: queries.New(tx.DBTX())}
}

func (r employee) Find(ctx context.Context, id uuid.UUID) (model.Employee, error) {
	row, err := r.q.GetEmployee(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Employee{}, aerrors.NotFound("employee")
	}
	if err != nil {
		return model.Employee{}, fmt.Errorf("get employee: %w", err)
	}
	return model.Employee{
		ID:      row.ID,
		Code:    row.Code,
		Name:    row.Name,
		Email:   row.Email,
		HiredOn: row.HiredOn,
	}, nil
}

func (r employee) Names(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	rows, err := r.q.GetEmployeeNames(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get employee names: %w", err)
	}

	names := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		names[row.ID] = row.Name
	}
	return names, nil
}

func (r employee) List(ctx context.Context, afterID uuid.UUID, limit int32) ([]model.Employee, error) {
	rows, err := r.q.ListEmployees(ctx, queries.ListEmployeesParams{
		AfterID:  afterID,
		PageSize: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}

	employees := make([]model.Employee, len(rows))
	for i, row := range rows {
		employees[i] = model.Employee{
			ID:      row.ID,
			Code:    row.Code,
			Name:    row.Name,
			Email:   row.Email,
			HiredOn: row.HiredOn,
		}
	}
	return employees, nil
}

func (r employee) Create(ctx context.Context, e model.Employee) error {
	err := r.q.CreateEmployee(ctx, queries.CreateEmployeeParams{
		ID:    e.ID,
		Code:  e.Code,
		Name:  e.Name,
		Email: e.Email,
	})
	if db.IsUniqueViolation(err) {
		return aerrors.Conflict("employee code or email")
	}
	if err != nil {
		return fmt.Errorf("create employee: %w", err)
	}
	return nil
}

func (r employee) CreateHire(ctx context.Context, hire model.EmployeeHire) error {
	err := r.q.CreateEmployeeHire(ctx, queries.CreateEmployeeHireParams{
		ID:         hire.ID,
		EmployeeID: hire.EmployeeID,
		HiredOn:    hire.HiredOn,
	})
	if err != nil {
		return fmt.Errorf("create employee hire: %w", err)
	}
	return nil
}
