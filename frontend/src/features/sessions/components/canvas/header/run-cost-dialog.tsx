import type {ReactNode} from 'react'

import {formatMoment, formatTokens, formatUSD} from '@/shared/lib/format'
import {sessionRunWindow} from '@/features/sessions/graph'
import {useElapsed} from '@/shared/hooks/use-elapsed'
import {Progress} from '@/shared/ui/progress'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/shared/ui/dialog'
import type {Session, Spend, Task} from '@/features/sessions/types'

const NOTHING_SPENT: Spend = {input: 0, cached: 0, output: 0}

type NodeSpend = {
    id: string
    title: string
    spent: Spend
    measure: number
}

function ledger(spent: Spend) {
    return `${formatTokens(spent.input)} in · ${formatTokens(spent.cached)} cached · ${formatTokens(spent.output)} out`
}

function nodeSpends(tasks: Task[], priced: boolean): NodeSpend[] {
    return tasks
        .flatMap((task) => {
            const spent = task.run?.spent
            if (!spent) return []

            return [
                {
                    id: task.id,
                    title: task.title,
                    spent,
                    measure: priced ? (task.run?.costUsd ?? 0) : spent.output,
                },
            ]
        })
        .sort((left, right) => right.measure - left.measure)
}

export function RunCostDialog({session, onClose}: {session: Session; onClose: () => void}) {
    const {startedAt, finishedAt} = sessionRunWindow(session)
    const elapsed = useElapsed(startedAt, finishedAt)

    const priced = session.priced ?? false
    const spent = session.spent ?? NOTHING_SPENT

    const nodes = nodeSpends(session.tasks, priced)
    const largest = nodes[0]?.measure ?? 0

    const change = (open: boolean) => {
        if (!open) onClose()
    }

    return (
        <Dialog open onOpenChange={change}>
            <DialogContent className="p-5">
                <DialogHeader className="pr-8">
                    <DialogTitle>Run cost</DialogTitle>
                    <DialogDescription className="truncate">{session.name}</DialogDescription>
                </DialogHeader>

                <div className="mt-4 grid grid-cols-2 gap-3">
                    <Readout
                        label="Cost"
                        value={priced ? formatUSD(session.costUsd ?? 0) : '—'}
                        detail={
                            priced ? (
                                <span className="font-mono tabular-nums">{ledger(spent)}</span>
                            ) : (
                                'Add prices in Settings → Preferences to see this run in dollars.'
                            )
                        }
                    />
                    <Readout
                        label="Elapsed"
                        value={elapsed ?? '—'}
                        detail={
                            finishedAt
                                ? `Finished ${formatMoment(finishedAt)}`
                                : `Started ${formatMoment(startedAt)}`
                        }
                    />
                </div>

                <div className="mt-6">
                    <div className="flex items-baseline justify-between gap-3">
                        <span className="micro-label">By node</span>
                        <span className="text-sm text-muted-foreground">Every attempt</span>
                    </div>

                    {nodes.length === 0 && (
                        <p className="mt-3 text-sm text-muted-foreground">
                            No node has spent anything yet.
                        </p>
                    )}

                    {nodes.length > 0 && (
                        <ul className="mt-3 flex max-h-[264px] flex-col gap-3 overflow-y-auto">
                            {nodes.map((node) => (
                                <li key={node.id} className="flex flex-col gap-2">
                                    <div className="flex items-baseline justify-between gap-3">
                                        <span className="truncate">{node.title}</span>
                                        <span className="shrink-0 font-mono text-sm tabular-nums text-muted-foreground">
                                            {priced
                                                ? formatUSD(node.measure)
                                                : formatTokens(node.measure)}
                                        </span>
                                    </div>
                                    <Progress
                                        className="h-1.5"
                                        value={largest ? (node.measure / largest) * 100 : 0}
                                    />
                                    <span className="font-mono text-sm tabular-nums text-muted-foreground">
                                        {ledger(node.spent)}
                                    </span>
                                </li>
                            ))}
                        </ul>
                    )}
                </div>
            </DialogContent>
        </Dialog>
    )
}

function Readout({label, value, detail}: {label: string; value: string; detail: ReactNode}) {
    return (
        <div className="rounded-md border border-border bg-muted px-3 py-3">
            <span className="micro-label">{label}</span>
            <p className="mt-2 font-mono text-xl font-medium tabular-nums text-foreground">
                {value}
            </p>
            <p className="mt-2 text-sm text-muted-foreground">{detail}</p>
        </div>
    )
}
