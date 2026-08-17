import type {TaskState} from '@/shared/lib/enums'
import type {HandoverDoc, Session, Task, TaskReport} from '@/features/sessions/types'

const secondsAgo = (seconds: number) => new Date(Date.now() - seconds * 1000).toISOString()

const TEMPLATE_IDS = {
    reviewer: '0192f3a1-0001-7000-8000-000000000001',
    tests: '0192f3a1-0002-7000-8000-000000000002',
    docs: '0192f3a1-0003-7000-8000-000000000003',
    ticket: '0192f3a1-0004-7000-8000-000000000004',
}

function handover(
    doc: Partial<HandoverDoc> & Pick<HandoverDoc, 'task' | 'tldr' | 'outcome'>,
): HandoverDoc {
    return {
        blockers: {},
        approvedDecisions: {},
        rejectedDecisions: {},
        currentBehaviors: {},
        changedBehaviors: {},
        mustAvoid: {},
        nuances: {},
        knownGaps: {},
        ...doc,
    }
}

export const MOCK_SESSIONS: Session[] = [
    {
        id: '0198e3a1-0000-7000-8000-000000000001',
        name: 'Coordinator port',
        createdAt: '2026-07-30T09:12:00Z',
        finalized: false,
        started: false,
        cancelled: false,
        workingDir: '/Users/rye/dev/agent-harness',
        contextDir: '/Users/rye/dev/agent-harness/.harness/context',
        tasks: [
            {
                id: 'task-1',
                title: 'Map the task graph ports',
                prompt: 'Read internal/interface/core and list every port the coordinator will need to call. Report the gaps.',
                state: 'idle',
                position: {x: 0, y: 120},
                dependsOn: [],
                templateId: TEMPLATE_IDS.reviewer,
                values: {},
            },
            {
                id: 'task-2',
                title: 'Draft the coordinator interface',
                prompt: 'Using the handover, add a Coordinator port to internal/interface/core. Keep it minimal.',
                state: 'idle',
                position: {x: 340, y: 0},
                dependsOn: ['task-1'],
                templateId: TEMPLATE_IDS.tests,
                values: {
                    package_path: 'internal/interface/core',
                    max_cases: 6,
                    focus: 'The port compiles and nothing in implementation/ imports it yet.',
                },
            },
            {
                id: 'task-3',
                title: 'Write the handover builder',
                prompt: 'Build the HandoverDoc composer that turns a finished node into the next node prompt.',
                state: 'idle',
                position: {x: 340, y: 240},
                dependsOn: ['task-1'],
                templateId: TEMPLATE_IDS.docs,
                values: {doc_path: 'docs/handover-format.md'},
            },
            {
                id: 'task-15',
                title: 'Read the tracker ticket',
                prompt: 'Fetch the ticket at https://atlassian/tickets/{{ticket_id}} and restate it as a task with a definition of done.',
                state: 'idle',
                position: {x: 340, y: 480},
                dependsOn: ['task-1'],
                templateId: TEMPLATE_IDS.ticket,
                values: {ticket_id: '', include_comments: true},
            },
            {
                id: 'task-4',
                title: 'Wire it into wire.go',
                prompt: 'Compose the coordinator in wire.go. It is the only composition root.',
                state: 'idle',
                position: {x: 680, y: 120},
                dependsOn: ['task-2', 'task-3'],
                templateId: TEMPLATE_IDS.reviewer,
                values: {},
            },
            {
                id: 'task-5',
                title: 'Typecheck and vet',
                prompt: 'Run go build ./..., go vet ./..., and npx tsc --noEmit in frontend/. Report anything that fails.',
                state: 'idle',
                position: {x: 1020, y: 120},
                dependsOn: ['task-4'],
                templateId: TEMPLATE_IDS.tests,
                values: {package_path: './...', max_cases: 0, focus: 'Nothing new fails.'},
            },
        ],
    },
    {
        id: '0198e3a1-0000-7000-8000-000000000007',
        name: 'Streaming errors',
        createdAt: '2026-08-05T16:05:00Z',
        finalized: true,
        started: false,
        cancelled: false,
        workingDir: '/Users/rye/dev/agent-harness',
        contextDir: '/Users/rye/dev/agent-harness/.harness/context',
        tasks: [
            {
                id: 'task-21',
                title: 'Read the streaming path',
                prompt: 'Follow a streamed MCP response from the proxy to the agent and say where the error is dropped.',
                state: 'idle',
                position: {x: 0, y: 120},
                dependsOn: [],
                templateId: TEMPLATE_IDS.reviewer,
                values: {},
            },
            {
                id: 'task-22',
                title: 'Pull the ticket that reported it',
                prompt: 'Fetch the ticket at https://atlassian/tickets/{{ticket_id}} and quote the acceptance criteria verbatim.',
                state: 'idle',
                position: {x: 340, y: 0},
                dependsOn: ['task-21'],
                templateId: TEMPLATE_IDS.ticket,
                values: {ticket_id: '', include_comments: true},
            },
            {
                id: 'task-23',
                title: 'Cover the dropped error',
                prompt: 'Write the tests that prove a stream failing after its first chunk still fails the call.',
                state: 'idle',
                position: {x: 340, y: 240},
                dependsOn: ['task-21'],
                templateId: TEMPLATE_IDS.tests,
                values: {
                    package_path: '',
                    max_cases: 6,
                    focus: 'A stream that dies mid-body fails the call instead of ending it.',
                },
            },
            {
                id: 'task-24',
                title: 'Say so in the proxy docs',
                prompt: 'Bring the streaming paragraph of the README in line with what the tests now prove.',
                state: 'idle',
                position: {x: 680, y: 120},
                dependsOn: ['task-22', 'task-23'],
                templateId: TEMPLATE_IDS.docs,
                values: {doc_path: 'README.md'},
            },
        ],
    },
    {
        id: '0198e3a1-0000-7000-8000-000000000006',
        name: 'Approval gate',
        createdAt: '2026-08-04T11:15:00Z',
        finalized: true,
        started: true,
        cancelled: false,
        workingDir: '/Users/rye/dev/agent-harness',
        contextDir: '/Users/rye/dev/agent-harness/.harness/context',
        tasks: [
            {
                id: 'task-19',
                title: 'Review the proxy deadline change',
                prompt: 'Read the deadline change on the proxy and report the defects you can point to a line for.',
                state: 'awaiting_accept',
                position: {x: 0, y: 60},
                dependsOn: [],
                templateId: TEMPLATE_IDS.reviewer,
                values: {},
                run: {
                    startedAt: secondsAgo(430),
                    finishedAt: secondsAgo(295),
                    context: {used: 61200, total: 200000},
                },
                report: {
                    status: 'awaiting_accept',
                    handoverDocs: [
                        handover({
                            task: 'Review the proxy deadline change',
                            tldr: 'I read through a change that makes calls to outside tools give up after a set waiting time, and it works, but nobody has decided what should happen when no waiting time is configured.',
                            outcome:
                                'The deadline lands where it should and the typed error is right. One thing I could not settle on my own: the default is zero, which reads as no deadline at all, and the docs node downstream will repeat whatever I say here. Read it before I hand it on.',
                            approvedDecisions: {
                                typed_error:
                                    'A timeout surfaces as enums.ErrorTypeUpstream, so the coordinator can tell it apart from a bad request.',
                                config_source:
                                    'The deadline comes from config rather than a constant, so an operator can raise it per install.',
                            },
                            rejectedDecisions: {
                                retry_in_proxy:
                                    'I dropped the retry the prompt hinted at. Retries belong to the session manager, not the proxy.',
                            },
                            currentBehaviors: {
                                default:
                                    'A missing mcp_proxy block leaves ForwardDeadline at zero, which means no deadline at all.',
                            },
                            changedBehaviors: {
                                forward:
                                    'Forward() wraps the request in a context with p.deadline and returns a typed error on expiry.',
                            },
                            mustAvoid: {
                                shared_context:
                                    'Do not reuse p.ctx for a call. Cancelling it would kill the proxy, not the request.',
                            },
                            nuances: {
                                streaming:
                                    'Streaming calls take the same deadline, so a long stream now dies at 20s. That may be wrong.',
                            },
                            knownGaps: {
                                zero_default:
                                    'Nobody decided whether zero should mean no deadline or a 30s default. The next node will document whichever you accept.',
                            },
                        }),
                    ],
                },
            },
            {
                id: 'task-20',
                title: 'Write up the deadline',
                prompt: 'Bring the proxy section of the README in line with the review you were handed.',
                state: 'blocked',
                position: {x: 340, y: 60},
                dependsOn: ['task-19'],
                templateId: TEMPLATE_IDS.docs,
                values: {doc_path: 'README.md'},
            },
        ],
    },
    {
        id: '0198e3a1-0000-7000-8000-000000000004',
        name: 'Retry policy',
        createdAt: '2026-08-01T07:20:00Z',
        finalized: true,
        started: true,
        cancelled: false,
        workingDir: '/Users/rye/dev/agent-harness',
        contextDir: '/Users/rye/dev/agent-harness/.harness/context',
        tasks: [
            {
                id: 'task-12',
                title: 'Read the retry paths',
                prompt: 'Find every place a failed task is retried today and say which of them the coordinator will own.',
                state: 'done',
                position: {x: 0, y: 60},
                dependsOn: [],
                templateId: TEMPLATE_IDS.reviewer,
                values: {},
                run: {
                    startedAt: secondsAgo(280),
                    finishedAt: secondsAgo(196),
                    context: {used: 38400, total: 200000},
                },
                report: {
                    status: 'done',
                    handoverDocs: [
                        handover({
                            task: 'Read the retry paths',
                            tldr: 'I searched the project for every place where failed work is automatically tried again, and wrote down which of them a single new manager should take over.',
                            outcome:
                                'Retries live in two places today. The rename moves the owner into session_manager, where the coordinator can reach it.',
                            currentBehaviors: {
                                retry_owner:
                                    'task_manager/v1.go retries in place, with no backoff and no cap.',
                                heartbeat:
                                    'A missed heartbeat kills the agent but never schedules a retry.',
                            },
                            knownGaps: {
                                backoff: 'No backoff table exists yet. The next node adds one.',
                            },
                        }),
                    ],
                },
            },
            {
                id: 'task-13',
                title: 'Add the backoff table',
                prompt: 'Add a backoff table to the session manager and use it before every retry. Cap the wait.',
                state: 'running',
                position: {x: 340, y: 60},
                dependsOn: ['task-12'],
                templateId: TEMPLATE_IDS.tests,
                values: {
                    package_path: 'internal/implementation/session_manager',
                    max_cases: 8,
                    focus: 'The wait never exceeds the last entry in the table.',
                },
                run: {startedAt: secondsAgo(14), context: {used: 74600, total: 200000}},
            },
            {
                id: 'task-14',
                title: 'Wire retries into the coordinator',
                prompt: 'Call the retry policy from the coordinator when a node reports failed.',
                state: 'blocked',
                position: {x: 680, y: 60},
                dependsOn: ['task-13'],
                templateId: TEMPLATE_IDS.reviewer,
                values: {},
            },
        ],
    },
    {
        id: '0198e3a1-0000-7000-8000-000000000002',
        name: 'MCP proxy hardening',
        createdAt: '2026-07-28T14:40:00Z',
        finalized: true,
        started: true,
        cancelled: false,
        workingDir: '/Users/rye/dev/agent-harness',
        contextDir: '/Users/rye/dev/agent-harness/.harness/context',
        tasks: [
            {
                id: 'task-6',
                title: 'Audit the forwarding path',
                prompt: 'Trace every request the proxy forwards and note where errors are swallowed.',
                state: 'done',
                position: {x: 0, y: 60},
                dependsOn: [],
                templateId: TEMPLATE_IDS.reviewer,
                values: {},
                run: {
                    startedAt: '2026-07-28T14:41:02Z',
                    finishedAt: '2026-07-28T14:47:39Z',
                    context: {used: 41200, total: 200000},
                },
                report: {
                    status: 'done',
                    handoverDocs: [
                        handover({
                            task: 'Audit the forwarding path',
                            tldr: 'I read the code that passes requests out to other tools, changing nothing, and found that two of the three routes quietly hide failures instead of reporting them.',
                            outcome:
                                'Read-only pass. Three forwarding paths, two of them swallow the upstream error and return an empty response.',
                            currentBehaviors: {
                                forward:
                                    'Forward() returns the transport error verbatim, untyped, with no deadline.',
                                stream: 'Streaming calls drop the error once the first chunk is written.',
                            },
                            mustAvoid: {
                                retry_here:
                                    'Do not retry inside the proxy. Retries belong to the session manager.',
                            },
                            knownGaps: {
                                deadline:
                                    'Nothing bounds a forwarded call. The next node adds one.',
                            },
                        }),
                    ],
                },
            },
            {
                id: 'task-7',
                title: 'Add timeouts',
                prompt: 'Every forwarded call gets a deadline from config. Fail the call, not the node.',
                state: 'done',
                position: {x: 340, y: 60},
                dependsOn: ['task-6'],
                templateId: TEMPLATE_IDS.tests,
                values: {
                    package_path: 'internal/implementation/mcp_proxy',
                    max_cases: 8,
                    focus: 'A server that never answers fails the call inside the deadline.',
                },
                run: {
                    startedAt: '2026-07-28T14:48:10Z',
                    finishedAt: '2026-07-28T15:02:55Z',
                    retryCount: 1,
                    context: {used: 152800, total: 200000},
                },
                report: {
                    status: 'done',
                    handoverDocs: [
                        handover({
                            task: 'Add timeouts',
                            tldr: 'I gave every request to an outside tool a waiting time set by the operator, so a tool that never answers now fails in about twenty seconds instead of freezing the step forever.',
                            outcome:
                                'Forwarded calls now take a deadline from config. A hung server returns a typed error in 20s instead of pinning the node.',
                            approvedDecisions: {
                                deadline_source:
                                    'The deadline is read from config, not hard-coded, so an operator can raise it per install.',
                                error_type:
                                    'A timeout surfaces as enums.ErrorTypeUpstream so the coordinator can tell it apart from a bad request.',
                            },
                            rejectedDecisions: {
                                per_server_deadline:
                                    'One deadline for every server. Per-server config was more surface than the problem needs.',
                            },
                            changedBehaviors: {
                                forward:
                                    'Forward() wraps the request in a context with p.deadline and returns a typed error on expiry.',
                            },
                            mustAvoid: {
                                shared_context:
                                    'Do not reuse p.ctx directly for a call. Cancelling it would kill the proxy, not the request.',
                            },
                            nuances: {
                                default:
                                    'A missing mcp_proxy block leaves ForwardDeadline at zero, which means no deadline. The doc node has to say so.',
                            },
                            knownGaps: {
                                streaming:
                                    'Streaming calls still drop the error after the first chunk. Out of scope here.',
                            },
                        }),
                    ],
                },
            },
            {
                id: 'task-8',
                title: 'Say so in the README',
                prompt: 'Bring the proxy section of the README in line with what the code now does.',
                state: 'done',
                position: {x: 680, y: 60},
                dependsOn: ['task-7'],
                templateId: TEMPLATE_IDS.docs,
                values: {doc_path: 'README.md'},
                run: {
                    startedAt: '2026-07-28T15:03:20Z',
                    finishedAt: '2026-07-28T15:06:01Z',
                    context: {used: 22400, total: 200000},
                },
                report: {
                    status: 'done',
                    handoverDocs: [
                        handover({
                            task: 'Say so in the README',
                            tldr: 'I rewrote the part of the project handbook that covers talking to outside tools so it now explains the new waiting time and what happens when a tool stops answering.',
                            outcome:
                                'The proxy section now names the config key and says what happens when a server hangs.',
                            changedBehaviors: {
                                readme: 'The paragraph on forwarding describes the deadline.',
                            },
                        }),
                    ],
                },
            },
        ],
    },
    {
        id: '0198e3a1-0000-7000-8000-000000000003',
        name: 'WAL replay spike',
        createdAt: '2026-07-31T08:05:00Z',
        finalized: true,
        started: true,
        cancelled: false,
        workingDir: '/Users/rye/dev/agent-harness',
        contextDir: '/Users/rye/dev/agent-harness/.harness/context',
        tasks: [
            {
                id: 'task-9',
                title: 'Replay a truncated log',
                prompt: 'Prove the WAL can replay a log whose last write never landed.',
                state: 'done',
                position: {x: 0, y: 60},
                dependsOn: [],
                templateId: TEMPLATE_IDS.tests,
                values: {
                    package_path: 'internal/implementation/wal',
                    max_cases: 5,
                    focus: 'Replay stops at the torn tail and keeps every whole record before it.',
                },
                run: {
                    startedAt: '2026-07-31T08:06:12Z',
                    finishedAt: '2026-07-31T08:14:47Z',
                    context: {used: 88900, total: 200000},
                },
                report: {
                    status: 'done',
                    handoverDocs: [
                        handover({
                            task: 'Replay a truncated log',
                            tldr: 'I made the crash-recovery journal stop cleanly at a half-written entry at the end of the file instead of rejecting the whole file, so everything saved before the crash survives.',
                            outcome:
                                'Replay now stops at a torn tail instead of failing the whole log. Every whole record before it survives.',
                            approvedDecisions: {
                                tail_only:
                                    'A torn record is only ever the tail, so replay breaks out rather than skipping and continuing.',
                            },
                            changedBehaviors: {
                                replay: 'Replay returns the records it could read and logs the offset it stopped at.',
                            },
                            knownGaps: {
                                mid_log:
                                    'A torn record in the middle of the log is still treated as the tail. That is the case the next node has to break.',
                            },
                        }),
                    ],
                },
            },
            {
                id: 'task-10',
                title: 'Recover from a torn record mid-log',
                prompt: 'Handle a torn record that is not the tail. Recover the records after it.',
                state: 'failed',
                position: {x: 340, y: 60},
                dependsOn: ['task-9'],
                templateId: TEMPLATE_IDS.tests,
                values: {
                    package_path: 'internal/implementation/wal',
                    max_cases: 8,
                    focus: 'Records after a torn one are recovered, not dropped.',
                },
                run: {
                    startedAt: '2026-07-31T08:15:30Z',
                    finishedAt: '2026-07-31T08:29:04Z',
                    retryCount: 2,
                    context: {used: 191300, total: 200000},
                },
                report: {
                    status: 'failed',
                    handoverDocs: [
                        handover({
                            task: 'Recover from a torn record mid-log',
                            tldr: 'I tried to make crash recovery carry on past a damaged entry in the middle of the journal, but gave up after two attempts because the entries carry no marker to find where the next one starts.',
                            outcome:
                                'Stopped after two retries. Resynchronising mid-log needs a framing change the record format does not carry.',
                            blockers: {
                                no_sync_marker:
                                    'Records carry a length but no magic prefix, so there is nothing to scan forward to after a tear.',
                                context:
                                    'Ran out of context re-reading the encoder on the third attempt.',
                            },
                            currentBehaviors: {
                                replay: 'Replay still stops at the first torn record, tail or not.',
                            },
                            mustAvoid: {
                                guess_length:
                                    'Do not infer the next offset from the declared length of a torn record. It is the field that is wrong.',
                            },
                            knownGaps: {
                                format: 'The fix is a format change: a 4-byte marker before every record.',
                            },
                        }),
                    ],
                },
            },
            {
                id: 'task-11',
                title: 'Write up the recovery rules',
                prompt: 'Document what replay guarantees after a crash.',
                state: 'blocked',
                position: {x: 680, y: 60},
                dependsOn: ['task-10'],
                templateId: TEMPLATE_IDS.docs,
                values: {doc_path: 'docs/wal.md'},
            },
        ],
    },
    {
        id: '0198e3a1-0000-7000-8000-000000000005',
        name: 'Heartbeat recovery',
        createdAt: '2026-08-02T10:30:00Z',
        finalized: true,
        started: true,
        cancelled: true,
        workingDir: '/Users/rye/dev/agent-harness',
        contextDir: '/Users/rye/dev/agent-harness/.harness/context',
        tasks: [
            {
                id: 'task-16',
                title: 'Trace the heartbeat path',
                prompt: 'Follow a heartbeat from the agent to the session manager and say what a missed one does today.',
                state: 'done',
                position: {x: 0, y: 60},
                dependsOn: [],
                templateId: TEMPLATE_IDS.reviewer,
                values: {},
                run: {
                    startedAt: '2026-08-02T10:31:04Z',
                    finishedAt: '2026-08-02T10:38:22Z',
                    context: {used: 46700, total: 200000},
                },
                report: {
                    status: 'done',
                    handoverDocs: [
                        handover({
                            task: 'Trace the heartbeat path',
                            tldr: 'I followed how a working assistant checks in to say it is still alive, and found that when it stops checking in it gets shut down but its step is left looking like it is still running.',
                            outcome:
                                'Read-only pass. A missed heartbeat kills the agent process and leaves the node in running forever.',
                            currentBehaviors: {
                                sweep: 'The heartbeat ticker cancels the agent context but never writes a terminal state.',
                                restart:
                                    'On restart the node is still marked running, so nothing picks it back up.',
                            },
                            mustAvoid: {
                                kill_on_first_miss:
                                    'One missed tick is normal under load. Only a run of them means the agent is gone.',
                            },
                            knownGaps: {
                                terminal_state:
                                    'Nothing marks the node failed when its agent disappears. The next node adds that.',
                            },
                        }),
                    ],
                },
            },
            {
                id: 'task-17',
                title: 'Mark the node failed on a lost agent',
                prompt: 'When the heartbeat sweep kills an agent, write the terminal state so the coordinator can move on.',
                state: 'cancelled',
                position: {x: 340, y: 60},
                dependsOn: ['task-16'],
                templateId: TEMPLATE_IDS.tests,
                values: {
                    package_path: 'internal/implementation/session_manager',
                    max_cases: 6,
                    focus: 'A node whose agent stops answering ends up failed, not running.',
                },
                run: {startedAt: '2026-08-02T10:39:10Z', finishedAt: '2026-08-02T10:44:51Z'},
            },
            {
                id: 'task-18',
                title: 'Write up the heartbeat contract',
                prompt: 'Document how many missed ticks kill an agent and what the node looks like afterwards.',
                state: 'cancelled',
                position: {x: 680, y: 60},
                dependsOn: ['task-17'],
                templateId: TEMPLATE_IDS.docs,
                values: {doc_path: 'docs/heartbeats.md'},
            },
        ],
    },
]

