import type {MCPAuthKind, TaskLevel, ThinkingLevel} from '@/shared/lib/enums'

export type MCPServer = {
    id: string
    name: string
    url: string
    authorized: boolean
    authorizedAt?: string
    /** Who the server says the stored credential belongs to — set once authorized. */
    account?: string
    kind: MCPAuthKind
}

/**
 * What the vendor charges for this task level, in US dollars per million tokens,
 * held as the text the user typed. An empty string is a blank field, which is not
 * the same as '0' — a free model really does cost nothing.
 */
export type TokenPrices = {
    input: string
    cachedInput: string
    output: string
}

export type AgentDefault = {
    taskLevel: TaskLevel
    model: string
    modelLabel: string
    thinkingLevel: ThinkingLevel
    prices: TokenPrices
}

export type ModelOption = {
    model: string
    label: string
    /** Which harness runs this model — the DTO form of enums.ModelName.HarnessTool(). */
    harness: string
}

export type AgentDefaultOptions = {
    taskLevels: TaskLevel[]
    models: ModelOption[]
    thinkingLevels: ThinkingLevel[]
}
