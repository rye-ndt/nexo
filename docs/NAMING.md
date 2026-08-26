# Naming

One word per concept, the same word in Go, in SQL, in TypeScript, and on screen.
If you are adding a concept, add it here first. If you are renaming one, this
file is the only place that decides what it becomes.

## The model, in one sentence

A **workflow** is **steps**; each step has a **role**, runs on an **agent**, and
leaves a **handoff** for the next step.

## Vocabulary

| Term | Is | Not |
| --- | --- | --- |
| **Workflow** | The graph you author, lock, and run. Bound to one project folder. | Not "session", not "run", not "job". |
| **Step** | One scoped unit of work in the graph. What the canvas draws as a box. | Not "node" (graph jargon), not "task" (was the internal name). |
| **Role** | A reusable definition of a kind of worker: what it does, its inputs, its effort, its report format. A step starts from one. | Not "template", not "preset". |
| **Agent** | The vendor CLI that executes a step — Claude Code, OpenCode, Codex. Installed and logged in under Settings › Agents. | Never the step, never the role. This is the only meaning of "agent". |
| **Agent defaults** | The ordered `agent_defaults:` list in `config.yaml` — per harness, the model and thinking level for each effort. Its order *is* the priority: a step with no explicit model runs on the first listed harness that is logged in. | Not "fallback chain", not "ranking". There is no separate priority field to add — reorder the list. |
| **Handoff** | What a finished step writes down for the steps after it. The product of a step. | Not "handover doc", not "report". |
| **Result** | Everything a finished step produced: its status, its handoffs, its context usage, its activity. Contains handoffs. | Not "report". |
| **Lock** | Freezing a workflow's graph and folder so it can run. Reversible by unlocking, which discards the run. | Not "finalize". No longer permanent — it was, until Unlock existed. |
| **Unlock** | Reopening a locked workflow for editing. Discards the current run: the project folder goes back to the baseline the run started from, and the run's results move to history. | Not "revert" — that rewinds files while a run continues. Not "reset": the run is archived, not erased. |
| **Run** | One execution of a locked workflow, from pressing Run until it settles. A workflow can have several over its life. | Not the workflow itself, not a "job". `Step.run` is a step's slice of the *current* run. |
| **Past run** | A run that unlocking archived. Read-only; the last few are kept. | Not "version", not "revision": nothing about the graph is restored from it. |
| **Approval** | The agent stopped mid-step to ask you something it should not decide alone. | Not "question", not "gate". Kept as-is: once the accept gate became *review*, nothing collides with it. |
| **Recommendation** | The one option an approval singles out as the agent's own pick. Every approval has exactly one; it is listed first, and under autopilot it is the answer taken without asking. | Not "default", not "suggestion": the agent is answering its own question, not offering a fallback. Exactly one, never zero and never several — autopilot has to have a single answer to take. |
| **Review** | A step finished and is holding everything downstream until you accept or reject it. | Not "accept gate", not "manual acceptance". |
| **Effort** | How hard a role tries: Quick, Standard, Deep, Exhaustive. | Not "task level". "Daily" is gone; it described frequency, not effort. |
| **Input** | A named value a step fills in before it runs. Declared by the role. | Not "param". The kind of value is an `InputType`. |
| **Knowledge base** | The folder agents share across every workflow on a project — `AGENTS.md`, the glossary, the gotchas. Derived from the project folder by `helpers.KnowledgeDir`. | Not a handoff. Handoffs never land here; they travel in the prompt. |
| **Instructions** | The role's named prompt blocks, composed into what the agent receives. | Not "system prompts". The step's own free text stays **Prompt**. |
| **Project folder** | The checkout agents read and change. | Not "working directory". |
| **Duplicate** | Copy a workflow, new ids, run history cleared, project folder empty. Asks once whether the values typed into its steps come across. For trying another approach beside the original — changing a locked workflow is Unlock's job now. | Not "clone". Both words existed; only this one survives. Not the way out of a lock. |
| **Tour** | The one-time walkthrough a new user gets after onboarding: three surfaces, explained where they live. | Not "tutorial", not "walkthrough", not "coach marks". |
| **Stop** | One beat of the tour — the surface it rings and the sentence it says about it. | Not "step": a step is work an agent runs, and the tour runs nothing. |
| **Conflict aware** | A role's switch, on for a new role, that gives it one more instruction: other agents are working the same project folder right now, so write and edit in ways that cannot collide. Held as an instruction under the reserved key `conflict_awareness`, not as a field of its own. | Not "parallel safe", not "locking": nothing is locked, the agent is only told what else is running. |
| **Max running agents** | How many agents may run at once across every workflow. `config.yaml` sets the default; Settings › Preferences owns it after that. | Not the per-harness `max_instance`, which caps one vendor CLI underneath this. |
| **Language** | The language Nexo's own interface speaks — English or Vietnamese. Picked on the first screen of onboarding, changeable in Settings › Preferences. | Not "locale": nothing about dates, numbers or currency changes with it. Not the language an agent answers in — that follows the prompt. |

