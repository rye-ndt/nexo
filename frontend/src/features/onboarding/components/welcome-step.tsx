import {Check} from 'lucide-react'

import {OnboardingStep} from '@/features/onboarding/components/onboarding-step'
import {OnboardingPhase} from '@/features/onboarding/types'
import {useLanguageChoice} from '@/features/settings/use-language'
import {LANGUAGES, LANGUAGE_NAMES, type Language} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'
import {Button} from '@/shared/ui/button'
import {DialogFooter} from '@/shared/ui/dialog'
import type {Onboarding} from '@/features/onboarding/use-onboarding'

export function WelcomeStep({onboarding}: {onboarding: Onboarding}) {
    const {language, setLanguage} = useLanguageChoice()

    return (
        <OnboardingStep
            phase={OnboardingPhase.Welcome}
            eyebrow={t('onboarding.welcome.eyebrow')}
            title={t('onboarding.welcome.title')}
        >
            <p className="text-base text-muted-foreground">{t('onboarding.welcome.body')}</p>

            <div className="flex flex-col gap-2">
                <span className="micro-label">{t('onboarding.welcome.language')}</span>

                <div role="radiogroup" className="flex gap-2">
                    {LANGUAGES.map((option) => (
                        <LanguageCard
                            key={option}
                            language={option}
                            selected={option === language}
                            onSelect={setLanguage}
                        />
                    ))}
                </div>
            </div>

            <DialogFooter>
                <Button autoFocus className="min-w-36" onClick={onboarding.toAgent}>
                    {t('onboarding.welcome.next')}
                </Button>
            </DialogFooter>
        </OnboardingStep>
    )
}

function LanguageCard({
    language,
    selected,
    onSelect,
}: {
    language: Language
    selected: boolean
    onSelect: (language: Language) => void
}) {
    return (
        <button
            type="button"
            role="radio"
            aria-checked={selected}
            onClick={() => onSelect(language)}
            className={cn(
                'flex flex-1 items-center justify-between gap-2 rounded-xl border px-4 py-3 text-left text-base outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring/50',
                selected
                    ? 'border-live bg-live-tint font-medium'
                    : 'border-border hover:border-border-strong',
            )}
        >
            {LANGUAGE_NAMES[language]}
            {selected && <Check className="size-4 shrink-0 text-live" />}
        </button>
    )
}
