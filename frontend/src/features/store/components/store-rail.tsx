import type {ReactNode} from 'react'
import {LayoutTemplate, Users} from 'lucide-react'
import type {LucideIcon} from 'lucide-react'

import {HelpTip} from '@/shared/components/help-tip'
import {StoreSection} from '@/features/store/types'
import {cn} from '@/shared/lib/utils'
import {t, type MessageKey} from '@/shared/lib/i18n'

type Entry = {id: StoreSection; label: MessageKey; icon: LucideIcon; count: number}

export function StoreRail({
    nav,
    section,
    workflowCount,
    roleCount,
    onSelect,
}: {
    nav: ReactNode
    section: StoreSection
    workflowCount: number
    roleCount: number
    onSelect: (section: StoreSection) => void
}) {
    const entries: Entry[] = [
        {
            id: StoreSection.Workflows,
            label: 'store.rail.workflows',
            icon: LayoutTemplate,
            count: workflowCount,
        },
        {id: StoreSection.Roles, label: 'store.rail.roles', icon: Users, count: roleCount},
    ]

    return (
        <aside className="surface-card flex h-full w-[280px] shrink-0 flex-col overflow-hidden ring-1 ring-border-strong">
            {nav}

            <div className="flex h-14 shrink-0 items-center gap-2 border-b border-border px-4">
                <span className="micro-label">{t('store.rail.title')}</span>
                <HelpTip term="store" side="bottom" />
            </div>

            <div className="flex flex-col gap-1 p-2">
                {entries.map((entry) => (
                    <SectionRow
                        key={entry.id}
                        entry={entry}
                        active={entry.id === section}
                        onSelect={onSelect}
                    />
                ))}
            </div>
        </aside>
    )
}

function SectionRow({
    entry,
    active,
    onSelect,
}: {
    entry: Entry
    active: boolean
    onSelect: (section: StoreSection) => void
}) {
    const Icon = entry.icon

    return (
        <button
            type="button"
            aria-current={active}
            onClick={() => onSelect(entry.id)}
            className={cn(
                'relative flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-base outline-none transition-colors duration-[120ms] focus-visible:ring-2 focus-visible:ring-ring/50',
                active
                    ? 'bg-live-tint font-medium text-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground',
            )}
        >
            {active && (
                <span
                    aria-hidden="true"
                    className="pointer-events-none absolute inset-0 overflow-hidden rounded-lg"
                >
                    <span className="absolute inset-y-0 left-0 w-1 bg-live" />
                </span>
            )}
            <Icon className="size-4 shrink-0" />
            <span className="min-w-0 flex-1 truncate">{t(entry.label)}</span>
            <span className="font-mono text-sm tabular-nums text-muted-foreground">
                {entry.count}
            </span>
        </button>
    )
}
