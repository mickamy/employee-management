# employee-management

An architecture design showcase built around the employee management domain.

The claim is not "which pattern is best" but that **the nature of the problem picks the pattern** — demonstrated within
a single system:

- **employee** — basic attributes of an employee (PII). History has no intrinsic value here, so plain CRUD + DDD
- **assignment** — assignment starts/releases and manager history. Querying history is the requirement itself, so CQRS +
  Event Sourcing
- **organization** — departments. Flat CRUD for now; the generation-based tree is a future slice

The reasoning behind each choice lives in [docs/context-map.md](docs/context-map.md) and docs/adr/. Persistence
follows Kawashima's immutable data model — see [docs/data-modeling.md](docs/data-modeling.md).

## Stack

- Backend: Go (modular monolith) + PostgreSQL
- Contract: Protocol Buffers + [Connect RPC](https://connectrpc.com/), managed with [buf](https://buf.build/)
- Frontend: Remix v3 (BFF)

## Structure

```
.
├── proto/          # Connect RPC contracts (single source of truth)
├── gen/            # buf generate output (Go)
├── cmd/server/     # entry point
├── internal/
│   ├── feature/    # employee / organization / assignment (package by feature)
│   ├── server/     # composition root wiring Connect handlers
│   └── storage/    # storage clients: db (PostgreSQL)
├── frontend/       # Remix v3 (BFF)
├── spec/           # implementation-agnostic acceptance tests
└── docs/
```

## Development

```sh
envsubst < .env.example > .env
make compose-up-d # start dependencies via docker compose
make tdb-migrate   # apply migrations (goose)
go run ./cmd/server
```
