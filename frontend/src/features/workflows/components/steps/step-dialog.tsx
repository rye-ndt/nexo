import {StepStatusDialog} from '@/features/workflows/components/inspector/step-status-dialog'
import {EditStepDialog} from '@/features/workflows/components/steps/edit-step-dialog'
import {StepInputsDialog} from '@/features/workflows/components/steps/step-inputs-dialog'
import type {Workflow, Step} from '@/features/workflows/types'
import type {InputValue} from '@/features/roles/types'

export function StepDialog({
    workflow,
    step,
    savingInputs,
    reverting,
    onSave,
    onSaveInputs,
    onRevert,
    onDelete,
    onClose,
}: {
    workflow: Workflow
    step: Step
    savingInputs: boolean
    reverting: boolean
    onSave: (patch: Partial<Step>) => void
    onSaveInputs: (values: Record<string, InputValue>) => void
    onRevert: () => void
    onDelete: () => void
    onClose: () => void
}) {
    if (workflow.started)
        return (
            <StepStatusDialog
                workflow={workflow}
                step={step}
                reverting={reverting}
                onRevert={onRevert}
                onClose={onClose}
            />
        )

    if (workflow.locked)
        return (
            <StepInputsDialog
                step={step}
                busy={savingInputs}
                onSave={onSaveInputs}
                onClose={onClose}
            />
        )

    return <EditStepDialog step={step} onSave={onSave} onDelete={onDelete} onClose={onClose} />
}
