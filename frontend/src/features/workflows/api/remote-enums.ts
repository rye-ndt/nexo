/** The status strings the Go side puts on the wire, named so no caller spells one out. */

export const RemoteStepStatus = {
    Processing: 'processing',
    AwaitingReview: 'awaiting_review',
    Completed: 'completed',
    Failed: 'failed',
    Cancelled: 'cancelled',
} as const

export type RemoteStepStatus = (typeof RemoteStepStatus)[keyof typeof RemoteStepStatus]

export const RemoteChangeType = {
    Added: 'added',
    Modified: 'modified',
    Deleted: 'deleted',
    Renamed: 'renamed',
} as const

export type RemoteChangeType = (typeof RemoteChangeType)[keyof typeof RemoteChangeType]
