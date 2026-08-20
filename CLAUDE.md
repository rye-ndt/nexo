# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# What this project is

A desktop control plane for orchestrating a fleet of coding agents. Work is
authored as a **graph of steps**; each step is a scoped unit of work that starts
from a role and runs on an agent, steps chain together, and a `Handoff` carries
what one step learned into the next step's prompt. Read `README.md` before
designing anything — the section "Why Nexo exists" explains why pieces that look
over-engineered (heartbeats, unified diffs, the MCP proxy) exist.

`docs/NAMING.md` is the authority on vocabulary: one word per concept, the same
word in Go, in SQL, in TypeScript, and on screen. Consult it before introducing
or renaming a concept, and record the decision there first — it is what keeps
"workflow", "step", "role", "agent", "handoff" and "result" meaning exactly one
thing each.

Two things to hold onto:

- **Context engineering is the product**, not scheduling. `Handoff` is the most
  important type in the repo.
- **Simplicity is the top priority**, explicitly so this stays maintainable for
  years. Prefer the simplest thing that works; flag anything that adds ongoing
  maintenance cost and offer the simpler alternative first.

# Commands

Run everything from the repo root. `config.yaml` is compiled into the binary by
the `//go:embed` in `wire.go`, so editing it needs a rebuild, not a restart.

```sh
make dev      # run the app; frontend hot-reloads, Go changes need a restart
make dev-go   # same, but rebuilds on Go changes (macOS steals focus each time)
make build    # production .app bundle
make run      # build, then launch the bundle
make test     # go test ./...
```

`go run .` fails on a Wails build tag — always use `make dev`.

A single Go test: `go test ./internal/implementation/core/workflow_manager -run TestGatedStepHoldsDownstreamUntilAccepted -v`.
Tests live beside the code they cover (`v1_test.go`, `litesql_test.go`); the
suites that matter are `core/workflow_manager` (review gating, retry, cancel),
`core/mcp_proxy` (credential revoke, account lookup) and `input/storage`.

In `frontend/`:

```sh
npm run dev        # vite on port 8888, pinned in vite.config.ts — no Wails runtime
npx tsc --noEmit   # typecheck
npm run lint       # eslint
npm run format     # prettier
```

Regenerate bindings with `~/go/bin/wails generate module` (wails is usually not
on PATH). Never hand-edit `frontend/wailsjs/`.

# Architecture

Hexagonal, with this project's own naming: ports in
`internal/interface/{core,input,output}` (packages `core_itf` / `input_itf` /
`output_itf`), one package per technology in `internal/implementation/`, and
`wire.go` as the only composition root. Add the port first, then the
implementation, then wire it.

`interface/` knows no technology. `implementation/` packages depend on
`interface/`, never on each other. `main.go` stays thin and imports no Wails.

## The run loop

The piece that takes reading several files to see is how a workflow actually
advances. Four collaborators, wired in that order in `wire.go`:

- **`WorkflowManager`** (`core/workflow_manager`) owns graph state: steps, their
  dependencies, readiness, retries, reviews, heartbeat deadlines. It hands out
  `ReadySteps`, takes `Report` from a finished agent, and publishes a
  `WorkflowProgress` channel from `Execute`. It does not know how to start an
  agent.
- **`Coordinator`** (`core/coordinator`) is the loop. `Run` calls
  `workflows.Execute`, then a goroutine reacts to each progress event and to a
  15s ticker by calling `schedule`: for every ready spec it requests an agent
  from `AgentManager`, `Assign`s it, builds the prompt, and sends it. It also
  spawns a per-agent heartbeat watcher that reports the step failed if the agent
  goes quiet. Kills happen here, not in the workflow manager.
- **`AgentManager`** (`core/agent_manager`) owns instances on top of one
  `AgentHarness` per vendor (`input/harness/claude_code`, `.../open_code`,
  `.../codex`), which are the things that actually spawn a subprocess and stream
  output.
