-- +goose Up
-- +goose StatementBegin
DO
$$
    BEGIN
        -- Role creation here serves local development and CI. In environments where
        -- the migration runner lacks CREATEROLE (managed databases), pre-provision
        -- these roles via infrastructure tooling; creation is then skipped and only
        -- the grants below apply. Passwords are local/CI values — production
        -- rotates them out-of-band via ALTER ROLE.
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_writer') THEN
            CREATE ROLE app_writer LOGIN PASSWORD 'password';
        END IF;
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_reader') THEN
            CREATE ROLE app_reader LOGIN PASSWORD 'password';
        END IF;
    END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
DO
$$
    BEGIN
        EXECUTE 'GRANT CONNECT ON DATABASE ' || quote_ident(current_database()) || ' TO app_writer, app_reader';
    END
$$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO app_writer, app_reader;

-- Tables and sequences created by the role running migrations pick these up automatically.
-- The events table later revokes UPDATE/DELETE from app_writer to make events INSERT-only.
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_writer;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO app_writer;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO app_reader;

-- +goose Down
-- DROP OWNED revokes every privilege granted to these roles in the current
-- database and on shared objects (CONNECT), including per-object grants and
-- default-privilege entries that schema-level REVOKE would miss. The roles
-- own no objects here. Existence checks keep the down runnable when roles
-- were removed out-of-band.
-- +goose StatementBegin
DO
$$
    BEGIN
        IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_writer') THEN
            DROP OWNED BY app_writer;
        END IF;
        IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_reader') THEN
            DROP OWNED BY app_reader;
        END IF;
    END
$$;
-- +goose StatementEnd
DROP ROLE IF EXISTS app_writer;
DROP ROLE IF EXISTS app_reader;
