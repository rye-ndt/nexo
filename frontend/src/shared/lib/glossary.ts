import type {MessageKey} from '@/shared/lib/i18n'

export type GlossaryEntry = {
    title: MessageKey
    what: MessageKey
    why: MessageKey
}

export const GLOSSARY = {
    workflow: {
        title: 'glossary.workflow.title',
        what: 'glossary.workflow.what',
        why: 'glossary.workflow.why',
    },
    step: {
        title: 'glossary.step.title',
        what: 'glossary.step.what',
        why: 'glossary.step.why',
    },
    role: {
        title: 'glossary.role.title',
        what: 'glossary.role.what',
        why: 'glossary.role.why',
    },
    store: {
        title: 'glossary.store.title',
        what: 'glossary.store.what',
        why: 'glossary.store.why',
    },
    agent: {
        title: 'glossary.agent.title',
        what: 'glossary.agent.what',
        why: 'glossary.agent.why',
    },
    handoff: {
        title: 'glossary.handoff.title',
        what: 'glossary.handoff.what',
        why: 'glossary.handoff.why',
    },
    result: {
        title: 'glossary.result.title',
        what: 'glossary.result.what',
        why: 'glossary.result.why',
    },
    lock: {
        title: 'glossary.lock.title',
        what: 'glossary.lock.what',
        why: 'glossary.lock.why',
    },
    approval: {
        title: 'glossary.approval.title',
        what: 'glossary.approval.what',
        why: 'glossary.approval.why',
    },
    review: {
        title: 'glossary.review.title',
        what: 'glossary.review.what',
        why: 'glossary.review.why',
    },
    effort: {
        title: 'glossary.effort.title',
        what: 'glossary.effort.what',
        why: 'glossary.effort.why',
    },
    input: {
        title: 'glossary.input.title',
        what: 'glossary.input.what',
        why: 'glossary.input.why',
    },
    instructions: {
        title: 'glossary.instructions.title',
        what: 'glossary.instructions.what',
        why: 'glossary.instructions.why',
    },
    prompt: {
        title: 'glossary.prompt.title',
        what: 'glossary.prompt.what',
        why: 'glossary.prompt.why',
    },
    projectFolder: {
        title: 'glossary.projectFolder.title',
        what: 'glossary.projectFolder.what',
        why: 'glossary.projectFolder.why',
    },
    reportFormat: {
        title: 'glossary.reportFormat.title',
        what: 'glossary.reportFormat.what',
        why: 'glossary.reportFormat.why',
    },
    retry: {
        title: 'glossary.retry.title',
        what: 'glossary.retry.what',
        why: 'glossary.retry.why',
    },
    revert: {
        title: 'glossary.revert.title',
        what: 'glossary.revert.what',
        why: 'glossary.revert.why',
    },
    duplicate: {
        title: 'glossary.duplicate.title',
        what: 'glossary.duplicate.what',
        why: 'glossary.duplicate.why',
    },
    context: {
        title: 'glossary.context.title',
        what: 'glossary.context.what',
        why: 'glossary.context.why',
    },
    autopilot: {
        title: 'glossary.autopilot.title',
        what: 'glossary.autopilot.what',
        why: 'glossary.autopilot.why',
    },
    mcp: {
        title: 'glossary.mcp.title',
        what: 'glossary.mcp.what',
        why: 'glossary.mcp.why',
    },
    thinking: {
        title: 'glossary.thinking.title',
        what: 'glossary.thinking.what',
        why: 'glossary.thinking.why',
    },
} as const satisfies Record<string, GlossaryEntry>

export type GlossaryTerm = keyof typeof GLOSSARY
