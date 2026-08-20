import {OnboardingStep} from '@/features/onboarding/components/onboarding-step'
import {OnboardingPhase} from '@/features/onboarding/types'
import {RECOMMENDED_AGENT_ID} from '@/features/onboarding/use-onboarding'
import {StatusChip} from '@/shared/components/status-chip'
import {t, type MessageKey} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'
import {Button} from '@/shared/ui/button'
import {DialogFooter} from '@/shared/ui/dialog'
import type {Agent} from '@/features/agents/types'
import type {Onboarding} from '@/features/onboarding/use-onboarding'

const BLURBS: Record<string, MessageKey> = {
    claude_code: 'onboarding.agent.claude_code.blurb',
    codex: 'onboarding.agent.codex.blurb',
    open_code: 'onboarding.agent.open_code.blurb',
}

export function AgentStep({onboarding}: {onboarding: Onboarding}) {
    const chosen = onboarding.agent
    const name = chosen?.name ?? ''

    return (
        <OnboardingStep
            phase={OnboardingPhase.Agent}
            eyebrow={t('onboarding.agent.eyebrow')}
            title={t('onboarding.agent.title')}
        >
            <div className="flex flex-col gap-3">
                <p className="text-sm text-muted-foreground">{t('onboarding.agent.body')}</p>

                <div
                    role="radiogroup"
                    className="divide-y divide-border overflow-hidden rounded-xl border border-border"
                >
                    {onboarding.agents.map((agent) => (
                        <AgentChoice
                            key={agent.id}
                            agent={agent}
                            selected={agent.id === chosen?.id}
                            onSelect={onboarding.pick}
                        />
                    ))}
                </div>
            </div>

            <DialogFooter>
                <Button
                    autoFocus
                    className="min-w-36"
                    disabled={!chosen}
                    onClick={onboarding.confirmAgent}
                >
                    {onboarding.installedAlready
                        ? t('onboarding.agent.continue', {name})
                        : t('onboarding.agent.install', {name})}
                </Button>
            </DialogFooter>
        </OnboardingStep>
    )
}

function AgentChoice({
    agent,
    selected,
    onSelect,
}: {
    agent: Agent
    selected: boolean
    onSelect: (agentId: string) => void
}) {
    const blurb = BLURBS[agent.id]

    return (
        <button
            type="button"
            role="radio"
            aria-checked={selected}
            onClick={() => onSelect(agent.id)}
            className={cn(
                'flex w-full items-center gap-3 px-4 py-3 text-left outline-none transition-colors -outline-offset-2 focus-visible:outline-2 focus-visible:outline-ring',
                selected ? 'bg-live-tint' : 'hover:bg-muted',
            )}
        >
            <span
                className={cn(
                    'flex size-4 shrink-0 items-center justify-center rounded-full border-2 transition-colors',
                    selected ? 'border-live' : 'border-border-strong',
                )}
            >
                {selected && <span className="size-1.5 rounded-full bg-live" />}
            </span>

            <span className="flex min-w-0 flex-1 flex-col gap-1">
                <span className="truncate font-mono text-base font-medium">{agent.name}</span>
                {blurb && <span className="text-sm text-muted-foreground">{t(blurb)}</span>}
            </span>

            {agent.installed ? (
                <StatusChip tone="done">{t('onboarding.agent.installed')}</StatusChip>
            ) : (
                agent.id === RECOMMENDED_AGENT_ID && (
                    <StatusChip tone="outline">{t('onboarding.agent.recommended')}</StatusChip>
                )
            )}
        </button>
    )
}
