import {Boxes} from 'lucide-react'

import {AgentActionButton} from '@/features/agents/components/agent-action-button'
import {AgentLogin} from '@/features/agents/components/agent-login'
import {HelpTip} from '@/shared/components/help-tip'
import {AgentAction} from '@/shared/lib/enums'
import {isAgentReady, type RequiredAgent} from '@/features/workflows/use-required-agents'
import {formatAgentName} from '@/shared/lib/format'
import {t} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'
import type {AgentControls} from '@/features/agents/controls'

export function RequiredAgents({
    required,
    loading,
    controls,
}: {
    required: RequiredAgent[]
    loading: boolean
    controls: AgentControls
}) {
    if (!loading && required.length === 0) return <NoAgentLoggedIn />

    return (
        <div className="flex flex-col">
            <p className="px-4 py-4 text-base text-muted-foreground">
                {t('workflow.agents.intro')}
            </p>

            {loading ? (
                <p className="px-4 py-3 text-base text-muted-foreground">
                    {t('workflow.agents.checking')}
                </p>
            ) : (
                <div className="divide-y divide-border border-t border-border">
                    {required.map((entry) => (
                        <RequiredAgentRow key={entry.harness} entry={entry} controls={controls} />
                    ))}
                </div>
            )}
        </div>
    )
}

function NoAgentLoggedIn() {
    return (
        <div className="flex flex-col items-start gap-2 px-4 py-5">
            <span className="flex items-center gap-2 text-base font-medium">
                <Boxes className="size-4 shrink-0 text-muted-foreground" />
                {t('workflow.agents.noneTitle')}
                <HelpTip term="agent" />
            </span>

            <p className="text-sm text-muted-foreground">{t('workflow.agents.noneBody')}</p>
        </div>
    )
}

function statusLine(entry: RequiredAgent) {
    if (!entry.agent) return t('workflow.agents.notAvailable')
    if (!entry.agent.installed) return t('workflow.agents.notInstalled')
    if (!entry.agent.loggedIn) return t('workflow.agents.notLoggedIn')
    return t('workflow.agents.ready', {version: entry.agent.version})
}

function RequiredAgentRow({entry, controls}: {entry: RequiredAgent; controls: AgentControls}) {
    const {agent} = entry
    const ready = isAgentReady(entry)
    const authUrl = agent ? controls.authUrlOf(agent.id) : null

    return (
        <div className="flex flex-col gap-3 px-4 py-3">
            <div className="flex items-center gap-3">
                <span
                    className={cn(
                        'size-2 shrink-0 rounded-full',
                        ready ? 'bg-state-done' : 'bg-state-approval',
                    )}
                />

                <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <p className="truncate text-base font-medium">
                        {agent?.name ?? formatAgentName(entry.harness)}
                    </p>
                    <p className="truncate font-mono text-sm text-muted-foreground">
                        {statusLine(entry)}
                    </p>
                </div>

                {agent && !agent.installed && (
                    <AgentActionButton
                        action={AgentAction.Install}
                        agentId={agent.id}
                        controls={controls}
                        className="shrink-0"
                    />
                )}

                {agent?.installed && !agent.loggedIn && (
                    <AgentActionButton
                        action={AgentAction.LogIn}
                        agentId={agent.id}
                        controls={controls}
                        className="shrink-0"
                    />
                )}
            </div>

            <p className="text-sm text-muted-foreground">
                {t('workflow.agents.models', {models: entry.modelLabels.join(', ')})}
            </p>

            {agent && authUrl && !agent.loggedIn && (
                <AgentLogin agentId={agent.id} authUrl={authUrl} controls={controls} />
            )}
        </div>
    )
}