- **`MCPProxy`** (`core/mcp_proxy`) closes the loop. It serves a local MCP
  server (`constances.GatewayLocalServer`, `"harness"`) exposing two tools:
  `request_approval` and `report_step`. An agent finishing its work calls
  `report_step`, which lands back in `WorkflowManager.Report` — that is the only
  way a step completes. The proxy also holds the OAuth credentials for remote
  MCP servers and swaps a placeholder for the real key on the way out.

`coordinator.buildPrompt` is where context engineering actually happens: it
composes the step's guidance, its write allowance, the `Handoff`s of every
dependency, and the instruction to call `report_step` exactly once. Read it
before changing anything about how steps talk to each other.

Note the deliberate cycle break: `WorkflowManager` needs live context usage and
activity from `AgentManager`, but `AgentManager` is built on top of the MCP
gateway that reports into `WorkflowManager`. So the workflow manager takes the
reader separately, after construction, via `TrackLiveAgents` (a narrow
`LiveAgentReader`).

## Durability

Every step write goes straight to SQLite (`input/storage`) in one transaction,
inside `WorkflowManager.persist` — a failed write rolls the in-memory change
back and surfaces as an error on the call that made it, so memory and disk never
disagree. Agents send heartbeats; one that goes quiet past the deadline has its
step dropped back into the pool. File changes are stored as unified diffs so a
revert never depends on git.

# Conventions

- Constructors are `New()` (or `InitV1()` for versioned implementations),
  disambiguated by package path.
- All ids are UUIDv7 via `github.com/google/uuid`, called inline at the call
  site — no `newID()` wrapper.
- Errors go through `custom_error`: `Critical`, `Bypass`, or `TypedCritical`
  with an `enums.ErrorType`.
- **Do not add new comments on code changes, except when explicitly asked or in
  the rare case that code cannot reflect its intent after renaming or splitting.**
  This holds for every language and every comment form — `//`, `/* */`, `/** */`
  doc blocks, JSDoc, TSDoc, Go doc comments. Comments already in the file stay
  unless the code under them changes.
- **Commit and push straight to `main`.** No feature branches, no pull requests,
  no worktrees — this is a solo repo with no reviewer on the other end. Commit
  only when asked, then push to `origin main`. Messages end at their last real
  line: no `Co-Authored-By`, no "Generated with" trailer.

# Frontend rules

- **Always run the `frontend-design` skill before frontend work.** Every new
  screen, component, or reshape of an existing one starts there, so the result
  is a deliberate design rather than framework defaults.
- **The feature `api` module is the only seam that touches Go.** Every feature
  now calls real bindings *and* keeps a fallback: the api module checks
  `hasWailsRuntime()` and either goes over the bindings or serves the feature's
  mock (`src/features/<domain>/mock-*.ts`, shared ones in `src/shared/api/`).
  That is what keeps port 8888 usable — under a plain vite server there is no
  Wails runtime. Nothing above the api layer may import a mock or a binding
  directly; ESLint `no-restricted-imports` enforces the mock half.
- **Workflows is the worked example.** `features/workflows/api/index.ts` holds
  graph editing, `api/remote-run.ts` builds a `RunWorkflowSpec`, starts the real
  run and polls `WorkflowStatus` until it settles, and `api/simulated-run.ts` is
  the same state machine without a backend. A new backed flow follows that split.
- **Types are DTOs, not component props.** Declare the shape in
  `src/features/<domain>/types.ts` as what the Go side returns (string ids, ISO
  timestamps), and map to view shapes at the api boundary.
- **Every flow must be end-to-end testable in the browser.** Mocks model state
  transitions and latency, not just a static payload — a run that starts must
  progress and finish, an approval must resolve, a failure path must be
  reachable. If a flow can only be demonstrated by editing a fixture by hand,
  it is not wired.
- **Irreducible domain terms get a `HelpTip`.** A word a newcomer cannot be
  expected to know — effort, handoff, review, autopilot — ships with
  `<HelpTip term="…" />` (`src/shared/components/help-tip.tsx`) next to it rather
  than a rename. `HelpTip` is keyed by term against
  `src/shared/lib/glossary.ts`, so putting such a term on screen without adding
  its glossary entry will not typecheck.
