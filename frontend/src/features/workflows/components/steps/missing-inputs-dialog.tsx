import {useState} from 'react'

import {DialogShell} from '@/shared/components/dialog-shell'
import {InputFields} from '@/features/roles/components/input-fields'
import {MissingInputsNote} from '@/features/workflows/components/steps/missing-inputs'
import {Button} from '@/shared/ui/button'
import {t, tn} from '@/shared/lib/i18n'
import {toFieldValues, toInputValues} from '@/features/roles/role'
import {pendingInputs} from '@/features/workflows/step-inputs'
import type {MissingStepInputs} from '@/features/workflows/step-inputs'
import type {FieldValue, InputValue} from '@/features/roles/types'

type StepEdits = Record<string, Record<string, FieldValue>>

export function MissingInputsDialog({
    entries,
    onRun,
    onClose,
}: {
    entries: MissingStepInputs[]
    onRun: (values: Record<string, Record<string, InputValue>>) => void
    onClose: () => void
}) {
    const [edits, setEdits] = useState<StepEdits>({})

    const filled = entries.map((entry) => {
        const values = {
            ...toFieldValues(entry.role, entry.step.values),
            ...edits[entry.step.id],
        }
        return {
            entry,
            values,
            stored: toInputValues({...entry.role, inputs: entry.inputs}, values),
            pending: pendingInputs({...entry.step, values}, entry.role),
        }
    })

    const remaining = filled.reduce((count, step) => count + step.pending.length, 0)

    const editValue = (stepId: string, key: string, value: FieldValue) =>
        setEdits((current) => ({...current, [stepId]: {...current[stepId], [key]: value}}))

    const run = () =>
        onRun(Object.fromEntries(filled.map((step) => [step.entry.step.id, step.stored])))

    return (
        <DialogShell
            onClose={onClose}
            title={t('step.missing.title')}
            description={tn('step.missing.asking.one', 'step.missing.asking.other', entries.length)}
            footer={
                <>
                    <MissingInputsNote count={remaining} />
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        {t('step.missing.cancel')}
                    </Button>
                    <Button size="sm" onClick={run}>
                        {t('step.missing.run')}
                    </Button>
                </>
            }
        >
            <div className="flex flex-col gap-6 p-4">
                {filled.map((step) => (
                    <section
                        key={step.entry.step.id}
                        className="flex flex-col gap-3 border-t border-border pt-6 first:border-t-0 first:pt-0"
                    >
                        <div className="flex items-baseline justify-between gap-3">
                            <h3 className="truncate text-base font-medium">
                                {step.entry.step.title || t('step.untitled')}
                            </h3>
                            <span className="micro-label shrink-0">
                                {tn(
                                    'step.missing.count.one',
                                    'step.missing.count.other',
                                    step.entry.inputs.length,
                                )}
                            </span>
                        </div>

                        <InputFields
                            inputs={step.entry.inputs}
                            values={step.values}
                            onChange={(key, value) => editValue(step.entry.step.id, key, value)}
                        />
                    </section>
                ))}
            </div>
        </DialogShell>
    )
}
