/**
 * The only place the generated Wails bindings for agent defaults and the
 * autopilot switch are touched. Inside the webview the defaults and the models
 * they can pick from come from the Go user config. Under the plain vite dev
 * server there is no Go side, so both live in this module's memory on top of
 * src/lib/mock-preferences.ts.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {TASK_LEVELS, THINKING_LEVELS, type TaskLevel, type ThinkingLevel} from '@/shared/lib/enums'
import {
    MOCK_AGENT_DEFAULT_OPTIONS,
    MOCK_AGENT_DEFAULTS,
    MOCK_AUTOPILOT,
} from '@/features/settings/mock-preferences'
import type {AgentDefault, AgentDefaultOptions, TokenPrices} from '@/features/settings/types'
import {
    AgentDefaultOptions as FetchAgentDefaultOptions,
    AgentDefaults as FetchAgentDefaults,
    Autopilot as FetchAutopilot,
    SetAgentDefault as SaveAgentDefault,
    SetAgentDefaultPrices as SaveAgentDefaultPrices,
    SetAutopilot as SaveAutopilot,
} from '@wailsjs/go/wails_api/API'

const ROUNDTRIP_MS = 400

let defaults: AgentDefault[] = structuredClone(MOCK_AGENT_DEFAULTS)

let autopilotOn = MOCK_AUTOPILOT

async function roundtrip() {
    await new Promise((resolve) => setTimeout(resolve, ROUNDTRIP_MS))
}

function modelLabel(model: string) {
    const option = MOCK_AGENT_DEFAULT_OPTIONS.models.find((candidate) => candidate.model === model)
    if (!option) throw new Error(`${model} is not a model this harness can run.`)
    return option.label
}

export async function listAgentDefaults(): Promise<AgentDefault[]> {
    if (!hasWailsRuntime()) return structuredClone(defaults)

    const infos = await bridge(FetchAgentDefaults)
    const stored: AgentDefault[] = []

    for (const info of infos) {
        if (!isTaskLevel(info.task_level) || !isThinkingLevel(info.thinking_level)) continue

        stored.push({
            taskLevel: info.task_level,
            model: info.model,
            modelLabel: info.model_label,
            thinkingLevel: info.thinking_level,
            prices: {
                input: info.input_price ?? '',
                cachedInput: info.cached_input_price ?? '',
                output: info.output_price ?? '',
            },
        })
    }

    return stored
}

export async function agentDefaultOptions(): Promise<AgentDefaultOptions> {
    if (!hasWailsRuntime()) return structuredClone(MOCK_AGENT_DEFAULT_OPTIONS)

    const info = await bridge(FetchAgentDefaultOptions)

    return {
        taskLevels: (info.task_levels ?? []).filter(isTaskLevel),
        models: (info.models ?? []).map((model) => ({
            model: model.model,
            label: model.label,
            harness: model.harness,
        })),
        thinkingLevels: (info.thinking_levels ?? []).filter(isThinkingLevel),
    }
}

export async function setAgentDefault(
    taskLevel: string,
    model: string,
    thinkingLevel: string,
): Promise<void> {
    if (!isTaskLevel(taskLevel)) throw new Error(`${taskLevel} is not a task level.`)
    if (!isThinkingLevel(thinkingLevel)) throw new Error(`${thinkingLevel} is not an effort level.`)

    if (hasWailsRuntime()) {
        await bridge(() => SaveAgentDefault(taskLevel, model, thinkingLevel))
        return
    }

    const label = modelLabel(model)
    await roundtrip()

    defaults = defaults.map((current) =>
        current.taskLevel === taskLevel
            ? {...current, model, modelLabel: label, thinkingLevel}
            : current,
    )
}

export async function setAgentDefaultPrices(taskLevel: string, prices: TokenPrices): Promise<void> {
    if (!isTaskLevel(taskLevel)) throw new Error(`${taskLevel} is not a task level.`)

    const trimmed: TokenPrices = {
        input: prices.input.trim(),
        cachedInput: prices.cachedInput.trim(),
        output: prices.output.trim(),
    }

    if (hasWailsRuntime()) {
        await bridge(() =>
            SaveAgentDefaultPrices(taskLevel, trimmed.input, trimmed.cachedInput, trimmed.output),
        )
        return
    }

    for (const price of Object.values(trimmed)) {
        if (price !== '' && !(Number(price) >= 0))
            throw new Error(`${price} is not a price. Enter dollars per million tokens.`)
    }

    await roundtrip()

    defaults = defaults.map((current) =>
        current.taskLevel === taskLevel ? {...current, prices: trimmed} : current,
    )
}

export async function autopilot(): Promise<boolean> {
    if (hasWailsRuntime()) autopilotOn = await bridge(FetchAutopilot)
    return autopilotOn
}

export async function setAutopilot(on: boolean): Promise<void> {
    if (hasWailsRuntime()) {
        await bridge(() => SaveAutopilot(on))
        autopilotOn = on
        return
    }

    await roundtrip()
    autopilotOn = on
}

export function cachedAutopilot(): boolean {
    return autopilotOn
}

/** What the vite mock has been told so far, for the simulated run's price lookup. */
export function cachedAgentDefaults(): AgentDefault[] {
    return defaults
}

function isTaskLevel(value: string): value is TaskLevel {
    return TASK_LEVELS.some((level) => level === value)
}

function isThinkingLevel(value: string): value is ThinkingLevel {
    return THINKING_LEVELS.some((level) => level === value)
}
