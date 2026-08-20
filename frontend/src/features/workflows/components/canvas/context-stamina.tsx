import {contextDrained, contextLeft, contextRingClass} from '@/features/workflows/context-ring'
import {formatPercent, formatTokens} from '@/shared/lib/format'
import {cn} from '@/shared/lib/utils'
import {t} from '@/shared/lib/i18n'

const RADIUS = 8

const CIRCUMFERENCE = 2 * Math.PI * RADIUS

/** Drains as the step spends its context window: full ring at the start, empty when it runs out. */
export function ContextStamina({used, total}: {used: number; total: number}) {
    const left = contextLeft(used, total)
    const arc = CIRCUMFERENCE * left
    const reading = t('canvas.context.reading', {
        used: formatTokens(used),
        total: formatTokens(total),
    })

    return (
        <span className="relative flex size-5 shrink-0 items-center justify-center" title={reading}>
            <svg viewBox="0 0 20 20" aria-hidden="true" className="size-full -rotate-90">
                <circle
                    cx="10"
                    cy="10"
                    r={RADIUS}
                    fill="none"
                    strokeWidth="2.5"
                    className="stroke-progress-track"
                />
                <circle
                    cx="10"
                    cy="10"
                    r={RADIUS}
                    fill="none"
                    strokeWidth="2.5"
                    strokeLinecap="round"
                    strokeDasharray={`${arc} ${CIRCUMFERENCE}`}
                    className={cn(
                        'transition-[stroke-dasharray,stroke] duration-300 ease-out motion-reduce:transition-none',
                        contextRingClass(left),
                        contextDrained(left) && 'animate-pulse motion-reduce:animate-none',
                    )}
                />
            </svg>

            <span className="sr-only">
                {t('canvas.context.share', {percent: formatPercent(used, total)})}
            </span>
        </span>
    )
}
