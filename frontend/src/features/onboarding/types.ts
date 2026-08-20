export const OnboardingPhase = {
    Checking: 'checking',
    Welcome: 'welcome',
    Agent: 'agent',
    Installing: 'installing',
    Login: 'login',
    Tour: 'tour',
    Done: 'done',
} as const

export type OnboardingPhase = (typeof OnboardingPhase)[keyof typeof OnboardingPhase]

/** The beats the spine draws, in order. Phases outside it show no spine. */
export const ONBOARDING_BEATS: OnboardingPhase[] = [
    OnboardingPhase.Welcome,
    OnboardingPhase.Agent,
    OnboardingPhase.Installing,
    OnboardingPhase.Login,
]

export type Rect = {top: number; left: number; width: number; height: number}
