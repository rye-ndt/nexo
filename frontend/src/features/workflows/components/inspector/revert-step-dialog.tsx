import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {descendantsOf} from '@/features/workflows/graph'
import {t} from '@/shared/lib/i18n'
import type {Workflow, Step} from '@/features/workflows/types'

function describe(workflow: Workflow, step: Step) {
    const later = descendantsOf(workflow.steps, step.id)
    const started = workflow.steps.filter((other) => later.has(other.id) && other.run).length
    const dir = workflow.projectDir

    if (started === 0) return t('inspector.revert.none', {dir})
    if (started === 1) return t('inspector.revert.one', {dir})

    return t('inspector.revert.many', {dir, count: started})
}

export function RevertStepDialog({
    workflow,
    step,
    busy,
    onConfirm,
    onClose,
}: {
    workflow: Workflow
    step: Step
    busy: boolean
    onConfirm: () => void
    onClose: () => void
}) {
    return (
        <ConfirmDialog
            title={t('inspector.revert.title', {name: step.title || t('step.untitled')})}
            description={describe(workflow, step)}
            confirmLabel={t('inspector.revert.confirm')}
            destructive
            busy={busy}
            onConfirm={onConfirm}
            onClose={onClose}
        />
    )
}