const ARCHIVE_VERSION = 1
const SESSION_FILE_INVALID = 'err_session_file_invalid'

export function mockSessionArchive(body: string): string {
    const archive = {
        version: ARCHIVE_VERSION,
        exported_at: new Date().toISOString(),
        session: JSON.parse(body),
    }

    return `${JSON.stringify(archive, null, 2)}\n`
}

export function mockSessionBody(file: string): string {
    let archive: {version?: number; session?: unknown}

    try {
        archive = JSON.parse(file)
    } catch {
        throw new Error(
            `[Err] Type: ${SESSION_FILE_INVALID} - Message: that file is not a session file - Critical: critical`,
        )
    }

    if (archive?.version !== ARCHIVE_VERSION)
        throw new Error(
            `[Err] Type: ${SESSION_FILE_INVALID} - Message: that file is a version ${archive?.version} session file, and this app reads version ${ARCHIVE_VERSION} - Critical: critical`,
        )

    if (!archive.session)
        throw new Error(
            `[Err] Type: ${SESSION_FILE_INVALID} - Message: that file holds no session - Critical: critical`,
        )

    return JSON.stringify(archive.session)
}

function seedOf(id: string) {
    let seed = 0
    for (let index = 0; index < id.length; index += 1)
        seed = (seed * 31 + id.charCodeAt(index)) % 99991
    return seed
}

