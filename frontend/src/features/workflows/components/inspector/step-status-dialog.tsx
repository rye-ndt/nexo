import {ContextDonut} from '@/features/workflows/components/inspector/context-donut'
import {Handoffs} from '@/features/workflows/components/inspector/handoffs'
import {RevertStepDialog} from '@/features/workflows/components/inspector/revert-step-dialog'
import {StepDiff} from '@/features/workflows/components/inspector/step-diff'
import type {Workflow, Step} from '@/features/workflows/types'
import {roleOf} from '@/features/workflows/step-inputs'
import {effortOf} from '@/features/workflows/step-spec'
import {agentDefaultFor} from '@/features/settings/agent-default'
import type {AgentDefault} from '@/features/settings/types'
import {useAgentDefaults} from '@/features/settings/use-agent-defaults'
import {InputFields} from '@/features/roles/components/input-fields'
import {toFieldValues} from '@/features/roles/role'
import type {Role} from '@/features/roles/types'
import {useRoles} from '@/features/roles/use-roles'
import {DialogShell} from '@/shared/components/dialog-shell'
import {HelpTip} from '@/shared/components/help-tip'
import {StateBadge} from '@/shared/components/step-state'
import {useElapsed} from '@/shared/hooks/use-elapsed'
import {useToggle} from '@/shared/hooks/use-toggle'
import {EFFORT_LABELS, type Effort, StepState, THINKING_LEVEL_LABELS} from '@/shared/lib/enums'
import {formatMoment, formatPercent} from '@/shared/lib/format'
import {t} from '@/shared/lib/i18n'
import {Button} from '@/shared/ui/button'

/** A step only has a snapshot to go back to once it has reported. */
const REVERTABLE = new Set<StepState>([StepState.Done, StepState.Failed])

function summarize(effort: Effort | null, role?: Role, agentDefault?: AgentDefault) {
    return [
        role?.name,
        effort && t(EFFORT_LABELS[effort]),
        agentDefault &&
            t('inspector.status.model', {
                model: agentDefault.modelLabel,
                thinking: t(THINKING_LEVEL_LABELS[agentDefault.thinkingLevel]),
            }),
    ]
        .filter(Boolean)
        .join(' · ')
}

function emptyLine(step: Step) {
    if (step.state === StepState.Cancelled) return t('inspector.status.emptyCancelled')
    if (step.state === StepState.Blocked) return t('inspector.status.emptyBlocked')

    return t('inspector.status.empty')
}

function pendingLine(step: Step) {
    if (step.state === StepState.Cancelled) return t('inspector.status.pendingCancelled')

    return t('inspector.status.pending')
}

function ignoreChange() {}

function ReadOnlyInputs({role, step}: {role: Role; step: Step}) {
    return (
        <section className="flex flex-col gap-3 px-4 py-4">
            <span className="micro-label">{t('inspector.status.inputs')}</span>
            <InputFields
                inputs={role.inputs}
                values={toFieldValues(role, step.values)}
                disabled
                onChange={ignoreChange}
            />
        </section>
    )
}

function Stat({label, value}: {label: string; value: string}) {
    return (
        <div className="flex min-w-0 flex-col gap-2 rounded-xl border border-border p-3">
            <span className="micro-label">{label}</span>
            <span className="truncate font-mono text-base">{value}</span>
        </div>
    )
}

export function StepStatusDialog({
    workflow,
    step,
    reverting,
    onRevert,
    onClose,
}: {
    workflow: Workflow
    step: Step
    reverting: boolean
    onRevert: () => void
    onClose: () => void
}) {
    const {defaults} = useAgentDefaults()
    const {roles} = useRoles()
    const role = roleOf(step, roles)
    const effort = effortOf(step, roles)
    const agentDefault = agentDefaultFor(defaults, effort ?? undefined)
    const confirm = useToggle()

    const run = step.run
    const context = run?.context
    const elapsed = useElapsed(run?.startedAt, run?.finishedAt)
    const retried = run?.retryCount !== undefined && run.retryCount > 0

    const revert = () => {
        confirm.close()
        onRevert()
    }

    return (
        <>
            <DialogShell
                onClose={onClose}
                title={step.title || t('step.untitled')}
                description={summarize(effort, role, agentDefault)}
                aside={<StateBadge state={step.state} />}
                footer={
                    <>
                        <span className="min-w-0 flex-1 truncate font-mono text-xs text-muted-foreground">
                            {step.id}
                        </span>
                        {REVERTABLE.has(step.state) && (
                            <Button
                                variant="destructive"
                                size="sm"
                                disabled={reverting}
                                onClick={confirm.open}
                            >
                                {reverting
                                    ? t('inspector.status.reverting')
                                    : t('inspector.status.revert')}
                            </Button>
                        )}
                        <Button variant="outline" size="sm" onClick={onClose}>
                            {t('inspector.status.close')}
                        </Button>
                    </>
                }
            >
                <div className="divide-y divide-border">
                    {run ? (
                        <>
                            <div className="grid grid-cols-2 gap-2 px-4 py-4">
                                <Stat
                                    label={t('inspector.status.started')}
                                    value={formatMoment(run.startedAt)}
                                />
                                <Stat
                                    label={t(
                                        step.state === StepState.Cancelled
                                            ? 'inspector.status.stopped'
                                            : 'inspector.status.finished',
                                    )}
                                    value={formatMoment(run.finishedAt)}
                                />
                                <Stat
                                    label={t('inspector.status.elapsed')}
                                    value={elapsed ?? '—'}
                                />
                                {retried && (
                                    <Stat
                                        label={t('inspector.status.retries')}
                                        value={String(run.retryCount)}
                                    />
                                )}
                            </div>

                            {context && (
                                <div className="flex items-center gap-4 px-4 py-4">
                                    <ContextDonut used={context.used} total={context.total} />
                                    <div className="flex min-w-0 flex-col gap-2">
                                        <span className="flex items-center gap-2">
                                            <span className="micro-label">
                                                {t('inspector.status.context')}
                                            </span>
                                            <HelpTip term="context" />
                                        </span>
                                        <span className="font-mono text-lg leading-none">
                                            {t('inspector.status.contextUsed', {
                                                percent: formatPercent(context.used, context.total),
                                            })}
                                        </span>
                                    </div>
                                </div>
                            )}
                        </>
                    ) : (
                        <p className="px-4 py-4 text-base text-muted-foreground">
                            {emptyLine(step)}
                        </p>
                    )}

                    {role && role.inputs.length > 0 && <ReadOnlyInputs role={role} step={step} />}

                    {run &&
                        (step.report ? (
                            <>
                                <div className="px-4 py-4">
                                    <StepDiff workflowId={workflow.id} stepId={step.id} />
                                </div>
                                <div className="px-4 py-4">
                                    <Handoffs docs={step.report.handoffs} />
                                </div>
                            </>
                        ) : (
                            <p className="px-4 py-4 text-base text-muted-foreground">
                                {pendingLine(step)}
                            </p>
                        ))}
                </div>
            </DialogShell>

            {confirm.on && (
                <RevertStepDialog
                    workflow={workflow}
                    step={step}
                    busy={reverting}
                    onConfirm={revert}
                    onClose={confirm.close}
                />
            )}
        </>
    )
}
