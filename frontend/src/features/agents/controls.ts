import type {AgentAction} from '@/shared/lib/enums'

/** What the settings roster and the new-workflow preflight both drive an agent through. */
export type AgentControls = {
    busy: boolean
    actionOf: (agentId: string) => AgentAction | null
    progressOf: (agentId: string) => string | null
    authUrlOf: (agentId: string) => string | null
    install: (agentId: string) => void
    uninstall: (agentId: string) => void
    logIn: (agentId: string) => void
    submitAuthCode: (agentId: string, code: string) => void
    openAuthUrl: (url: string) => void
}
