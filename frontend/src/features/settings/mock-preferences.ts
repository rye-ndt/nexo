import {EFFORTS, Effort, ThinkingLevel} from '@/shared/lib/enums'
import {mockLoggedInAgentIds} from '@/features/agents/mock-agents'
import type {
    AgentDefault,
    AgentDefaultOptions,
    ModelOption,
    ModelPrice,
    TokenPrices,
} from '@/features/settings/types'

export const MOCK_AUTOPILOT = false

export const MOCK_MAX_RUNNING_AGENTS = 3

const MOCK_MODEL_OPTIONS: ModelOption[] = [
    {model: 'fable', label: 'Claude Fable', harness: 'claude_code'},
    {model: 'opus', label: 'Claude Opus', harness: 'claude_code'},
    {model: 'sonnet', label: 'Claude Sonnet', harness: 'claude_code'},
    {model: 'haiku', label: 'Claude Haiku', harness: 'claude_code'},
    {model: 'gpt-5.6-sol', label: 'GPT-5.6 Sol', harness: 'codex'},
    {model: 'gpt-5.6-terra', label: 'GPT-5.6 Terra', harness: 'codex'},
    {model: 'gpt-5.6-luna', label: 'GPT-5.6 Luna', harness: 'codex'},
    {
        model: 'opencode/deepseek-v4-flash-free',
        label: 'Deepseek V4 Flash',
        harness: 'open_code',
    },
]

type HarnessDefaults = {
    harness: string
    models: Record<Effort, {model: string; thinkingLevel: ThinkingLevel}>
}

const MOCK_HARNESS_DEFAULTS: HarnessDefaults[] = [
    {
        harness: 'claude_code',
        models: {
            [Effort.Quick]: {model: 'haiku', thinkingLevel: ThinkingLevel.Low},
            [Effort.Standard]: {model: 'sonnet', thinkingLevel: ThinkingLevel.Medium},
            [Effort.Deep]: {model: 'opus', thinkingLevel: ThinkingLevel.High},
            [Effort.Exhaustive]: {model: 'fable', thinkingLevel: ThinkingLevel.Max},
        },
    },
    {
        harness: 'codex',
        models: {
            [Effort.Quick]: {model: 'gpt-5.6-luna', thinkingLevel: ThinkingLevel.Low},
            [Effort.Standard]: {model: 'gpt-5.6-luna', thinkingLevel: ThinkingLevel.Medium},
            [Effort.Deep]: {model: 'gpt-5.6-terra', thinkingLevel: ThinkingLevel.High},
            [Effort.Exhaustive]: {model: 'gpt-5.6-sol', thinkingLevel: ThinkingLevel.Max},
        },
    },
    {
        harness: 'open_code',
        models: {
            [Effort.Quick]: {
                model: 'opencode/deepseek-v4-flash-free',
                thinkingLevel: ThinkingLevel.Low,
            },
            [Effort.Standard]: {
                model: 'opencode/deepseek-v4-flash-free',
                thinkingLevel: ThinkingLevel.Medium,
            },
            [Effort.Deep]: {
                model: 'opencode/deepseek-v4-flash-free',
                thinkingLevel: ThinkingLevel.High,
            },
            [Effort.Exhaustive]: {
                model: 'opencode/deepseek-v4-flash-free',
                thinkingLevel: ThinkingLevel.Max,
            },
        },
    },
]

function loggedInModels(): ModelOption[] {
    const loggedIn = mockLoggedInAgentIds()

    return MOCK_MODEL_OPTIONS.filter((option) => loggedIn.includes(option.harness))
}

export function mockModelOption(model: string): ModelOption | undefined {
    return loggedInModels().find((option) => option.model === model)
}

export function mockAgentDefaultOptions(): AgentDefaultOptions {
    return {
        efforts: [...EFFORTS],
        models: loggedInModels(),
        thinkingLevels: [
            ThinkingLevel.Low,
            ThinkingLevel.Medium,
            ThinkingLevel.High,
            ThinkingLevel.XHigh,
            ThinkingLevel.Max,
        ],
    }
}

export function mockAgentDefaults(picked: Partial<Record<Effort, AgentDefault>>): AgentDefault[] {
    const loggedIn = mockLoggedInAgentIds()
    const table = MOCK_HARNESS_DEFAULTS.find((entry) => loggedIn.includes(entry.harness))
    if (!table) return []

    return EFFORTS.map((effort) => {
        const pick = picked[effort]
        if (pick && mockModelOption(pick.model)) return {...pick}

        const {model, thinkingLevel} = table.models[effort]

        return {
            effort,
            model,
            modelLabel: mockModelOption(model)?.label ?? model,
            thinkingLevel,
        }
    })
}

/**
 * The real app ships every price blank and waits to be told. The mock fills them in
 * so a run under the plain vite server shows what it cost; clearing a field is how
 * the unpriced state is reached from here.
 */
export const MOCK_MODEL_PRICES: ModelPrice[] = [
    {model: 'fable', modelLabel: 'Claude Fable', prices: price('25', '2.5', '125')},
    {model: 'opus', modelLabel: 'Claude Opus', prices: price('15', '1.5', '75')},
    {model: 'sonnet', modelLabel: 'Claude Sonnet', prices: price('3', '0.3', '15')},
    {model: 'haiku', modelLabel: 'Claude Haiku', prices: price('1', '0.1', '5')},
    {model: 'gpt-5.6-sol', modelLabel: 'GPT-5.6 Sol', prices: price('12', '1.2', '60')},
    {model: 'gpt-5.6-terra', modelLabel: 'GPT-5.6 Terra', prices: price('4', '0.4', '20')},
    {model: 'gpt-5.6-luna', modelLabel: 'GPT-5.6 Luna', prices: price('1.2', '0.12', '6')},
    {
        model: 'opencode/deepseek-v4-flash-free',
        modelLabel: 'Deepseek V4 Flash',
        prices: price('0', '', '0'),
    },
]

function price(input: string, cachedInput: string, output: string): TokenPrices {
    return {input, cachedInput, output}
}
