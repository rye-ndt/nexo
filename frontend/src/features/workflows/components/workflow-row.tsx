import type {MouseEvent} from 'react'
import {CircleStop, Lock, MoreHorizontal, Pause} from 'lucide-react'
import type {LucideIcon} from 'lucide-react'

import {WORKFLOW_TITLE_CLASSES} from '@/features/workflows/workflow-status'
import {cancelledHint, lockedHint, pausedHint} from '@/features/workflows/workflow-copy'
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
import {formatRelative, formatTokens, formatUSD} from '@/shared/lib/format'
import {t} from '@/shared/lib/i18n'
import {hasActiveStep, workflowRunWindow, workflowStatus} from '@/features/workflows/graph'
import {useElapsed} from '@/shared/hooks/use-elapsed'
import {cn} from '@/shared/lib/utils'
import type {Workflow} from '@/features/workflows/types'

type Marker = {icon: LucideIcon; label: string; hint: string; className: string}

function markerFor(workflow: Workflow, status: WorkflowStatus): Marker | null {
    if (status === WorkflowStatus.Cancelled)
        return {
            icon: CircleStop,
            label: t('workflow.row.cancelled'),
            hint: cancelledHint(),
            className: 'text-state-idle',
        }

    if (status === WorkflowStatus.Paused)
        return {
            icon: Pause,
            label: t('workflow.row.paused'),
            hint: pausedHint(),
            className: 'text-state-approval',
        }

    if (workflow.locked)
        return {
            icon: Lock,
            label: t('workflow.row.locked'),
            hint: lockedHint(),
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
        {label: t('workflow.row.duplicate'), run: () => onDuplicate(workflow.id)},
        {label: t('workflow.row.export'), run: () => onExport(workflow.id)},
        {label: t('workflow.row.delete'), destructive: true, run: () => onDelete(workflow.id)},
    ]

    return (
        <ContextMenu>
            <ContextMenuTrigger asChild>
                <div data-tour={active ? 'workflow' : undefined} className="group relative min-w-0">
                    <button
                        type="button"
                        onClick={select}
                        aria-current={active}
                        className={cn(
                            'flex w-full min-w-0 flex-col gap-2 rounded-xl px-3 py-3 text-left ring-1 ring-border transition-colors duration-[120ms] outline-none hover:bg-muted hover:ring-border-strong focus-visible:ring-2 focus-visible:ring-ring/50',
                            active &&
                                'bg-live-tint ring-live/35 hover:bg-live-tint hover:ring-live/35',
                        )}
                    >
                        <span className="grid w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-2 pr-5">
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

                        <WorkflowTelemetry workflow={workflow} />

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
                                aria-label={t('workflow.row.options', {name: workflow.name})}
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

function WorkflowTelemetry({workflow}: {workflow: Workflow}) {
    const {startedAt, finishedAt} = workflowRunWindow(workflow)
    const timed = hasActiveStep(workflow) || Boolean(finishedAt)
    const elapsed = useElapsed(timed ? startedAt : undefined, finishedAt)

    if (!workflow.startedAt)
        return <span className="text-sm text-muted-foreground">{t('workflow.row.notRun')}</span>

    const spent = workflow.spent
    const tokens = spent ? spent.input + spent.cached + spent.output : 0

    return (
        <span className="flex items-center gap-1.5 font-mono text-sm tabular-nums text-muted-foreground">
            <span aria-label={t('workflow.row.tokens', {tokens: formatTokens(tokens)})}>
                {formatTokens(tokens)}
            </span>
            <span aria-hidden>·</span>
            <span
                aria-label={
                    elapsed ? t('workflow.row.ranFor', {elapsed}) : t('workflow.row.noRunTime')
                }
            >
                {elapsed ?? '—'}
            </span>
            <span aria-hidden>·</span>
            <span
                className="text-foreground"
                aria-label={workflow.priced ? t('workflow.row.cost') : t('workflow.row.notPriced')}
            >
                {workflow.priced ? formatUSD(workflow.costUsd ?? 0) : '—'}
            </span>
        </span>
    )
}

function WorkflowMeta({workflow}: {workflow: Workflow}) {
    const relative = formatRelative(workflow.createdAt)

    if (!relative) return null

    return (
        <span className="w-full border-t border-border pt-2 font-mono text-sm text-muted-foreground">
            {relative}
        </span>
    )
}
