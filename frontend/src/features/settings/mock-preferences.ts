import {Effort, ThinkingLevel} from '@/shared/lib/enums'
import type {
    AgentDefault,
    AgentDefaultOptions,
    ModelPrice,
    TokenPrices,
} from '@/features/settings/types'

export const MOCK_AUTOPILOT = false

const MOCK_MODEL_OPTIONS = [
    {model: 'fable', label: 'Claude Fable', harness: 'claude_code'},
    {model: 'opus', label: 'Claude Opus', harness: 'claude_code'},
    {model: 'sonnet', label: 'Claude Sonnet', harness: 'claude_code'},
    {model: 'haiku', label: 'Claude Haiku', harness: 'claude_code'},
    {
        model: 'opencode/deepseek-v4-flash-free',
        label: 'Deepseek V4 Flash',
        harness: 'open_code',
    },
]

export const MOCK_AGENT_DEFAULT_OPTIONS: AgentDefaultOptions = {
    efforts: [Effort.Quick, Effort.Standard, Effort.Deep, Effort.Exhaustive],
    models: MOCK_MODEL_OPTIONS,
    thinkingLevels: [
        ThinkingLevel.Low,
        ThinkingLevel.Medium,
        ThinkingLevel.High,
        ThinkingLevel.XHigh,
        ThinkingLevel.Max,
    ],
}

export const MOCK_AGENT_DEFAULTS: AgentDefault[] = [
    {
        effort: Effort.Quick,
        model: 'haiku',
        modelLabel: 'Claude Haiku',
        thinkingLevel: ThinkingLevel.Low,
    },
    {
        effort: Effort.Standard,
        model: 'sonnet',
        modelLabel: 'Claude Sonnet',
        thinkingLevel: ThinkingLevel.Medium,
    },
    {
        effort: Effort.Deep,
        model: 'opus',
        modelLabel: 'Claude Opus',
        thinkingLevel: ThinkingLevel.High,
    },
    {
        effort: Effort.Exhaustive,
        model: 'opus',
        modelLabel: 'Claude Opus',
        thinkingLevel: ThinkingLevel.Max,
    },
]

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
    {
        model: 'opencode/deepseek-v4-flash-free',
        modelLabel: 'Deepseek V4 Flash',
        prices: price('0', '', '0'),
    },
]

function price(input: string, cachedInput: string, output: string): TokenPrices {
    return {input, cachedInput, output}
}
