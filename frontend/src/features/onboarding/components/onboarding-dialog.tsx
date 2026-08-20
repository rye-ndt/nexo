import {AgentStep} from '@/features/onboarding/components/agent-step'
import {InstallStep} from '@/features/onboarding/components/install-step'
import {LoginStep} from '@/features/onboarding/components/login-step'
import {WelcomeStep} from '@/features/onboarding/components/welcome-step'
import {OnboardingPhase} from '@/features/onboarding/types'
import {Dialog, DialogContent} from '@/shared/ui/dialog'
import type {Onboarding} from '@/features/onboarding/use-onboarding'

const stayOpen = (event: Event | KeyboardEvent) => event.preventDefault()

export function OnboardingDialog({onboarding}: {onboarding: Onboarding}) {
    const step = stepOf(onboarding)
    if (!step) return null

    return (
        <Dialog open>
            <DialogContent
                showCloseButton={false}
                className="flex flex-col gap-6 rounded-[20px] bg-card p-8 outline-none sm:max-w-lg"
                onEscapeKeyDown={stayOpen}
                onPointerDownOutside={stayOpen}
                onInteractOutside={stayOpen}
            >
                {step}
            </DialogContent>
        </Dialog>
    )
}

function stepOf(onboarding: Onboarding) {
    switch (onboarding.phase) {
        case OnboardingPhase.Welcome:
            return <WelcomeStep onboarding={onboarding} />
        case OnboardingPhase.Agent:
            return <AgentStep onboarding={onboarding} />
        case OnboardingPhase.Installing:
            return <InstallStep onboarding={onboarding} />
        case OnboardingPhase.Login:
            return <LoginStep onboarding={onboarding} />
        default:
            return null
    }
}
