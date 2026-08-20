import {useMemo, useState} from 'react'

import {DialogShell} from '@/shared/components/dialog-shell'
import {MissingInputs} from '@/features/workflows/components/steps/missing-inputs'
import {InputFields} from '@/features/roles/components/input-fields'
import {PromptPreview} from '@/features/roles/components/prompt-preview'
import {Button} from '@/shared/ui/button'
import {useRole} from '@/features/roles/use-roles'
import {toFieldValues, toInputValues} from '@/features/roles/role'
import {inputRefs} from '@/features/roles/input-refs'
import {pendingInputs} from '@/features/workflows/step-inputs'
import type {Step} from '@/features/workflows/types'
import {t} from '@/shared/lib/i18n'
import type {FieldValue, InputValue} from '@/features/roles/types'

export function StepInputsDialog({
    step,
    busy,
    onSave,
    onClose,
}: {
    step: Step
    busy: boolean
    onSave: (values: Record<string, InputValue>) => void
    onClose: () => void
}) {
    const role = useRole(step.roleId)

    const [edits, setEdits] = useState<Record<string, FieldValue>>({})

    const stored = useMemo(
        () => (role ? toFieldValues(role, step.values) : {}),
        [role, step.values],
    )

    const values = {...stored, ...edits}
    const pending = role ? pendingInputs({...step, values: toInputValues(role, values)}, role) : []
    const embedded = inputRefs(step.prompt).length > 0

    const editValue = (key: string, value: FieldValue) =>
        setEdits((current) => ({...current, [key]: value}))

    const save = () => {
        if (!role || busy) return

        onSave(toInputValues(role, values))
        onClose()
    }

    return (
        <DialogShell
            onClose={onClose}
            title={step.title || t('step.untitled')}
            description={t('step.locked.description')}
            footer={
                <>
                    <span className="flex-1" />
                    <MissingInputs count={pending.length} />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        {t('step.locked.cancel')}
                    </Button>
                    <Button size="sm" disabled={busy} onClick={save}>
                        {busy ? t('step.locked.saving') : t('step.locked.save')}
                    </Button>
                </>
            }
        >
            <div className="flex flex-col gap-6 p-4">
                {role ? (
                    <InputFields inputs={role.inputs} values={values} onChange={editValue} />
                ) : (
                    <p className="text-base text-muted-foreground">{t('step.locked.roleGone')}</p>
                )}

                {embedded && (
                    <section className="flex flex-col gap-3">
                        <span className="micro-label">{t('step.locked.promptPreview')}</span>
                        <PromptPreview prompt={step.prompt} values={values} />
                    </section>
                )}
            </div>
        </DialogShell>
    )
}
