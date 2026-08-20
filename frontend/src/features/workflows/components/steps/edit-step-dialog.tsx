import {useMemo, useState} from 'react'
import {Lock, Trash2} from 'lucide-react'

import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {DialogShell} from '@/shared/components/dialog-shell'
import {HelpTip} from '@/shared/components/help-tip'
import {InheritedAgent} from '@/features/workflows/components/steps/inherited-agent'
import {MissingInputsNote} from '@/features/workflows/components/steps/missing-inputs'
import {StepForm} from '@/features/workflows/components/steps/step-form'
import {Button} from '@/shared/ui/button'
import {useRoles} from '@/features/roles/use-roles'
import {useToggle} from '@/shared/hooks/use-toggle'
import {missingRequired, toFieldValues, toInputValues} from '@/features/roles/role'
import {roleOf} from '@/features/workflows/step-inputs'
import {effortOf} from '@/features/workflows/step-spec'
import type {Step} from '@/features/workflows/types'
import {t} from '@/shared/lib/i18n'
import type {FieldValue} from '@/features/roles/types'

export function EditStepDialog({
    step,
    onSave,
    onDelete,
    onClose,
}: {
    step: Step
    onSave: (patch: Partial<Step>) => void
    onDelete: () => void
    onClose: () => void
}) {
    const {roles} = useRoles()
    const role = roleOf(step, roles)

    const [title, setTitle] = useState(step.title)
    const [prompt, setPrompt] = useState(step.prompt)
    const [edits, setEdits] = useState<Record<string, FieldValue>>({})
    const confirmingDelete = useToggle()

    const stored = useMemo(
        () => (role ? toFieldValues(role, step.values) : {}),
        [role, step.values],
    )

    const values = {...stored, ...edits}
    const missing = role ? missingRequired(role, values) : []
    const ready = title.trim().length > 0

    const editValue = (key: string, value: FieldValue) =>
        setEdits((current) => ({...current, [key]: value}))

    const save = () => {
        if (!ready) return

        onSave({
            title: title.trim(),
            prompt: prompt.trim(),
            values: role ? toInputValues(role, values) : step.values,
        })
        onClose()
    }

    return (
        <DialogShell
            onClose={onClose}
            title={title.trim() || t('step.untitled')}
            footer={
                <>
                    <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                        onClick={confirmingDelete.open}
                    >
                        <Trash2 />
                        {t('step.edit.deleteStep')}
                    </Button>
                    <span className="flex-1" />
                    <MissingInputsNote count={missing.length} />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        {t('step.edit.cancel')}
                    </Button>
                    <Button size="sm" disabled={!ready} onClick={save}>
                        {t('step.edit.save')}
                    </Button>
                </>
            }
        >
            <div className="flex flex-col gap-2 border-b border-border px-4 py-3">
                <div className="flex items-center justify-between gap-3">
                    <span className="flex items-center gap-2">
                        <span className="micro-label">{t('step.edit.role')}</span>
                        <HelpTip term="role" />
                    </span>
                    <span className="flex min-w-0 items-center gap-2 text-muted-foreground">
                        <Lock className="size-3 shrink-0" aria-hidden="true" />
                        <span className="truncate font-mono text-base text-foreground">
                            {role?.name ?? t('step.edit.roleGone')}
                        </span>
                    </span>
                </div>
                <p className="text-sm text-muted-foreground">{t('step.edit.roleFixed')}</p>
            </div>

            <InheritedAgent effort={effortOf(step, roles)} fromRole={Boolean(role)} />

            <StepForm
                inputs={role?.inputs ?? []}
                title={title}
                prompt={prompt}
                values={values}
                onTitleChange={setTitle}
                onPromptChange={setPrompt}
                onValueChange={editValue}
            />

            {confirmingDelete.on && (
                <ConfirmDialog
                    title={t('step.edit.deleteTitle', {
                        name: step.title || t('step.edit.deleteThisStep'),
                    })}
                    description={t('step.edit.deleteBody')}
                    confirmLabel={t('step.edit.deleteStep')}
                    destructive
                    onConfirm={onDelete}
                    onClose={confirmingDelete.close}
                />
            )}
        </DialogShell>
    )
}