## Removed concepts

These are gone from the product, not renamed. Do not reintroduce them.

- **Context directory.** Always derived from the project folder. One function
  owns the path; nothing asks the user for it and nothing stores it.
- **Finalize as a distinct idea.** It is locking, and it is called locking.
- **Handover doc *and* report as separate nouns on screen.** A step produces a
  result; the part that travels is its handoffs.
- **The "Daily" effort level.** It named a cadence in a scale about effort.

## Rename map

Applied across Go, SQL, and TypeScript. Left column must return zero hits.

### Types and identifiers

| Was | Is |
| --- | --- |
| `Session`, `SessionEntity`, `SessionInfo`, … | `Workflow`, `WorkflowEntity`, `WorkflowInfo`, … |
| `Task`, `TaskEntity`, `TaskSpec`, `TaskStatus`, … | `Step`, `StepEntity`, `StepSpec`, `StepStatus`, … |
| `TaskReport` | `StepResult` |
| `TaskLevel` | `Effort` |
| `Template`, `TemplateEntity`, `AgentTemplateManager`, … | `Role`, `RoleEntity`, `RoleManager`, … |
| `TemplateParam` | `RoleInput` |
| `HandoverDoc` | `Handoff` |
| `ManualAcceptRequired` | `PauseForReview` |
| `WorkingDirPath` | `ProjectDirPath` |
| `ContextDirPath` | *derived, no field* |
| `Clone…` | `Duplicate…` |
| `ParamType`, `TextParam`, … | `InputType`, `TextInput`, … |
| `SystemPrompts` | `Instructions` |
| `Params` | `Inputs` |

### Effort values

| Was | Is |
| --- | --- |
| `lightweight_task` | `quick` |
| `daily_task` | `standard` |
| `heavy_task` | `deep` |
| `maximum_effort_task` | `exhaustive` |

### Packages and paths

| Was | Is |
| --- | --- |
| `core/session_manager` | `core/workflow_manager` |
| `core/session_control` | `core/workflow_control` |
| `core/template_manager` | `core/role_manager` |
| `input/session_archive` | `input/workflow_archive` |
| `input/template_archive` | `input/role_archive` |
| `features/sessions/` | `features/workflows/` |
| `features/templates/` | `features/roles/` |

### SQL tables

Renamed by appended migrations, never by editing an existing one.

| Was | Is |
| --- | --- |
| `sessions` | `workflows` |
| `tasks` | `steps` |
| `task_reports` | `step_results` |
| `agent_templates` | `roles` |
| `session_drafts` | `workflow_drafts` |

### MCP tools

Every tool the app serves, on both servers. Verified against
`mcp_proxy/v1_local.go` and `mcp_proxy/v1_control.go` — if you change a tool
name, change it here in the same commit.

| Was | Is |
| --- | --- |
| `report_task` | `report_step` |
| `report_template` | `report_role` |
| `list_templates` | `list_roles` |
| `create_session` | `create_workflow` |
| `start_session` | `start_workflow` |
| `pause_session` | `pause_workflow` |
| `cancel_session` | `cancel_workflow` |
| `session_status` | `workflow_status` |
| `list_sessions` | `list_workflows` |
| `answer_acceptance` | `answer_review` |

