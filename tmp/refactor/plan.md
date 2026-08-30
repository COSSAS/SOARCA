# SOARCA restructure: orchestrator / transport separation

Working document. Lives in `./tmp` so it is not committed; move it into the repo if you
want it tracked.

Companion: `durable-execution-design.md` covers the future execution model (durable runs,
resume, parallel steps). This document is only about the current restructure.

## Root cause

`*runtime.Runtime` is a service locator that is passed *into* the things it constructs.
Because the container flows downward, the dependency graph has no direction.

```
Runtime ──constructs──> finRegistry, playbookService, manualInbox, ...
Runtime ──passed into──> WorkflowFactory ──> f.runtime.GetInteraction(), GetCache(), ...
Runtime ──passed into──> execution.Service ──> s.runtime.GetCache().GetExecutionReport()
```

Every other symptom follows from this:

- the "circular dependency" that forced `SetExecutionRuntime` (which also secretly
  constructs `TriggerService`)
- `runtime.something.something` in `internal/services/execution/runtime.go`
- transport constructing the app (`httptransport.New` calls `bootstrap.New`)
- `bootstrap.Container` holding an HTTP handler (`FinHandler`)
- ceremony interfaces (`controller/database`, `controller/informer`) that exist only to
  break import cycles caused by the above
- three unrelated meanings of "controller", three of "runtime"

There is no real cycle in the domain. `Executions -> Engine -> {Interaction, Cache,
FinQueue, PlaybookStore}` is a DAG.

**Therefore: invert dependencies first, rename and move files last.**

## Target boundary

Transport never sees the orchestrator. It receives a flat value struct of interfaces:

```go
// pkg/soarca
type Operations struct {
    Playbooks  playbooks.Library
    Executions executions.Runner
    Fins       fins.Registry
    Work       fins.Dispatch
    Manual     manual.Inbox
}

func New(cfg Config) (*Orchestrator, error)
func (o *Orchestrator) Operations() Operations
func (o *Orchestrator) Close() error
```

```go
// cmd/soarca/main.go
cfg := config.Load()
orc := soarca.New(cfg); defer orc.Close()
srv := httpapi.New(orc.Operations(), cfg.HTTP)
srv.Run()
```

`Operations` has no behaviour and no infrastructure getters, so `runtime.x.y` is
structurally impossible and depth is capped at `ops.Executions.Start(ctx, ...)`.

### What may cross the boundary

| Concern | Owner |
|---|---|
| routes, verbs, status codes, DTOs, json/example tags, auth, CORS, TLS | transport |
| client poll-interval hints | transport |
| `context.Context` | shared (stdlib, not HTTP) |
| domain args (`cacao.Playbook`, `cacao.Variables`, `uuid.UUID`) | core |
| typed domain errors (`playbooks.ErrNotFound`, ...) | core; transport maps to codes |
| blocking until work available (`PollJob`) | core |

## Target layout

```
pkg/                       interfaces, domain types, errors. NO logic.
  soarca/                  Orchestrator, Operations, Config
  cacao/                   CACAO model (json + validate + example)
  playbooks/               Library + errors
  executions/              Runner + status/metadata types
  manual/                  Inbox, PendingStep, Response
  fins/                    Registry, Dispatch, Record
  fins/protocol/           wire types, dependency-light (stdlib + uuid only)
internal/
  config/
  orchestrator/            composition root
  playbooks/               (1) persistence + CRUD
  executions/              (3) start/resume/status + (6) status reporting read side
    engine/                decomposer construction (was bootstrap/workflow_factory.go)
  fins/registry/           (2) identity, tokens, persistence   [state: FinStore]
  fins/dispatch/           (4) queue, leases, long-poll        [state: Queue]
  manual/                  (5) inbox + interaction registry
  adapters/storage/{memory,mongodb}
  adapters/thehive/        (7) plugs into the engine reporter chain
  transport/http/          server.go, routes.go, handlers/*
cmd/soarca/main.go
```

Everything else under today's `pkg/` (`core/`, `reporting/`, `api/`, `utils/`,
`integration/`, rest of `models/`) moves to `internal/`.

Three merges do most of the work:

- `trigger` + `execution` + `reporter` -> `internal/executions` (kills the fake cycle)
- `fin/registry` + `fin/work` + `capability/fin/queue` -> `internal/fins/*`
- `controller/database`, `controller/informer` -> deleted

## Naming

