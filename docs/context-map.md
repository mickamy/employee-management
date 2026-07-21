# Context Map

The employee management domain is split into bounded contexts, each with an architecture pattern chosen for the nature
of its problem. The claim this repository makes is not that one pattern is superior, but that the problem picks the
pattern.

## Contexts

| Context      | Classification | Pattern                                   | Rationale                                                                                                                |
|--------------|----------------|-------------------------------------------|--------------------------------------------------------------------------------------------------------------------------|
| employee     | supporting     | CRUD + plain DDD (aggregate + repository) | Basic attributes (PII) only have value in their latest state. Don't bring ES where there is no history requirement       |
| assignment   | **core**       | CQRS + Event Sourcing                     | Assignment starts/releases and manager history make querying history the requirement itself. Events are the primary data |
| organization | supporting     | CRUD (flat); generation-based tree later  | Reorganization history is desirable, but its change frequency and complexity don't justify the investment in ES          |

## Boundary rules

- Each feature exposes a public API at its package root; everything else is feature-internal. Depending on another
  feature means calling that public API, never reaching into its internals
- Synchronous calls between features are allowed, but the dependency edges are explicit and acyclic:
    - assignment → employee, assignment → organization (existence checks on commands, resolving names at query time)
    - employee and organization depend on no other feature
- Events are for reactions, not request/response: updating projections, cleanup after a deletion — chains where the
  caller does not wait for a result
- assignment stores only employee_id; PII (names, etc.) stays within the employee context and is never replicated into
  another feature's storage. Read models store IDs and resolve display names at query time
    - A deletion request from a former employee is satisfied by deleting on the employee side; the event log stays
      immutable and becomes meaningless

## Design decisions

### Immutable data modeling

Persistence follows Kawashima's immutable data model ([docs/data-modeling.md](data-modeling.md)): entities are either
resources or events, and events are INSERT-only with exactly one occurred-on attribute. Contexts differ only in how far
they take the principle — employee applies it to plain relational tables, assignment goes all the way to event
sourcing.

### Managers are derived from assignments

A manager is not stored as direct data. It is derived from the MANAGER assignment in the same department.

- Manager history = a projection composed from the employee's own assignment history and the MANAGER assignment history
  of their departments
- A change of department head is just another assignment event, so "does a reorganization show up in manager history?"
  is YES by construction
- The alternative — a direct manager relation between employees — was rejected for its data-entry burden and
  inconsistency risk on org changes

### Command/query service split: the decision rule

A context splits its proto into CommandService / QueryService if and only if it has derived read models — shapes that
can be queried but that no command writes (assignment's ManagerTenure). The split then communicates something true:
everything on the query side may lag behind an accepted command (hence revision / min_revision), and everything on the
command side is authoritative.

A context without derived read models gets a single service (employee, organization). Splitting there would be a false
signal: identical shape, different semantics — its query side would be immediately consistent, so readers could no
longer infer consistency guarantees from the contract's shape. What is uniform across contexts is this rule, not the
shape it produces. Method-level discipline (commands mutate, queries don't) applies everywhere regardless.

### Read-your-writes

Command responses return the event stream revision, and queries accept min_revision. Waiting for the projection to catch
up gives the UI a path that hides eventual consistency.

## Out of scope (future slices)

- Concurrent assignments (an employee holding multiple posts at once). v1 assumes a single assignment per employee
- The department tree and reorganizations (generation-based, with as-of queries; gets its own slice), and denormalizing
  department names into read models
- payroll (salary revisions; bitemporal effective/recorded dates)
- A formal shape for published events between contexts, separate from internal events
