import {useState, type ReactNode} from 'react'

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
import {t, tn, type MessageKey} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'
import type {Workflow, Spend, Step} from '@/features/workflows/types'

const NOTHING_SPENT: Spend = {input: 0, cached: 0, output: 0}

type SpendRow = {id: string; title: string; spent: Spend; measure: number; note?: string}

const PANEL_ID = 'run-cost-breakdown'

const Breakdown = {Step: 'step', Model: 'model'} as const

type Breakdown = (typeof Breakdown)[keyof typeof Breakdown]

function tabId(breakdown: Breakdown) {
    return `${PANEL_ID}-${breakdown}`
}

const TABS: {breakdown: Breakdown; label: MessageKey}[] = [
    {breakdown: Breakdown.Step, label: 'canvas.cost.byStep'},
    {breakdown: Breakdown.Model, label: 'canvas.cost.byModel'},
]

function ledger(spent: Spend) {
    return t('canvas.cost.ledger', {
        input: formatTokens(spent.input),
        cached: formatTokens(spent.cached),
        output: formatTokens(spent.output),
    })
}

function amount(measure: number, priced: boolean) {
    return priced ? formatUSD(measure) : formatTokens(measure)
}

function measureOf(step: Step, priced: boolean) {
    return priced ? (step.run?.costUsd ?? 0) : (step.run?.spent?.output ?? 0)
}

function added(left: Spend, right: Spend): Spend {
    return {
        input: left.input + right.input,
        cached: left.cached + right.cached,
        output: left.output + right.output,
    }
}

function bySpend(rows: SpendRow[]) {
    return rows.sort((left, right) => right.measure - left.measure)
}

function stepSpends(steps: Step[], priced: boolean): SpendRow[] {
    return bySpend(
        steps.flatMap((step) => {
            const spent = step.run?.spent
            if (!spent) return []

            return [{id: step.id, title: step.title, spent, measure: measureOf(step, priced)}]
        }),
    )
}

/** Grouped by the model id, not its label: two models the app cannot name share one label. */
function modelSpends(steps: Step[], priced: boolean): SpendRow[] {
    const groups = new Map<string, SpendRow & {steps: number}>()

    for (const step of steps) {
        const spent = step.run?.spent
        if (!spent) continue

        const id = step.run?.model ?? ''
        const group = groups.get(id)
        const counted = (group?.steps ?? 0) + 1

        groups.set(id, {
            id,
            title: step.run?.modelLabel || id || t('canvas.cost.unknownModel'),
            spent: group ? added(group.spent, spent) : spent,
            measure: (group?.measure ?? 0) + measureOf(step, priced),
            steps: counted,
            note: tn('canvas.cost.stepCount.one', 'canvas.cost.stepCount.other', counted),
        })
    }

    return bySpend([...groups.values()])
}

export function RunCostDialog({workflow, onClose}: {workflow: Workflow; onClose: () => void}) {
    const [breakdown, setBreakdown] = useState<Breakdown>(Breakdown.Step)

    const {startedAt, finishedAt} = workflowRunWindow(workflow)
    const elapsed = useElapsed(startedAt, finishedAt)

    const priced = workflow.priced ?? false
    const spent = workflow.spent ?? NOTHING_SPENT

    const steps = stepSpends(workflow.steps, priced)
    const models = modelSpends(workflow.steps, priced)

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
                    <div className="flex items-center justify-between gap-3">
                        <BreakdownTabs picked={breakdown} onPick={setBreakdown} />
                        <span className="text-sm text-muted-foreground">
                            {t('canvas.cost.everyAttempt')}
                        </span>
                    </div>

                    <div
                        role="tabpanel"
                        id={PANEL_ID}
                        aria-labelledby={tabId(breakdown)}
                        className="max-h-[264px] overflow-y-auto"
                    >
                        {breakdown === Breakdown.Step && <SpendList rows={steps} priced={priced} />}

                        {breakdown === Breakdown.Model && (
                            <SpendList rows={models} priced={priced} />
                        )}
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    )
}

function BreakdownTabs({
    picked,
    onPick,
}: {
    picked: Breakdown
    onPick: (breakdown: Breakdown) => void
}) {
    return (
        <div
            role="tablist"
            aria-label={t('canvas.cost.breakdown')}
            className="flex gap-0.5 rounded-md bg-muted p-0.5"
        >
            {TABS.map((tab) => (
                <button
                    key={tab.breakdown}
                    type="button"
                    role="tab"
                    id={tabId(tab.breakdown)}
                    aria-controls={PANEL_ID}
                    aria-selected={tab.breakdown === picked}
                    onClick={() => onPick(tab.breakdown)}
                    className={cn(
                        'rounded-sm px-3 py-1 text-sm font-medium whitespace-nowrap transition-colors outline-none focus-visible:ring-3 focus-visible:ring-ring/50',
                        tab.breakdown === picked
                            ? 'bg-card text-foreground shadow-[0_1px_3px_rgba(27,28,30,0.08)]'
                            : 'text-muted-foreground hover:text-foreground',
                    )}
                >
                    {t(tab.label)}
                </button>
            ))}
        </div>
    )
}

function SpendList({rows, priced}: {rows: SpendRow[]; priced: boolean}) {
    const largest = rows[0]?.measure ?? 0

    if (rows.length === 0)
        return <p className="mt-3 text-sm text-muted-foreground">{t('canvas.cost.nothingSpent')}</p>

    return (
        <ul className="mt-3 flex flex-col gap-3">
            {rows.map((row) => (
                <li key={row.id} className="flex flex-col gap-2">
                    <div className="flex items-baseline justify-between gap-3">
                        <span className="truncate">{row.title}</span>
                        <span className="shrink-0 font-mono text-sm tabular-nums text-muted-foreground">
                            {amount(row.measure, priced)}
                        </span>
                    </div>
                    <Progress
                        className="h-1.5"
                        value={largest ? (row.measure / largest) * 100 : 0}
                    />
                    <div className="flex items-baseline justify-between gap-3">
                        <span className="font-mono text-sm tabular-nums text-muted-foreground">
                            {ledger(row.spent)}
                        </span>
                        {row.note && (
                            <span className="shrink-0 text-sm text-muted-foreground">
                                {row.note}
                            </span>
                        )}
                    </div>
                </li>
            ))}
        </ul>
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
