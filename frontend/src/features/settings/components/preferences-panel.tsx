import {Boxes} from 'lucide-react'

import {
    AgentDefaultRow,
    MODEL_COLUMN,
    THINKING_COLUMN,
} from '@/features/settings/components/agent-default-row'
import {ModelPriceRow, PRICE_BAND} from '@/features/settings/components/model-price-row'
import type {AgentDefault, ModelPrice, TokenPrices} from '@/features/settings/types'
import {useAgentDefaults} from '@/features/settings/use-agent-defaults'
import {useAutopilot} from '@/features/settings/use-autopilot'
import {useLanguageChoice} from '@/features/settings/use-language'
import {useModelPrices} from '@/features/settings/use-model-prices'
import {HelpTip} from '@/shared/components/help-tip'
import {LANGUAGES, LANGUAGE_NAMES, type Language, type ThinkingLevel} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'
import {Button} from '@/shared/ui/button'
import {Switch} from '@/shared/ui/switch'

function LanguageSection() {
    const {language, saving, setLanguage} = useLanguageChoice()

    return (
        <section className="flex items-start justify-between gap-4 border-b border-border px-4 py-4">
            <div className="flex min-w-0 flex-col gap-1">
                <h3 className="text-lg font-medium">{t('settings.language.title')}</h3>
                <p className="text-sm text-muted-foreground">{t('settings.language.hint')}</p>
            </div>

            <div
                role="radiogroup"
                className="mt-1 flex shrink-0 rounded-lg border border-border p-0.5"
            >
                {LANGUAGES.map((option) => (
                    <LanguageOption
                        key={option}
                        language={option}
                        selected={option === language}
                        disabled={saving}
                        onSelect={setLanguage}
                    />
                ))}
            </div>
        </section>
    )
}

function LanguageOption({
    language,
    selected,
    disabled,
    onSelect,
}: {
    language: Language
    selected: boolean
    disabled: boolean
    onSelect: (language: Language) => void
}) {
    return (
        <button
            type="button"
            role="radio"
            aria-checked={selected}
            disabled={disabled}
            onClick={() => onSelect(language)}
            className={cn(
                'rounded-md px-3 py-1.5 text-base outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring/50 disabled:opacity-60',
                selected
                    ? 'bg-live-tint font-medium text-foreground'
                    : 'text-muted-foreground hover:text-foreground',
            )}
        >
            {LANGUAGE_NAMES[language]}
        </button>
    )
}

function AutopilotSection() {
    const {autopilot, loading, saving, setAutopilot} = useAutopilot()

    return (
        <section className="flex items-start justify-between gap-4 border-b border-border px-4 py-4">
            <div className="flex min-w-0 flex-col gap-1">
                <h3 id="autopilot-label" className="text-lg font-medium">
                    {t('settings.autopilot.title')}
                </h3>
                <p className="text-sm text-muted-foreground">{t('settings.autopilot.hint')}</p>
            </div>

            <Switch
                aria-labelledby="autopilot-label"
                className="mt-1 shrink-0"
                checked={autopilot}
                disabled={loading || saving}
                onCheckedChange={setAutopilot}
            />
        </section>
    )
}

function ModelPricingSection() {
    const {modelPrices, loading, setModelPrices} = useModelPrices()

    const changePrices = (modelPrice: ModelPrice) => (prices: TokenPrices) =>
        setModelPrices({model: modelPrice.model, prices})

    return (
        <section className="flex flex-col border-t border-border">
            <div className="flex flex-col gap-1 px-4 pt-4 pb-3">
                <h3 className="text-lg font-medium">{t('settings.prices.title')}</h3>
                <p className="text-sm text-muted-foreground">{t('settings.prices.hint')}</p>
            </div>

            <div className="flex items-center gap-3 border-y border-border px-4 py-2">
                <span className="micro-label min-w-0 flex-1">
                    {t('settings.prices.modelColumn')}
                </span>
                <span className={cn('micro-label shrink-0', PRICE_BAND)}>
                    {t('settings.prices.band')}
                </span>
            </div>

            {loading ? (
                <p className="px-4 py-3 text-base text-muted-foreground">
                    {t('settings.prices.loading')}
                </p>
            ) : (
                <div className="divide-y divide-border">
                    {modelPrices.map((modelPrice) => (
                        <ModelPriceRow
                            key={modelPrice.model}
                            modelPrice={modelPrice}
                            onChangePrices={changePrices(modelPrice)}
                        />
                    ))}
                </div>
            )}
        </section>
    )
}

function NoAgentLoggedIn({onShowAgents}: {onShowAgents: () => void}) {
    return (
        <div className="flex flex-col items-start gap-2 border-t border-border px-4 py-5">
            <span className="flex items-center gap-2 text-base font-medium">
                <Boxes className="size-4 shrink-0 text-muted-foreground" />
                {t('settings.defaults.emptyTitle')}
                <HelpTip term="agent" />
            </span>

            <p className="text-sm text-muted-foreground">{t('settings.defaults.emptyBody')}</p>

            <Button variant="outline" size="sm" className="mt-1" onClick={onShowAgents}>
                {t('settings.defaults.openAgents')}
            </Button>
        </div>
    )
}

export function PreferencesPanel({onShowAgents}: {onShowAgents: () => void}) {
    const {defaults, options, loading, pendingEffort, setAgentDefault} = useAgentDefaults()

    const empty = !loading && defaults.length === 0

    const changeModel = (agentDefault: AgentDefault) => (model: string) =>
        setAgentDefault({
            effort: agentDefault.effort,
            model,
            thinkingLevel: agentDefault.thinkingLevel,
        })

    const changeThinkingLevel = (agentDefault: AgentDefault) => (thinkingLevel: ThinkingLevel) =>
        setAgentDefault({
            effort: agentDefault.effort,
            model: agentDefault.model,
            thinkingLevel,
        })

    return (
        <div className="flex flex-col">
            <LanguageSection />

            <AutopilotSection />

            <section className="flex flex-col">
                <div className="flex flex-col gap-1 px-4 pt-4 pb-3">
                    <h3 className="text-lg font-medium">{t('settings.defaults.title')}</h3>
                    <p className="text-sm text-muted-foreground">{t('settings.defaults.hint')}</p>
                </div>

                {!empty && (
                    <div className="flex items-center gap-3 border-y border-border px-4 py-2">
                        <span className="micro-label min-w-0 flex-1">
                            {t('settings.defaults.effortColumn')}
                        </span>
                        <span className={cn('micro-label shrink-0', MODEL_COLUMN)}>
                            {t('settings.defaults.modelColumn')}
                        </span>
                        <span className={cn('flex shrink-0 items-center gap-2', THINKING_COLUMN)}>
                            <span className="micro-label">
                                {t('settings.defaults.thinkingColumn')}
                            </span>
                            <HelpTip term="thinking" />
                        </span>
                    </div>
                )}

                {loading || !options ? (
                    <p className="px-4 py-3 text-base text-muted-foreground">
                        {t('settings.defaults.loading')}
                    </p>
                ) : empty ? (
                    <NoAgentLoggedIn onShowAgents={onShowAgents} />
                ) : (
                    <div className="divide-y divide-border">
                        {defaults.map((agentDefault) => (
                            <AgentDefaultRow
                                key={agentDefault.effort}
                                agentDefault={agentDefault}
                                options={options}
                                saving={pendingEffort === agentDefault.effort}
                                onChangeModel={changeModel(agentDefault)}
                                onChangeThinkingLevel={changeThinkingLevel(agentDefault)}
                            />
                        ))}
                    </div>
                )}
            </section>

            <ModelPricingSection />
        </div>
    )
}
