import {Plus} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {TemplateGraph} from '@/features/store/components/template-graph'
import {t, tn} from '@/shared/lib/i18n'
import type {StoreTemplate} from '@/features/store/types'

export function TemplateCard({
    template,
    roleCount,
    onOpen,
}: {
    template: StoreTemplate
    roleCount: number
    onOpen: (template: StoreTemplate) => void
}) {
    const open = () => onOpen(template)

    return (
        <article className="group surface-card flex flex-col overflow-hidden ring-1 ring-border transition-shadow duration-[120ms] hover:ring-border-strong">
            <button
                type="button"
                aria-label={t('store.card.openWorkflow', {name: template.name})}
                onClick={open}
                className="flex flex-1 flex-col text-left outline-none focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:ring-inset"
            >
                <span className="flex h-16 shrink-0 items-center justify-center border-b border-border bg-accent px-4 py-3">
                    <TemplateGraph steps={template.steps} />
                </span>

                <span className="flex flex-1 flex-col gap-1.5 p-3">
                    <span className="text-lg font-medium">{template.name}</span>
                    <span className="line-clamp-3 text-sm text-muted-foreground">
                        {template.description}
                    </span>
                </span>
            </button>

            <div className="flex h-12 shrink-0 items-center gap-2 border-t border-border px-3">
                <span className="min-w-0 flex-1 truncate font-mono text-sm tabular-nums text-muted-foreground">
                    {tn('store.card.steps.one', 'store.card.steps.other', template.steps.length)}
                    <span aria-hidden> · </span>
                    {tn('store.card.usesRoles.one', 'store.card.usesRoles.other', roleCount)}
                </span>

                <Button
                    size="sm"
                    aria-label={t('store.card.addWorkflow', {name: template.name})}
                    onClick={open}
                >
                    <Plus />
                    {t('store.card.add')}
                </Button>
            </div>
        </article>
    )
}
