import {useState} from 'react'
import {Check} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {DialogShell} from '@/shared/components/dialog-shell'
import {TaskLevelTag} from '@/shared/components/task-level-tag'
import {cn} from '@/shared/lib/utils'
import {pluralize} from '@/shared/lib/format'
import type {Template} from '@/features/templates/types'

export function TemplateExportDialog({
    templates,
    onExport,
    onClose,
}: {
    templates: Template[]
    onExport: (templateIds: string[]) => void
    onClose: () => void
}) {
    const [pickedIds, setPickedIds] = useState<string[]>([])

    const toggle = (templateId: string) =>
        setPickedIds((current) =>
            current.includes(templateId)
                ? current.filter((id) => id !== templateId)
                : [...current, templateId],
        )

    const all = () => setPickedIds(templates.map((template) => template.id))
    const none = () => setPickedIds([])
    const send = () => onExport(pickedIds)

    const everyone = pickedIds.length === templates.length

    return (
        <DialogShell
            onClose={onClose}
            title="Export templates"
            description="They go into one .json file you choose the place for."
            aside={
                <Button variant="ghost" size="sm" onClick={everyone ? none : all}>
                    {everyone ? 'Clear' : 'Select all'}
                </Button>
            }
            footer={
                <>
                    <p className="min-w-0 flex-1 text-sm text-muted-foreground">
                        {pickedIds.length === 0
                            ? 'Tick the ones to take with you.'
                            : `${pickedIds.length} of ${templates.length} selected`}
                    </p>
                    <Button variant="ghost" size="sm" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button size="sm" disabled={pickedIds.length === 0} onClick={send}>
                        Export {pluralize(pickedIds.length, 'template')}
                    </Button>
                </>
            }
        >
            <div className="flex flex-col gap-2 p-4">
                {templates.map((template) => (
                    <ExportRow
                        key={template.id}
                        template={template}
                        picked={pickedIds.includes(template.id)}
                        onToggle={toggle}
                    />
                ))}
            </div>
        </DialogShell>
    )
}

function ExportRow({
    template,
    picked,
    onToggle,
}: {
    template: Template
    picked: boolean
    onToggle: (templateId: string) => void
}) {
    return (
        <button
            type="button"
            aria-pressed={picked}
            onClick={() => onToggle(template.id)}
            className={cn(
                'flex w-full items-start gap-3 rounded-xl border p-3 text-left transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring',
                picked
                    ? 'border-live bg-live-tint'
                    : 'border-border bg-card hover:border-foreground/25 hover:bg-muted/40',
            )}
        >
            <span
                aria-hidden="true"
                className={cn(
                    'mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-[5px] border transition-colors',
                    picked
                        ? 'border-live bg-live text-white'
                        : 'border-border-strong bg-background',
                )}
            >
                {picked && <Check className="size-3" strokeWidth={3} />}
            </span>

            <span className="flex min-w-0 flex-1 flex-col gap-1">
                <span className="flex items-center gap-2">
                    <span className="truncate text-base font-medium">{template.name}</span>
                    <TaskLevelTag taskLevel={template.taskLevel} />
                </span>

                <span className="line-clamp-1 text-sm text-muted-foreground">
                    {template.role || 'No role set.'}
                </span>
            </span>
        </button>
    )
}
