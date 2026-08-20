import type {MCPAuthKind, Effort, ThinkingLevel} from '@/shared/lib/enums'

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
 * What the vendor charges for one model, in US dollars per million tokens, held as
 * the text the user typed. An empty string is a blank field, which is not the same
 * as '0' — a free model really does cost nothing.
 */
export type TokenPrices = {
    input: string
    cachedInput: string
    output: string
}

export type ModelPrice = {
    model: string
    modelLabel: string
    prices: TokenPrices
}

export type AgentDefault = {
    effort: Effort
    model: string
    modelLabel: string
    thinkingLevel: ThinkingLevel
}

export type ModelOption = {
    model: string
    label: string
    /** Which harness runs this model — the DTO form of enums.ModelName.HarnessTool(). */
    harness: string
}

export type AgentDefaultOptions = {
    efforts: Effort[]
    models: ModelOption[]
    thinkingLevels: ThinkingLevel[]
}
