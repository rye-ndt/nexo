import {TaskLevel, ThinkingLevel} from '@/shared/lib/enums'
import type {AgentDefault, AgentDefaultOptions} from '@/features/settings/types'

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
    taskLevels: [TaskLevel.Lightweight, TaskLevel.Daily, TaskLevel.Heavy, TaskLevel.MaximumEffort],
    models: MOCK_MODEL_OPTIONS,
    thinkingLevels: [
        ThinkingLevel.Low,
        ThinkingLevel.Medium,
        ThinkingLevel.High,
        ThinkingLevel.XHigh,
        ThinkingLevel.Max,
    ],
}

/**
 * The real app ships every price blank and waits to be told. The mock fills them in
 * so a run under the plain vite server shows what it cost; clearing a field is how
 * the unpriced state is reached from here.
 */
export const MOCK_AGENT_DEFAULTS: AgentDefault[] = [
    {
        taskLevel: TaskLevel.Lightweight,
        model: 'haiku',
        modelLabel: 'Claude Haiku',
        thinkingLevel: ThinkingLevel.Low,
        prices: {input: '1', cachedInput: '0.1', output: '5'},
    },
    {
        taskLevel: TaskLevel.Daily,
        model: 'sonnet',
        modelLabel: 'Claude Sonnet',
        thinkingLevel: ThinkingLevel.Medium,
        prices: {input: '3', cachedInput: '0.3', output: '15'},
    },
    {
        taskLevel: TaskLevel.Heavy,
        model: 'opus',
        modelLabel: 'Claude Opus',
        thinkingLevel: ThinkingLevel.High,
        prices: {input: '15', cachedInput: '1.5', output: '75'},
    },
    {
        taskLevel: TaskLevel.MaximumEffort,
        model: 'opus',
        modelLabel: 'Claude Opus',
        thinkingLevel: ThinkingLevel.Max,
        prices: {input: '15', cachedInput: '1.5', output: '75'},
    },
]
