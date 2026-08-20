import type {ReactNode} from 'react'

import {ONBOARDING_BEATS, type OnboardingPhase} from '@/features/onboarding/types'
import {StepSpine} from '@/shared/components/step-spine'
import {DialogHeader, DialogTitle} from '@/shared/ui/dialog'

/**
 * The sheet every setup beat is drawn on. The spine at the top is the same chain
 * the tour uses and the same shape the canvas draws: setup is itself a workflow,
 * and the user walks one before being told what one is.
 */
export function OnboardingStep({
    phase,
    eyebrow,
    title,
    children,
}: {
    phase: OnboardingPhase
    eyebrow: string
    title: string
    children: ReactNode
}) {
    const beat = ONBOARDING_BEATS.indexOf(phase)

    return (
        <>
            <DialogHeader className="gap-4">
                <StepSpine
                    total={ONBOARDING_BEATS.length}
                    current={beat}
                    className="justify-start"
                />

                <div className="flex flex-col gap-2">
                    <span className="micro-label">{eyebrow}</span>
                    <DialogTitle className="text-xl">{title}</DialogTitle>
                </div>
            </DialogHeader>

            {children}
        </>
    )
}
