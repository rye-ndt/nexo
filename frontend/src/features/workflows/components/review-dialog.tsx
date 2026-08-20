import {DialogShell} from '@/shared/components/dialog-shell'
import {Button} from '@/shared/ui/button'
import {useDecision} from '@/shared/hooks/use-decision'
import {pluralize} from '@/shared/lib/format'
import type {Handoff, Step} from '@/features/workflows/types'

type DocSection = Exclude<keyof Handoff, 'step' | 'tldr' | 'outcome'>

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

function HandoffBrief({doc, index, total}: {doc: Handoff; index: number; total: number}) {
    const sections = SECTIONS.map(([field, label]) => ({
        label,
        entries: Object.entries(doc[field]),
    })).filter((section) => section.entries.length > 0)

    return (
        <article className="flex flex-col gap-6">
            {total > 1 && (
                <span className="micro-label">
                    Handoff {index + 1} of {total}
                </span>
            )}

            {doc.tldr && (
                <section className="flex flex-col gap-2">
                    <span className="micro-label">In short</span>
                    <p className="text-xl leading-[1.5]">{doc.tldr}</p>
                </section>
            )}

            {doc.step && (
                <section className="flex flex-col gap-2">
                    <span className="micro-label">Step</span>
                    <p className="text-sm text-muted-foreground">{doc.step}</p>
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

export function ReviewDialog({
    step,
    waiting,
    busy,
    onAnswer,
    onClose,
}: {
    step: Step
    waiting: number
    busy: boolean
    onAnswer: (accepted: boolean) => void
    onClose: () => void
}) {
    const decision = useDecision(busy, onAnswer)
    const docs = step.report?.handoffs ?? []

    const reject = () => decision.answer(false)
    const confirm = () => decision.answer(true)

    return (
        <DialogShell
            onClose={onClose}
            title={step.title || 'Untitled step'}
            description="Finished. Nothing downstream runs until you decide."
            term="review"
            aside={
                waiting > 1 ? (
                    <span className="micro-label shrink-0">
                        {pluralize(waiting, 'step')} waiting
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
                        This step carried nothing forward. Confirm only if you have checked its work
                        yourself.
                    </p>
                ) : (
                    docs.map((doc, index) => (
                        <HandoffBrief key={index} doc={doc} index={index} total={docs.length} />
                    ))
                )}
            </div>
        </DialogShell>
    )
}
