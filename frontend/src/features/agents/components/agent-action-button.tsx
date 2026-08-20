import {Button} from '@/shared/ui/button'
import {AGENT_ACTION_BUSY_LABELS, AGENT_ACTION_LABELS, AgentAction} from '@/shared/lib/enums'
import type {AgentControls} from '@/features/agents/controls'

const RUN: Record<AgentAction, (controls: AgentControls, agentId: string) => void> = {
    [AgentAction.Install]: (controls, agentId) => controls.install(agentId),
    [AgentAction.Uninstall]: (controls, agentId) => controls.uninstall(agentId),
    [AgentAction.LogIn]: (controls, agentId) => controls.logIn(agentId),
    [AgentAction.LogOut]: (controls, agentId) => controls.logOut(agentId),
    [AgentAction.Verify]: () => {},
}

export function AgentActionButton({
    action,
    agentId,
    controls,
    variant,
    className,
}: {
    action: AgentAction
    agentId: string
    controls: AgentControls
    variant?: 'default' | 'ghost'
    className?: string
}) {
    const running = controls.actionOf(agentId) === action
    const label = running
        ? (controls.progressOf(agentId) ?? AGENT_ACTION_BUSY_LABELS[action])
        : AGENT_ACTION_LABELS[action]

    return (
        <Button
            size="sm"
            variant={variant}
            className={className}
            disabled={controls.busy}
            onClick={() => RUN[action](controls, agentId)}
        >
            {label}
        </Button>
    )
}
