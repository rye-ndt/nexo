import type {FileChangeType, TaskState} from '@/shared/lib/enums'
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

export type Run = {
    startedAt?: string
    finishedAt?: string
    retryCount?: number
    context?: ContextUsage
}

export type Task = {
    id: string
    title: string
    prompt: string
    state: TaskState
    position: Point
    dependsOn: string[]
    templateId?: string
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

export type Session = {
    id: string
    name: string
    createdAt: string
    finalized: boolean
    started: boolean
    cancelled: boolean
    workingDir: string
    contextDir: string
    tasks: Task[]
    /** Every token the session's nodes have spent, kept once a node is done with them. */
    tokensUsed?: number
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
