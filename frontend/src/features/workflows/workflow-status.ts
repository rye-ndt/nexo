import type {ChipTone} from '@/shared/components/status-chip'
import {WorkflowStatus} from '@/shared/lib/enums'

export const WORKFLOW_CHIP_TONES: Record<WorkflowStatus, ChipTone> = {
    [WorkflowStatus.Empty]: 'muted',
    [WorkflowStatus.Draft]: 'muted',
    [WorkflowStatus.Ready]: 'info',
    [WorkflowStatus.Running]: 'running',
    [WorkflowStatus.Paused]: 'stopped',
    [WorkflowStatus.Done]: 'done',
    [WorkflowStatus.Failed]: 'failed',
    [WorkflowStatus.Cancelled]: 'stopped',
}

export const WORKFLOW_TITLE_CLASSES: Record<WorkflowStatus, string> = {
    [WorkflowStatus.Empty]: 'text-muted-foreground',
    [WorkflowStatus.Draft]: 'text-state-approval',
    [WorkflowStatus.Ready]: 'text-state-approval',
    [WorkflowStatus.Running]: 'text-info',
    [WorkflowStatus.Paused]: 'text-state-approval',
    [WorkflowStatus.Done]: 'text-state-done',
    [WorkflowStatus.Failed]: 'text-state-failed',
    [WorkflowStatus.Cancelled]: 'text-muted-foreground',
}
