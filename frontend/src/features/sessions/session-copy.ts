export const FINALIZED_HINT = 'Finalized — duplicate to make changes.'

export const CANCELLED_HINT = 'Run cancelled — duplicate to start over.'

export const FINALIZE_CONFIRM = {
    description:
        'The graph and its directories lock for good. To change anything after this you have to duplicate the session. Nothing runs until you press Run.',
    confirmLabel: 'Finalize session',
}

export const PAUSED_HINT = 'Paused — resume when you are ready.'

export const PAUSE_CONFIRM = {
    title: 'Pause the run?',
    description:
        'Nodes running right now lose their work and start over when you resume — there will be no report for this attempt. Finished nodes keep theirs, and you can close the app or shut the machine down; the run picks up where it stopped.',
    confirmLabel: 'Pause run',
    dismissLabel: 'Keep running',
}

export const CANCEL_CONFIRM = {
    title: 'Cancel the run?',
    description:
        'The node running right now loses its work — there will be no report for it. Finished nodes keep theirs, and the session stays on the rail. You cannot resume a cancelled run; duplicate it to start over.',
    dismissLabel: 'Keep running',
}
