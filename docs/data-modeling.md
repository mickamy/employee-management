# Data Modeling

Persistence in this repository follows the **immutable data model** advocated by Yoshitaka Kawashima
([@kawasima](https://scrapbox.io/kawasima/%E3%82%A4%E3%83%9F%E3%83%A5%E3%83%BC%E3%82%BF%E3%83%96%E3%83%AB%E3%83%87%E3%83%BC%E3%82%BF%E3%83%A2%E3%83%87%E3%83%AB)).
The premise: UPDATE is what makes systems complex, so design the data model to eliminate it.

## Rules

1. Every entity is either a **resource** or an **event**. An event is an entity extracted from a verb — it records a
   business activity happening at a point in time, and carries the datetime at which the activity occurred.
2. Planned dates, expiry dates, and effective dates are **not** occurred-on attributes. Only the datetime of the
   activity itself classifies an entity as an event.
3. An event has exactly **one** occurred-on attribute. More than one means multiple events hide in one entity — split
   them. When a business event spans a period, model the detail events INSERT-only, and if a current status is needed,
   keep it in a separate long-term entity — the only place an update happens.
4. Events are INSERT-only. A mistake is corrected by inserting a compensating event, never by UPDATE or DELETE.
5. Wanting created_at/updated_at on a resource is the litmus test for hidden events: enumerate the concrete update
   cases and extract each as its own event with its own attributes. Blanket created_at/updated_at columns are rejected
   on both resources and events.
6. Do not embed a foreign key for a non-dependent relationship. An employee must not carry department_id — that forces
   a nullable FK updated on assignment. Put an intersection entity between the two resources, and keep the
   current-state intersection (affiliation) separate from the events that change it.
7. These rules govern the system of record. Derived read models (projections) are rebuildable caches and are exempt.

Infrastructure note: the event store records plumbing metadata per event (global position, persisted-at) for
projections and replay. That is not a domain attribute and never appears in the domain model — consistent with rule 5,
no domain table gets blanket audit columns. Where transaction time is a genuine business requirement (retroactive
payroll corrections), it will be modeled deliberately when that slice lands, not as a default column.

## Future-dated changes

A future date is never an event's occurred-on attribute. It travels as a regular attribute of a decision that occurs
now, and state is evaluated against a date at query time. Registering on 3/20 an assignment that takes effect on 4/1:

1. 3/20 — INSERT the decision event carrying assigned_on = 4/1. The activity that occurred is the decision itself.
2. A query before 4/1 evaluates assigned_on <= today and still returns the previous assignment.
3. A query on or after 4/1 returns the new assignment from the same, unchanged data. Nothing is updated when the date
   arrives; only "today" changed.
4. Revoking the decision before it takes effect is another INSERT (rule 4), which the projection cancels out.

Current state is a function — f(events, as-of date) — not a stored value, so there is no batch job rewriting rows when
an effective date arrives. Side effects wanted on that date (notifications, etc.) are scheduled reactions to the event,
not data mutations.

In Kawashima's terms this is the generation problem: future changes are generations, not history. The organization
resource will use one of his generation patterns directly; in the event-sourced assignment context, the projection
plays the role of the effective-generation view — derived and rebuildable, and therefore the one place updates are
allowed (rule 7).

One naming consequence: strictly, what occurs is the decision, so write-side event types lean that way (e.g.,
AssignmentDecided rather than AssignmentStarted), and assigned_on is the date the decision takes effect.

## Compensating events

A revocation is a first-class domain event, not a generic tombstone:

```
assignment_decided { id: A1, employee_id, department_id, assigned_on: 4/1 }   -- occurs 3/20
assignment_revoked { id: A2, revokes: A1, reason }                            -- occurs 3/28
```

assignment_revoked records the decision to revoke — with its own occurred-on, a reference to what it revokes, and its
own attributes (who, why). The original event is never touched; "revoked" is an interpretation made by projections,
which stop counting A1 toward current state. History read models can render both a clean view (revoked decisions
hidden) and an audit view (shown struck through) from the same events.

A data-entry mistake (wrong department) is represented as revoke + re-issue rather than a dedicated correction event,
keeping the event vocabulary small. Revocation applies only to decisions that have not yet taken effect; undoing a
decision already in effect is a different business activity — a release, or a retroactive correction, which belongs to
the bitemporal payroll slice.

## Entity classification

| Entity                    | Kind     | Occurred on |
|---------------------------|----------|-------------|
| employee                  | resource | —           |
| department                | resource | —           |
| hire                      | event    | hired_on    |
| assignment start          | event    | assigned_on |
| assignment release        | event    | released_on |
| salary revision (planned) | event    | revised_on  |

The effective date of a salary revision (when the new salary starts to apply) is a lifecycle attribute, not its
occurred-on (rule 2). Future-dated changes are a generation problem and will use one of Kawashima's generation
patterns when the payroll slice lands.

## How each context applies the rules

The three contexts apply the same principle at different intensities, which is part of what this repository wants to
show:

- **employee** — the immutable data model on plain relational tables, without event sourcing machinery. The employee
  resource is updatable (name, email); the hire fact is a separate INSERT-only event table. Current state is read
  directly with SQL.
- **assignment** — full event sourcing. Events are the only source of truth, so rules 1–4 hold by construction. The
  current-assignment projection is Kawashima's current-state intersection entity (rule 6) and the long-term status
  entity (rule 3) at once — and because it is a derived read model, the one place updates happen is exempt by rule 7.
- **organization** — starts as a flat, updatable resource; renaming in place is the same documented trade-off as
  employee. Reorganizations and the department tree arrive in a dedicated slice using Kawashima's generation patterns,
  where changes become new interval rows instead of destructive updates.

## Consequences for the current design

- `Assignment` in the proto contract carries both assigned_on and released_on, so by rule 3 it is **not** a stored
  entity. The write side stores two events — assignment start and assignment release — and `Assignment` is a read model
  composed from them. Contracts describe interface shapes, not storage.
- `Employee.hired_on` in the contract is likewise a read shape. Hire is an event (the verb test: "to hire" works), so
  the employees table has no hired_on column; it comes from the hire event.
- Because employee never carries department_id (rule 6), a new hire before their first assignment and an officer who
  belongs to no department are both representable without nullable FKs.
- Updating employee.name and employee.email in place is a deliberate decision, not a shortcut: name history is PII we
  choose not to keep. Former names are a liability (retention minimization; people who change their name rarely want
  the old one shown), and change events carrying name values would spread PII beyond the employee row, complicating
  deletion. A document that needs the name as of issuance (a pay statement, an appointment letter) freezes it into its
  own record at that moment — the issuance is the event, the name is its attribute. If auditing the fact of a change
  becomes a requirement, extract an event without name values (employee_name_changed carrying only employee_id and
  changed_on) per rule 5.
- Deleting the employee resource stays possible (rule 4 applies to events, not resources). Events reference employee_id
  only, which is what makes the PII deletion story in the context map work. When the retirement slice lands, active and
  retired employees become subtypes by classification, keeping sensitive attributes erasable for retired ones.

## References

- [イミュータブルデータモデル — kawasima](https://scrapbox.io/kawasima/%E3%82%A4%E3%83%9F%E3%83%A5%E3%83%BC%E3%82%BF%E3%83%96%E3%83%AB%E3%83%87%E3%83%BC%E3%82%BF%E3%83%A2%E3%83%87%E3%83%AB)
- [イミュータブルデータモデルの極意 — kawasima](https://www.slideshare.net/slideshow/ss-250716400/250716400)
- [5分でわかる イミュータブル データモデル](https://speakerdeck.com/shimomura/5fen-dewakaru-imiyutaburu-detamoderu)
