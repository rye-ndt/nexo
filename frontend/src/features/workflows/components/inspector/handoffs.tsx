import {ReportSection} from '@/features/workflows/components/inspector/report-section'
import type {Handoff} from '@/features/workflows/types'

type MapField = Exclude<keyof Handoff, 'step' | 'tldr' | 'outcome'>

const MAP_FIELDS: [MapField, string][] = [
    ['blockers', 'Blockers'],
    ['approvedDecisions', 'Approved decisions'],
    ['rejectedDecisions', 'Rejected decisions'],
    ['currentBehaviors', 'Current behaviors'],
    ['changedBehaviors', 'Changed behaviors'],
    ['mustAvoid', 'Must avoid'],
    ['nuances', 'Nuances'],
    ['knownGaps', 'Known gaps'],
]

function docText(doc: Handoff) {
    const blocks = [`STEP\n${doc.step}`, `OUTCOME\n${doc.outcome}`]

    for (const [field, label] of MAP_FIELDS) {
        const entries = Object.entries(doc[field])
        if (entries.length === 0) continue

        const lines = entries.map(([key, value]) => `${key}: ${value}`).join('\n')
        blocks.push(`${label.toUpperCase()}\n${lines}`)
    }

    return blocks.join('\n\n')
}

export function Handoffs({docs}: {docs: Handoff[]}) {
    const summaries = docs.map((doc) => doc.tldr).filter(Boolean)

    return (
        <div className="flex flex-col gap-3">
            {summaries.map((tldr, index) => (
                <p key={index} className="text-base leading-[1.7]">
                    {tldr}
                </p>
            ))}

            <ReportSection
                label="Handoff"
                term="handoff"
                count={docs.length}
                empty="None. Nothing was carried forward."
                trailing={
                    <span className="text-sm text-muted-foreground">Goes to the next step</span>
                }
            >
                {docs.map((doc, index) => (
                    <p
                        key={index}
                        className="rounded-xl border border-border bg-muted/30 p-3 font-mono text-xs leading-[1.8] whitespace-pre-wrap"
                    >
                        {docText(doc)}
                    </p>
                ))}
            </ReportSection>
        </div>
    )
}
