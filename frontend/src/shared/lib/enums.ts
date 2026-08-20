export const StepState = {
    Idle: 'idle',
    Blocked: 'blocked',
    Queued: 'queued',
    Running: 'running',
    AwaitingApproval: 'awaiting_approval',
    AwaitingReview: 'awaiting_review',
    Done: 'done',
    Failed: 'failed',
    Cancelled: 'cancelled',
} as const

export type StepState = (typeof StepState)[keyof typeof StepState]

export const STEP_STATE_LABELS: Record<StepState, string> = {
    [StepState.Idle]: 'Idle',
    [StepState.Blocked]: 'Blocked',
    [StepState.Queued]: 'Ready',
    [StepState.Running]: 'Running',
    [StepState.AwaitingApproval]: 'Needs approval',
    [StepState.AwaitingReview]: 'Needs review',
    [StepState.Done]: 'Done',
    [StepState.Failed]: 'Failed',
    [StepState.Cancelled]: 'Cancelled',
}

export const WorkflowStatus = {
    Empty: 'empty',
    Draft: 'draft',
    Ready: 'ready',
    Running: 'running',
    Paused: 'paused',
    Done: 'done',
    Failed: 'failed',
    Cancelled: 'cancelled',
} as const

export type WorkflowStatus = (typeof WorkflowStatus)[keyof typeof WorkflowStatus]

export const WORKFLOW_STATUS_LABELS: Record<WorkflowStatus, string> = {
    [WorkflowStatus.Empty]: 'Empty',
    [WorkflowStatus.Draft]: 'Draft',
    [WorkflowStatus.Ready]: 'Ready to run',
    [WorkflowStatus.Running]: 'Running',
    [WorkflowStatus.Paused]: 'Paused',
    [WorkflowStatus.Done]: 'Done',
    [WorkflowStatus.Failed]: 'Failed',
    [WorkflowStatus.Cancelled]: 'Cancelled',
}

export const FileChangeType = {
    Created: 'created',
    Modified: 'modified',
    Deleted: 'deleted',
    Renamed: 'renamed',
} as const

export type FileChangeType = (typeof FileChangeType)[keyof typeof FileChangeType]

export const Effort = {
    Quick: 'quick',
    Standard: 'standard',
    Deep: 'deep',
    Exhaustive: 'exhaustive',
} as const

export type Effort = (typeof Effort)[keyof typeof Effort]

export const EFFORTS: Effort[] = [Effort.Quick, Effort.Standard, Effort.Deep, Effort.Exhaustive]

export const EFFORT_LABELS: Record<Effort, string> = {
    [Effort.Quick]: 'Quick',
    [Effort.Standard]: 'Standard',
    [Effort.Deep]: 'Deep',
    [Effort.Exhaustive]: 'Exhaustive',
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
    LogOut: 'log_out',
    Verify: 'verify',
} as const

export type AgentAction = (typeof AgentAction)[keyof typeof AgentAction]

export const AGENT_ACTION_LABELS: Record<AgentAction, string> = {
    [AgentAction.Install]: 'Install',
    [AgentAction.Uninstall]: 'Uninstall',
    [AgentAction.LogIn]: 'Log in',
    [AgentAction.LogOut]: 'Log out',
    [AgentAction.Verify]: 'Verify',
}

export const AGENT_ACTION_BUSY_LABELS: Record<AgentAction, string> = {
    [AgentAction.Install]: 'Installing',
    [AgentAction.Uninstall]: 'Uninstalling',
    [AgentAction.LogIn]: 'Logging in',
    [AgentAction.LogOut]: 'Logging out',
    [AgentAction.Verify]: 'Verifying',
}

/** Completed by the agent's name to title the error dialog when the action fails. */
export const AGENT_ACTION_FAILURES: Record<AgentAction, string> = {
    [AgentAction.Install]: 'Could not install',
    [AgentAction.Uninstall]: 'Could not uninstall',
    [AgentAction.LogIn]: 'Could not start the login for',
    [AgentAction.LogOut]: 'Could not log out of',
    [AgentAction.Verify]: 'Could not verify that code for',
}

export const InputType = {
    Text: 'text',
    Textarea: 'textarea',
    Number: 'number',
    Boolean: 'boolean',
    Select: 'select',
    MultiSelect: 'multiselect',
    File: 'file',
} as const

export type InputType = (typeof InputType)[keyof typeof InputType]

export const INPUT_TYPES: InputType[] = [
    InputType.Text,
    InputType.Textarea,
    InputType.Number,
    InputType.Boolean,
    InputType.Select,
    InputType.MultiSelect,
    InputType.File,
]

export const INPUT_TYPE_LABELS: Record<InputType, string> = {
    [InputType.Text]: 'Text',
    [InputType.Textarea]: 'Long text',
    [InputType.Number]: 'Number',
    [InputType.Boolean]: 'Boolean',
    [InputType.Select]: 'Choice',
    [InputType.MultiSelect]: 'Multiple choices',
    [InputType.File]: 'File',
}
