import {Field} from '@/shared/components/field'
import {HelpTip} from '@/shared/components/help-tip'
import {InputFields} from '@/features/roles/components/input-fields'
import {PromptField} from '@/features/roles/components/prompt-field'
import {INPUT_REF_HINT} from '@/features/roles/input-refs'
import {Input} from '@/shared/ui/input'
import type {FieldValue, RoleInput} from '@/features/roles/types'

export function StepForm({
    inputs,
    title,
    prompt,
    values,
    onTitleChange,
    onPromptChange,
    onValueChange,
}: {
    inputs: RoleInput[]
    title: string
    prompt: string
    values: Record<string, FieldValue>
    onTitleChange: (title: string) => void
    onPromptChange: (prompt: string) => void
    onValueChange: (key: string, value: FieldValue) => void
}) {
    return (
        <div className="flex flex-col gap-6 p-4">
            <Field htmlFor="step-title" label="Title">
                <Input
                    id="step-title"
                    value={title}
                    placeholder="Review the auth change"
                    onChange={(event) => onTitleChange(event.target.value)}
                />
            </Field>

            <Field
                htmlFor="step-prompt"
                label="Prompt"
                hint={`Starts as the role's instructions. Edit it to change this step only. ${INPUT_REF_HINT}`}
            >
                <PromptField
                    id="step-prompt"
                    className="max-h-64 min-h-32"
                    value={prompt}
                    inputs={inputs}
                    onChange={onPromptChange}
                />
            </Field>

            {inputs.length > 0 && (
                <section className="flex flex-col gap-3">
                    <span className="flex items-center gap-2">
                        <span className="micro-label">Inputs</span>
                        <HelpTip term="input" />
                    </span>
                    <InputFields inputs={inputs} values={values} onChange={onValueChange} />
                </section>
            )}
        </div>
    )
}
