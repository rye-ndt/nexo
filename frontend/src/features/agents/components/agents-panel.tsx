import {AgentActionButton} from '@/features/agents/components/agent-action-button'
import {AgentLogin} from '@/features/agents/components/agent-login'
import {HelpTip} from '@/shared/components/help-tip'
import {AgentAction} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {useAgents} from '@/features/agents/use-agents'
import type {Agent} from '@/features/agents/types'
import type {AgentControls} from '@/features/agents/controls'

function metaLine(agent: Agent) {
    if (!agent.installed) return t('agent.panel.notInstalled')

    return t('agent.panel.meta', {
        version: agent.version.replace(/^v/, ''),
        count: agent.instanceCount,
    })
}

export function AgentsPanel() {
    const controls = useAgents()

    return (
        <section className="flex flex-col">
            <div className="flex flex-col gap-1 px-4 pt-4 pb-3">
                <div className="flex items-center gap-2">
                    <h3 className="text-lg font-medium">{t('agent.panel.title')}</h3>
                    <HelpTip term="agent" />
                </div>
                <p className="text-sm text-muted-foreground">{t('agent.panel.hint')}</p>
            </div>

            <div className="divide-y divide-border border-t border-border">
                {controls.agents.map((agent) => (
                    <AgentRow key={agent.id} agent={agent} controls={controls} />
                ))}

                {controls.agents.length === 0 && (
                    <p className="px-4 py-3 text-base text-muted-foreground">
                        {t('agent.panel.empty')}
                    </p>
                )}
            </div>
        </section>
    )
}

function AgentRow({agent, controls}: {agent: Agent; controls: AgentControls}) {
    const authUrl = controls.authUrlOf(agent.id)

    return (
        <div className="flex flex-col gap-3 px-4 py-3">
            <div className="flex items-center gap-3">
                <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <p className="truncate font-mono text-base font-medium">{agent.name}</p>
                    <p className="truncate font-mono text-sm text-muted-foreground">
                        {metaLine(agent)}
                    </p>
                </div>

                {agent.installed ? (
                    <div className="flex shrink-0 gap-2">
                        <AgentActionButton
                            action={agent.loggedIn ? AgentAction.LogOut : AgentAction.LogIn}
                            agentId={agent.id}
                            controls={controls}
                            variant={agent.loggedIn ? 'ghost' : undefined}
                        />
                        <AgentActionButton
                            action={AgentAction.Uninstall}
                            agentId={agent.id}
                            controls={controls}
                            variant="ghost"
                            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                        />
                    </div>
                ) : (
                    <AgentActionButton
                        action={AgentAction.Install}
                        agentId={agent.id}
                        controls={controls}
                        className="shrink-0"
                    />
                )}
            </div>

            {authUrl && !agent.loggedIn && (
                <AgentLogin agentId={agent.id} authUrl={authUrl} controls={controls} />
            )}
        </div>
    )
}
