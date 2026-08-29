import type {LucideIcon} from 'lucide-react'

import {t, type MessageKey} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'

export type RailNavItem<TId extends string> = {
    id: TId
    label: MessageKey
    icon: LucideIcon
    tour?: string
}

export function RailNav<TId extends string>({
    items,
    activeId,
    onSelect,
}: {
    items: readonly RailNavItem<TId>[]
    activeId: TId
    onSelect: (id: TId) => void
}) {
    return (
        <nav className="flex h-12 shrink-0 items-center gap-1 border-b border-border px-2">
            {items.map((item) => {
                const Icon = item.icon
                const active = item.id === activeId

                return (
                    <button
                        key={item.id}
                        type="button"
                        data-tour={item.tour}
                        aria-current={active}
                        onClick={() => onSelect(item.id)}
                        className={cn(
                            'relative flex flex-1 items-center justify-center gap-2 rounded-lg px-3 py-2 text-base outline-none transition-colors duration-[120ms] focus-visible:ring-2 focus-visible:ring-ring/50',
                            active
                                ? 'bg-live-tint font-medium text-foreground'
                                : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                        )}
                    >
                        <Icon className="size-4 shrink-0" />
                        {t(item.label)}
                        {active && (
                            <span
                                aria-hidden="true"
                                className="pointer-events-none absolute inset-x-3 bottom-0 h-0.5 rounded-full bg-live"
                            />
                        )}
                    </button>
                )
            })}
        </nav>
    )
}
