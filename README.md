# Nexo — Nexo EXecution Orchestrator

**Streamlined, debuggable, retryable agentic workflow orchestration — that you can set up yourself.**

[![License](https://img.shields.io/github/license/rye-ndt/nexo)](LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/rye-ndt/nexo)](go.mod)
[![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)](#requirements)
[![Stars](https://img.shields.io/github/stars/rye-ndt/nexo?style=social)](https://github.com/rye-ndt/nexo/stargazers)

![Nexo running a graph of agents](docs/demo.gif)

You draw the work as a graph. Each node is one scoped task: a role, a model, the
files it's allowed to touch, plus whatever guidance it needs. Nexo runs agents
through the graph, watches them, captures the diff, and feeds what each node
learned into the next node's prompt.

---

## Contents

- [Why Nexo exists](#why-nexo-exists)
- [Requirements](#requirements)
- [Quick start](#quick-start)
- [Config](#config)
- [Architecture](#architecture)
- [Driving Nexo from Claude Code](#driving-nexo-from-claude-code)
- [FAQ](#faq)
- [Support](#support)
- [License](#license)

---

## Why Nexo exists

I believe the limitation of LLMs on long, multi-step work isn't the model — it's
the context. An agent that starts a task knowing only the task will improvise the
rest. So Nexo makes the handover between steps the thing you engineer.

Claude Code and OpenCode are great at one task, and Nexo runs on top of them
anyway. But:

- can you delegate a 10-step job, walk away, and trust it without reading every line?
- can you make it follow _your_ procedure instead of improvising one?
- can you see why it chose what it chose?
- what happens when your vendor triples the price?
- can you undo step 4 and keep steps 1 through 3?

That's what this is for. Concretely, four things separate it from a wrapper:

**Nodes are a DAG, not a queue.** Chaining is the entire point.

**The handover doc is the product.** A finished node writes down what it did,
what blocked it, what got approved, what got _rejected_, what to avoid, and
what's still broken. That doc becomes the next node's context. Telling the next
agent what **not** to do matters as much as telling it what happened.

**It assumes agents fail.** Every task write commits to SQLite as it happens, so
a crash costs you nothing that was already reported. Agents send heartbeats — go
quiet and your task drops back into the pool. Every file change is stored as a
unified diff, so you can revert without depending on git.

**Agents never see your secrets.** Nexo does the OAuth once, encrypts the token,
and hands agents a placeholder name. The proxy swaps in the real key on the way
out, and refuses anything aimed at an auth endpoint.

## Requirements

- macOS. Windows is half-built and untested.
- Go 1.25+
- Node 20+
- Wails: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

Wails installs to `~/go/bin`, which might not be on your PATH. The Makefile
finds it either way.

## Quick start

```sh
git clone https://github.com/rye-ndt/nexo.git
cd nexo
make dev
```

Then, in the app:

1. install and sign in to Claude Code / Open Code in settings
2. authorize any MCP servers you need
3. start a session
4. make templates — each one is a kind of agent that does a kind of work
5. drop nodes on the canvas and wire them together, mixing agent types freely
6. finalize the session, run it
7. go play pickleball
8. come back, review, revert what you don't like
9. commit

Two things that will bite you: `config.yaml` is compiled into the binary, so
editing it needs a rebuild rather than a restart. And use `make dev` — `go run .`
dies on a Wails build tag.

## Config

`config.yaml` holds window size, timeouts, model lists, and the MCP servers you
can authorize.

**Change `encode_key` before you use this for real.** It encrypts your stored
OAuth tokens, and the default value is committed to this repo, so it isn't a
secret. Also: a typo anywhere in that file surfaces as a nil panic instead of a
readable error. Blame viper.

## Architecture

Hexagonal, and enforced. Ports live in `internal/interface/{core,input,output}`
and know no technology. Each package under `internal/implementation/` is one
technology, depends only on `interface/`, and never on a sibling. `wire.go` is
the only composition root; `main.go` imports no Wails.

| Layer | Path | Holds |
| --- | --- | --- |
| Ports | `internal/interface/` | `core_itf`, `input_itf`, `output_itf` — pure contracts |
| Core | `internal/implementation/core/` | session manager, coordinator, agent manager, MCP proxy, template manager |
| Input | `internal/implementation/input/` | SQLite storage, agent harnesses, config, archives |
| Output | `internal/implementation/output/` | frontend API, logger, message queue, user config |
| UI | `frontend/` | React + TypeScript + React Flow canvas |
| Shim | `cmd/nexo-mcp/` | stdio MCP server for driving Nexo from another agent |

### The run loop

Four collaborators, wired in this order:

- **SessionManager** owns graph state — tasks, dependencies, readiness, retries,
  accept gates, heartbeat deadlines. It hands out ready tasks and takes reports.
  It does not know how to start an agent.
- **Coordinator** is the loop. It reacts to progress events and a 15s ticker,
  requests an agent per ready task, builds the prompt, and watches heartbeats.
  `buildPrompt` is where context engineering actually happens: task guidance,
  write allowance, and the handover doc of every dependency.
- **AgentManager** owns instances on top of one harness per vendor
  (`claude_code`, `open_code`) — the things that spawn a subprocess and stream
  output.
- **MCPProxy** closes the loop. It serves a local MCP server exposing
  `request_approval` and `report_task`. An agent calling `report_task` is the
  only way a node completes.

### Working on it

```sh
make dev      # run the app; frontend hot-reloads, Go changes need a restart
make dev-go   # same, but rebuilds on Go changes
make build    # production .app bundle
make run      # build, then launch the bundle
make test     # go test ./...
```

In `frontend/`: `npm run dev` (vite on 8888, no Wails runtime), `npx tsc --noEmit`,
`npm run lint`, `npm run format`. Regenerate bindings with
`wails generate module` — never hand-edit `frontend/wailsjs/`.

The suites that matter are `core/session_manager` (gating, retry, cancel),
`core/mcp_proxy` (credential revoke, account lookup) and `input/storage`.

## Driving Nexo from Claude Code

Nexo serves an MCP endpoint on a random port behind a token that changes every
launch, so a static HTTP registration would go stale daily. `cmd/nexo-mcp` is a
tiny stdio server that fixes that: it reads the current port and token from the
file the app writes and forwards each call there. Register it once and it keeps
working across app restarts.

```sh
make mcp-shim
claude mcp add nexo -- /full/path/to/nexo/build/bin/nexo-mcp
```

Use the absolute path — Claude Code does not resolve it against your project.

Eight tools come across: list the templates you have, create a session from one,
start it, pause it, cancel it, read its status, list your sessions, and answer an
acceptance gate a node is waiting on. So Claude Code can hand a whole graph of
work to Nexo and check back on it, while Nexo remains the thing that runs the
agents.

Which is also the catch: **none of it works while the Nexo app is closed.** The
shim only forwards; the app is what spawns agents, holds the graph and writes to
the database. Add `-launch` to have the shim open Nexo for you and wait for it to
come up, rather than failing the call:

```sh
claude mcp add nexo -- /full/path/to/nexo/build/bin/nexo-mcp -launch
```

## FAQ

**Is this vibe-coded?** No. I built the backend and the architecture by hand, so
I know how it works. Claude helped with the frontend.

**Is this another LLM wrapper?** See [Why Nexo exists](#why-nexo-exists). Short
version: the graph, the handover doc, the durability and the secret handling are
the product. The model call is the easy part.

## Support

Questions and bug reports go to [GitHub Issues](https://github.com/rye-ndt/nexo/issues).

Want help setting this up for your team, custom templates, or tuning it for your
company? Contact nduytung.1611@gmail.com.

If Nexo saves you time, you can buy me a coffee:

- BTC `bc1q5kaek7dnkfah0h958ys4lzn574yu6lpu2tzwsr`
- EVM `0xeb8d8C405C4d377fcFB61FFF3b0Bb730B103A51f`
- SVM `rypBpVeH5HmMXKkMRHy3W4fWRhCUSPHsxD3hDy46QPG`

## License

[MIT](LICENSE) · v0.0.1 · last big update Aug 4, 2026
