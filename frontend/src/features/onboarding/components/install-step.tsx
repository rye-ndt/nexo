import {OnboardingStep} from '@/features/onboarding/components/onboarding-step'
import {OnboardingPhase} from '@/features/onboarding/types'
import {t} from '@/shared/lib/i18n'
import {Button} from '@/shared/ui/button'
import {DialogFooter} from '@/shared/ui/dialog'
import {Progress} from '@/shared/ui/progress'
import type {Onboarding} from '@/features/onboarding/use-onboarding'

export function InstallStep({onboarding}: {onboarding: Onboarding}) {
    const name = onboarding.agent?.name ?? ''
    const percent = Math.round(onboarding.ratio * 100)

    const label = onboarding.failed
        ? t('onboarding.install.failed', {name})
        : t('onboarding.install.busy', {name})

    return (
        <OnboardingStep
            phase={OnboardingPhase.Installing}
            eyebrow={t('onboarding.install.eyebrow')}
            title={t('onboarding.install.title', {name})}
        >
            <div className="flex flex-col gap-2">
                <div className="flex items-baseline justify-between">
                    <span className="text-sm text-muted-foreground">{label}</span>
                    <span className="font-mono text-xs font-semibold text-muted-foreground">
                        {percent}%
                    </span>
                </div>

                <Progress
                    aria-label={label}
                    value={percent}
                    className="h-2 rounded-full bg-muted"
                    indicatorClassName={onboarding.failed ? 'bg-state-failed' : 'bg-brand'}
                />
            </div>

            {onboarding.failed && (
                <DialogFooter>
                    <Button variant="ghost" className="min-w-36" onClick={onboarding.toAgent}>
                        {t('onboarding.install.pickAnother')}
                    </Button>

                    <Button autoFocus className="min-w-36" onClick={onboarding.retryInstall}>
                        {t('onboarding.install.retry')}
                    </Button>
                </DialogFooter>
            )}
        </OnboardingStep>
    )
}
