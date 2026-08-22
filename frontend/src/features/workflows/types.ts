import type {FileChangeType, Effort, StepState} from '@/shared/lib/enums'
import type {InputValue} from '@/features/roles/types'

export type Point = {
    x: number
    y: number
}

export type ContextUsage = {
    used: number
    total: number
}

export type FileChange = {
    path: string
    oldPath: string
    changeType: FileChangeType
    additions: number
    deletions: number
    unifiedDiff: string
}

export type Handoff = {
    step: string
    tldr: string
    outcome: string
    blockers: Record<string, string>
    approvedDecisions: Record<string, string>
    rejectedDecisions: Record<string, string>
    currentBehaviors: Record<string, string>
    changedBehaviors: Record<string, string>
    mustAvoid: Record<string, string>
    nuances: Record<string, string>
    knownGaps: Record<string, string>
}

/** One short sentence a running step emitted about what it is doing. */
export type ActivityLine = {
    seq: number
    at: string
    text: string
}

export type StepResult = {
    status: StepState
    handoffs: Handoff[]
}

/**
 * Every token this step has cost, all of its attempts summed. Input is prompt read
 * at full price, cached is prompt read back out of the vendor's cache at a fraction
 * of it, output is what the agent wrote.
 */
export type Spend = {
    input: number
    cached: number
    output: number
}

export type Run = {
    startedAt?: string
    finishedAt?: string
    retryCount?: number
    context?: ContextUsage
    spent?: Spend
    /** What `spent` comes to at the prices of the model this step ran on. */
    costUsd?: number
    priced?: boolean
}

/** What a role would have told the run, carried by a step that no longer has one. */
export type StepSpec = {
    effort: Effort
    instructions: string[]
    outputStructure: string
    pauseForReview: boolean
}

export type Step = {
    id: string
    title: string
    prompt: string
    state: StepState
    position: Point
    dependsOn: string[]
    roleId?: string
    spec?: StepSpec
    values?: Record<string, InputValue>
    agentId?: string
    run?: Run
    report?: StepResult
}

export type StepDraft = {
    title: string
    prompt: string
    roleId: string
    values: Record<string, InputValue>
}

export type RemoteRunIds = {
    workflowId: string
    stepIds: Record<string, string>
}

export type PastRun = {
    runId: string
    remote?: RemoteRunIds
    startedAt?: string
    finishedAt?: string
    spent?: Spend
    costUsd?: number
    priced?: boolean
    steps: {
        id: string
        title: string
        state: StepState
        agentId?: string
        run?: Run
        report?: StepResult
    }[]
}

export type Workflow = {
    id: string
    name: string
    createdAt: string
    railRank?: number
    locked: boolean
    started: boolean
    cancelled: boolean
    paused: boolean
    projectDir: string
    steps: Step[]
    remote?: RemoteRunIds
    /** What every step has spent, summed — the steps and the total always agree. */
    spent?: Spend
    costUsd?: number
    /** False as soon as one step that spent something ran on a model with no prices. */
    priced?: boolean
    /** When the first step was assigned, and when the last one settled. */
    startedAt?: string
    finishedAt?: string
    history?: PastRun[]
}

export type WorkflowLocations = {
    projectDir: string
}

export type WorkflowDraft = {
    name: string
    projectDir: string
}
