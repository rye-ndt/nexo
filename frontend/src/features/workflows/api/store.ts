/**
 * Workflows are authored in this module's memory. Inside the Wails webview they
 * are hydrated from the stored drafts and every change is written back as one
 * JSON doc; under the plain vite dev server the mock workflows are the drafts.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {WORKFLOW_LIFECYCLES, WorkflowLifecycle} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {
    archiveRun,
    byRailRank,
    haltSteps,
    isLocked,
    isPausable,
    pauseRun,
} from '@/features/workflows/graph'
import {MOCK_WORKFLOWS} from '@/features/workflows/mock-workflows'
import type {Workflow} from '@/features/workflows/types'
import {DeleteWorkflowDraft, SaveWorkflowDraft, WorkflowDrafts} from '@wailsjs/go/wails_api/API'

type DraftStore = {
    seed(): Workflow[]
    load(): Promise<Workflow[]>
    save(workflow: Workflow): Promise<void>
    remove(workflowId: string): Promise<void>
}

const storedDrafts: DraftStore = {
    seed: () => [],
    load: async () => {
        const stored = await bridge(WorkflowDrafts)
        return stored.map((draft) => restore(JSON.parse(draft.doc) as Workflow))
    },
    save: async (workflow) => {
        await bridge(() => SaveWorkflowDraft(workflow.id, JSON.stringify(workflow)))
    },
    remove: async (workflowId) => {
        await bridge(() => DeleteWorkflowDraft(workflowId))
    },
}

const mockDrafts: DraftStore = {
    seed: () => structuredClone(MOCK_WORKFLOWS),
    load: async () => workflows,
    save: async () => {},
    remove: async () => {},
}

const drafts: DraftStore = hasWailsRuntime() ? storedDrafts : mockDrafts

export let workflows: Workflow[] = drafts.seed()

let hydrated = false

export function setWorkflows(next: Workflow[]) {
    workflows = next
}

export function prependWorkflow(workflow: Workflow): Workflow {
    const top = Math.min(0, ...workflows.map((existing) => existing.railRank ?? 0))
    const ranked = {...workflow, railRank: top - 1}

    setWorkflows([ranked, ...workflows])

    return ranked
}

export function findWorkflow(workflowId: string) {
    const workflow = workflows.find((workflow) => workflow.id === workflowId)
    if (!workflow) throw new Error(t('workflow.api.gone'))
    return workflow
}

export function findOpenWorkflow(workflowId: string) {
    const workflow = findWorkflow(workflowId)
    if (isLocked(workflow)) throw new Error(t('workflow.api.locked'))
    return workflow
}

export function findStep(workflow: Workflow, stepId: string) {
    const step = workflow.steps.find((step) => step.id === stepId)
    if (!step) throw new Error(t('workflow.api.stepGone'))
    return step
}

export function replaceWorkflow(next: Workflow) {
    workflows = workflows.map((workflow) => (workflow.id === next.id ? next : workflow))
    void saveDraft(next).catch(() => {})
    return structuredClone(next)
}

export async function saveDraft(workflow: Workflow) {
    await drafts.save(workflow)
}

export async function deleteDraft(workflowId: string) {
    await drafts.remove(workflowId)
}

export async function hydrate() {
    if (hydrated) return

    workflows = (await drafts.load()).sort(byRailRank)
    hydrated = true
}

/**
 * No agent survives the process, so a run that was still going comes back paused —
 * the same steps it would have lost to a pause. A cancelled workflow is terminal, so
 * it comes back exactly as it was stored, and a doc written before the lifecycle
 * existed comes back as a draft rather than crashing the rail.
 */
function restore(stored: Workflow): Workflow {
    if (!WORKFLOW_LIFECYCLES.includes(stored.lifecycle)) return archiveRun(stored)

    const lifecycle = stored.lifecycle

    if (lifecycle === WorkflowLifecycle.Cancelled) return {...stored, lifecycle}

    const halted = haltSteps({...stored, lifecycle})

    return isPausable(halted) ? pauseRun(halted) : halted
}
