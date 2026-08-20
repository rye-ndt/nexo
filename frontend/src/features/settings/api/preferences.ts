/**
 * The only place the generated Wails bindings for agent defaults, model prices and
 * the autopilot switch are touched. Inside the webview the defaults, the models they
 * can pick from and what those models cost come from the Go user config. Under the
 * plain vite dev server there is no Go side, so they live in this module's memory on
 * top of src/features/settings/mock-preferences.ts.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {
    EFFORTS,
    LANGUAGES,
    Language,
    THINKING_LEVELS,
    type Effort,
    type ThinkingLevel,
} from '@/shared/lib/enums'
import {applyLanguage, t} from '@/shared/lib/i18n'
import {
    MOCK_AUTOPILOT,
    MOCK_MODEL_PRICES,
    mockAgentDefaultOptions,
    mockAgentDefaults,
    mockModelOption,
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
    Language as FetchLanguage,
    ModelPrices as FetchModelPrices,
    SetAgentDefault as SaveAgentDefault,
    SetAutopilot as SaveAutopilot,
    SetLanguage as SaveLanguage,
    SetModelPrices as SaveModelPrices,
} from '@wailsjs/go/wails_api/API'

const ROUNDTRIP_MS = 400

let picked: Partial<Record<Effort, AgentDefault>> = {}

let modelPrices: ModelPrice[] = structuredClone(MOCK_MODEL_PRICES)

let autopilotOn = MOCK_AUTOPILOT

let chosenLanguage: Language = Language.En

async function roundtrip() {
    await new Promise((resolve) => setTimeout(resolve, ROUNDTRIP_MS))
}

function modelLabel(model: string) {
    const option = mockModelOption(model)
    if (!option) throw new Error(t('settings.error.unknownModel', {model}))
    return option.label
}

export async function listAgentDefaults(): Promise<AgentDefault[]> {
    if (!hasWailsRuntime()) return mockAgentDefaults(picked)

    const infos = await bridge(FetchAgentDefaults)
    const stored: AgentDefault[] = []

    for (const info of infos) {
        if (!isEffort(info.effort) || !isThinkingLevel(info.thinking_level)) continue

        stored.push({
            effort: info.effort,
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
    if (!hasWailsRuntime()) return mockAgentDefaultOptions()

    const info = await bridge(FetchAgentDefaultOptions)

    return {
        efforts: (info.efforts ?? []).filter(isEffort),
        models: (info.models ?? []).map((model) => ({
            model: model.model,
            label: model.label,
            harness: model.harness,
        })),
        thinkingLevels: (info.thinking_levels ?? []).filter(isThinkingLevel),
    }
}

export async function setAgentDefault(
    effort: string,
    model: string,
    thinkingLevel: string,
): Promise<void> {
    if (!isEffort(effort)) throw new Error(t('settings.error.unknownEffort', {effort}))
    if (!isThinkingLevel(thinkingLevel))
        throw new Error(t('settings.error.unknownThinkingLevel', {level: thinkingLevel}))

    if (hasWailsRuntime()) {
        await bridge(() => SaveAgentDefault(effort, model, thinkingLevel))
        return
    }

    const label = modelLabel(model)
    await roundtrip()

    picked = {...picked, [effort]: {effort, model, modelLabel: label, thinkingLevel}}
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

/**
 * Read once before the first render so the app never paints English and then
 * corrects itself. Applying is this module's job, not the caller's — every path
 * that changes the language goes through here.
 */
export async function loadLanguage(): Promise<Language> {
    if (hasWailsRuntime()) {
        const stored = await bridge(FetchLanguage).catch(() => Language.En)
        chosenLanguage = isLanguage(stored) ? stored : Language.En
    }

    applyLanguage(chosenLanguage)

    return chosenLanguage
}

export async function setLanguage(language: Language): Promise<void> {
    if (hasWailsRuntime()) await bridge(() => SaveLanguage(language))
    else await roundtrip()

    chosenLanguage = language
    applyLanguage(language)
}

export function cachedLanguage(): Language {
    return chosenLanguage
}

export function cachedAutopilot(): boolean {
    return autopilotOn
}

/** What the vite mock has been told so far, for the simulated run's price lookup. */
export function cachedAgentDefaults(): AgentDefault[] {
    return mockAgentDefaults(picked)
}

export function cachedModelPrices(): ModelPrice[] {
    return modelPrices
}

function isLanguage(value: string): value is Language {
    return LANGUAGES.some((language) => language === value)
}

function isEffort(value: string): value is Effort {
    return EFFORTS.some((level) => level === value)
}

function isThinkingLevel(value: string): value is ThinkingLevel {
    return THINKING_LEVELS.some((level) => level === value)
}
