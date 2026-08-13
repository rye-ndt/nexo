import {useState} from 'react'

import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {DialogShell} from '@/shared/components/dialog-shell'
import {RefineButton} from '@/features/templates/components/refine-button'
import {TemplateForm} from '@/features/templates/components/template-form'
import {Button} from '@/shared/ui/button'
import {useTemplateHelper} from '@/features/templates/use-template-helper'
import {useTemplates} from '@/features/templates/use-templates'
import {emptyTemplate, templateIssues} from '@/features/templates/template'
import type {Template, TemplateDraft} from '@/features/templates/types'

export function TemplateFormDialog({
    template,
    onClose,
}: {
    template: Template | null
    onClose: () => void
}) {
    const {saveTemplate, saving} = useTemplates()
    const [draft, setDraft] = useState<TemplateDraft>(() =>
        template ? structuredClone(template) : emptyTemplate(),
    )
    const [opened] = useState(() => JSON.stringify(draft))
    const [asking, setAsking] = useState<'save' | 'discard' | null>(null)
    /** What the form held before the agent filled it in, so the user can get it back. */
    const [beforeFill, setBeforeFill] = useState<TemplateDraft | null>(null)

    const helper = useTemplateHelper((filled, sent) => {
        setBeforeFill(sent)
        setDraft(filled)
    })

    const issues = templateIssues(draft)

    const dismiss = () => setAsking(null)

    const undoFill = () => {
        if (beforeFill) setDraft(beforeFill)
        setBeforeFill(null)
    }

    const commit = () => {
        setAsking(null)
        saveTemplate(draft, {onSuccess: onClose})
    }

    const save = () => {
        if (issues.length > 0) return
        if (template) setAsking('save')
        else commit()
    }

    const requestClose = () => {
        if (JSON.stringify(draft) !== opened) setAsking('discard')
        else onClose()
    }

    const saveLabel = template ? 'Save changes' : 'Create template'

    /** There is a name and a role, so the agent has something to work from. */
    const fillable = Boolean(draft.name.trim() && draft.role.trim())

    // A blocked helper takes the hint only once the user could plausibly reach for
    // it, since that is the moment they find the button dead and want to know why.
    const hint = helper.filling
        ? 'An agent is writing this. It can take a minute.'
        : (fillable ? helper.blocked : '') || (issues[0] ?? '')

    return (
        <DialogShell
            onClose={requestClose}
            size="wide"
            title={template ? 'Edit template' : 'New template'}
            footer={
                <>
                    <p className="min-w-0 flex-1 truncate text-sm text-muted-foreground">{hint}</p>
                    <RefineButton
                        disabled={!fillable || Boolean(helper.blocked)}
                        reason={helper.blocked}
                        working={helper.filling}
                        onFill={() => helper.fillIn(draft)}
                    />
                    {beforeFill && !helper.filling && (
                        <Button variant="ghost" size="sm" onClick={undoFill}>
                            Undo
                        </Button>
                    )}
                    <Button variant="ghost" size="sm" onClick={requestClose}>
                        Cancel
                    </Button>
                    <Button
                        size="sm"
                        disabled={saving || helper.filling || issues.length > 0}
                        onClick={save}
                    >
                        {saving ? 'Saving' : saveLabel}
                    </Button>
                </>
            }
        >
            <TemplateForm draft={draft} onChange={setDraft} />

            {asking === 'save' && template && (
                <ConfirmDialog
                    title={`Save changes to “${template.name}”?`}
                    description="The template is overwritten. Nodes already built from it keep their prompt."
                    confirmLabel="Save changes"
                    busy={saving}
                    onConfirm={commit}
                    onClose={dismiss}
                />
            )}

            {asking === 'discard' && (
                <ConfirmDialog
                    destructive
                    title={
                        template
                            ? `Discard changes to “${template.name}”?`
                            : 'Discard this template?'
                    }
                    description={
                        template
                            ? 'The edits made here are lost and the saved template stays as it was.'
                            : 'Nothing is saved, and everything filled in here is lost.'
                    }
                    confirmLabel="Discard"
                    dismissLabel="Keep editing"
                    onConfirm={onClose}
                    onClose={dismiss}
                />
            )}
        </DialogShell>
    )
}
