import {ChevronRight, Plus} from 'lucide-react'
import {useState, type ChangeEvent, type ReactNode} from 'react'

import {Field} from '@/shared/components/field'
import {HelpTip} from '@/shared/components/help-tip'
import {InputEditor} from '@/features/roles/components/input-editor'
import {PromptEditor} from '@/features/roles/components/prompt-editor'
import {Button} from '@/shared/ui/button'
import {Collapsible, CollapsibleContent, CollapsibleTrigger} from '@/shared/ui/collapsible'
import {Input} from '@/shared/ui/input'
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/shared/ui/select'
import {Switch} from '@/shared/ui/switch'
import {Textarea} from '@/shared/ui/textarea'
import type {GlossaryTerm} from '@/shared/lib/glossary'
import {t, tn} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'
import {EFFORTS, EFFORT_LABELS, Effort} from '@/shared/lib/enums'
import {isConflictAware, withConflictAwareness} from '@/features/roles/conflict-awareness'
import {advancedIssues, emptyInput, hasAdvancedSettings} from '@/features/roles/role'
import {INPUT_REF_HINT} from '@/features/roles/input-refs'
import type {Instruction, RoleDraft, RoleInput} from '@/features/roles/types'

const EMPTY_PROMPT: Instruction = {key: '', value: ''}

const STRUCTURE_EXAMPLE = `summary: one paragraph a non-programmer can follow
findings:
  - title: short label for this finding
    severity: high | medium | low
    detail: what was found and where
next_steps: what should happen now`

function replaceAt<T>(list: T[], index: number, fields: Partial<T>) {
    return list.map((item, at) => (at === index ? {...item, ...fields} : item))
}

function removeAt<T>(list: T[], index: number) {
    return list.filter((_, at) => at !== index)
}

function advancedSummary(draft: RoleDraft) {
    const parts = [
        draft.inputs.length > 0
            ? tn('role.form.inputCountOne', 'role.form.inputCountOther', draft.inputs.length)
            : t('role.card.noInputs'),
    ]

    if (draft.effort !== Effort.Standard) parts.push(t(EFFORT_LABELS[draft.effort]))
    if (draft.outputStructure.trim()) parts.push(t('role.card.structured'))
    if (!draft.retryable) parts.push(t('role.card.noRetry'))
    if (draft.pauseForReview) parts.push(t('role.form.reviewTag'))

    return parts.join(' · ')
}

export function RoleForm({
    draft,
    onChange,
}: {
    draft: RoleDraft
    onChange: (draft: RoleDraft) => void
}) {
    const [showAdvanced, setShowAdvanced] = useState(() => hasAdvancedSettings(draft))

    const patch = (fields: Partial<RoleDraft>) => onChange({...draft, ...fields})

    const changeName = (event: ChangeEvent<HTMLInputElement>) => patch({name: event.target.value})
    const changeDescription = (event: ChangeEvent<HTMLTextAreaElement>) =>
        patch({description: event.target.value})
    const changeLevel = (value: string) => patch({effort: value as Effort})
    const changeRetryable = (retryable: boolean) => patch({retryable})
    const changeManualAccept = (pauseForReview: boolean) => patch({pauseForReview})
    const changeConflictAware = (on: boolean) => onChange(withConflictAwareness(draft, on))
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

    // Something in here is blocking the save, so it cannot stay behind a closed panel.
    const advanced = showAdvanced || advancedIssues(draft).length > 0

    return (
        <div className="grid min-h-full grid-cols-1 gap-x-6 gap-y-6 p-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.15fr)]">
            <div className="flex min-h-0 flex-col gap-6">
                <Field htmlFor="role-name" label={t('role.form.name')}>
                    <Input
                        id="role-name"
                        value={draft.name}
                        placeholder={t('role.form.namePlaceholder')}
                        onChange={changeName}
                    />
                </Field>

                <Field htmlFor="role-description" label={t('role.form.description')}>
                    <Textarea
                        id="role-description"
                        rows={2}
                        value={draft.description}
                        placeholder={t('role.form.descriptionPlaceholder')}
                        onChange={changeDescription}
                    />
                </Field>

                <Collapsible
                    open={advanced}
                    onOpenChange={setShowAdvanced}
                    className="flex flex-col gap-4"
                >
                    <CollapsibleTrigger className="h-11 shrink-0 justify-between border-t border-border px-1">
                        <span className="flex items-center gap-2">
                            <ChevronRight
                                aria-hidden="true"
                                className={cn(
                                    'size-3.5 shrink-0 text-muted-foreground transition-transform duration-120 motion-reduce:transition-none',
                                    advanced && 'rotate-90',
                                )}
                            />
                            <span className="micro-label">{t('role.form.advanced')}</span>
                        </span>
                        {!advanced && (
                            <span className="min-w-0 truncate text-sm text-muted-foreground">
                                {hasAdvancedSettings(draft)
                                    ? advancedSummary(draft)
                                    : t('role.form.advancedHint')}
                            </span>
                        )}
                    </CollapsibleTrigger>

                    <CollapsibleContent className="flex flex-col gap-6">
                        <div className="flex items-end gap-3">
                            <Field
                                htmlFor="role-level"
                                label={t('role.form.effort')}
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
                                                {t(EFFORT_LABELS[level])}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </Field>

                            <label
                                htmlFor="role-retryable"
                                className="flex h-9 shrink-0 items-center gap-3 rounded-lg border border-border px-3"
                            >
                                <span className="text-base font-medium">
                                    {t('role.form.retryable')}
                                </span>
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
                                <span className="text-base font-medium">
                                    {t('role.form.pauseForReview')}
                                </span>
                                <HelpTip term="review" />
                            </span>
                            <Switch
                                id="role-manual-accept"
                                checked={draft.pauseForReview}
                                onCheckedChange={changeManualAccept}
                            />
                        </label>

                        <Section
                            title={t('role.form.inputs')}
                            term="input"
                            hint={t('role.form.inputsHint')}
                            onAdd={addInput}
                            addLabel={t('role.form.addInput')}
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
                            label={t('role.form.reportFormat')}
                            term="reportFormat"
                            hint={t('role.form.reportFormatHint')}
                        >
                            <Textarea
                                id="role-structure"
                                value={draft.outputStructure}
                                placeholder={STRUCTURE_EXAMPLE}
                                spellCheck={false}
                                className="min-h-44 font-mono"
                                onChange={changeStructure}
                            />
                        </Field>
                    </CollapsibleContent>
                </Collapsible>
            </div>

            <div className="flex min-h-0 flex-col gap-6">
                <label
                    htmlFor="role-conflict-aware"
                    className="flex h-11 items-center justify-between gap-3 rounded-lg border border-border px-3"
                >
                    <span className="text-base font-medium">{t('role.form.conflictAware')}</span>
                    <Switch
                        id="role-conflict-aware"
                        checked={isConflictAware(draft)}
                        onCheckedChange={changeConflictAware}
                    />
                </label>

                <Section
                    className="min-h-0 flex-1"
                    title={t('role.form.instructions')}
                    term="instructions"
                    hint={t(
                        draft.instructions.length === 0
                            ? 'role.issue.noInstructions'
                            : INPUT_REF_HINT,
                    )}
                    onAdd={addPrompt}
                    addLabel={t('role.form.addInstruction')}
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
