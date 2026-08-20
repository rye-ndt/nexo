import {ReportSection} from '@/features/workflows/components/inspector/report-section'
import {t, type MessageKey} from '@/shared/lib/i18n'
import type {Handoff} from '@/features/workflows/types'

type MapField = Exclude<keyof Handoff, 'step' | 'tldr' | 'outcome'>

const MAP_FIELDS: [MapField, MessageKey][] = [
    ['blockers', 'inspector.handoff.blockers'],
    ['approvedDecisions', 'inspector.handoff.approvedDecisions'],
    ['rejectedDecisions', 'inspector.handoff.rejectedDecisions'],
    ['currentBehaviors', 'inspector.handoff.currentBehaviors'],
    ['changedBehaviors', 'inspector.handoff.changedBehaviors'],
    ['mustAvoid', 'inspector.handoff.mustAvoid'],
    ['nuances', 'inspector.handoff.nuances'],
    ['knownGaps', 'inspector.handoff.knownGaps'],
]

function docText(doc: Handoff) {
    const blocks = [
        `${t('inspector.handoff.step').toUpperCase()}\n${doc.step}`,
        `${t('inspector.handoff.outcome').toUpperCase()}\n${doc.outcome}`,
    ]

    for (const [field, label] of MAP_FIELDS) {
        const entries = Object.entries(doc[field])
        if (entries.length === 0) continue

        const lines = entries.map(([key, value]) => `${key}: ${value}`).join('\n')
        blocks.push(`${t(label).toUpperCase()}\n${lines}`)
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
                label={t('inspector.handoff.label')}
                term="handoff"
                count={docs.length}
                empty={t('inspector.handoff.empty')}
                trailing={
                    <span className="text-sm text-muted-foreground">
                        {t('inspector.handoff.goesNext')}
                    </span>
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
