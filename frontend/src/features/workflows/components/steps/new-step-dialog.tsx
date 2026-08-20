import {useState} from 'react'
import {ArrowLeft} from 'lucide-react'

import {DialogShell} from '@/shared/components/dialog-shell'
import {StepSpine} from '@/shared/components/step-spine'
import {InheritedAgent} from '@/features/workflows/components/steps/inherited-agent'
import {MissingInputsNote} from '@/features/workflows/components/steps/missing-inputs'
import {StepForm} from '@/features/workflows/components/steps/step-form'
import {RolesPanel} from '@/features/roles/components/roles-panel'
import {Button} from '@/shared/ui/button'
import {useRoles} from '@/features/roles/use-roles'
import {
    defaultFieldValues,
    missingRequired,
    promptFromRole,
    toInputValues,
} from '@/features/roles/role'
import type {StepDraft} from '@/features/workflows/types'
import type {DraftContext, FieldValue, Role} from '@/features/roles/types'

const STEP_COUNT = 2

export function NewStepDialog({
    context,
    onCreate,
    onClose,
}: {
    context: DraftContext
    onCreate: (draft: StepDraft) => void
    onClose: () => void
}) {
    const {roles} = useRoles()

    const [chosenId, setChosenId] = useState<string | null>(null)
    const [title, setTitle] = useState('')
    const [prompt, setPrompt] = useState('')
    const [values, setValues] = useState<Record<string, FieldValue>>({})

    const chosen = roles.find((role) => role.id === chosenId)
    const missing = chosen ? missingRequired(chosen, values) : []
    const ready = Boolean(chosen) && title.trim().length > 0

    const choose = (role: Role) => {
        setChosenId(role.id)
        setTitle(role.name)
        setPrompt(promptFromRole(role))
        setValues(defaultFieldValues(role))
    }

    const back = () => setChosenId(null)

    const changeValue = (key: string, value: FieldValue) =>
        setValues((current) => ({...current, [key]: value}))

    const create = () => {
        if (!chosen || !ready) return

        onCreate({
            title: title.trim(),
            prompt: prompt.trim(),
            roleId: chosen.id,
            values: toInputValues(chosen, values),
        })
    }

    return (
        <DialogShell
            onClose={onClose}
            title={chosen ? chosen.name : 'New step'}
            term={chosen ? undefined : 'step'}
            aside={<StepSpine total={STEP_COUNT} current={chosen ? 1 : 0} />}
            footer={
                chosen ? (
                    <FillFooter
                        missingCount={missing.length}
                        ready={ready}
                        onBack={back}
                        onCreate={create}
                    />
                ) : (
                    <PickFooter onCancel={onClose} />
                )
            }
        >
            {chosen ? (
                <>
                    <InheritedAgent effort={chosen.effort} fromRole />
                    <StepForm
                        key={chosen.id}
                        inputs={chosen.inputs}
                        title={title}
                        prompt={prompt}
                        values={values}
                        onTitleChange={setTitle}
                        onPromptChange={setPrompt}
                        onValueChange={changeValue}
                    />
                </>
            ) : (
                <RolesPanel context={context} onPick={choose} />
            )}
        </DialogShell>
    )
}

function FillFooter({
    missingCount,
    ready,
    onBack,
    onCreate,
}: {
    missingCount: number
    ready: boolean
    onBack: () => void
    onCreate: () => void
}) {
    return (
        <>
            <Button variant="ghost" size="sm" onClick={onBack}>
                <ArrowLeft />
                Roles
            </Button>
            <span className="flex-1" />
            <MissingInputsNote count={missingCount} />
            <Button size="sm" disabled={!ready} onClick={onCreate}>
                Create step
            </Button>
        </>
    )
}

function PickFooter({onCancel}: {onCancel: () => void}) {
    return (
        <>
            <span className="flex-1" />
            <Button variant="outline" size="sm" onClick={onCancel}>
                Cancel
            </Button>
        </>
    )
}
