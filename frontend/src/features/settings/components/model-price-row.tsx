import {useState, type ChangeEvent, type KeyboardEvent} from 'react'

import {t, type MessageKey} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'
import type {ModelPrice, TokenPrices} from '@/features/settings/types'

export const PRICE_BAND = 'w-[276px]'

const PRICE_FIELDS: {field: keyof TokenPrices; label: MessageKey; description: MessageKey}[] = [
    {field: 'input', label: 'settings.prices.input', description: 'settings.prices.inputFor'},
    {
        field: 'cachedInput',
        label: 'settings.prices.cachedInput',
        description: 'settings.prices.cachedInputFor',
    },
    {field: 'output', label: 'settings.prices.output', description: 'settings.prices.outputFor'},
]

/**
 * The three prices of one model are edited apart but saved together: a save carries
 * the whole set, so a second field committed before the first one's save has come
 * back cannot write the stale blank it would otherwise still be reading.
 *
 * The typed text is the truth until the stored prices actually change. Comparing
 * against the previous prop rather than the current one is what keeps a save in
 * flight — during which the prop still holds the old value — from pulling the field
 * back to what the user just replaced.
 */
export function ModelPriceRow({
    modelPrice,
    onChangePrices,
}: {
    modelPrice: ModelPrice
    onChangePrices: (prices: TokenPrices) => void
}) {
    const {modelLabel: label, prices} = modelPrice

    const [draft, setDraft] = useState(prices)
    const [stored, setStored] = useState(prices)
    const [editing, setEditing] = useState(false)

    if (!samePrices(stored, prices)) {
        setStored(prices)
        if (!editing) setDraft(prices)
    }

    const commit = () => {
        setEditing(false)

        if (!PRICE_FIELDS.every(({field}) => isPrice(draft[field]))) {
            setDraft(prices)
            return
        }

        const trimmed = trimPrices(draft)
        setDraft(trimmed)

        if (!samePrices(trimmed, prices)) onChangePrices(trimmed)
    }

    return (
        <div className="flex items-center gap-3 px-4 py-3">
            <span className="min-w-0 flex-1 truncate text-base font-medium">{label}</span>

            <div
                className={cn(
                    'flex h-12 shrink-0 divide-x divide-border rounded-lg border border-input bg-card',
                    PRICE_BAND,
                )}
            >
                {PRICE_FIELDS.map(({field, label: fieldLabel, description}) => {
                    const invalid = !isPrice(draft[field])

                    return (
                        <label
                            key={field}
                            className={cn(
                                'flex flex-1 flex-col justify-center gap-0.5 rounded-[9px] px-2.5 transition-colors',
                                'focus-within:relative focus-within:z-10 focus-within:ring-3 focus-within:ring-ring/50',
                                invalid &&
                                    'bg-destructive/5 text-destructive focus-within:ring-destructive/30',
                            )}
                        >
                            <span
                                className={cn(
                                    'text-xs leading-none select-none',
                                    invalid ? 'text-destructive' : 'text-muted-foreground',
                                )}
                            >
                                {t(fieldLabel)}
                            </span>
                            <span className="flex min-w-0 items-center font-mono text-base leading-none">
                                <span
                                    className={
                                        invalid ? 'text-destructive' : 'text-muted-foreground'
                                    }
                                >
                                    $
                                </span>
                                <input
                                    value={draft[field]}
                                    aria-label={t(description, {model: label})}
                                    aria-invalid={invalid}
                                    inputMode="decimal"
                                    placeholder="—"
                                    autoCapitalize="off"
                                    autoCorrect="off"
                                    className="w-full min-w-0 bg-transparent tabular-nums outline-none placeholder:text-muted-foreground"
                                    onFocus={() => setEditing(true)}
                                    onChange={(event: ChangeEvent<HTMLInputElement>) =>
                                        setDraft({...draft, [field]: event.target.value})
                                    }
                                    onBlur={commit}
                                    onKeyDown={(event: KeyboardEvent<HTMLInputElement>) => {
                                        if (event.key === 'Enter') event.currentTarget.blur()
                                    }}
                                />
                            </span>
                        </label>
                    )
                })}
            </div>
        </div>
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
