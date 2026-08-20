import {useCallback, useEffect, useMemo, useState} from 'react'
import {useQuery, useQueryClient} from '@tanstack/react-query'

import {installAgent, listAgents, onInstallProgress} from '@/features/agents/api'
import {completeOnboarding, hasOnboarded} from '@/features/onboarding/api'
import {OnboardingPhase} from '@/features/onboarding/types'
import {installRatio} from '@/features/onboarding/install'
import {reportError} from '@/shared/lib/error-bus'
import {t} from '@/shared/lib/i18n'
import type {Agent, InstallProgress} from '@/features/agents/types'

const AGENTS_KEY = ['agents']
const ONBOARDED_KEY = ['onboarded']

/** The one Nexo supports best, so it is what a fresh machine lands on. */
export const RECOMMENDED_AGENT_ID = 'claude_code'

/**
 * The whole first run, as one machine. The user picks a language, picks one
 * agent, and only that agent is installed — the rest are a Settings visit away.
 * Nothing is written to the user config until the tour ends, so quitting halfway
 * starts over rather than dropping the user into an app nobody explained.
 */
export function useOnboarding() {
    const queryClient = useQueryClient()

    const {data, isPending} = useQuery({queryKey: AGENTS_KEY, queryFn: listAgents})
    const onboarded = useQuery({queryKey: ONBOARDED_KEY, queryFn: hasOnboarded})

    const [phase, setPhase] = useState<OnboardingPhase | null>(null)
    const [pickedId, setPickedId] = useState<string | null>(null)
    const [progress, setProgress] = useState<InstallProgress | null>(null)
    const [installedIds, setInstalledIds] = useState<string[]>([])
    const [failed, setFailed] = useState(false)

    const agents = useMemo(() => data ?? [], [data])

    const agent = useMemo(() => {
        const chosen = agents.find((one) => one.id === pickedId)
        if (chosen) return chosen

        return agents.find((one) => one.id === RECOMMENDED_AGENT_ID) ?? agents[0] ?? null
    }, [agents, pickedId])

    const isInstalled = useCallback(
        (one: Agent) => one.installed || installedIds.includes(one.id),
        [installedIds],
    )

    useEffect(
        () =>
            onInstallProgress((agentId, update) => {
                if (agentId === agent?.id) setProgress(update)
            }),
        [agent?.id],
    )

    const install = useCallback(
        async (target: Agent) => {
            setPhase(OnboardingPhase.Installing)
            setFailed(false)
            setProgress(null)

            try {
                await installAgent(target.id)
                setInstalledIds((current) => [...current, target.id])
                await queryClient.invalidateQueries({queryKey: AGENTS_KEY})
                setPhase(OnboardingPhase.Login)
            } catch (cause) {
                reportError(cause, t('onboarding.install.failed', {name: target.name}))
                setFailed(true)
            }
        },
        [queryClient],
    )

    /**
     * Where a first run starts, before anyone has pressed anything. Derived
     * rather than assigned, so the two queries settling in either order land on
     * the same phase and no render is spent correcting the previous one. An
     * empty roster — including a roster that failed to load — skips onboarding
     * rather than trapping the user in a modal with nothing to sign in to.
     */
    const opening = (): OnboardingPhase => {
        if (isPending || onboarded.isPending) return OnboardingPhase.Checking
        if (onboarded.data) return OnboardingPhase.Done
        if (agents.length === 0) return OnboardingPhase.Done

        return OnboardingPhase.Welcome
    }

    const current = phase ?? opening()

    const installedAlready = agent ? isInstalled(agent) : false

    const confirmAgent = () => {
        if (!agent) return
        if (isInstalled(agent)) return setPhase(OnboardingPhase.Login)

        void install(agent)
    }

    const finish = useCallback(() => {
        setPhase(OnboardingPhase.Done)
        void completeOnboarding().catch(() => {})
        queryClient.setQueryData(ONBOARDED_KEY, true)
    }, [queryClient])

    return {
        phase: current,
        agent,
        agents,
        installedAlready,
        ratio: installedAlready ? 1 : installRatio(progress),
        failed,
        pick: setPickedId,
        confirmAgent,
        retryInstall: () => agent && void install(agent),
        toAgent: () => setPhase(OnboardingPhase.Agent),
        toTour: () => setPhase(OnboardingPhase.Tour),
        finish,
    }
}

export type Onboarding = ReturnType<typeof useOnboarding>
