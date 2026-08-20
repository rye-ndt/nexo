import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {workflowProgress} from '@/features/workflows/graph'
import {t, tn} from '@/shared/lib/i18n'
import type {Workflow} from '@/features/workflows/types'

function stakes(workflow: Workflow) {
    const {done, total} = workflowProgress(workflow)

    if (total === 0) return t('workflow.delete.noSteps')
    if (done === 0) return tn('workflow.delete.steps.one', 'workflow.delete.steps.other', total)

    return t('workflow.delete.finished', {total, done})
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
            title={t('workflow.delete.title', {name: workflow.name})}
            description={stakes(workflow)}
            confirmLabel={t('workflow.delete.confirm')}
            destructive
            onConfirm={onConfirm}
            onClose={onClose}
        />
    )
}
