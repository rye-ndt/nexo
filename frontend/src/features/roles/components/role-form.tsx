import {Plus} from 'lucide-react'
import type {ChangeEvent, ReactNode} from 'react'

import {Field} from '@/shared/components/field'
import {HelpTip} from '@/shared/components/help-tip'
import {InputEditor} from '@/features/roles/components/input-editor'
import {PromptEditor} from '@/features/roles/components/prompt-editor'
import {Button} from '@/shared/ui/button'
import {Input} from '@/shared/ui/input'
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/shared/ui/select'
import {Switch} from '@/shared/ui/switch'
import {Textarea} from '@/shared/ui/textarea'
import type {GlossaryTerm} from '@/shared/lib/glossary'
import {cn} from '@/shared/lib/utils'
import type {Effort} from '@/shared/lib/enums'
import {EFFORTS, EFFORT_LABELS} from '@/shared/lib/enums'
import {emptyInput, NO_INSTRUCTIONS_ISSUE} from '@/features/roles/role'
import {INPUT_REF_HINT} from '@/features/roles/input-refs'
import type {Instruction, RoleDraft, RoleInput} from '@/features/roles/types'

const EMPTY_PROMPT: Instruction = {key: '', value: ''}

const STRUCTURE_EXAMPLE = `summary: one paragraph a non-programmer can follow
findings:
  - title: short label for this finding
    severity: high | medium | low
    detail: what was found and where
next_steps: what should happen now`

const STRUCTURE_HINT =
    'One field per line, written as name: what goes in it. Indent two spaces to nest, and start a line with a dash to describe one element of a list. Leave this empty and the step reports in its own words.'

function replaceAt<T>(list: T[], index: number, fields: Partial<T>) {
    return list.map((item, at) => (at === index ? {...item, ...fields} : item))
}

function removeAt<T>(list: T[], index: number) {
    return list.filter((_, at) => at !== index)
}

