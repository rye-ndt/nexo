/** The status strings the Go side puts on the wire, named so no caller spells one out. */

export const RemoteTaskStatus = {
    Processing: 'processing',
    AwaitingAccept: 'awaiting_accept',
    Completed: 'completed',
    Failed: 'failed',
    Cancelled: 'cancelled',
} as const

export type RemoteTaskStatus = (typeof RemoteTaskStatus)[keyof typeof RemoteTaskStatus]

export const RemoteChangeType = {
    Added: 'added',
    Modified: 'modified',
    Deleted: 'deleted',
    Renamed: 'renamed',
} as const

export type RemoteChangeType = (typeof RemoteChangeType)[keyof typeof RemoteChangeType]