- **Written for a human maintainer, not a bot.** Small components with one job,
  names that say what the thing is, no cleverness that needs a comment to
  survive. Pure presentational components take props and render; state lives in
  hooks (the feature's `use-*.ts`, shared ones in `src/shared/hooks/`) and
  TanStack Query. Duplicate before you abstract — a wrong abstraction costs
  more than a repeated block.

# Performance

The hot path is agent output: a subprocess streams lines, Go forwards them over
the bridge, React renders them. It runs continuously for the life of a run, and
everything below exists because of it. Measure before optimizing anything not
on that path — `net/http/pprof` behind a debug flag for Go, Safari Web
Inspector against the WebContent process for the frontend.

- **Coalesce events on the Go side, never the JS side.** Throttle `EventsEmit`
  to ≤10Hz and batch whatever accumulated since the last tick into one payload.
  An agent that prints fast must not become one IPC message per line.
- **No N+1 across the bridge.** One call returning many, never a loop of bound
  calls — per-agent polling in a `for` loop is the shape to avoid. Return page
  slices and DTOs holding the fields the view needs, not whole tables or fat
  nested structs.
- **Write on commit, not on every frame.** Continuous gestures — step drag,
  search-as-you-type, resize — render from local state and cross the bridge
  once, when the gesture ends.
- **Every goroutine has an exit path; every ticker has a `Stop()`.** Output
  pumps and heartbeats outlive the call that started them, so they exit on
  context cancel or channel close. `runtime.NumGoroutine()` stays flat at idle.
- **Every `EventsOn` returns its unsubscribe from `useEffect`.** A leaked
  listener pins the closure and everything it captured.
- **Bound anything that grows for the length of a run.** Run logs are the case
  that matters: cap the lines held in React state and let SQLite keep the rest.
  Virtualize a list past ~100 rows instead of mounting all of it.
- Set `GOMEMLIMIT` via `debug.SetMemoryLimit` in `main()`. Preallocate when the
  size is known (`make([]T, 0, n)`).

# Gotchas

- Run from the repo root. `config.yaml` is embedded at compile time, so an edit
  to it takes effect on the next build — there is no file read at startup.
- `go run .` fails with a Wails build-tag error — use `make dev`.
- The SQLite driver name is `sqlite` (modernc), not `sqlite3`.
- `storage.New` opens the DB through `dsn()`, which sets `busy_timeout` and
  `journal_mode(WAL)`. Do not drop those. `database/sql` pools connections and
  modernc installs no default busy handler, so without them the second writer of
  any overlapping pair fails instantly with `SQLITE_BUSY` — and a failed step
  write makes `WorkflowManager` roll back a result that already happened.
- `internal/implementation/input/harness/**` says "session" on purpose, and it
  stays. There it means the *vendor's* session — OpenCode's REST `/session`
  endpoint, Claude Code's login session — not a workflow of ours. Do not "fix"
  it to `workflow`; everywhere else "session" is a stale name and should be one
  of `docs/NAMING.md`'s terms instead.
- `viper.Read()` returns nil on unmarshal error, so a config typo surfaces as a
  nil dereference rather than a clear error.
- `config.yaml`'s `app.data_dir` is the storage key for every on-disk path.
  Renaming it orphans the database, the installed harnesses and the user config.
- The generated bindings in `frontend/wailsjs/` dereference `window.go` /
  `window.runtime` **synchronously**, so outside the Wails webview they throw
  rather than reject. `.catch()` on the returned promise never fires — go
  through `shared/api/bridge.ts` (`bridge` / `hasWailsRuntime`) or the component
  crashes the whole React tree under a plain `vite` dev server.
- `@xyflow/react/dist/base.css` must be imported with `layer(base)`. Unlayered
  CSS beats layered CSS at any specificity, so without it React Flow's defaults
  silently override the project's `@layer components` edge and handle styles.