| Now | Target |
|---|---|
| `internal/runtime.Runtime` | `orchestrator.Orchestrator` |
| `services.ExecutionRuntime` | `executions.Runner` |
| `bootstrap.Container` | deleted |
| `controller.Initialize`, `decomposer_controller.IController`, `database.IController` | deleted |
| `TriggerService.ExecutePlaybook` | `executions.Runner.Start` (route stays `/trigger`) |
| `Decomposer` | `WorkflowRunner` |
| `WorkflowFactory.NewDecomposer()` | `type NewRunner func() WorkflowRunner` |
| `IDecomposer`, `ICapability`, `IActionExecutor`, `IWorkflowReporter` | drop `I` prefix |
| `InteractionController`/`ManualInbox`/`CommandInfo`/`InteractionResponse` | `manual.Inbox`, `manual.PendingStep`, `manual.Response` |
| `FinRegistry` / `FinWorkService` | `fins.Registry` / `fins.Dispatch` |
| `PlaybookService` | `playbooks.Library` |

No documentation updates for now (per decision).

## Phases

Each phase compiles, keeps tests green, and is independently mergeable.

### Phase 0 — safety net (DONE)

See "Phase 0 results" below.

### Phase 1 — stop passing the container (DONE)

No files moved. `WorkflowFactory` and `execution.Service` now take explicit dependency
structs instead of `*appruntime.Runtime`:

```go
newWorkflowFactory(EngineDeps{Interaction, Cache, FinQueue, FinStore, PlaybookStore, Config})
execservice.New(wf, interaction, cache)
```

`execution.Service` declares its own narrow consumer interfaces (`ManualResumer`,
`ExecutionReports`) rather than reaching through the container. `bootstrap.New` is now
the only place that reads runtime getters, which is legitimate for a composition root.

No service or engine holds `*Runtime` any more. It survives only in `bootstrap.New`,
`controller.go`, `httptransport.New` and tests — all removed in Phase 3.

### Phase 2 — collapse the fake cycle (DONE)

Merged `trigger` + `execution` + `reporter` into `internal/executions`, exposing one
interface:

```go
type Runner interface {
	Start(ctx, playbook, variables) (uuid.UUID, error)
	StartByID(ctx, playbookID, variables) (uuid.UUID, error)
	List(ctx) ([]cache.ExecutionEntry, error)
	Report(ctx, executionID) (cache.ExecutionEntry, error)
}
```

The engine moved to `internal/executions/engine` — required, not cosmetic: `bootstrap`
imports `runtime`, so `runtime` could not import the factory while it lived in
`bootstrap`. With the engine outside, `Runtime` now constructs strictly top-down:

```
storage -> cache/interaction/queue -> engine.New(deps) -> executions.New(engine, store, cache)
```

Deleted: `SetExecutionRuntime`, `runtime.triggerService = nil`, `GetExecutionRuntime`,
`GetTriggerService`, `GetReporterService`, `internal/services/{execution,trigger,reporter}`,
`internal/controller/informer`, `internal/bootstrap/workflow_factory.go`, and the
`ExecutionRuntime` / `TriggerService` / `ReporterService` interfaces.

`bootstrap.New` no longer constructs anything; it only builds `TransportOptions`.
Phase 3 deletes it entirely.

Dead code removed in the process (no production callers, found by usage audit):

- `ExecutionRuntime.ResumeManualStep` — duplicated `ManualInbox.ContinuePendingCommand`;
  both ended at `interaction.PostContinue`. The manual handler only ever used the inbox.
- `ExecutionRuntime.GetExecutionStatus` — duplicated `ReporterService.GetExecutionReport`.

Also fixed here (were listed as known defects):

- the discarded `variables` argument — `Start` now applies the variables it is given
  instead of `_ = variables`
- the dead channel-filter loop over a buffered size-1 channel written by exactly one
  decomposer — replaced with a plain select

Note: `runtime.Options` gained `HTTP` and `TheHive`, and `controller.Initialize` now
passes them. Without that, `SkipCertValidation` and the whole TheHive integration would
have silently stopped being configured.

### Phase 3 — install the boundary (DONE)

`Runtime` now exposes exactly one surface:

```go
type Operations struct {
	Playbooks  services.PlaybookService
	Executions executions.Runner
	Fins       services.FinRegistry
	Work       services.FinWorkService
	Manual     services.ManualInbox
}

func (r *Runtime) Operations() Operations
```

