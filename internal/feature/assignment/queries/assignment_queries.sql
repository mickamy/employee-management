-- name: InsertAssignmentView :exec
INSERT INTO assignment_view (id, employee_id, department_id, position, assigned_on)
VALUES ($1, $2, $3, $4, $5);

-- name: ReleaseAssignmentView :exec
UPDATE assignment_view
SET released_on = $2
WHERE id = $1;

-- name: ClearAssignmentViewRelease :exec
UPDATE assignment_view
SET released_on = NULL
WHERE id = $1;

-- name: DeleteAssignmentView :exec
DELETE
FROM assignment_view
WHERE id = $1;

-- name: GetCurrentAssignment :one
SELECT id, employee_id, department_id, position, assigned_on, released_on
FROM assignment_view
WHERE employee_id = $1
  AND assigned_on <= sqlc.arg(today)
  AND (released_on IS NULL OR released_on > sqlc.arg(today))
ORDER BY assigned_on DESC
LIMIT 1;

-- name: ListAssignmentHistory :many
SELECT id, employee_id, department_id, position, assigned_on, released_on
FROM assignment_view
WHERE employee_id = $1
ORDER BY assigned_on, id;

-- name: FindAssignmentStream :one
SELECT stream_id
FROM assignment_events
WHERE event_type = 'AssignmentDecided'
  AND payload ->> 'assignment_id' = sqlc.arg(assignment_id)::text;
