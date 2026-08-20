import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {workflowProgress} from '@/features/workflows/graph'
import type {Workflow} from '@/features/workflows/types'

function stakes(workflow: Workflow) {
    const {done, total} = workflowProgress(workflow)

    if (total === 0) return 'It has no steps yet.'
    if (done === 0) return `Its ${total} ${total === 1 ? 'step goes' : 'steps go'} with it.`

    return `Its ${total} steps go with it, including ${done} finished and their results.`
}

export function DeleteWorkflowDialog({
    workflow,
    onConfirm,
    onClose,
}: {
    workflow: Workflow
    onConfirm: () => void
    onClose: () => void
}) {
    return (
        <ConfirmDialog
            title={`Delete “${workflow.name}”?`}
            description={`${stakes(workflow)} This cannot be undone.`}
            confirmLabel="Delete workflow"
            destructive
            onConfirm={onConfirm}
            onClose={onClose}
        />
    )
}
