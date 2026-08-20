export type GlossaryEntry = {
    title: string
    what: string
    why: string
}

export const GLOSSARY = {
    workflow: {
        title: 'Workflow',
        what: 'A graph of steps bound to one project folder.',
        why: 'You build it, lock it, run it, and review what it changed.',
    },
    step: {
        title: 'Step',
        what: 'One scoped piece of work, run by one agent.',
        why: 'Steps chain, so each one starts with what the ones before it learned.',
    },
    role: {
        title: 'Role',
        what: 'A reusable kind of worker: what it does, what it needs, how hard it tries.',
        why: 'Define a reviewer once and every step that reviews starts from it.',
    },
    agent: {
        title: 'Agent',
        what: 'The coding CLI that runs a step — Claude Code, OpenCode, or Codex.',
        why: 'Install and sign in once under Settings; steps pick which one they use.',
    },
    handoff: {
        title: 'Handoff',
        what: 'What a finished step writes down for the steps after it.',
        why: 'The next agent reads it instead of guessing, including what not to do.',
    },
    result: {
        title: 'Result',
        what: 'Everything a finished step produced: its handoffs, files, and cost.',
        why: 'It is the record you check before letting the run continue.',
    },
    lock: {
        title: 'Locking',
        what: 'Freezes the graph and the project folder so the workflow can run.',
        why: 'A run that could be edited underneath itself could not be trusted or replayed.',
    },
    approval: {
        title: 'Approval',
        what: 'An agent stopped mid-step to ask you something it should not decide alone.',
        why: 'It waits for your answer, then keeps going.',
    },
    review: {
        title: 'Review',
        what: 'A step finished and is holding everything downstream until you accept it.',
        why: 'Catches a bad handoff before it becomes the next three steps’ context.',
    },
    effort: {
        title: 'Effort',
        what: 'How hard the agent tries: Quick, Standard, Deep, or Exhaustive.',
        why: 'Sets the model and thinking budget. Higher costs more and takes longer.',
    },
    input: {
        title: 'Inputs',
        what: 'Named values a step fills in before it runs.',
        why: 'The role declares them once so each step only supplies what changes.',
    },
    instructions: {
        title: 'Instructions',
        what: 'The role’s prompt blocks, composed into what the agent receives.',
        why: 'Keeps standing rules in the role instead of retyping them per step.',
    },
    prompt: {
        title: 'Prompt',
        what: 'This step’s own wording, on top of its role’s instructions.',
        why: 'Change it to steer one step without touching every step sharing the role.',
    },
    projectFolder: {
        title: 'Project folder',
        what: 'The checkout agents read and change.',
        why: 'Everything a run touches lives here, so a revert never leaves the folder.',
    },
    reportFormat: {
        title: 'Report format',
        what: 'The shape you want the step to report back in.',
        why: 'A fixed shape is readable by the next step. Leave it empty for prose.',
    },
    retry: {
        title: 'Retry on failure',
        what: 'A step that fails is put back in the pool and tried again.',
        why: 'Turn it off for anything that should not run twice.',
    },
    revert: {
        title: 'Revert',
        what: 'Undoes every file change from this step onward.',
        why: 'Changes are stored as diffs, so this works without git.',
    },
    duplicate: {
        title: 'Duplicate',
        what: 'Copies the workflow with fresh ids and no run history.',
        why: 'The only way to change a locked workflow.',
    },
    context: {
        title: 'Context',
        what: 'How much of its context window the agent has used.',
        why: 'An agent that runs out mid-step loses the thread. Watch it on long steps.',
    },
    autopilot: {
        title: 'Autopilot',
        what: 'Answers routine approvals for you.',
        why: 'Lets a run finish unattended. Reviews still stop and wait.',
    },
    mcp: {
        title: 'MCP',
        what: 'A standard way for agents to call outside services.',
        why: 'Authorize a server once here; agents get a placeholder, never your key.',
    },
    thinking: {
        title: 'Thinking',
        what: 'How much reasoning the model does before it answers.',
        why: 'Set per effort level so you tune it once, not per step.',
    },
} as const satisfies Record<string, GlossaryEntry>

export type GlossaryTerm = keyof typeof GLOSSARY
