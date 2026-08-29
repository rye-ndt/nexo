import {AppShell} from '@/app/app-shell'
import {OnboardingDialog} from '@/features/onboarding/components/onboarding-dialog'
import {TourLayer} from '@/features/onboarding/components/tour-layer'
import {TourProvider} from '@/features/onboarding/components/tour-provider'
import {OnboardingPhase} from '@/features/onboarding/types'
import {useOnboarding} from '@/features/onboarding/use-onboarding'
import {useLanguage} from '@/shared/hooks/use-language'

function App() {
    useLanguage()

    const onboarding = useOnboarding()

    return (
        <TourProvider active={onboarding.phase === OnboardingPhase.Tour} onDone={onboarding.finish}>
            <AppShell />

            <OnboardingDialog onboarding={onboarding} />

            <TourLayer />
        </TourProvider>
    )
}

export default App
