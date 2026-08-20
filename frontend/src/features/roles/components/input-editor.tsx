import {useState, type ChangeEvent} from 'react'
import {X} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {Input} from '@/shared/ui/input'
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/shared/ui/select'
import {Switch} from '@/shared/ui/switch'
import {InputType, INPUT_TYPES, INPUT_TYPE_LABELS} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import type {RoleInput} from '@/features/roles/types'

export function InputEditor({
    index,
    input,
    onChange,
    onRemove,
}: {
    index: number
    input: RoleInput
    onChange: (index: number, fields: Partial<RoleInput>) => void
    onRemove: (index: number) => void
}) {
    const hasOptions = input.type === InputType.Select || input.type === InputType.MultiSelect
    const isBoolean = input.type === InputType.Boolean

    const change = (fields: Partial<RoleInput>) => onChange(index, fields)
    const remove = () => onRemove(index)

    const changeKey = (event: ChangeEvent<HTMLInputElement>) => change({key: event.target.value})
    const changeLabel = (event: ChangeEvent<HTMLInputElement>) =>
        change({label: event.target.value})
    const changeDefault = (event: ChangeEvent<HTMLInputElement>) =>
        change({default: event.target.value})

    const changeType = (value: string) =>
        change({type: value as InputType, default: undefined, options: undefined})

    const changeOptions = (options: string[]) => change({options})
    const changeBooleanDefault = (value: string) => change({default: value})
    const changeRequired = (required: boolean) => change({required})

    const booleanDefault = input.default === 'true' ? 'true' : 'false'

    return (
        <div className="flex flex-col gap-2 rounded-xl border border-border bg-muted/40 p-3">
            <div className="flex items-center gap-2">
                <Input
                    value={input.key}
                    placeholder="target_dir"
                    aria-label={t('role.input.key')}
                    className="h-8 flex-1 bg-background font-mono"
                    onChange={changeKey}
                />
                <Select value={input.type} onValueChange={changeType}>
                    <SelectTrigger
                        aria-label={t('role.input.type')}
                        className="h-8 w-32 shrink-0 bg-background"
                    >
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        {INPUT_TYPES.map((type) => (
                            <SelectItem key={type} value={type}>
                                {t(INPUT_TYPE_LABELS[type])}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={t('role.input.remove')}
                    onClick={remove}
                >
                    <X />
                </Button>
            </div>

            <Input
                value={input.label}
                placeholder={t('role.input.labelPlaceholder')}
                aria-label={t('role.input.label')}
                className="h-8 bg-background"
                onChange={changeLabel}
            />

            {hasOptions && <OptionsInput options={input.options ?? []} onChange={changeOptions} />}

            <div className="flex items-center gap-2">
                {isBoolean ? (
                    <Select value={booleanDefault} onValueChange={changeBooleanDefault}>
                        <SelectTrigger
                            aria-label={t('role.input.default')}
                            className="h-8 flex-1 bg-background font-mono"
                        >
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="false">false</SelectItem>
                            <SelectItem value="true">true</SelectItem>
                        </SelectContent>
                    </Select>
                ) : (
                    <Input
                        value={input.default ?? ''}
                        placeholder={t(
                            input.type === InputType.MultiSelect
                                ? 'role.input.multiDefault'
                                : 'role.input.default',
                        )}
                        aria-label={t('role.input.default')}
                        className="h-8 flex-1 bg-background font-mono"
                        onChange={changeDefault}
                    />
                )}

                <label className="flex h-8 shrink-0 items-center gap-3 rounded-lg border border-border bg-background px-3 text-base">
                    <span>{t('role.input.required')}</span>
                    <Switch checked={input.required} onCheckedChange={changeRequired} />
                </label>
            </div>
        </div>
    )
}

function OptionsInput({
    options,
    onChange,
}: {
    options: string[]
    onChange: (options: string[]) => void
}) {
    const [text, setText] = useState(() => options.join(', '))

    const commit = (event: ChangeEvent<HTMLInputElement>) => {
        const next = event.target.value
        setText(next)
        onChange(
            next
                .split(',')
                .map((option) => option.trim())
                .filter(Boolean),
        )
    }

    return (
        <Input
            value={text}
            placeholder="lenient, normal, strict"
            aria-label={t('role.input.options')}
            className="h-8 bg-background font-mono"
            onChange={commit}
        />
    )
}
