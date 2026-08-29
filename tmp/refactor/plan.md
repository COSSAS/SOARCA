# SOARCA restructure: orchestrator / transport separation

Working document. Lives in `./tmp` so it is not committed; move it into the repo if you
want it tracked.

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

### Phase 1 — stop passing the container

No files move. `WorkflowFactory` and `execution.Service` take explicit dependency
structs instead of `*appruntime.Runtime`:

```go
engine.New(engine.Deps{Interaction, Cache, FinQueue, FinStore, PlaybookStore, Config})
```

`Runtime` becomes construction-only. This is the phase that actually fixes the
architecture.

### Phase 2 — collapse the fake cycle

Merge trigger/execution/reporter into `internal/executions`. Delete
`SetExecutionRuntime`, `runtime.triggerService = nil`, `controller/informer`.
Construction becomes a straight top-down sequence with no setters.

### Phase 3 — install the boundary

Introduce `Operations`. `httptransport.New` takes `Operations` + config instead of
`*Runtime` + `Config`. Move `FinHandler` construction into transport. Delete
`bootstrap.Container` and `internal/controller/controller.go` (logic moves to `main.go`).

Proof: a ~30-line `soarca run playbook.json` CLI that builds an orchestrator, takes
`Operations()`, calls `Executions.Start` + `Executions.Status`, and imports nothing
under `transport/`.

### Phase 4 — renames and moves only

`git mv` + import rewrites, one PR per slice. Zero behaviour change.

### Phase 5 — enforce

A `go test` that shells `go list -deps` and fails if any package under
`internal/{orchestrator,playbooks,executions,fins,manual}` transitively imports
`net/http`, `gin`, or `soarca/internal/transport/...`.

### Phase 6 — optional, non-blocking

Strip `bson:` tags from domain models; the mongo adapter owns document types and
mapping. Notably `cacao.Playbook.ID` is `bson:"_id" json:"id"` today.

## Known defects found while planning

- [x] `config.Load()` panicked: `v.SetEnvKeyReplacer(nil)` overwrote viper's default
      replacer, so `getEnv` nil-dereferenced. Viper's default is already a no-op.
      **Production bug**, fixed in Phase 0.
- [ ] `execution.Service.StartExecution` does `_ = variables`, silently discarding the
      argument. Harmless today only because `trigger.Service` pre-merges into
      `playbook.PlaybookVariables` and passes `cacao.Variables{}`. Remove or honour it.
- [ ] Same function loops on `details.PlaybookId != playbook.ID` over a buffered
      size-1 channel written by exactly one decomposer. Dead logic from a shared-channel
      design.
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

Coverage gaps to be aware of during the refactor: there is no integration coverage for
the FIN routes (`/fin/register`, `/poll`, `/jobs/:id`, admin) or `/status`. Phase 3
changes FIN handler construction, so FIN route tests are worth adding first.
