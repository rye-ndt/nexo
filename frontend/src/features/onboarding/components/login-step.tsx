import {AgentActionButton} from '@/features/agents/components/agent-action-button'
import {AgentLogin} from '@/features/agents/components/agent-login'
import {OnboardingStep} from '@/features/onboarding/components/onboarding-step'
import {OnboardingPhase} from '@/features/onboarding/types'
import {StatusChip} from '@/shared/components/status-chip'
import {AgentAction} from '@/shared/lib/enums'
import {useAgents} from '@/features/agents/use-agents'
import {t} from '@/shared/lib/i18n'
import {Button} from '@/shared/ui/button'
import {DialogFooter} from '@/shared/ui/dialog'
import type {Onboarding} from '@/features/onboarding/use-onboarding'

export function LoginStep({onboarding}: {onboarding: Onboarding}) {
    const controls = useAgents()
    const agent = controls.agents.find((one) => one.id === onboarding.agent?.id)
    const authUrl = agent ? controls.authUrlOf(agent.id) : null
    const ready = Boolean(agent?.loggedIn)
    const name = agent?.name ?? onboarding.agent?.name ?? ''

    return (
        <OnboardingStep
            phase={OnboardingPhase.Login}
            eyebrow={t('onboarding.login.eyebrow')}
            title={t('onboarding.login.title', {name})}
        >
            <div className="flex flex-col gap-3">
                <p className="text-sm text-muted-foreground">
                    {t('onboarding.login.body', {name})}
                </p>

                {agent && (
                    <div className="flex flex-col gap-3 rounded-xl border border-border px-4 py-3">
                        <div className="flex items-center gap-3">
                            <div className="flex min-w-0 flex-1 flex-col gap-1">
                                <p className="truncate font-mono text-base font-medium">
                                    {agent.name}
                                </p>
                                <p className="truncate font-mono text-sm text-muted-foreground">
                                    {ready
                                        ? t('onboarding.login.signedIn')
                                        : t('onboarding.login.notSignedIn')}
                                </p>
                            </div>

                            {ready ? (
                                <StatusChip tone="done">{t('onboarding.login.ready')}</StatusChip>
                            ) : (
                                <AgentActionButton
                                    action={AgentAction.LogIn}
                                    agentId={agent.id}
                                    controls={controls}
                                    className="shrink-0"
                                />
                            )}
                        </div>

                        {authUrl && !ready && (
                            <AgentLogin agentId={agent.id} authUrl={authUrl} controls={controls} />
                        )}
                    </div>
                )}
            </div>

            <DialogFooter className="justify-between">
                {ready ? (
                    <span />
                ) : (
                    <p className="text-sm text-muted-foreground">{t('onboarding.login.hint')}</p>
                )}

                <span className="flex gap-2">
                    <Button variant="ghost" onClick={onboarding.toAgent}>
                        {t('onboarding.install.pickAnother')}
                    </Button>

                    <Button className="min-w-36" disabled={!ready} onClick={onboarding.toTour}>
                        {t('onboarding.login.start')}
                    </Button>
                </span>
            </DialogFooter>
        </OnboardingStep>
    )
}
