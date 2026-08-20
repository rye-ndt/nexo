import {EFFORTS, type Effort} from '@/shared/lib/enums'
import type {AgentDefault, ModelOption} from '@/features/settings/types'

/**
 * A step never stores a model. The Go side resolves one from the step level
 * when the workflow runs, so the frontend resolves the same way when it draws.
 */
export function agentDefaultFor(defaults: AgentDefault[], effort: Effort | undefined) {
    if (!effort) return undefined
    return defaults.find((agentDefault) => agentDefault.effort === effort)
}

/** Where a level sits on the ladder, 1 for the lightest. */
export function effortWeight(effort: Effort) {
    return EFFORTS.indexOf(effort) + 1
}

type HarnessDemand = {
    harness: string
    modelLabels: string[]
}

/**
 * Which harnesses the current preferences commit the user to, and the models
 * that pull each one in. A workflow cannot start until every one of these is
 * installed and logged in, because any step level may be used once steps exist.
 */
export function requiredHarnesses(
    defaults: AgentDefault[],
    models: ModelOption[],
): HarnessDemand[] {
    const harnessOf = new Map(models.map((option) => [option.model, option.harness]))
    const demands = new Map<string, string[]>()

    for (const agentDefault of defaults) {
        const harness = harnessOf.get(agentDefault.model)
        if (!harness) continue

        const labels = demands.get(harness) ?? []
        if (!labels.includes(agentDefault.modelLabel)) labels.push(agentDefault.modelLabel)
        demands.set(harness, labels)
    }

    return [...demands].map(([harness, modelLabels]) => ({harness, modelLabels}))
}
