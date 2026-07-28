-- +goose Up
-- Flat for now; the department tree and reorganizations (generation-based)
-- are a future slice. Renaming in place is the same documented trade-off as
-- employee (docs/data-modeling.md).
CREATE TABLE departments
(
    id   uuid PRIMARY KEY,
    code text NOT NULL UNIQUE,
    name text NOT NULL
);

-- +goose Down
DROP TABLE departments;
