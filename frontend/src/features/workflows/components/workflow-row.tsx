import type {MouseEvent} from 'react'
import {CircleStop, Lock, MoreHorizontal, Pause} from 'lucide-react'
import type {LucideIcon} from 'lucide-react'

import {WorkflowSpine} from '@/features/workflows/components/workflow-spine'
import {WORKFLOW_TITLE_CLASSES} from '@/features/workflows/workflow-status'
import {CANCELLED_HINT, LOCKED_HINT, PAUSED_HINT} from '@/features/workflows/workflow-copy'
import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuLabel,
    ContextMenuTrigger,
} from '@/shared/ui/context-menu'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/shared/ui/dropdown-menu'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {WorkflowStatus} from '@/shared/lib/enums'
import {formatRelative} from '@/shared/lib/format'
import {workflowProgress, workflowStatus} from '@/features/workflows/graph'
import {cn} from '@/shared/lib/utils'
import type {Workflow} from '@/features/workflows/types'

type Marker = {icon: LucideIcon; label: string; hint: string; className: string}

function markerFor(workflow: Workflow, status: WorkflowStatus): Marker | null {
    if (status === WorkflowStatus.Cancelled)
        return {
            icon: CircleStop,
            label: 'Cancelled',
            hint: CANCELLED_HINT,
            className: 'text-state-idle',
        }

    if (status === WorkflowStatus.Paused)
        return {
            icon: Pause,
            label: 'Paused',
            hint: PAUSED_HINT,
            className: 'text-state-approval',
        }

    if (workflow.locked)
        return {
            icon: Lock,
            label: 'Locked',
            hint: LOCKED_HINT,
            className: 'text-muted-foreground',
        }

    return null
}

export function WorkflowRow({
    workflow,
    active,
    onSelect,
    onDuplicate,
    onExport,
    onDelete,
}: {
    workflow: Workflow
    active: boolean
    onSelect: (workflowId: string) => void
    onDuplicate: (workflowId: string) => void
    onExport: (workflowId: string) => void
    onDelete: (workflowId: string) => void
}) {
    const status = workflowStatus(workflow)
    const marker = markerFor(workflow, status)

    const select = () => onSelect(workflow.id)
    const stopPropagation = (event: MouseEvent<HTMLButtonElement>) => event.stopPropagation()

    const entries = [
        {label: 'Duplicate', run: () => onDuplicate(workflow.id)},
        {label: 'Export', run: () => onExport(workflow.id)},
        {label: 'Delete', destructive: true, run: () => onDelete(workflow.id)},
    ]

    return (
        <ContextMenu>
            <ContextMenuTrigger asChild>
                <div className="group relative min-w-0">
                    <button
                        type="button"
                        onClick={select}
                        aria-current={active}
                        className={cn(
                            'flex w-full min-w-0 flex-col gap-2 rounded-xl px-3 py-3 pr-8 text-left transition-colors duration-[120ms] outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring/50',
                            active && 'bg-live-tint hover:bg-live-tint',
                        )}
                    >
                        <span className="grid w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-2">
                            <span
                                className={cn(
                                    'truncate text-base font-medium',
                                    WORKFLOW_TITLE_CLASSES[status],
                                )}
                            >
                                {workflow.name}
                            </span>
                            {marker && <LockMarker marker={marker} />}
                        </span>

                        <WorkflowSpine workflow={workflow} />

                        <WorkflowMeta workflow={workflow} />
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
                                aria-label={`Options for ${workflow.name}`}
                                onClick={stopPropagation}
                                className="absolute top-2 right-1 flex size-7 items-center justify-center rounded-md text-muted-foreground opacity-0 transition-opacity duration-[120ms] outline-none hover:bg-muted hover:text-foreground focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-ring/50 group-hover:opacity-100 aria-expanded:opacity-100"
                            >
                                <MoreHorizontal className="size-3.5" />
                            </button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-40">
                            {entries.map(({label, destructive, run}) => (
                                <DropdownMenuItem
                                    key={label}
                                    variant={destructive ? 'destructive' : 'default'}
                                    onSelect={run}
                                >
                                    {label}
                                </DropdownMenuItem>
                            ))}
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>
            </ContextMenuTrigger>

            <ContextMenuContent className="max-w-64">
                <ContextMenuLabel>{workflow.name}</ContextMenuLabel>
                {entries.map(({label, destructive, run}) => (
                    <ContextMenuItem
                        key={label}
                        variant={destructive ? 'destructive' : 'default'}
                        onSelect={run}
                    >
                        {label}
                    </ContextMenuItem>
                ))}
            </ContextMenuContent>
        </ContextMenu>
    )
}

function LockMarker({marker}: {marker: Marker}) {
    const {icon: Icon, label, hint, className} = marker

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <span
                    role="img"
                    aria-label={label}
                    className={cn('flex shrink-0 justify-self-end', className)}
                >
                    <Icon className="size-3" />
                </span>
            </TooltipTrigger>
            <TooltipContent side="bottom">{hint}</TooltipContent>
        </Tooltip>
    )
}

function WorkflowMeta({workflow}: {workflow: Workflow}) {
    const {done, total} = workflowProgress(workflow)
    const relative = formatRelative(workflow.createdAt)

    return (
        <span className="flex items-center gap-1 text-sm text-muted-foreground">
            {total === 0 ? (
                <span>No steps</span>
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
