/**
 * The only place the generated Wails bindings for agent defaults, model prices and
 * the autopilot switch are touched. Inside the webview the defaults, the models they
 * can pick from and what those models cost come from the Go user config. Under the
 * plain vite dev server there is no Go side, so they live in this module's memory on
 * top of src/features/settings/mock-preferences.ts.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {TASK_LEVELS, THINKING_LEVELS, type TaskLevel, type ThinkingLevel} from '@/shared/lib/enums'
import {
    MOCK_AGENT_DEFAULT_OPTIONS,
    MOCK_AGENT_DEFAULTS,
    MOCK_AUTOPILOT,
    MOCK_MODEL_PRICES,
} from '@/features/settings/mock-preferences'
import type {
    AgentDefault,
    AgentDefaultOptions,
    ModelPrice,
    TokenPrices,
} from '@/features/settings/types'
import {
    AgentDefaultOptions as FetchAgentDefaultOptions,
    AgentDefaults as FetchAgentDefaults,
    Autopilot as FetchAutopilot,
    ModelPrices as FetchModelPrices,
    SetAgentDefault as SaveAgentDefault,
    SetAutopilot as SaveAutopilot,
    SetModelPrices as SaveModelPrices,
} from '@wailsjs/go/wails_api/API'

const ROUNDTRIP_MS = 400

let defaults: AgentDefault[] = structuredClone(MOCK_AGENT_DEFAULTS)

let modelPrices: ModelPrice[] = structuredClone(MOCK_MODEL_PRICES)

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
        })
    }

    return stored
}

export async function listModelPrices(): Promise<ModelPrice[]> {
    if (!hasWailsRuntime()) return structuredClone(modelPrices)

    const infos = await bridge(FetchModelPrices)

    return infos.map((info) => ({
        model: info.model,
        modelLabel: info.model_label,
        prices: {
            input: info.input_price ?? '',
            cachedInput: info.cached_input_price ?? '',
            output: info.output_price ?? '',
        },
    }))
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

export async function setModelPrices(model: string, prices: TokenPrices): Promise<void> {
    if (hasWailsRuntime()) {
        await bridge(() => SaveModelPrices(model, prices.input, prices.cachedInput, prices.output))
        return
    }

    await roundtrip()

    modelPrices = modelPrices.map((current) =>
        current.model === model ? {...current, prices} : current,
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

export function cachedModelPrices(): ModelPrice[] {
    return modelPrices
}

function isTaskLevel(value: string): value is TaskLevel {
    return TASK_LEVELS.some((level) => level === value)
}

function isThinkingLevel(value: string): value is ThinkingLevel {
    return THINKING_LEVELS.some((level) => level === value)
}
