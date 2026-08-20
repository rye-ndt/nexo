import {useId} from 'react'
import type {ChangeEvent} from 'react'
import {Check, ChevronDown, File} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/shared/ui/dropdown-menu'
import {Input} from '@/shared/ui/input'
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/shared/ui/select'
import {Textarea} from '@/shared/ui/textarea'
import {chooseFile} from '@/shared/api/dialogs'
import {reportError} from '@/shared/lib/error-bus'
import {InputType} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {chosenOptions, joinOptions} from '@/features/roles/role'
import type {FieldValue, RoleInput} from '@/features/roles/types'

export function InputFields({
    inputs,
    values,
    disabled = false,
    onChange,
}: {
    inputs: RoleInput[]
    values: Record<string, FieldValue>
    disabled?: boolean
    onChange: (key: string, value: FieldValue) => void
}) {
    return (
        <div className="flex flex-col gap-3">
            {inputs.map((input) => (
                <InputField
                    key={input.key}
                    input={input}
                    value={values[input.key]}
                    disabled={disabled}
                    onChange={onChange}
                />
            ))}
        </div>
    )
}

function InputField({
    input,
    value,
    disabled,
    onChange,
}: {
    input: RoleInput
    value: FieldValue | undefined
    disabled: boolean
    onChange: (key: string, value: FieldValue) => void
}) {
    const id = useId()
    const label = input.label.trim() || input.key
    const text = String(value ?? '')

    const change = (next: FieldValue) => onChange(input.key, next)
    const changeEvent = (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
        change(event.target.value)

    return (
        <div className="flex flex-col gap-2">
            <div className="flex items-baseline justify-between gap-3">
                <label htmlFor={id} className="text-base font-medium">
                    {label}
                    {input.required && <span className="text-destructive"> *</span>}
                </label>
                <span className="shrink-0 font-mono text-sm text-muted-foreground">
                    {input.key}
                </span>
            </div>

            {input.type === InputType.Textarea && (
                <Textarea
                    id={id}
                    rows={3}
                    value={text}
                    disabled={disabled}
                    onChange={changeEvent}
                />
            )}

            {input.type === InputType.Boolean && (
                <Select
                    value={value === true ? 'true' : 'false'}
                    disabled={disabled}
                    onValueChange={(next) => change(next === 'true')}
                >
                    <SelectTrigger id={id}>
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="true">true</SelectItem>
                        <SelectItem value="false">false</SelectItem>
                    </SelectContent>
                </Select>
            )}

            {input.type === InputType.Select && (
                <Select value={text} disabled={disabled} onValueChange={change}>
                    <SelectTrigger id={id}>
                        <SelectValue placeholder={t('role.field.pickOne')} />
                    </SelectTrigger>
                    <SelectContent>
                        {(input.options ?? []).map((option) => (
                            <SelectItem key={option} value={option}>
                                {option}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            )}

            {input.type === InputType.MultiSelect && (
                <MultiChoiceField
                    id={id}
                    options={input.options ?? []}
                    value={text}
                    disabled={disabled}
                    onChange={change}
                />
            )}

            {input.type === InputType.File && (
                <FileField
                    id={id}
                    label={label}
                    value={text}
                    disabled={disabled}
                    onChange={change}
                />
            )}

            {(input.type === InputType.Text || input.type === InputType.Number) && (
                <Input
                    id={id}
                    type={input.type}
                    className="font-mono"
                    value={text}
                    disabled={disabled}
                    onChange={changeEvent}
                />
            )}
        </div>
    )
}

function MultiChoiceField({
    id,
    options,
    value,
    disabled,
    onChange,
}: {
    id: string
    options: string[]
    value: string
    disabled: boolean
    onChange: (value: string) => void
}) {
    const picked = chosenOptions(value)

    const toggle = (option: string) =>
        onChange(
            joinOptions(
                picked.includes(option)
                    ? picked.filter((other) => other !== option)
                    : options.filter((other) => other === option || picked.includes(other)),
            ),
        )

    return (
        <DropdownMenu>
            <DropdownMenuTrigger
                id={id}
                disabled={disabled}
                className="flex h-9 w-full items-center justify-between gap-2 rounded-lg border border-input bg-card px-3 py-1 text-base transition-colors outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-input/30"
            >
                <span className={picked.length ? 'truncate' : 'truncate text-muted-foreground'}>
                    {picked.length ? joinOptions(picked) : t('role.field.pickAny')}
                </span>
                <ChevronDown className="size-3.5 shrink-0 text-muted-foreground" />
            </DropdownMenuTrigger>

            <DropdownMenuContent className="p-1">
                {options.map((option) => (
                    <DropdownMenuItem
                        key={option}
                        className="justify-between py-2 pr-2 pl-2"
                        onSelect={(event) => {
                            event.preventDefault()
                            toggle(option)
                        }}
                    >
                        {option}
                        {picked.includes(option) && <Check className="size-3.5" />}
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    )
}

function FileField({
    id,
    label,
    value,
    disabled,
    onChange,
}: {
    id: string
    label: string
    value: string
    disabled: boolean
    onChange: (path: string) => void
}) {
    const choose = async () => {
        try {
            const path = await chooseFile(t('role.field.filePicker', {label}))
            if (path) onChange(path)
        } catch (cause) {
            reportError(cause, t('role.error.filePicker'))
        }
    }

    return (
        <div className="flex items-center gap-2">
            <span
                className={
                    value
                        ? 'min-w-0 flex-1 truncate rounded-lg border border-input px-3 py-2 font-mono text-base'
                        : 'min-w-0 flex-1 truncate rounded-lg border border-dashed border-input px-3 py-2 text-base text-muted-foreground'
                }
            >
                {value || t('role.field.noFile')}
            </span>

            <Button
                id={id}
                variant="outline"
                size="sm"
                className="shrink-0"
                disabled={disabled}
                onClick={choose}
            >
                <File />
                {value ? t('role.field.changeFile') : t('role.field.chooseFile')}
            </Button>
        </div>
    )
}
