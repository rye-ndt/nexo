import {contextLeft, contextRingClass} from '@/features/workflows/context-ring'
import {formatPercent, formatTokens} from '@/shared/lib/format'
import {cn} from '@/shared/lib/utils'

const RADIUS = 32

const CIRCUMFERENCE = 2 * Math.PI * RADIUS

/** Drains like the stamina ring on the step: the arc is the context this step has left. */
export function ContextDonut({used, total}: {used: number; total: number}) {
    const left = contextLeft(used, total)
    const arc = CIRCUMFERENCE * left

    return (
        <div className="relative size-[88px] shrink-0">
            <svg viewBox="0 0 80 80" aria-hidden="true" className="size-full -rotate-90">
                <circle
                    cx="40"
                    cy="40"
                    r={RADIUS}
                    fill="none"
                    stroke="var(--progress-track)"
                    strokeWidth="7"
                />
                <circle
                    cx="40"
                    cy="40"
                    r={RADIUS}
                    fill="none"
                    strokeWidth="7"
                    strokeLinecap="round"
                    strokeDasharray={`${arc} ${CIRCUMFERENCE}`}
                    className={cn(
                        'transition-[stroke-dasharray,stroke] duration-300 ease-out motion-reduce:transition-none',
                        contextRingClass(left),
                    )}
                />
            </svg>

            <span className="absolute inset-0 flex flex-col items-center justify-center gap-1">
                <span className="font-mono text-base leading-none">{formatTokens(used)}</span>
                <span className="h-px w-5 bg-border" />
                <span className="font-mono text-xs leading-none text-muted-foreground">
                    {formatTokens(total)}
                </span>
            </span>

            <span className="sr-only">
                {formatPercent(used, total)}% of the context window used
            </span>
        </div>
    )
}
