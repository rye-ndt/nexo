import {StatusChip} from '@/shared/components/status-chip'
import {WORKFLOW_STATUS_LABELS} from '@/shared/lib/enums'
import {WORKFLOW_CHIP_TONES} from '@/features/workflows/workflow-status'
import {workflowProgress, workflowStatus} from '@/features/workflows/graph'
import type {Workflow} from '@/features/workflows/types'

export function WorkflowIdentity({workflow}: {workflow: Workflow}) {
    const status = workflowStatus(workflow)
    const {done, total} = workflowProgress(workflow)

    return (
        <span className="flex shrink-0 items-center gap-3">
            <StatusChip tone={WORKFLOW_CHIP_TONES[status]}>
                {WORKFLOW_STATUS_LABELS[status]}
            </StatusChip>

            <span className="hidden font-mono text-sm text-muted-foreground lg:block">
                {done}/{total}
            </span>
        </span>
    )
}
