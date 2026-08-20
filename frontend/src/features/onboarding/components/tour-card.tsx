import {useEffect, useRef} from 'react'

import {HelpTip} from '@/shared/components/help-tip'
import {StepSpine} from '@/shared/components/step-spine'
import {t} from '@/shared/lib/i18n'
import {Button} from '@/shared/ui/button'
import {cn} from '@/shared/lib/utils'
import type {TourStop} from '@/features/onboarding/tour'

export function TourCard({
    stop,
    index,
    total,
    onNext,
    onSkip,
    className = 'surface-card p-4 ring-1 ring-border-strong',
}: {
    stop: TourStop
    index: number
    total: number
    onNext: () => void
    onSkip: () => void
    className?: string
}) {
    const advance = useRef<HTMLButtonElement>(null)
    const last = index === total - 1

    // Runs after the dialog a docked stop sits in has claimed focus for its first
    // field, so the tour's own action is what the keyboard lands on.
    useEffect(() => advance.current?.focus(), [stop.id])

    return (
        <div className={cn('flex flex-col gap-3', className)}>
            <div className="flex items-center gap-2">
                <h2 className="text-lg font-medium">{t(stop.title)}</h2>
                <HelpTip term={stop.term} />
            </div>

            <p className="text-sm text-muted-foreground">{t(stop.body)}</p>

            <div className="flex items-center gap-2">
                <StepSpine total={total} current={index} />
                <span className="flex-1" />
                <Button variant="ghost" size="sm" onClick={onSkip}>
                    {t('onboarding.tour.skip')}
                </Button>
                <Button ref={advance} size="sm" onClick={onNext}>
                    {last ? t('onboarding.tour.done') : t('onboarding.tour.next')}
                </Button>
            </div>
        </div>
    )
}
