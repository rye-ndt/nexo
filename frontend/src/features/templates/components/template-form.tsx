import {Plus} from 'lucide-react'
import type {ChangeEvent, ReactNode} from 'react'

import {Field} from '@/shared/components/field'
import {ParamEditor} from '@/features/templates/components/param-editor'
import {PromptEditor} from '@/features/templates/components/prompt-editor'
import {Button} from '@/shared/ui/button'
import {Input} from '@/shared/ui/input'
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/shared/ui/select'
import {Switch} from '@/shared/ui/switch'
import {Textarea} from '@/shared/ui/textarea'
import type {TaskLevel} from '@/shared/lib/enums'
import {TASK_LEVELS, TASK_LEVEL_LABELS} from '@/shared/lib/enums'
import {emptyParam, NO_PROMPTS_ISSUE} from '@/features/templates/template'
import {PARAM_REF_HINT} from '@/features/templates/param-refs'
import type {SystemPrompt, TemplateDraft, TemplateParam} from '@/features/templates/types'

const EMPTY_PROMPT: SystemPrompt = {key: '', value: ''}

const STRUCTURE_EXAMPLE = `summary: one paragraph a non-programmer can follow
findings:
  - title: short label for this finding
    severity: high | medium | low
    detail: what was found and where
next_steps: what should happen now`

const STRUCTURE_HINT =
    'One field per line, written as name: what goes in it. Indent two spaces to nest, and start a line with a dash to describe one element of a list. Leave this empty and the node reports in its own words.'

function replaceAt<T>(list: T[], index: number, fields: Partial<T>) {
    return list.map((item, at) => (at === index ? {...item, ...fields} : item))
}

function removeAt<T>(list: T[], index: number) {
    return list.filter((_, at) => at !== index)
}

export function TemplateForm({
    draft,
    onChange,
}: {
    draft: TemplateDraft
    onChange: (draft: TemplateDraft) => void
}) {
    const patch = (fields: Partial<TemplateDraft>) => onChange({...draft, ...fields})

    const changeName = (event: ChangeEvent<HTMLInputElement>) => patch({name: event.target.value})
    const changeRole = (event: ChangeEvent<HTMLTextAreaElement>) =>
        patch({role: event.target.value})
    const changeLevel = (value: string) => patch({taskLevel: value as TaskLevel})
    const changeRetryable = (retryable: boolean) => patch({retryable})
    const changeManualAccept = (manualAcceptRequired: boolean) => patch({manualAcceptRequired})
    const changeStructure = (event: ChangeEvent<HTMLTextAreaElement>) =>
        patch({outputStructure: event.target.value})

    const addParam = () => patch({params: [...draft.params, emptyParam()]})
    const changeParam = (index: number, fields: Partial<TemplateParam>) =>
        patch({params: replaceAt(draft.params, index, fields)})
    const removeParam = (index: number) => patch({params: removeAt(draft.params, index)})

    const addPrompt = () => patch({systemPrompts: [...draft.systemPrompts, EMPTY_PROMPT]})
    const changePrompt = (index: number, fields: Partial<SystemPrompt>) =>
        patch({systemPrompts: replaceAt(draft.systemPrompts, index, fields)})
    const removePrompt = (index: number) =>
        patch({systemPrompts: removeAt(draft.systemPrompts, index)})

    return (
        <div className="flex flex-col gap-6 p-4">
            <Field htmlFor="template-name" label="Name">
                <Input
                    id="template-name"
                    value={draft.name}
                    placeholder="Code reviewer"
                    onChange={changeName}
                />
            </Field>

            <Field
                htmlFor="template-role"
                label="Role"
                hint="Shows on every node built from this template."
            >
                <Textarea
                    id="template-role"
                    rows={2}
                    value={draft.role}
                    placeholder="Reads a diff and reports the defects it can prove."
                    onChange={changeRole}
                />
            </Field>

            <div className="flex flex-col gap-2">
                <div className="flex items-end gap-3">
                    <Field htmlFor="template-level" label="Effort" className="min-w-0 flex-1">
                        <Select value={draft.taskLevel} onValueChange={changeLevel}>
                            <SelectTrigger id="template-level">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {TASK_LEVELS.map((level) => (
                                    <SelectItem key={level} value={level}>
                                        {TASK_LEVEL_LABELS[level]}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </Field>

                    <label
                        htmlFor="template-retryable"
                        className="flex h-9 shrink-0 items-center gap-3 rounded-lg border border-border px-3"
                    >
                        <span className="text-base font-medium">Retry on failure</span>
                        <Switch
                            id="template-retryable"
                            checked={draft.retryable}
                            onCheckedChange={changeRetryable}
                        />
                    </label>
                </div>

                <label
                    htmlFor="template-manual-accept"
                    className="flex items-center justify-between gap-3 rounded-lg border border-border px-3 py-2"
                >
                    <span className="flex min-w-0 flex-col gap-1">
                        <span className="text-base font-medium">Require manual acceptance</span>
                        <span className="text-sm text-muted-foreground">
                            The run stops here when this node finishes and waits for you to read the
                            handover and confirm. Autopilot ignores it and runs straight through.
                        </span>
                    </span>
                    <Switch
                        id="template-manual-accept"
                        checked={draft.manualAcceptRequired}
                        onCheckedChange={changeManualAccept}
                    />
                </label>
            </div>

            <Section
                title="Inputs"
                hint="What a node fills in. An input left empty reaches the agent as the reference itself."
                onAdd={addParam}
                addLabel="Add input"
            >
                {draft.params.map((param, index) => (
                    <ParamEditor
                        key={index}
                        index={index}
                        param={param}
                        onChange={changeParam}
                        onRemove={removeParam}
                    />
                ))}
            </Section>

            <Section
                title="Prompts"
                hint={
                    draft.systemPrompts.length === 0
                        ? NO_PROMPTS_ISSUE
                        : `${PARAM_REF_HINT} An input left empty reaches the agent as the reference itself.`
                }
                onAdd={addPrompt}
                addLabel="Add prompt"
            >
                {draft.systemPrompts.map((prompt, index) => (
                    <PromptEditor
                        key={index}
                        index={index}
                        prompt={prompt}
                        params={draft.params}
                        onChange={changePrompt}
                        onRemove={removePrompt}
                    />
                ))}
            </Section>

            <Field htmlFor="template-structure" label="Output" hint={STRUCTURE_HINT}>
                <Textarea
                    id="template-structure"
                    value={draft.outputStructure}
                    placeholder={STRUCTURE_EXAMPLE}
                    spellCheck={false}
                    className="min-h-44 font-mono"
                    onChange={changeStructure}
                />
            </Field>
        </div>
    )
}

function Section({
    title,
    hint,
    addLabel,
    onAdd,
    children,
}: {
    title: string
    hint?: string
    addLabel: string
    onAdd: () => void
    children: ReactNode
}) {
    return (
        <section className="flex flex-col gap-3">
            <div className="flex items-center justify-between gap-3">
                <div className="flex min-w-0 flex-col gap-2">
                    <span className="micro-label">{title}</span>
                    {hint && <span className="text-sm text-muted-foreground">{hint}</span>}
                </div>
                <Button variant="outline" size="sm" className="shrink-0" onClick={onAdd}>
                    <Plus />
                    {addLabel}
                </Button>
            </div>
            {children}
        </section>
    )
}
