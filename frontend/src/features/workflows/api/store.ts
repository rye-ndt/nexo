/**
 * Workflows are authored in this module's memory. Inside the Wails webview they
 * are hydrated from the stored drafts and every change is written back as one
 * JSON doc.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {t} from '@/shared/lib/i18n'
import {byRailRank, isPausable, pauseRun} from '@/features/workflows/graph'
import {MOCK_WORKFLOWS} from '@/features/workflows/mock-workflows'
import type {Workflow} from '@/features/workflows/types'
import {SaveWorkflowDraft, WorkflowDrafts} from '@wailsjs/go/wails_api/API'

export let workflows: Workflow[] = hasWailsRuntime() ? [] : structuredClone(MOCK_WORKFLOWS)

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
    if (workflow.locked) throw new Error(t('workflow.api.locked'))
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
    if (!hasWailsRuntime()) return

    await bridge(() => SaveWorkflowDraft(workflow.id, JSON.stringify(workflow)))
}

export async function hydrate() {
    if (hydrated || !hasWailsRuntime()) return

    const drafts = await bridge(WorkflowDrafts)
    workflows = drafts.map((draft) => restore(JSON.parse(draft.doc) as Workflow)).sort(byRailRank)
    hydrated = true
}

/**
 * No agent survives the process, so a run that was still going comes back paused —
 * the same steps it would have lost to a pause. A cancelled workflow is terminal, so
 * it comes back exactly as it was stored.
 */
function restore(stored: Workflow): Workflow {
    if (stored.cancelled) return {...stored, cancelled: true}

    const halted = {...stored, cancelled: false, paused: false}

    return {...pauseRun(halted), paused: isPausable(halted)}
}
