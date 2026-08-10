import {useState} from 'react'

import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {DialogShell} from '@/shared/components/dialog-shell'
import {TemplateForm} from '@/features/templates/components/template-form'
import {Button} from '@/shared/ui/button'
import {useTemplates} from '@/features/templates/use-templates'
import {useToggle} from '@/shared/hooks/use-toggle'
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
    const confirming = useToggle()

    const issues = templateIssues(draft)

    const commit = () => {
        confirming.close()
        saveTemplate(draft, {onSuccess: onClose})
    }

    const save = () => {
        if (issues.length > 0) return
        if (template) confirming.open()
        else commit()
    }

    const saveLabel = template ? 'Save changes' : 'Create template'

    return (
        <DialogShell
            onClose={onClose}
            title={template ? 'Edit template' : 'New template'}
            footer={
                <>
                    <p className="min-w-0 flex-1 text-sm text-muted-foreground">
                        {issues[0] ?? ''}
                    </p>
                    <Button variant="ghost" size="sm" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button size="sm" disabled={saving || issues.length > 0} onClick={save}>
                        {saving ? 'Saving' : saveLabel}
                    </Button>
                </>
            }
        >
            <TemplateForm draft={draft} onChange={setDraft} />

            {confirming.on && template && (
                <ConfirmDialog
                    title={`Save changes to “${template.name}”?`}
                    description="The template is overwritten. Nodes already built from it keep their prompt."
                    confirmLabel="Save changes"
                    busy={saving}
                    onConfirm={commit}
                    onClose={confirming.close}
                />
            )}
        </DialogShell>
    )
}
