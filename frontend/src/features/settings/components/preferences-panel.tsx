import {
    AgentDefaultRow,
    EFFORT_COLUMN,
    MODEL_COLUMN,
} from '@/features/settings/components/agent-default-row'
import {ModelPriceRow, PRICE_BAND} from '@/features/settings/components/model-price-row'
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
                    A node whose template requires manual acceptance stops when it finishes and
                    waits for you to read its handover. Autopilot ignores that flag everywhere:
                    every node hands straight to the next one and the session runs to the end
                    unattended. Leave it off while you still want to read the work.
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

export function PreferencesPanel() {
    const {defaults, options, loading, pendingTaskLevel, setAgentDefault} = useAgentDefaults()

    const changeModel = (agentDefault: AgentDefault) => (model: string) =>
        setAgentDefault({
            taskLevel: agentDefault.taskLevel,
            model,
            thinkingLevel: agentDefault.thinkingLevel,
        })

    const changeThinkingLevel = (agentDefault: AgentDefault) => (thinkingLevel: ThinkingLevel) =>
        setAgentDefault({
            taskLevel: agentDefault.taskLevel,
            model: agentDefault.model,
            thinkingLevel,
        })

    return (
        <div className="flex flex-col">
            <AutopilotSection />

            <section className="flex flex-col">
                <div className="flex flex-col gap-1 px-4 pt-4 pb-3">
                    <h3 className="text-lg font-medium">Model per task level</h3>
                    <p className="text-sm text-muted-foreground">
                        Every node inherits its model and effort from its task level. Change a row
                        and every node at that level follows, including nodes you have already
                        drawn.
                    </p>
                </div>

                <div className="flex items-center gap-3 border-y border-border px-4 py-2">
                    <span className="micro-label min-w-0 flex-1">Task level</span>
                    <span className={cn('micro-label shrink-0', MODEL_COLUMN)}>Model</span>
                    <span className={cn('micro-label shrink-0', EFFORT_COLUMN)}>Effort</span>
                </div>

                {loading || !options ? (
                    <p className="px-4 py-3 text-base text-muted-foreground">
                        Loading preferences…
                    </p>
                ) : (
                    <div className="divide-y divide-border">
                        {defaults.map((agentDefault) => (
                            <AgentDefaultRow
                                key={agentDefault.taskLevel}
                                agentDefault={agentDefault}
                                options={options}
                                saving={pendingTaskLevel === agentDefault.taskLevel}
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
