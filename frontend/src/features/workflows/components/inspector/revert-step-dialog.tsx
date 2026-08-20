import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {descendantsOf} from '@/features/workflows/graph'
import type {Workflow, Step} from '@/features/workflows/types'

function undoneAfter(workflow: Workflow, step: Step) {
    const later = descendantsOf(workflow.steps, step.id)
    const started = workflow.steps.filter((other) => later.has(other.id) && other.run).length

    if (started === 0) return 'No step after it has run, so nothing else is undone.'
    if (started === 1) return 'The one step after it is undone and goes back to not started.'

    return `The ${started} steps after it are undone and go back to not started.`
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
            title={`Revert to “${step.title || 'Untitled step'}”?`}
            description={`Files in ${workflow.projectDir} go back to how this step left them, and anything written after it is lost. ${undoneAfter(workflow, step)} The run stops, and you start it again when you are ready.`}
            confirmLabel="Revert to this step"
            destructive
            busy={busy}
            onConfirm={onConfirm}
            onClose={onClose}
        />
    )
}
