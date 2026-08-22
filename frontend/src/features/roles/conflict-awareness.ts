/**
 * Conflict awareness is not a field on a role — it is one of the role's own
 * instructions, under a reserved key. The toggle in the role form adds or removes
 * that block, so it travels through storage, export and the run prompt with every
 * other instruction and needs nothing of its own anywhere else.
 */

import type {Instruction, RoleDraft} from '@/features/roles/types'

const CONFLICT_KEY = 'conflict_awareness'

const CONFLICT_INSTRUCTION: Instruction = {
    key: CONFLICT_KEY,
    value: `Other agents are working in this project folder at the same time as you, some of them filling the same role on other steps. Nothing is locked for you, so assume a file you did not create may be read or rewritten by someone else while you work.

Work so your changes cannot collide with theirs. Keep to the files your step actually needs. Prefer adding a new file over rewriting a shared one, and name what you create after your step so two agents never write the same path. Re-read a file immediately before you change it rather than trusting a copy you read earlier, and make each edit narrow enough that someone else's edit to the same file still applies. If your step cannot avoid changing something another agent is likely holding, say so in your handoff instead of forcing it through.`,
}

export function isConflictAware(draft: RoleDraft): boolean {
    return draft.instructions.some((instruction) => instruction.key === CONFLICT_KEY)
}

export function withConflictAwareness(draft: RoleDraft, on: boolean): RoleDraft {
    const without = draft.instructions.filter(
        (instruction) =>
            instruction.key !== CONFLICT_KEY && instruction.value !== CONFLICT_INSTRUCTION.value,
    )

    return {...draft, instructions: on ? [...without, CONFLICT_INSTRUCTION] : without}
}
