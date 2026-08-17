import {ContextDonut} from '@/features/sessions/components/inspector/context-donut'
import {HandoverDocs} from '@/features/sessions/components/inspector/handover-docs'
import {RevertStepDialog} from '@/features/sessions/components/inspector/revert-step-dialog'
import {TaskDiff} from '@/features/sessions/components/inspector/task-diff'
import type {Session, Task} from '@/features/sessions/types'
import {templateOf} from '@/features/sessions/task-inputs'
import {taskLevelOf} from '@/features/sessions/task-spec'
import {agentDefaultFor} from '@/features/settings/agent-default'
import type {AgentDefault} from '@/features/settings/types'
import {useAgentDefaults} from '@/features/settings/use-agent-defaults'
import {ParamFields} from '@/features/templates/components/param-fields'
import {toFieldValues} from '@/features/templates/template'
import type {Template} from '@/features/templates/types'
import {useTemplates} from '@/features/templates/use-templates'
import {DialogShell} from '@/shared/components/dialog-shell'
import {StateBadge} from '@/shared/components/task-state'
import {useElapsed} from '@/shared/hooks/use-elapsed'
import {useToggle} from '@/shared/hooks/use-toggle'
import {
    TASK_LEVEL_LABELS,
    type TaskLevel,
    TaskState,
    THINKING_LEVEL_LABELS,
} from '@/shared/lib/enums'
import {formatMoment, formatPercent} from '@/shared/lib/format'
import {Button} from '@/shared/ui/button'

/** A step only has a snapshot to go back to once it has reported. */
const REVERTABLE = new Set<TaskState>([TaskState.Done, TaskState.Failed])

function summarize(taskLevel: TaskLevel | null, template?: Template, agentDefault?: AgentDefault) {
    return [
        template?.name,
        taskLevel && TASK_LEVEL_LABELS[taskLevel],
        agentDefault &&
            `${agentDefault.modelLabel} · ${THINKING_LEVEL_LABELS[agentDefault.thinkingLevel]}`,
    ]
        .filter(Boolean)
        .join(' · ')
}

function emptyLine(task: Task) {
    if (task.state === TaskState.Cancelled)
        return 'This node never started. The run was cancelled before its turn came.'
    if (task.state === TaskState.Blocked)
        return 'This node has not run. It is waiting on the nodes upstream of it.'
    return 'This node has not run yet.'
}

function pendingLine(task: Task) {
    if (task.state === TaskState.Cancelled)
        return 'You cancelled the run while this node was working. Its work was discarded, so there is no report.'
    return 'The report lands when this node stops.'
}

function ignoreChange() {}

function ReadOnlyInputs({template, task}: {template: Template; task: Task}) {
    return (
        <section className="flex flex-col gap-3 px-4 py-4">
            <span className="micro-label">Inputs</span>
            <ParamFields
                params={template.params}
                values={toFieldValues(template, task.values)}
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

export function TaskStatusDialog({
    session,
    task,
    reverting,
    onRevert,
    onClose,
}: {
    session: Session
    task: Task
    reverting: boolean
    onRevert: () => void
    onClose: () => void
}) {
    const {defaults} = useAgentDefaults()
    const {templates} = useTemplates()
    const template = templateOf(task, templates)
    const taskLevel = taskLevelOf(task, templates)
    const agentDefault = agentDefaultFor(defaults, taskLevel ?? undefined)
    const confirm = useToggle()

    const run = task.run
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
                title={task.title || 'Untitled node'}
                description={summarize(taskLevel, template, agentDefault)}
                aside={<StateBadge state={task.state} />}
                footer={
                    <>
                        <span className="min-w-0 flex-1 truncate font-mono text-xs text-muted-foreground">
                            {task.id}
                        </span>
                        {REVERTABLE.has(task.state) && (
                            <Button
                                variant="destructive"
                                size="sm"
                                disabled={reverting}
                                onClick={confirm.open}
                            >
                                {reverting ? 'Reverting…' : 'Revert to this step'}
                            </Button>
                        )}
                        <Button variant="outline" size="sm" onClick={onClose}>
                            Close
                        </Button>
                    </>
                }
            >
                <div className="divide-y divide-border">
                    {run ? (
                        <>
                            <div className="grid grid-cols-2 gap-2 px-4 py-4">
                                <Stat label="Started" value={formatMoment(run.startedAt)} />
                                <Stat
                                    label={
                                        task.state === TaskState.Cancelled ? 'Stopped' : 'Finished'
                                    }
                                    value={formatMoment(run.finishedAt)}
                                />
                                <Stat label="Elapsed" value={elapsed ?? '—'} />
                                {retried && <Stat label="Retries" value={String(run.retryCount)} />}
                            </div>

                            {context && (
                                <div className="flex items-center gap-4 px-4 py-4">
                                    <ContextDonut used={context.used} total={context.total} />
                                    <div className="flex min-w-0 flex-col gap-2">
                                        <span className="micro-label">Context</span>
                                        <span className="font-mono text-lg leading-none">
                                            {formatPercent(context.used, context.total)}% used
                                        </span>
                                    </div>
                                </div>
                            )}
                        </>
                    ) : (
                        <p className="px-4 py-4 text-base text-muted-foreground">
                            {emptyLine(task)}
                        </p>
                    )}

                    {template && template.params.length > 0 && (
                        <ReadOnlyInputs template={template} task={task} />
                    )}

                    {run &&
                        (task.report ? (
                            <>
                                <div className="px-4 py-4">
                                    <TaskDiff sessionId={session.id} taskId={task.id} />
                                </div>
                                <div className="px-4 py-4">
                                    <HandoverDocs docs={task.report.handoverDocs} />
                                </div>
                            </>
                        ) : (
                            <p className="px-4 py-4 text-base text-muted-foreground">
                                {pendingLine(task)}
                            </p>
                        ))}
                </div>
            </DialogShell>

            {confirm.on && (
                <RevertStepDialog
                    session={session}
                    task={task}
                    busy={reverting}
                    onConfirm={revert}
                    onClose={confirm.close}
                />
            )}
        </>
    )
}
