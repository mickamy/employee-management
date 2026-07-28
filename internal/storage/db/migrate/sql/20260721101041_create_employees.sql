-- +goose Up
CREATE TABLE employees (
    id uuid PRIMARY KEY,
    code text NOT NULL UNIQUE,
    name text NOT NULL,
    email text NOT NULL UNIQUE
);

-- Hire event. ON DELETE CASCADE serves the PII deletion story; UNIQUE
-- (employee_id) is a v1 simplification until retirement/rehire exists.
CREATE TABLE employee_hires (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL UNIQUE REFERENCES employees (id) ON DELETE CASCADE,
    hired_on date NOT NULL
);

REVOKE UPDATE, DELETE ON employee_hires FROM app_writer;

-- +goose Down
DROP TABLE employee_hires;
DROP TABLE employees;
