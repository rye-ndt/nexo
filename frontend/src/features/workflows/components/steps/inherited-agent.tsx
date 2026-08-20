import {SlidersHorizontal} from 'lucide-react'

import {EffortTag} from '@/shared/components/effort-tag'
import {HelpTip} from '@/shared/components/help-tip'
import {type Effort} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'

function explain(effort: Effort | null | undefined, fromRole: boolean) {
    if (!effort) return t('step.agent.noRole')
    if (fromRole) return t('step.agent.fromRole')

    return t('step.agent.fromExport')
}

export function InheritedAgent({
    effort,
    fromRole,
}: {
    effort: Effort | null | undefined
    fromRole: boolean
}) {
    return (
        <div className="flex flex-col gap-2 border-b border-border px-4 py-3">
            <div className="flex items-center justify-between gap-3">
                <span className="flex items-center gap-2">
                    <span className="micro-label">{t('step.agent.effort')}</span>
                    <HelpTip term="effort" />
                </span>
                <span className="flex min-w-0 items-center gap-2 text-muted-foreground">
                    <SlidersHorizontal className="size-3 shrink-0" aria-hidden="true" />
                    {effort ? (
                        <EffortTag effort={effort} />
                    ) : (
                        <span className="truncate text-base text-foreground">
                            {t('step.agent.notSet')}
                        </span>
                    )}
                </span>
            </div>
            <p className="text-sm text-muted-foreground">{explain(effort, fromRole)}</p>
        </div>
    )
}
