import type {ReactNode} from 'react'

import {formatMoment, formatTokens, formatUSD} from '@/shared/lib/format'
import {workflowRunWindow} from '@/features/workflows/graph'
import {useElapsed} from '@/shared/hooks/use-elapsed'
import {Progress} from '@/shared/ui/progress'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/shared/ui/dialog'
import {t} from '@/shared/lib/i18n'
import type {Workflow, Spend, Step} from '@/features/workflows/types'

const NOTHING_SPENT: Spend = {input: 0, cached: 0, output: 0}

type StepSpend = {
    id: string
    title: string
    spent: Spend
    measure: number
}

function ledger(spent: Spend) {
    return t('canvas.cost.ledger', {
        input: formatTokens(spent.input),
        cached: formatTokens(spent.cached),
        output: formatTokens(spent.output),
    })
}

function stepSpends(steps: Step[], priced: boolean): StepSpend[] {
    return steps
        .flatMap((step) => {
            const spent = step.run?.spent
            if (!spent) return []

            return [
                {
                    id: step.id,
                    title: step.title,
                    spent,
                    measure: priced ? (step.run?.costUsd ?? 0) : spent.output,
                },
            ]
        })
        .sort((left, right) => right.measure - left.measure)
}

export function RunCostDialog({workflow, onClose}: {workflow: Workflow; onClose: () => void}) {
    const {startedAt, finishedAt} = workflowRunWindow(workflow)
    const elapsed = useElapsed(startedAt, finishedAt)

    const priced = workflow.priced ?? false
    const spent = workflow.spent ?? NOTHING_SPENT

    const steps = stepSpends(workflow.steps, priced)
    const largest = steps[0]?.measure ?? 0

    const change = (open: boolean) => {
        if (!open) onClose()
    }

    return (
        <Dialog open onOpenChange={change}>
            <DialogContent className="p-5">
                <DialogHeader className="pr-8">
                    <DialogTitle>{t('canvas.cost.title')}</DialogTitle>
                    <DialogDescription className="truncate">{workflow.name}</DialogDescription>
                </DialogHeader>

                <div className="mt-4 grid grid-cols-2 gap-3">
                    <Readout
                        label={t('canvas.cost.cost')}
                        value={priced ? formatUSD(workflow.costUsd ?? 0) : '—'}
                        detail={
                            priced ? (
                                <span className="font-mono tabular-nums">{ledger(spent)}</span>
                            ) : (
                                t('canvas.cost.unpriced')
                            )
                        }
                    />
                    <Readout
                        label={t('canvas.cost.elapsed')}
                        value={elapsed ?? '—'}
                        detail={
                            finishedAt
                                ? t('canvas.cost.finishedAt', {moment: formatMoment(finishedAt)})
                                : t('canvas.cost.startedAt', {moment: formatMoment(startedAt)})
                        }
                    />
                </div>

                <div className="mt-6">
                    <div className="flex items-baseline justify-between gap-3">
                        <span className="micro-label">{t('canvas.cost.byStep')}</span>
                        <span className="text-sm text-muted-foreground">
                            {t('canvas.cost.everyAttempt')}
                        </span>
                    </div>

                    {steps.length === 0 && (
                        <p className="mt-3 text-sm text-muted-foreground">
                            {t('canvas.cost.nothingSpent')}
                        </p>
                    )}

                    {steps.length > 0 && (
                        <ul className="mt-3 flex max-h-[264px] flex-col gap-3 overflow-y-auto">
                            {steps.map((step) => (
                                <li key={step.id} className="flex flex-col gap-2">
                                    <div className="flex items-baseline justify-between gap-3">
                                        <span className="truncate">{step.title}</span>
                                        <span className="shrink-0 font-mono text-sm tabular-nums text-muted-foreground">
                                            {priced
                                                ? formatUSD(step.measure)
                                                : formatTokens(step.measure)}
                                        </span>
                                    </div>
                                    <Progress
                                        className="h-1.5"
                                        value={largest ? (step.measure / largest) * 100 : 0}
                                    />
                                    <span className="font-mono text-sm tabular-nums text-muted-foreground">
                                        {ledger(step.spent)}
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
