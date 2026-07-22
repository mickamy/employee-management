-- name: CreateEmployee :exec
INSERT INTO employees (id, code, name, email)
VALUES ($1, $2, $3, $4);

-- name: CreateEmployeeHire :exec
INSERT INTO employee_hires (id, employee_id, hired_on)
VALUES ($1, $2, $3);

-- name: GetEmployee :one
SELECT e.id, e.code, e.name, e.email, h.hired_on
FROM employees AS e
INNER JOIN employee_hires AS h ON h.employee_id = e.id
WHERE e.id = $1;

-- name: ListEmployees :many
SELECT e.id, e.code, e.name, e.email, h.hired_on
FROM employees AS e
INNER JOIN employee_hires AS h ON h.employee_id = e.id
WHERE e.id > sqlc.arg(after_id)
ORDER BY e.id
LIMIT sqlc.arg(page_size);
