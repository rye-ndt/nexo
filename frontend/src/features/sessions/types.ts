import type {FileChangeType, TaskLevel, TaskState} from '@/shared/lib/enums'
import type {ParamValue} from '@/features/templates/types'

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

export type HandoverDoc = {
    task: string
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

/** One short sentence a running node emitted about what it is doing. */
export type ActivityLine = {
    seq: number
    at: string
    text: string
}

export type TaskReport = {
    status: TaskState
    handoverDocs: HandoverDoc[]
}

/**
 * Every token this node has cost, all of its attempts summed. Input is prompt read
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
    /** What `spent` comes to at the prices of the model this node ran on. */
    costUsd?: number
    priced?: boolean
}

/** What a template would have told the run, carried by a node that no longer has one. */
export type TaskSpec = {
    taskLevel: TaskLevel
    systemPrompts: string[]
    outputStructure: string
    manualAcceptRequired: boolean
}

export type Task = {
    id: string
    title: string
    prompt: string
    state: TaskState
    position: Point
    dependsOn: string[]
    templateId?: string
    spec?: TaskSpec
    values?: Record<string, ParamValue>
    agentId?: string
    run?: Run
    report?: TaskReport
}

export type TaskDraft = {
    title: string
    prompt: string
    templateId: string
    values: Record<string, ParamValue>
}

export type RemoteRunIds = {
    sessionId: string
    taskIds: Record<string, string>
}

export type Session = {
    id: string
    name: string
    createdAt: string
    finalized: boolean
    started: boolean
    cancelled: boolean
    paused: boolean
    workingDir: string
    contextDir: string
    tasks: Task[]
    remote?: RemoteRunIds
    /** What every node has spent, summed — the nodes and the total always agree. */
    spent?: Spend
    costUsd?: number
    /** False as soon as one node that spent something ran on a model with no prices. */
    priced?: boolean
    /** When the first node was assigned, and when the last one settled. */
    startedAt?: string
    finishedAt?: string
}

export type SessionLocations = {
    workingDir: string
    contextDir: string
}

export type SessionDraft = {
    name: string
    workingDir: string
    contextDir: string
}
