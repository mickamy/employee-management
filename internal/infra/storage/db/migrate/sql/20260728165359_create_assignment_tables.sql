-- +goose Up
-- Event store for the assignment context. A stream is an employee's
-- assignment timeline; global_position and persisted_at are store plumbing,
-- not domain attributes (docs/data-modeling.md).
CREATE TABLE assignment_events (
    global_position bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    stream_id uuid NOT NULL,
    stream_revision bigint NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    persisted_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (stream_id, stream_revision)
);

-- Release/revoke commands address an assignment; this locates its stream.
CREATE INDEX assignment_events_assignment_id_idx
    ON assignment_events ((payload ->> 'assignment_id'));

REVOKE UPDATE, DELETE ON assignment_events FROM app_writer;

-- Rebuildable projection (clean view: revoked decisions are removed),
-- maintained in the same transaction as the event append. The one place
-- updates happen, exempt by rule 7 of docs/data-modeling.md.
CREATE TABLE assignment_view (
    id uuid PRIMARY KEY,
    employee_id uuid NOT NULL,
    department_id uuid NOT NULL,
    position text NOT NULL,
    assigned_on date NOT NULL,
    released_on date
);

CREATE INDEX assignment_view_employee_idx ON assignment_view (employee_id, assigned_on);
CREATE INDEX assignment_view_manager_idx ON assignment_view (department_id) WHERE position = 'MANAGER';

-- +goose Down
DROP TABLE assignment_view;
DROP TABLE assignment_events;
