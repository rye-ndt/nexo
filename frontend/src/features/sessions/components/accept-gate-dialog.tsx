import {DialogShell} from '@/shared/components/dialog-shell'
import {Button} from '@/shared/ui/button'
import {useDecision} from '@/shared/hooks/use-decision'
import {pluralize} from '@/shared/lib/format'
import type {HandoverDoc, Task} from '@/features/sessions/types'

type DocSection = Exclude<keyof HandoverDoc, 'task' | 'tldr' | 'outcome'>

const SECTIONS: [DocSection, string][] = [
    ['blockers', 'Blockers'],
    ['knownGaps', 'Known gaps'],
    ['mustAvoid', 'Must avoid'],
    ['rejectedDecisions', 'Rejected decisions'],
    ['approvedDecisions', 'Approved decisions'],
    ['changedBehaviors', 'Changed behaviors'],
    ['currentBehaviors', 'Current behaviors'],
    ['nuances', 'Nuances'],
]

function DocSectionGroup({label, entries}: {label: string; entries: [string, string][]}) {
    return (
        <section className="flex flex-col gap-2">
            <span className="micro-label">{label}</span>
            <dl className="flex flex-col gap-2 border-l border-border pl-3">
                {entries.map(([key, value]) => (
                    <div key={key} className="flex flex-col gap-1">
                        <dt className="font-mono text-xs text-muted-foreground">{key}</dt>
                        <dd className="text-base">{value}</dd>
                    </div>
                ))}
            </dl>
        </section>
    )
}

function HandoverBrief({doc, index, total}: {doc: HandoverDoc; index: number; total: number}) {
    const sections = SECTIONS.map(([field, label]) => ({
        label,
        entries: Object.entries(doc[field]),
    })).filter((section) => section.entries.length > 0)

    return (
        <article className="flex flex-col gap-6">
            {total > 1 && (
                <span className="micro-label">
                    Handover {index + 1} of {total}
                </span>
            )}

            {doc.tldr && (
                <section className="flex flex-col gap-2">
                    <span className="micro-label">In short</span>
                    <p className="text-xl leading-[1.5]">{doc.tldr}</p>
                </section>
            )}

            {doc.task && (
                <section className="flex flex-col gap-2">
                    <span className="micro-label">Task</span>
                    <p className="text-sm text-muted-foreground">{doc.task}</p>
                </section>
            )}

            <section className="flex flex-col gap-2">
                <span className="micro-label">Outcome</span>
                <p className="text-lg leading-[1.7] whitespace-pre-wrap">{doc.outcome}</p>
            </section>

            {sections.map((section) => (
                <DocSectionGroup
                    key={section.label}
                    label={section.label}
                    entries={section.entries}
                />
            ))}
        </article>
    )
}

export function AcceptGateDialog({
    task,
    waiting,
    busy,
    onAnswer,
    onClose,
}: {
    task: Task
    waiting: number
    busy: boolean
    onAnswer: (accepted: boolean) => void
    onClose: () => void
}) {
    const decision = useDecision(busy, onAnswer)
    const docs = task.report?.handoverDocs ?? []

    const reject = () => decision.answer(false)
    const confirm = () => decision.answer(true)

    return (
        <DialogShell
            onClose={onClose}
            title={task.title || 'Untitled node'}
            description="Finished. Nothing downstream runs until you decide."
            aside={
                waiting > 1 ? (
                    <span className="micro-label shrink-0">
                        {pluralize(waiting, 'node')} waiting
                    </span>
                ) : undefined
            }
            footer={
                <>
                    <Button variant="ghost" size="sm" onClick={onClose}>
                        Not now
                    </Button>
                    <span className="flex-1" />
                    <Button variant="destructive" size="sm" disabled={busy} onClick={reject}>
                        {decision.labelOf(false, 'Reject and stop', 'Rejecting…')}
                    </Button>
                    <Button size="sm" disabled={busy} onClick={confirm}>
                        {decision.labelOf(true, 'Confirm and continue', 'Confirming…')}
                    </Button>
                </>
            }
        >
            <div className="flex flex-col gap-8 p-4">
                {docs.length === 0 ? (
                    <p className="text-base text-muted-foreground">
                        This node carried nothing forward. Confirm only if you have checked its work
                        yourself.
                    </p>
                ) : (
                    docs.map((doc, index) => (
                        <HandoverBrief key={index} doc={doc} index={index} total={docs.length} />
                    ))
                )}
            </div>
        </DialogShell>
    )
}