export function RoleForm({
    draft,
    onChange,
}: {
    draft: RoleDraft
    onChange: (draft: RoleDraft) => void
}) {
    const patch = (fields: Partial<RoleDraft>) => onChange({...draft, ...fields})

    const changeName = (event: ChangeEvent<HTMLInputElement>) => patch({name: event.target.value})
    const changeDescription = (event: ChangeEvent<HTMLTextAreaElement>) =>
        patch({description: event.target.value})
    const changeLevel = (value: string) => patch({effort: value as Effort})
    const changeRetryable = (retryable: boolean) => patch({retryable})
    const changeManualAccept = (pauseForReview: boolean) => patch({pauseForReview})
    const changeStructure = (event: ChangeEvent<HTMLTextAreaElement>) =>
        patch({outputStructure: event.target.value})

    const addInput = () => patch({inputs: [...draft.inputs, emptyInput()]})
    const changeInput = (index: number, fields: Partial<RoleInput>) =>
        patch({inputs: replaceAt(draft.inputs, index, fields)})
    const removeInput = (index: number) => patch({inputs: removeAt(draft.inputs, index)})

    const addPrompt = () => patch({instructions: [...draft.instructions, EMPTY_PROMPT]})
    const changePrompt = (index: number, fields: Partial<Instruction>) =>
        patch({instructions: replaceAt(draft.instructions, index, fields)})
    const removePrompt = (index: number) =>
        patch({instructions: removeAt(draft.instructions, index)})

    return (
        <div className="grid min-h-full grid-cols-1 gap-x-6 gap-y-6 p-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.15fr)]">
            <div className="flex min-h-0 flex-col gap-6">
                <Field htmlFor="role-name" label="Name">
                    <Input
                        id="role-name"
                        value={draft.name}
                        placeholder="Code reviewer"
                        onChange={changeName}
                    />
                </Field>

                <Field htmlFor="role-description" label="What it does">
                    <Textarea
                        id="role-description"
                        rows={2}
                        value={draft.description}
                        placeholder="Reads a diff and reports the defects it can prove."
                        onChange={changeDescription}
                    />
                </Field>

                <div className="flex flex-col gap-2">
                    <div className="flex items-end gap-3">
                        <Field
                            htmlFor="role-level"
                            label="Effort"
                            term="effort"
                            className="min-w-0 flex-1"
                        >
                            <Select value={draft.effort} onValueChange={changeLevel}>
                                <SelectTrigger id="role-level">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {EFFORTS.map((level) => (
                                        <SelectItem key={level} value={level}>
                                            {EFFORT_LABELS[level]}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </Field>

                        <label
                            htmlFor="role-retryable"
                            className="flex h-9 shrink-0 items-center gap-3 rounded-lg border border-border px-3"
                        >
                            <span className="text-base font-medium">Retry on failure</span>
                            <Switch
                                id="role-retryable"
                                checked={draft.retryable}
                                onCheckedChange={changeRetryable}
                            />
                        </label>
                    </div>

                    <label
                        htmlFor="role-manual-accept"
                        className="flex h-11 items-center justify-between gap-3 rounded-lg border border-border px-3"
                    >
                        <span className="flex items-center gap-2">
                            <span className="text-base font-medium">Pause for my review</span>
                            <HelpTip term="review" />
                        </span>
                        <Switch
                            id="role-manual-accept"
                            checked={draft.pauseForReview}
                            onCheckedChange={changeManualAccept}
                        />
                    </label>
                </div>

                <Section
                    title="Inputs"
                    term="input"
                    hint="What a step fills in."
                    onAdd={addInput}
                    addLabel="Add input"
                >
                    {draft.inputs.map((input, index) => (
                        <InputEditor
                            key={index}
                            index={index}
                            input={input}
                            onChange={changeInput}
                            onRemove={removeInput}
                        />
                    ))}
                </Section>

                <Field
                    htmlFor="role-structure"
                    label="Report format"
                    term="reportFormat"
                    hint={STRUCTURE_HINT}
                    className="min-h-0 flex-1"
                >
                    <Textarea
                        id="role-structure"
                        value={draft.outputStructure}
                        placeholder={STRUCTURE_EXAMPLE}
                        spellCheck={false}
                        className="min-h-44 flex-1 font-mono"
                        onChange={changeStructure}
                    />
                </Field>
            </div>

            <div className="flex min-h-0 flex-col gap-6">
                <Section
                    className="min-h-0 flex-1"
                    title="Instructions"
                    term="instructions"
                    hint={draft.instructions.length === 0 ? NO_INSTRUCTIONS_ISSUE : INPUT_REF_HINT}
                    onAdd={addPrompt}
                    addLabel="Add instruction"
                >
                    {draft.instructions.map((prompt, index) => (
                        <PromptEditor
                            key={index}
                            index={index}
                            prompt={prompt}
                            inputs={draft.inputs}
                            onChange={changePrompt}
                            onRemove={removePrompt}
                        />
                    ))}
                </Section>
            </div>
        </div>
    )
}

function Section({
    title,
    term,
    hint,
    addLabel,
    className,
    onAdd,
    children,
}: {
    title: string
    term?: GlossaryTerm
    hint?: string
    addLabel: string
    className?: string
    onAdd: () => void
    children: ReactNode
}) {
    return (
        <section className={cn('flex flex-col gap-3', className)}>
            <div className="flex items-center justify-between gap-3">
                <div className="flex min-w-0 flex-col gap-2">
                    <span className="flex items-center gap-2">
                        <span className="micro-label">{title}</span>
                        {term && <HelpTip term={term} />}
                    </span>
                    {hint && <span className="text-sm text-muted-foreground">{hint}</span>}
                </div>
                <Button variant="outline" size="sm" className="shrink-0" onClick={onAdd}>
                    <Plus />
                    {addLabel}
                </Button>
            </div>
            <div className="flex min-h-0 flex-1 flex-col gap-3">{children}</div>
        </section>
    )
}
