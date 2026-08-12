import type {ChangeEvent} from 'react'
import {File} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {Input} from '@/shared/ui/input'
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/shared/ui/select'
import {Textarea} from '@/shared/ui/textarea'
import {chooseFile} from '@/shared/api/dialogs'
import {reportError} from '@/shared/lib/error-bus'
import {ParamType} from '@/shared/lib/enums'
import type {FieldValue, TemplateParam} from '@/features/templates/types'

export function ParamFields({
    params,
    values,
    disabled = false,
    onChange,
}: {
    params: TemplateParam[]
    values: Record<string, FieldValue>
    disabled?: boolean
    onChange: (key: string, value: FieldValue) => void
}) {
    return (
        <div className="flex flex-col gap-3">
            {params.map((param) => (
                <ParamField
                    key={param.key}
                    param={param}
                    value={values[param.key]}
                    disabled={disabled}
                    onChange={onChange}
                />
            ))}
        </div>
    )
}

function ParamField({
    param,
    value,
    disabled,
    onChange,
}: {
    param: TemplateParam
    value: FieldValue | undefined
    disabled: boolean
    onChange: (key: string, value: FieldValue) => void
}) {
    const id = `param-${param.key}`
    const label = param.label.trim() || param.key
    const text = String(value ?? '')

    const change = (next: FieldValue) => onChange(param.key, next)
    const changeEvent = (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
        change(event.target.value)

    return (
        <div className="flex flex-col gap-2">
            <div className="flex items-baseline justify-between gap-3">
                <label htmlFor={id} className="text-base font-medium">
                    {label}
                </label>
                <span className="shrink-0 font-mono text-sm text-muted-foreground">
                    {param.key}
                    {param.required && <span className="text-live">*</span>}
                </span>
            </div>

            {param.type === ParamType.Textarea && (
                <Textarea
                    id={id}
                    rows={3}
                    value={text}
                    disabled={disabled}
                    onChange={changeEvent}
                />
            )}

            {param.type === ParamType.Boolean && (
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

            {param.type === ParamType.Select && (
                <Select value={text} disabled={disabled} onValueChange={change}>
                    <SelectTrigger id={id}>
                        <SelectValue placeholder="Pick one" />
                    </SelectTrigger>
                    <SelectContent>
                        {(param.options ?? []).map((option) => (
                            <SelectItem key={option} value={option}>
                                {option}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            )}

            {param.type === ParamType.File && (
                <FileField
                    id={id}
                    label={label}
                    value={text}
                    disabled={disabled}
                    onChange={change}
                />
            )}

            {(param.type === ParamType.Text || param.type === ParamType.Number) && (
                <Input
                    id={id}
                    type={param.type}
                    className="font-mono"
                    value={text}
                    disabled={disabled}
                    onChange={changeEvent}
                />
            )}
        </div>
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
            const path = await chooseFile(`Choose a file for ${label}`)
            if (path) onChange(path)
        } catch (cause) {
            reportError(cause, 'Could not open the file picker')
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
                {value || 'No file chosen'}
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
                {value ? 'Change' : 'Choose'}
            </Button>
        </div>
    )
}
