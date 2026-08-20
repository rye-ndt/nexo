import {SlidersHorizontal} from 'lucide-react'

import {EffortTag} from '@/shared/components/effort-tag'
import {HelpTip} from '@/shared/components/help-tip'
import {type Effort} from '@/shared/lib/enums'

const CHOSEN_UNDER = 'Which model runs each effort level is set under Settings → Preferences.'

function explain(effort: Effort | null | undefined, fromRole: boolean) {
    if (!effort) return 'Set once this step has a role.'
    if (fromRole) return `Inherited from the role. ${CHOSEN_UNDER}`

    return `Set when this workflow was exported. ${CHOSEN_UNDER}`
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
                    <span className="micro-label">Effort</span>
                    <HelpTip term="effort" />
                </span>
                <span className="flex min-w-0 items-center gap-2 text-muted-foreground">
                    <SlidersHorizontal className="size-3 shrink-0" aria-hidden="true" />
                    {effort ? (
                        <EffortTag effort={effort} />
                    ) : (
                        <span className="truncate text-base text-foreground">Not set</span>
                    )}
                </span>
            </div>
            <p className="text-sm text-muted-foreground">{explain(effort, fromRole)}</p>
        </div>
    )
}
