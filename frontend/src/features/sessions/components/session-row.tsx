import type {MouseEvent} from 'react'
import {CircleStop, Lock, MoreHorizontal} from 'lucide-react'
import type {LucideIcon} from 'lucide-react'

import {SessionSpine} from '@/features/sessions/components/session-spine'
import {SESSION_TITLE_CLASSES} from '@/features/sessions/session-status'
import {CANCELLED_HINT, FINALIZED_HINT} from '@/features/sessions/session-copy'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/shared/ui/dropdown-menu'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {SessionStatus} from '@/shared/lib/enums'
import {formatRelative} from '@/shared/lib/format'
import {sessionProgress, sessionStatus} from '@/features/sessions/graph'
import {cn} from '@/shared/lib/utils'
import type {Session} from '@/features/sessions/types'

type Marker = {icon: LucideIcon; label: string; hint: string; className: string}

function markerFor(session: Session, status: SessionStatus): Marker | null {
    if (status === SessionStatus.Cancelled)
        return {
            icon: CircleStop,
            label: 'Cancelled',
            hint: CANCELLED_HINT,
            className: 'text-state-idle',
        }

    if (session.finalized)
        return {
            icon: Lock,
            label: 'Finalized',
            hint: FINALIZED_HINT,
            className: 'text-muted-foreground',
        }

    return null
}

export function SessionRow({
    session,
    active,
    onSelect,
    onClone,
    onDelete,
}: {
    session: Session
    active: boolean
    onSelect: (sessionId: string) => void
    onClone: (sessionId: string) => void
    onDelete: (sessionId: string) => void
}) {
    const status = sessionStatus(session)
    const marker = markerFor(session, status)

    const select = () => onSelect(session.id)
    const clone = () => onClone(session.id)
    const remove = () => onDelete(session.id)
    const stopPropagation = (event: MouseEvent<HTMLButtonElement>) => event.stopPropagation()

    return (
        <div className="group relative">
            <button
                type="button"
                onClick={select}
                aria-current={active}
                className={cn(
                    'flex w-full flex-col gap-2 rounded-xl px-3 py-3 pr-8 text-left transition-colors duration-[120ms] outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring/50',
                    active && 'bg-live-tint hover:bg-live-tint',
                )}
            >
                <span className="flex min-w-0 items-center gap-2">
                    <span
                        className={cn(
                            'truncate text-base font-medium',
                            SESSION_TITLE_CLASSES[status],
                        )}
                    >
                        {session.name}
                    </span>
                    {marker && <LockMarker marker={marker} />}
                </span>

                <SessionSpine session={session} />

                <SessionMeta session={session} />
            </button>

            {active && (
                <span className="pointer-events-none absolute inset-0 overflow-hidden rounded-xl">
                    <span className="absolute inset-y-0 left-0 w-1 bg-live" />
                </span>
            )}

            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <button
                        type="button"
                        aria-label={`Options for ${session.name}`}
                        onClick={stopPropagation}
                        className="absolute top-2 right-1 flex size-7 items-center justify-center rounded-md text-muted-foreground opacity-0 transition-opacity duration-[120ms] outline-none hover:bg-muted hover:text-foreground focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-ring/50 group-hover:opacity-100 aria-expanded:opacity-100"
                    >
                        <MoreHorizontal className="size-3.5" />
                    </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-40">
                    <DropdownMenuItem onSelect={clone}>Duplicate</DropdownMenuItem>
                    <DropdownMenuItem variant="destructive" onSelect={remove}>
                        Delete
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>
        </div>
    )
}

function LockMarker({marker}: {marker: Marker}) {
    const {icon: Icon, label, hint, className} = marker

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <span role="img" aria-label={label} className={cn('flex shrink-0', className)}>
                    <Icon className="size-3" />
                </span>
            </TooltipTrigger>
            <TooltipContent side="bottom">{hint}</TooltipContent>
        </Tooltip>
    )
}

function SessionMeta({session}: {session: Session}) {
    const {done, total} = sessionProgress(session)
    const relative = formatRelative(session.createdAt)

    return (
        <span className="flex items-center gap-1 text-sm text-muted-foreground">
            {total === 0 ? (
                <span>No tasks</span>
            ) : (
                <span>
                    <span className="font-mono">
                        {done}/{total}
                    </span>{' '}
                    done
                </span>
            )}
            {relative && (
                <>
                    <span aria-hidden>·</span>
                    <span className="font-mono">{relative}</span>
                </>
            )}
        </span>
    )
}
