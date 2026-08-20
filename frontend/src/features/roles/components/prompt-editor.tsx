import type {ChangeEvent} from 'react'
import {X} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {Input} from '@/shared/ui/input'
import {PromptField} from '@/features/roles/components/prompt-field'
import type {Instruction, RoleInput} from '@/features/roles/types'

export function PromptEditor({
    index,
    prompt,
    inputs,
    onChange,
    onRemove,
}: {
    index: number
    prompt: Instruction
    inputs: RoleInput[]
    onChange: (index: number, fields: Partial<Instruction>) => void
    onRemove: (index: number) => void
}) {
    const remove = () => onRemove(index)

    const changeKey = (event: ChangeEvent<HTMLInputElement>) =>
        onChange(index, {key: event.target.value})

    const changeValue = (value: string) => onChange(index, {value})

    return (
        <div className="flex min-h-0 flex-1 flex-col gap-2 rounded-xl border border-border bg-muted/40 p-3">
            <div className="flex items-center gap-2">
                <Input
                    value={prompt.key}
                    placeholder="base"
                    aria-label="Instruction key"
                    className="h-8 flex-1 bg-background font-mono"
                    onChange={changeKey}
                />
                <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label="Remove instruction"
                    onClick={remove}
                >
                    <X />
                </Button>
            </div>
            <PromptField
                fill
                className="min-h-56"
                value={prompt.value}
                inputs={inputs}
                ariaLabel="Instruction text"
                placeholder="Fetch the ticket at https://jira/browse/{{ticket_id}} and summarise it."
                onChange={changeValue}
            />
        </div>
    )
}
