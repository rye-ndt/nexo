import {useRef, useState, type ChangeEvent, type KeyboardEvent} from 'react'

import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/shared/ui/select'
import {
    TASK_LEVEL_LABELS,
    TASK_LEVELS,
    THINKING_LEVEL_LABELS,
    type ThinkingLevel,
} from '@/shared/lib/enums'
import {taskLevelWeight} from '@/features/settings/agent-default'
import {Input} from '@/shared/ui/input'
import {cn} from '@/shared/lib/utils'
import type {AgentDefault, AgentDefaultOptions, TokenPrices} from '@/features/settings/types'

export const MODEL_COLUMN = 'w-[148px]'
export const EFFORT_COLUMN = 'w-[116px]'

const PRICE_FIELDS: {field: keyof TokenPrices; name: string; suffix: string}[] = [
    {field: 'input', name: 'input', suffix: 'in'},
    {field: 'cachedInput', name: 'cached input', suffix: 'cached'},
    {field: 'output', name: 'output', suffix: 'out'},
]

export function AgentDefaultRow({
    agentDefault,
    options,
    saving,
    onChangeModel,
    onChangeThinkingLevel,
    onChangePrices,
}: {
    agentDefault: AgentDefault
    options: AgentDefaultOptions
    saving: boolean
    onChangeModel: (model: string) => void
    onChangeThinkingLevel: (thinkingLevel: ThinkingLevel) => void
    onChangePrices: (prices: TokenPrices) => void
}) {
    const {taskLevel, prices} = agentDefault
    const label = TASK_LEVEL_LABELS[taskLevel]

    return (
        <div className="flex flex-col gap-2 px-4 py-3">
            <div className="flex items-center gap-3">
                <span className="flex min-w-0 flex-1 items-center gap-2">
                    <span className="truncate text-base font-medium">{label}</span>
                    <LevelWeight weight={taskLevelWeight(taskLevel)} />
                </span>

                <Select value={agentDefault.model} disabled={saving} onValueChange={onChangeModel}>
                    <SelectTrigger
                        className={cn('shrink-0', MODEL_COLUMN)}
                        aria-label={`${label} model`}
                    >
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        {options.models.map((option) => (
                            <SelectItem key={option.model} value={option.model}>
                                {option.label}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>

                <Select
                    value={agentDefault.thinkingLevel}
                    disabled={saving}
                    onValueChange={onChangeThinkingLevel}
                >
                    <SelectTrigger
                        className={cn('shrink-0', EFFORT_COLUMN)}
                        aria-label={`${label} effort`}
                    >
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        {options.thinkingLevels.map((level) => (
                            <SelectItem key={level} value={level}>
                                {THINKING_LEVEL_LABELS[level]}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            <div className="flex items-center justify-end gap-3">
                <PriceFields prices={prices} label={label} onCommit={onChangePrices} />
            </div>
        </div>
    )
}

/**
 * The three prices of one level are edited apart but saved together: a save carries
 * the whole set, so a second field committed before the first one's save has come
 * back cannot write the stale blank it would otherwise still be reading.
 *
 * The typed text is the truth until the stored prices actually change. Comparing
 * against the previous prop rather than the current one is what keeps a save in
 * flight — during which the prop still holds the old value — from pulling the field
 * back to what the user just replaced.
 */
function PriceFields({
    prices,
    label,
    onCommit,
}: {
    prices: TokenPrices
    label: string
    onCommit: (prices: TokenPrices) => void
}) {
    const [draft, setDraft] = useState(prices)
    const [stored, setStored] = useState(prices)
    const [editing, setEditing] = useState(false)

    if (!samePrices(stored, prices)) {
        setStored(prices)
        if (!editing) setDraft(prices)
    }

    // Blurring runs the commit synchronously, so escaping has to say so before it blurs
    // rather than by putting the stored prices back in a draft the commit cannot see yet.
    const escaped = useRef(false)

    const commit = () => {
        setEditing(false)

        const abandon = escaped.current || !PRICE_FIELDS.every(({field}) => isPrice(draft[field]))
        escaped.current = false

        if (abandon) {
            setDraft(prices)
            return
        }

        const trimmed = trimPrices(draft)
        setDraft(trimmed)

        if (!samePrices(trimmed, prices)) onCommit(trimmed)
    }

    const discard = (input: HTMLInputElement) => {
        escaped.current = true
        input.blur()
    }

    return (
        <>
            {PRICE_FIELDS.map(({field, name, suffix}) => (
                <span key={field} className="flex items-center gap-2">
                    <span className="relative inline-flex items-center">
                        <span className="pointer-events-none absolute inset-y-0 left-2 flex items-center text-sm text-muted-foreground">
                            $
                        </span>
                        <Input
                            value={draft[field]}
                            aria-label={`${label} ${name} price per million tokens`}
                            aria-invalid={!isPrice(draft[field])}
                            inputMode="decimal"
                            placeholder=""
                            className="h-8 w-[84px] pr-2 pl-5 text-right font-mono text-sm tabular-nums"
                            onFocus={() => setEditing(true)}
                            onChange={(event: ChangeEvent<HTMLInputElement>) =>
                                setDraft({...draft, [field]: event.target.value})
                            }
                            onBlur={commit}
                            onKeyDown={(event: KeyboardEvent<HTMLInputElement>) => {
                                if (event.key === 'Enter') event.currentTarget.blur()
                                if (event.key === 'Escape') discard(event.currentTarget)
                            }}
                        />
                    </span>
                    <span className="text-sm text-muted-foreground">{suffix}</span>
                </span>
            ))}
        </>
    )
}

function trimPrices(prices: TokenPrices): TokenPrices {
    return {
        input: prices.input.trim(),
        cachedInput: prices.cachedInput.trim(),
        output: prices.output.trim(),
    }
}

function samePrices(left: TokenPrices, right: TokenPrices) {
    return PRICE_FIELDS.every(({field}) => left[field].trim() === right[field].trim())
}

function isPrice(text: string) {
    const trimmed = text.trim()
    if (trimmed === '') return true

    const value = Number(trimmed)
    return Number.isFinite(value) && value >= 0
}

/** Four rungs, filled up to this level's place on the ladder. */
function LevelWeight({weight}: {weight: number}) {
    return (
        <span className="flex shrink-0 items-end gap-0.75" aria-hidden="true">
            {TASK_LEVELS.map((level, index) => (
                <span
                    key={level}
                    className={cn(
                        'h-2.5 w-0.75 rounded-full',
                        index < weight ? 'bg-foreground/60' : 'bg-border',
                    )}
                />
            ))}
        </span>
    )
}