All infrastructure fields and getters went private (`playbookStore`, `finStore`,
`finQueue`, `cache`, `interaction`). `GetPlaybookStore`, `GetCache`, `GetFinQueue`,
`GetInteraction`, `GetFinStore` are gone, so transport cannot reach through the
container even by accident.

`httptransport.New(ops, opts)` now takes `Operations` plus its own narrow `Options`
(Server, Fin, Auth, CORS) instead of `*Runtime` + the whole `config.Config`. Note what
is no longer passed to transport: storage config, TheHive config and outbound TLS
settings — all orchestrator concerns. The transport also constructs its own FinHandler.

`internal/bootstrap` is deleted entirely (Container, TransportOptions, workflow factory).

`internal/controller/controller.go` was kept rather than folded into `main.go`: it is the
composition root and its `loadConfig`/`newRuntime`/`newTransport` seams are what make
`controller_test.go` possible. It should be renamed (`internal/app`) in Phase 4 rather
than deleted — the objection was three meanings of "controller", not this file's job.

`boundary_test.go` was rewritten: it asserted every getter returned non-nil, which pinned
the service-locator shape. It now asserts `Operations` is fully populated and documents
the getters that must not come back.

### Phase 4 — renames and moves only (IN PROGRESS)

#### 4a — test layout (DONE)

Idiomatic Go puts `foo_test.go` beside `foo.go` in the same directory; a top-level
`test/` tree is not a Go convention. 43 of 58 test files were already beside their code.

Tests needing external services are now selected by build tag rather than by directory:

- `//go:build integration` — needs `deployments/docker/testing` (httpbin, ssh)
- `//go:build manual` — needs a special environment (Windows/PowerShell host, live TheHive)

`go test ./...` is now green for the first time. Previously six suites failed by design
on any machine without those services, which trains everyone to ignore a red suite.
Makefile targets: `test` (default, no services), `integration-test`, `manual-test`.

Still to do in 4a: move the remaining `test/integration/api/**` suites next to the code
they exercise, turn `test/unittest/mocks` into per-package mocks, and move playbook JSON
fixtures into `testdata/` (which the go tool ignores by convention).

#### 4b — package renames (IN PROGRESS)

Decision: option 2 — `run` everywhere including the wire. The FIN protocol is alpha, so
breaking changes are acceptable and no compatibility shim is needed.

Done:

- `internal/executions` → `internal/runs`; `executions.Runner` → `runs.Runner`.
- Wire vocabulary renamed: `execution_id` → `run_id`, `step_execution_id` → `step_run_id`,
  `"execution_status"` → `"run_status"`. Route params `/manual/:exec_id/:step_execution_id`
  → `/manual/:run_id/:step_run_id`.
- API models: `PlaybookExecutionReport` → `PlaybookRunReport`, `StepExecutionReport` →
  `StepRunReport`, `api.Execution` → `api.RunStarted`.
- `fin.Job.ExecutionId`/`StepExecutionId` → `RunId`/`StepRunId` (FIN protocol break).
- Fixed the `json:"payload"` bug on the trigger response as part of the model rename;
  it now correctly serialises as `playbook_id`.
- Swagger regenerated: 0 occurrences of `execution_id`, 15 of `run_id`.

Remaining:

- `pkg/models/execution.Metadata` → run vocabulary (`RunId`, `StepRunId`). This is the
  deep one: ~150 references across decomposer, cache, reporters and capabilities.
- `Decomposer` → `WorkflowRunner`, `executors/` → `workflow/steps/`, drop `I` prefixes.
- `internal/runtime` → `internal/orchestrator`, `internal/controller` → `internal/app`.
- `internal/services/{fin,manual,playbook}` → `internal/fins/{registry,dispatch}`,
  `internal/manual`, `internal/playbooks`; `internal/storage` → `internal/store`.
- `pkg/` → `internal/` split and the `pkg/soarca` embeddable entrypoint.

Constraint that no longer applies: wire compatibility. Kept for the record because the
Python FIN package must be updated in lockstep with the `run_id`/`step_run_id` rename.

### Phase 5 — enforce (DONE, ahead of Phase 4)

`test/architecture/boundary_test.go` runs `go list -deps` over the orchestrator packages
and fails if any of them transitively depends on gin, gauth, swagger,
`soarca/internal/transport` or `soarca/pkg/api`. Currently passing: the core is genuinely
transport-free.