`request_approval` is the only tool whose name did not change.

`create_workflow` used to take a `context_dir_path`. It is derived now, and the
argument is ignored rather than refused, so a caller written against the old
schema keeps working.

## Where the old words are still correct

Two places keep the old vocabulary on purpose. Both are load-bearing; a future
sweep that "finishes the job" here would break the app.

**`internal/implementation/input/harness/**`** — every `session` in the harness
packages is the *vendor's* session, not ours: OpenCode's REST `/session`
endpoint and the id it returns, Claude Code's login session, the `role` field on
a chat message. These packages never touch our own types, so they were exempted
from the rename wholesale.

**`params` in wire protocols we do not own** — JSON-RPC's `params` member
(`mcp_proxy/v1_rpc.go`), Chrome DevTools Protocol's `params` (`helpers/cdp.go`),
and `urn:ietf:params:oauth:grant-type:device_code`. A grep for `params` can
never come back empty, and should not.

**External flags and APIs that merely contain one of our words** — `git init
--template=`, `init.templateDir`, `slices.Clone`, `structuredClone`, React's
`ReactNode`, and every `@xyflow/react` identifier (`Node`, `NodeProps`,
`NodeChange`, `nodeTypes`, `onNodesChange`, the `react-flow__*` CSS classes).
React Flow's word for a graph vertex stays React Flow's word; ours is *step*.

## Vietnamese

Every string the interface shows lives in `frontend/src/shared/lib/i18n/messages/`,
one file per feature, keyed `feature.area.thing` and carrying both languages side
by side so neither can drift from the other. `t('key')` reads whichever is
current; there is no second place a string may live.

The Vietnamese is written for a Vietnamese developer, not translated word for
word. Terms that developer already reads in English every day stay in English —
translating them makes the app *harder* to use, not easier. Left column is what
the interface says in Vietnamese; it is as binding as the English column above.

| English | Vietnamese | Note |
| --- | --- | --- |
| Nexo | Nexo | Never translated, never declined. |
| Workflow | workflow | Loanword. "Luồng công việc" reads like a translated manual. |
| Step | bước | |
| Role | vai trò | |
| Agent | agent | Loanword — it is a CLI they already have installed. |
| Handoff | bàn giao | |
| Result | kết quả | |
| Lock | khóa | |
| Unlock | mở khóa | |
| Run (noun) | lần chạy | The verb on the button stays "Chạy". |
| Past run | lần chạy trước | |
| Approval | phê duyệt | |
| Recommendation | đề xuất | The option is marked "Được đề xuất". |
| Review | kiểm duyệt | Deliberately a different word from approval; the two gates must not read alike. |
| Effort | mức nỗ lực | Quick / Standard / Deep / Exhaustive → Nhanh / Tiêu chuẩn / Sâu / Tối đa. |
| Input | đầu vào | |
| Knowledge base | kho kiến thức | |
| Instructions | chỉ dẫn | |
| Prompt | prompt | Loanword. |
| Project folder | thư mục dự án | |
| Duplicate | nhân bản | |
| Conflict aware | tránh xung đột | |
| Tour | hướng dẫn | |
| Stop | chặng | "bước" is taken by Step. |
| Model | model | Loanword. |
| Thinking | suy luận | |
| Revert | hoàn tác | |
| Autopilot, MCP, CLI, diff, token | unchanged | Product and protocol names. |

Two rules that are not in the table:

- **A language names itself in itself.** The picker says "English" and "Tiếng
  Việt" whichever language is active. `LANGUAGE_NAMES` is not a message key.
- **Go speaks English.** `custom_error` messages cross the bridge as they are;
  what the interface translates is the *reading* of an error code in
  `shared/lib/errors.ts`, not the sentence Go wrote.

## Terms that had to stay

Some words cannot be reduced to something a newcomer already knows. Those get a
`HelpTip` next to them on screen rather than a rename, and every one of them has
an entry in `frontend/src/shared/lib/glossary.ts`. Adding such a term to the UI
without a glossary entry is a bug: `HelpTip` is keyed by term, so a missing
entry will not typecheck.
