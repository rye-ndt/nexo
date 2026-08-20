/**
 * Outside the webview there is no agent to narrate itself, so a simulated run
 * writes its own feed: a fixed script of plausible steps, picked and paced from
 * the step id so the same step always says the same things in the same order.
 * A step the mock is going to fail spends its last lines struggling.
 */

import {StepState} from '@/shared/lib/enums'
import {mockOutcome} from '@/features/workflows/mock-workflows'
import type {ActivityLine, Step} from '@/features/workflows/types'

const LINE_GAP_MS = 1200

const MIN_LINES = 8

const MAX_LINES = 15

const FILES = [
    'internal/interface/core/workflow_manager.go',
    'internal/implementation/mcp_proxy/v1.go',
    'internal/implementation/wal/replay.go',
    'internal/implementation/agent_manager/v1.go',
    'internal/implementation/claude_code/v1.go',
    'internal/implementation/workflow_manager/v1.go',
    'internal/helpers/custom_error/errors.go',
    'frontend/src/features/workflows/graph.ts',
    'wire.go',
]

const SEARCHES = [
    'where the coordinator reads context usage',
    'where the heartbeat sweep writes a terminal state',
    'the places a Handoff is composed',
    'anything that still imports step_manager',
    'the config key the proxy reads its deadline from',
    'goroutines started without an exit path',
]

const CLOSING = [
    'Running go test ./...',
    'Tests pass.',
    'Writing up what the next step needs to know.',
]

const STRUGGLING = [
    'Running go test ./...',
    'One test still fails. Trying a narrower change.',
    'Still failing. The real fix is in a file this step cannot touch.',
]

function seedOf(id: string) {
    let seed = 0
    for (let index = 0; index < id.length; index += 1)
        seed = (seed * 31 + id.charCodeAt(index)) % 99991
    return seed
}

function spineOf(step: Step): string[] {
    const seed = seedOf(step.id)
    const files = Array.from({length: 4}, (_, index) => FILES[(seed + index * 2) % FILES.length])

    return [
        `Reading the prompt for "${step.title}".`,
        `Reading ${files[0]}.`,
        `Searching for ${SEARCHES[seed % SEARCHES.length]}.`,
        `Reading ${files[1]}.`,
        `Editing ${files[1]}.`,
        'Running go build ./...',
        `Reading ${files[2]}.`,
        `Editing ${files[2]}.`,
        `Reading ${files[3]}.`,
        `Editing ${files[3]}.`,
        'Re-reading the diff so far.',
        'Running gofmt on the files I touched.',
    ]
}

function lineCount(durationMs: number) {
    return Math.min(MAX_LINES, Math.max(MIN_LINES, Math.round(durationMs / LINE_GAP_MS)))
}

/** Every line the step should have said by `now`, oldest first. */
export function mockActivity(step: Step, now: number): ActivityLine[] {
    const outcome = mockOutcome(step)
    const tail = outcome.state === StepState.Failed ? STRUGGLING : CLOSING
    const count = lineCount(outcome.durationMs)
    const texts = [...spineOf(step).slice(0, count - tail.length), ...tail]

    const startedAt = step.run?.startedAt ? Date.parse(step.run.startedAt) : now
    const gap = outcome.durationMs / count
    const said = Math.min(texts.length, Math.max(1, Math.floor((now - startedAt) / gap) + 1))

    return Array.from({length: said}, (_, index) => ({
        seq: index + 1,
        at: new Date(startedAt + index * gap).toISOString(),
        text: texts[index],
    }))
}
