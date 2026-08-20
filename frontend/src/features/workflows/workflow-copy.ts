export const LOCKED_HINT = 'Locked — duplicate to make changes.'

export const CANCELLED_HINT = 'Run cancelled — duplicate to start over.'

export const LOCK_CONFIRM = {
    description:
        'The graph and its project folder lock for good. To change anything after this you have to duplicate the workflow. Nothing runs until you press Run.',
    confirmLabel: 'Lock workflow',
}

export const PAUSED_HINT = 'Paused — resume when you are ready.'

export const PAUSE_CONFIRM = {
    title: 'Pause the run?',
    description:
        'Steps running right now lose their work and start over when you resume — there will be no result for this attempt. Finished steps keep theirs, and you can close the app or shut the machine down; the run picks up where it stopped.',
    confirmLabel: 'Pause run',
    dismissLabel: 'Keep running',
}

export const CANCEL_CONFIRM = {
    title: 'Cancel the run?',
    description:
        'The step running right now loses its work — there will be no result for it. Finished steps keep theirs, and the workflow stays on the rail. You cannot resume a cancelled run; duplicate it to start over.',
    dismissLabel: 'Keep running',
}
