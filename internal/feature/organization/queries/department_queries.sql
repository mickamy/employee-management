-- name: CreateDepartment :exec
INSERT INTO departments (id, code, name)
VALUES ($1, $2, $3);

-- name: RenameDepartment :one
UPDATE departments
SET name = $2
WHERE id = $1
RETURNING id, code, name;

-- name: ListDepartments :many
SELECT id, code, name
FROM departments
ORDER BY code;

-- name: DepartmentExists :one
SELECT EXISTS (SELECT 1 FROM departments WHERE id = $1);
