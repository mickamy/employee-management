package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/mickamy/employee-management/internal/errors/aerrors"
	"github.com/mickamy/employee-management/internal/feature/organization/model"
	"github.com/mickamy/employee-management/internal/feature/organization/queries"
	"github.com/mickamy/employee-management/internal/infra/storage/db"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

type Department interface {
	Create(ctx context.Context, department model.Department) error
	Rename(ctx context.Context, id uuid.UUID, name string) (model.Department, error)
	List(ctx context.Context) ([]model.Department, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	Bind(tx tx.Tx) Department
}

type department struct {
	q *queries.Queries
}

var _ Department = department{}

func NewDepartment(reader db.Reader) Department {
	return department{q: queries.New(reader)}
}

func (r department) Bind(tx tx.Tx) Department {
	return department{q: queries.New(tx.DBTX())}
}

func (r department) Create(ctx context.Context, d model.Department) error {
	err := r.q.CreateDepartment(ctx, queries.CreateDepartmentParams{
		ID:   d.ID,
		Code: d.Code,
		Name: d.Name,
	})
	if db.IsUniqueViolation(err) {
		return aerrors.Conflict("department code")
	}
	if err != nil {
		return fmt.Errorf("create department: %w", err)
	}
	return nil
}

func (r department) Rename(ctx context.Context, id uuid.UUID, name string) (model.Department, error) {
	row, err := r.q.RenameDepartment(ctx, queries.RenameDepartmentParams{
		ID:   id,
		Name: name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Department{}, aerrors.NotFound("department")
	}
	if err != nil {
		return model.Department{}, fmt.Errorf("rename department: %w", err)
	}
	return model.Department{
		ID:   row.ID,
		Code: row.Code,
		Name: row.Name,
	}, nil
}

func (r department) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	exists, err := r.q.DepartmentExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("department exists: %w", err)
	}
	return exists, nil
}

func (r department) List(ctx context.Context) ([]model.Department, error) {
	rows, err := r.q.ListDepartments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}

	departments := make([]model.Department, len(rows))
	for i, row := range rows {
		departments[i] = model.Department{
			ID:   row.ID,
			Code: row.Code,
			Name: row.Name,
		}
	}
	return departments, nil
}