type MockOutcome = {
    state: TaskState
    durationMs: number
    contextTotal: number
    contextPeak: number
    report: TaskReport
}

export function mockOutcome(task: Task): MockOutcome {
    const seed = seedOf(task.id)
    const failed = seed % 5 === 4

    const doc = failed
        ? handover({
              task: task.title,
              tldr: 'I stopped before finishing this step, because doing it properly meant editing a file this step was not allowed to touch.',
              outcome:
                  'Stopped before finishing. The agent could not satisfy the scope it was given.',
              blockers: {
                  scope: 'The change needed a file outside the paths this node was allowed to write.',
              },
              currentBehaviors: {
                  unchanged: 'Nothing was merged. The tree is as the upstream node left it.',
              },
              knownGaps: {
                  retry: 'Duplicate the session and narrow the prompt before running it again.',
              },
          })
        : handover({
              task: task.title,
              tldr: 'I made the change this step asked for, and left the working directory with the files it names.',
              outcome: 'Finished. Open Files changed to read what landed on disk.',
              approvedDecisions: {
                  scope: 'Stayed inside the paths the prompt named.',
              },
              changedBehaviors: {
                  summary: task.prompt.slice(0, 160),
              },
              nuances: {
                  followup: 'The next node inherits this doc as the head of its prompt.',
              },
          })

    return {
        state: failed ? 'failed' : 'done',
        durationMs: 9000 + (seed % 7) * 1600,
        contextTotal: 200000,
        contextPeak: 0.42 + (seed % 47) / 100,
        report: {
            status: failed ? 'failed' : 'done',
            handoverDocs: [doc],
        },
    }
}
