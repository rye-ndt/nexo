import {Boxes} from 'lucide-react'

import {
    AgentDefaultRow,
    MODEL_COLUMN,
    THINKING_COLUMN,
} from '@/features/settings/components/agent-default-row'
import {Button} from '@/shared/ui/button'
import {ModelPriceRow, PRICE_BAND} from '@/features/settings/components/model-price-row'
import {HelpTip} from '@/shared/components/help-tip'
import {useAgentDefaults} from '@/features/settings/use-agent-defaults'
import {useAutopilot} from '@/features/settings/use-autopilot'
import {useModelPrices} from '@/features/settings/use-model-prices'
import {Switch} from '@/shared/ui/switch'
import type {ThinkingLevel} from '@/shared/lib/enums'
import {cn} from '@/shared/lib/utils'
import type {AgentDefault, ModelPrice, TokenPrices} from '@/features/settings/types'

function AutopilotSection() {
    const {autopilot, loading, saving, setAutopilot} = useAutopilot()

    return (
        <section className="flex items-start gap-4 border-b border-border px-4 py-4">
            <div className="flex min-w-0 flex-col gap-1">
                <h3 id="autopilot-label" className="text-lg font-medium">
                    Autopilot
                </h3>
                <p className="text-sm text-muted-foreground">
                    A step whose role pauses for your review stops when it finishes and waits for
                    you to read its handoff. Autopilot ignores that flag everywhere: every step
                    hands straight to the next one and the workflow runs to the end unattended.
                    Leave it off while you still want to read the work.
                </p>
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
                <h3 className="text-lg font-medium">What each model costs</h3>
                <p className="text-sm text-muted-foreground">
                    Prices are optional and only used to show what a run cost. A model is priced
                    once it carries both an input and an output price; leave a cached price blank to
                    charge cache reads at the input price.
                </p>
            </div>

            <div className="flex items-center gap-3 border-y border-border px-4 py-2">
                <span className="micro-label min-w-0 flex-1">Model</span>
                <span className={cn('micro-label shrink-0', PRICE_BAND)}>
                    US$ per million tokens
                </span>
            </div>

            {loading ? (
                <p className="px-4 py-3 text-base text-muted-foreground">Loading prices…</p>
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
                No agent is logged in
                <HelpTip term="agent" />
            </span>

            <p className="text-sm text-muted-foreground">
                Nexo fills the four levels from the first agent it finds logged in — Claude Code,
                then Codex, then Open Code. Log one in and the rows appear here, yours to change.
            </p>

            <Button variant="outline" size="sm" className="mt-1" onClick={onShowAgents}>
                Open Agents
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
            <AutopilotSection />

            <section className="flex flex-col">
                <div className="flex flex-col gap-1 px-4 pt-4 pb-3">
                    <h3 className="text-lg font-medium">Model per effort level</h3>
                    <p className="text-sm text-muted-foreground">
                        {empty
                            ? 'Every step inherits its model from the effort its role asks for. Which model that is comes from the agent you are logged in to.'
                            : 'Every step inherits its model from the effort its role asks for. Change a row and every step at that effort follows, including steps you have already drawn.'}
                    </p>
                </div>

                {!empty && (
                    <div className="flex items-center gap-3 border-y border-border px-4 py-2">
                        <span className="micro-label min-w-0 flex-1">Effort</span>
                        <span className={cn('micro-label shrink-0', MODEL_COLUMN)}>Model</span>
                        <span className={cn('flex shrink-0 items-center gap-2', THINKING_COLUMN)}>
                            <span className="micro-label">Thinking</span>
                            <HelpTip term="thinking" />
                        </span>
                    </div>
                )}

                {loading || !options ? (
                    <p className="px-4 py-3 text-base text-muted-foreground">
                        Loading preferences…
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