`net/http` is deliberately *not* forbidden — the http and openc2 capabilities make
outbound calls and legitimately need it. The rule targets inbound web framework, routing
and auth middleware.

A second test (`TestDetectorWorks`) asserts that the transport layer *does* depend on
gin, so the check cannot silently degrade into one that inspects nothing.

### Phase 6 — optional, non-blocking

Strip `bson:` tags from domain models; the mongo adapter owns document types and
mapping. Notably `cacao.Playbook.ID` is `bson:"_id" json:"id"` today.

## Known defects found while planning

- [x] `config.Load()` panicked: `v.SetEnvKeyReplacer(nil)` overwrote viper's default
      replacer, so `getEnv` nil-dereferenced. Viper's default is already a no-op.
      **Production bug**, fixed in Phase 0.
- [x] `execution.Service.StartExecution` did `_ = variables`, silently discarding the
      argument. Fixed in Phase 2.
- [x] Same function looped on `details.PlaybookId != playbook.ID` over a buffered
      size-1 channel written by exactly one decomposer. Dead logic, removed in Phase 2.
- [ ] `pkg/models/api.Execution.PlaybookId` is tagged `json:"payload"` — wrong wire name.
- [ ] `fin.Record` is simultaneously persistence type (`bson:"_id"`), admin wire type
      (`ListFins` returns it straight to a handler) and secret holder (`FinTokenHash`),
      kept off the wire only by a `json:"-"` tag. Split into public `fins.Record` and an
      internal persistence record.

## Phase 0 results

Goal: a safety net that tests the real wiring. It was red and it was pinning the old
architecture.

Fixed:

1. `internal/config/config.go` — removed `v.SetEnvKeyReplacer(nil)`. This panicked
   `config.Load()`, which took down the whole-app smoke test in
   `test/integration/api`.
2. `reporter_api_invocation_test.go` — fixture predated the per-invocation
   `StepExecutionId`. It also keyed `StepResults` by `StepId` despite
   `pkg/models/cache/cache.go` documenting the key as `StepExecutionId`. Both corrected.
3. `manual_api_test.go` — expected `null` for `commands`/`targets` where the handler
   emits `[]`.
4. Migrated `playbook_api_test.go` (6 sites) and `reporter_api_test.go` /
   `reporter_api_invocation_test.go` (4 sites) off the legacy controller-based route
   helpers onto `PlaybookRoutesWithService` / `ReporterRoutesWithService`. The tests
   were the *only* remaining users of the legacy path, which is why
   `controller/database` still existed.
5. Deleted from `pkg/api/api.go`: `Database()`, `Reporter()`, `Api()`, `PlaybookRoutes()`,
   `ReporterRoutes()`. Deleted `internal/controller/database/` and
   `test/unittest/mocks/mock_controller/database/`.

Status: `go vet ./...` clean. All packages pass except the pre-existing
external-dependency suites, which need containers and failed identically before these
changes:

```
pkg/utils/http                     (httpbin)
test/integration/capability/http   (httpbin)
test/integration/capability/ssh    (ssh server)
test/manual/powershell             (windows host)
test/manual/thehive_connector      (thehive)
test/manual/thehive_reporter       (thehive)
```

Coverage gaps to be aware of during the refactor: `/status` has no integration coverage.

## FIN route coverage (added)

`test/integration/api/routes/fin_api/fin_api_test.go` — 16 tests over the real registry
and work service backed by in-memory storage and a real queue, so the route/middleware/
service composition is covered before Phase 3 moves FIN handler construction into
transport. Covers registration (success, wrong token, no capabilities, registration
disabled), bearer auth (missing, unknown), poll with no work, result submission errors,
admin list/get/delete, and unregister.

Includes `TestListFinsDoesNotLeakTokenHash`, a regression guard for `fin.Record` doubling
as persistence type and admin wire type with only `json:"-"` keeping the credential
hash off the wire.

## Notes for later phases

- `internal/controller/boundary_test.go` asserted that every `Runtime` getter returns
  non-nil. It pinned the service-locator shape. Rewritten in Phase 3.
- Something in the editor/toolchain repeatedly prepends a duplicate `package X` line to
  newly created Go files, producing `expected declaration, found 'package'`. Hit on
  `fin_api_test.go`, `engine.go`, `service.go` and `boundary_test.go`. Check the first
  two lines of any new file before building.
