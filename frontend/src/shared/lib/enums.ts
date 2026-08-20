import type {MessageKey} from '@/shared/lib/i18n'

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

export const STEP_STATE_LABELS: Record<StepState, MessageKey> = {
    [StepState.Idle]: 'enum.stepState.idle',
    [StepState.Blocked]: 'enum.stepState.blocked',
    [StepState.Queued]: 'enum.stepState.queued',
    [StepState.Running]: 'enum.stepState.running',
    [StepState.AwaitingApproval]: 'enum.stepState.awaitingApproval',
    [StepState.AwaitingReview]: 'enum.stepState.awaitingReview',
    [StepState.Done]: 'enum.stepState.done',
    [StepState.Failed]: 'enum.stepState.failed',
    [StepState.Cancelled]: 'enum.stepState.cancelled',
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

export const WORKFLOW_STATUS_LABELS: Record<WorkflowStatus, MessageKey> = {
    [WorkflowStatus.Empty]: 'enum.workflowStatus.empty',
    [WorkflowStatus.Draft]: 'enum.workflowStatus.draft',
    [WorkflowStatus.Ready]: 'enum.workflowStatus.ready',
    [WorkflowStatus.Running]: 'enum.workflowStatus.running',
    [WorkflowStatus.Paused]: 'enum.workflowStatus.paused',
    [WorkflowStatus.Done]: 'enum.workflowStatus.done',
    [WorkflowStatus.Failed]: 'enum.workflowStatus.failed',
    [WorkflowStatus.Cancelled]: 'enum.workflowStatus.cancelled',
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

export const EFFORT_LABELS: Record<Effort, MessageKey> = {
    [Effort.Quick]: 'enum.effort.quick',
    [Effort.Standard]: 'enum.effort.standard',
    [Effort.Deep]: 'enum.effort.deep',
    [Effort.Exhaustive]: 'enum.effort.exhaustive',
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

export const THINKING_LEVEL_LABELS: Record<ThinkingLevel, MessageKey> = {
    [ThinkingLevel.Low]: 'enum.thinking.low',
    [ThinkingLevel.Medium]: 'enum.thinking.medium',
    [ThinkingLevel.High]: 'enum.thinking.high',
    [ThinkingLevel.XHigh]: 'enum.thinking.xhigh',
    [ThinkingLevel.Max]: 'enum.thinking.max',
}

export const InstallStage = {
    Queued: 'queued',
    Resolve: 'resolve',
    Download: 'download',
    Extract: 'extract',
    Done: 'done',
} as const

export type InstallStage = (typeof InstallStage)[keyof typeof InstallStage]

export const INSTALL_STAGE_LABELS: Record<InstallStage, MessageKey> = {
    [InstallStage.Queued]: 'enum.installStage.queued',
    [InstallStage.Resolve]: 'enum.installStage.resolve',
    [InstallStage.Download]: 'enum.installStage.download',
    [InstallStage.Extract]: 'enum.installStage.extract',
    [InstallStage.Done]: 'enum.installStage.done',
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

export const AGENT_ACTION_LABELS: Record<AgentAction, MessageKey> = {
    [AgentAction.Install]: 'enum.agentAction.install',
    [AgentAction.Uninstall]: 'enum.agentAction.uninstall',
    [AgentAction.LogIn]: 'enum.agentAction.logIn',
    [AgentAction.LogOut]: 'enum.agentAction.logOut',
    [AgentAction.Verify]: 'enum.agentAction.verify',
}

export const AGENT_ACTION_BUSY_LABELS: Record<AgentAction, MessageKey> = {
    [AgentAction.Install]: 'enum.agentActionBusy.install',
    [AgentAction.Uninstall]: 'enum.agentActionBusy.uninstall',
    [AgentAction.LogIn]: 'enum.agentActionBusy.logIn',
    [AgentAction.LogOut]: 'enum.agentActionBusy.logOut',
    [AgentAction.Verify]: 'enum.agentActionBusy.verify',
}

/** Titles the error dialog when the action fails; takes the agent's name as {name}. */
export const AGENT_ACTION_FAILURES: Record<AgentAction, MessageKey> = {
    [AgentAction.Install]: 'enum.agentActionFailure.install',
    [AgentAction.Uninstall]: 'enum.agentActionFailure.uninstall',
    [AgentAction.LogIn]: 'enum.agentActionFailure.logIn',
    [AgentAction.LogOut]: 'enum.agentActionFailure.logOut',
    [AgentAction.Verify]: 'enum.agentActionFailure.verify',
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

export const INPUT_TYPE_LABELS: Record<InputType, MessageKey> = {
    [InputType.Text]: 'enum.inputType.text',
    [InputType.Textarea]: 'enum.inputType.textarea',
    [InputType.Number]: 'enum.inputType.number',
    [InputType.Boolean]: 'enum.inputType.boolean',
    [InputType.Select]: 'enum.inputType.select',
    [InputType.MultiSelect]: 'enum.inputType.multiselect',
    [InputType.File]: 'enum.inputType.file',
}

export const Language = {
    En: 'en',
    Vi: 'vi',
} as const

export type Language = (typeof Language)[keyof typeof Language]

export const LANGUAGES: Language[] = [Language.En, Language.Vi]

/** Endonyms: a language names itself the same way whichever language is active. */
export const LANGUAGE_NAMES: Record<Language, string> = {
    [Language.En]: 'English',
    [Language.Vi]: 'Tiếng Việt',
}
