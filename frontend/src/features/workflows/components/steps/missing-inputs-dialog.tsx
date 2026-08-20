import {DialogShell} from '@/shared/components/dialog-shell'
import {Button} from '@/shared/ui/button'
import {pluralize} from '@/shared/lib/format'
import type {MissingStepInputs} from '@/features/workflows/step-inputs'

export function MissingInputsDialog({
    entries,
    onSelectStep,
    onRunAnyway,
    onClose,
}: {
    entries: MissingStepInputs[]
    onSelectStep: (stepId: string) => void
    onRunAnyway: () => void
    onClose: () => void
}) {
    return (
        <DialogShell
            onClose={onClose}
            title="Some inputs are empty"
            description={`Empty inputs on ${pluralize(entries.length, 'step')}. Pick one to fill, or run as is — the agent gets the prompt as written.`}
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button size="sm" onClick={onRunAnyway}>
                        Run anyway
                    </Button>
                </>
            }
        >
            <div className="flex flex-col gap-2 p-4">
                {entries.map((entry) => (
                    <MissingRow key={entry.step.id} entry={entry} onSelect={onSelectStep} />
                ))}
            </div>
        </DialogShell>
    )
}

function MissingRow({
    entry,
    onSelect,
}: {
    entry: MissingStepInputs
    onSelect: (stepId: string) => void
}) {
    const select = () => onSelect(entry.step.id)

    return (
        <button
            type="button"
            onClick={select}
            className="flex w-full flex-col gap-2 rounded-xl border border-border bg-card p-3 text-left transition-colors outline-none hover:border-foreground/25 hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-live"
        >
            <span className="truncate text-base font-medium">
                {entry.step.title || 'Untitled step'}
            </span>

            <span className="flex flex-wrap gap-1.5">
                {entry.inputs.map((input) => (
                    <span
                        key={input.key}
                        className="rounded-sm bg-state-approval-tint px-2 py-0.5 font-mono text-xs text-state-approval"
                    >
                        {input.label.trim() || input.key}
                    </span>
                ))}
            </span>
        </button>
    )
}
