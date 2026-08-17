export const TaskState = {
    Idle: 'idle',
    Blocked: 'blocked',
    Queued: 'queued',
    Running: 'running',
    AwaitingApproval: 'awaiting_approval',
    AwaitingAccept: 'awaiting_accept',
    Done: 'done',
    Failed: 'failed',
    Cancelled: 'cancelled',
} as const

export type TaskState = (typeof TaskState)[keyof typeof TaskState]

export const TASK_STATE_LABELS: Record<TaskState, string> = {
    [TaskState.Idle]: 'Idle',
    [TaskState.Blocked]: 'Blocked',
    [TaskState.Queued]: 'Ready',
    [TaskState.Running]: 'Running',
    [TaskState.AwaitingApproval]: 'Needs you',
    [TaskState.AwaitingAccept]: 'Needs review',
    [TaskState.Done]: 'Done',
    [TaskState.Failed]: 'Failed',
    [TaskState.Cancelled]: 'Cancelled',
}

export const SessionStatus = {
    Empty: 'empty',
    Draft: 'draft',
    Ready: 'ready',
    Running: 'running',
    Done: 'done',
    Failed: 'failed',
    Cancelled: 'cancelled',
} as const

export type SessionStatus = (typeof SessionStatus)[keyof typeof SessionStatus]

export const SESSION_STATUS_LABELS: Record<SessionStatus, string> = {
    [SessionStatus.Empty]: 'Empty',
    [SessionStatus.Draft]: 'Draft',
    [SessionStatus.Ready]: 'Ready to run',
    [SessionStatus.Running]: 'Running',
    [SessionStatus.Done]: 'Done',
    [SessionStatus.Failed]: 'Failed',
    [SessionStatus.Cancelled]: 'Cancelled',
}

export const FileChangeType = {
    Created: 'created',
    Modified: 'modified',
    Deleted: 'deleted',
    Renamed: 'renamed',
} as const

export type FileChangeType = (typeof FileChangeType)[keyof typeof FileChangeType]

export const TaskLevel = {
    Lightweight: 'lightweight_task',
    Daily: 'daily_task',
    Heavy: 'heavy_task',
    MaximumEffort: 'maximum_effort_task',
} as const

export type TaskLevel = (typeof TaskLevel)[keyof typeof TaskLevel]

export const TASK_LEVELS: TaskLevel[] = [
    TaskLevel.Lightweight,
    TaskLevel.Daily,
    TaskLevel.Heavy,
    TaskLevel.MaximumEffort,
]

export const TASK_LEVEL_LABELS: Record<TaskLevel, string> = {
    [TaskLevel.Lightweight]: 'Lightweight',
    [TaskLevel.Daily]: 'Daily',
    [TaskLevel.Heavy]: 'Heavy',
    [TaskLevel.MaximumEffort]: 'Maximum effort',
}

export const ThinkingLevel = {
    Low: 'low',
    Medium: 'medium',
    High: 'high',
    XHigh: 'xhigh',
    Max: 'max',
} as const

export type ThinkingLevel = (typeof ThinkingLevel)[keyof typeof ThinkingLevel]

export const THINKING_LEVELS: ThinkingLevel[] = [
    ThinkingLevel.Low,
    ThinkingLevel.Medium,
    ThinkingLevel.High,
    ThinkingLevel.XHigh,
    ThinkingLevel.Max,
]

export const THINKING_LEVEL_LABELS: Record<ThinkingLevel, string> = {
    [ThinkingLevel.Low]: 'Low',
    [ThinkingLevel.Medium]: 'Medium',
    [ThinkingLevel.High]: 'High',
    [ThinkingLevel.XHigh]: 'Very high',
    [ThinkingLevel.Max]: 'Maximum',
}

export const InstallStage = {
    Queued: 'queued',
    Resolve: 'resolve',
    Download: 'download',
    Extract: 'extract',
    Done: 'done',
} as const

export type InstallStage = (typeof InstallStage)[keyof typeof InstallStage]

export const INSTALL_STAGE_LABELS: Record<InstallStage, string> = {
    [InstallStage.Queued]: 'Waiting',
    [InstallStage.Resolve]: 'Finding release',
    [InstallStage.Download]: 'Downloading',
    [InstallStage.Extract]: 'Unpacking',
    [InstallStage.Done]: 'Ready',
}

export const ApprovalKind = {
    Decision: 'decision',
    Permission: 'permission',
} as const

export type ApprovalKind = (typeof ApprovalKind)[keyof typeof ApprovalKind]

export const MCPAuthKind = {
    DynamicRegistration: 'dcr',
    Device: 'device',
    Token: 'token',
    Enable: 'enable',
} as const

export type MCPAuthKind = (typeof MCPAuthKind)[keyof typeof MCPAuthKind]

export const AgentAction = {
    Install: 'install',
    Uninstall: 'uninstall',
    LogIn: 'log_in',
    Verify: 'verify',
} as const

export type AgentAction = (typeof AgentAction)[keyof typeof AgentAction]

export const AGENT_ACTION_LABELS: Record<AgentAction, string> = {
    [AgentAction.Install]: 'Install',
    [AgentAction.Uninstall]: 'Uninstall',
    [AgentAction.LogIn]: 'Log in',
    [AgentAction.Verify]: 'Verify',
}

export const AGENT_ACTION_BUSY_LABELS: Record<AgentAction, string> = {
    [AgentAction.Install]: 'Installing',
    [AgentAction.Uninstall]: 'Uninstalling',
    [AgentAction.LogIn]: 'Logging in',
    [AgentAction.Verify]: 'Verifying',
}

/** Completed by the agent's name to title the error dialog when the action fails. */
export const AGENT_ACTION_FAILURES: Record<AgentAction, string> = {
    [AgentAction.Install]: 'Could not install',
    [AgentAction.Uninstall]: 'Could not uninstall',
    [AgentAction.LogIn]: 'Could not start the login for',
    [AgentAction.Verify]: 'Could not verify that code for',
}

export const ParamType = {
    Text: 'text',
    Textarea: 'textarea',
    Number: 'number',
    Boolean: 'boolean',
    Select: 'select',
    File: 'file',
} as const

export type ParamType = (typeof ParamType)[keyof typeof ParamType]

export const PARAM_TYPES: ParamType[] = [
    ParamType.Text,
    ParamType.Textarea,
    ParamType.Number,
    ParamType.Boolean,
    ParamType.Select,
    ParamType.File,
]

export const PARAM_TYPE_LABELS: Record<ParamType, string> = {
    [ParamType.Text]: 'Text',
    [ParamType.Textarea]: 'Long text',
    [ParamType.Number]: 'Number',
    [ParamType.Boolean]: 'Boolean',
    [ParamType.Select]: 'Choice',
    [ParamType.File]: 'File',
}
