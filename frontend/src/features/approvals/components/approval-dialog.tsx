import {useState} from 'react'
import {Check} from 'lucide-react'

import {DialogShell} from '@/shared/components/dialog-shell'
import {Badge} from '@/shared/ui/badge'
import {Button} from '@/shared/ui/button'
import {Textarea} from '@/shared/ui/textarea'
import {useDecision} from '@/shared/hooks/use-decision'
import {useElapsed} from '@/shared/hooks/use-elapsed'
import {pluralize} from '@/shared/lib/format'
import {cn} from '@/shared/lib/utils'
import type {
    Approval,
    ApprovalAnswer,
    ApprovalKind,
    ApprovalOption,
} from '@/features/approvals/types'

const KIND_COPY: Record<ApprovalKind, {badge: string; title: string; blurb: string}> = {
    decision: {
        badge: 'Decision',
        title: 'The agent needs a decision',
        blurb: 'It reached a fork it should not pick on its own.',
    },
    permission: {
        badge: 'Permission',
        title: 'The agent needs permission',
        blurb: 'It wants to do something its step does not already cover.',
    },
}

function OptionRow({
    option,
    selected,
    multiSelect,
    disabled,
    onPick,
}: {
    option: ApprovalOption
    selected: boolean
    multiSelect: boolean
    disabled: boolean
    onPick: () => void
}) {
    return (
        <button
            type="button"
            aria-pressed={selected}
            disabled={disabled}
            onClick={onPick}
            className={cn(
                'flex w-full items-start gap-3 rounded-lg border p-3 text-left transition-colors outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50',
                selected
                    ? 'border-state-approval bg-state-approval-tint'
                    : 'border-border bg-card hover:border-border-strong hover:bg-muted',
            )}
        >
            <span
                aria-hidden="true"
                className={cn(
                    'mt-0.5 flex size-4 shrink-0 items-center justify-center border-[1.5px]',
                    multiSelect ? 'rounded-sm' : 'rounded-full',
                    selected
                        ? 'border-state-approval bg-state-approval text-white'
                        : 'border-border-strong',
                )}
            >
                {selected && <Check className="size-2.5" strokeWidth={3.5} />}
            </span>

            <span className="flex min-w-0 flex-1 flex-col gap-1">
                <span className="text-base font-medium">{option.label}</span>
                {option.description && (
                    <span className="text-sm text-muted-foreground">{option.description}</span>
                )}
            </span>
        </button>
    )
}

export function ApprovalDialog({
    approval,
    waiting,
    busy,
    onAnswer,
    onClose,
}: {
    approval: Approval
    waiting: number
    busy: boolean
    onAnswer: (answer: ApprovalAnswer) => void
    onClose: () => void
}) {
    const [picked, setPicked] = useState<string[]>([])
    const [guidance, setGuidance] = useState('')

    const copy = KIND_COPY[approval.kind]
    const blocked = useElapsed(approval.requestedAt)

    const decision = useDecision(busy, (approved) =>
        onAnswer({approved, optionIds: approved ? picked : [], guidance: guidance.trim()}),
    )

    const reject = () => decision.answer(false)
    const approve = () => decision.answer(true)

    const pick = (optionId: string) => {
        setPicked((current) => {
            if (!approval.multiSelect) return [optionId]

            return current.includes(optionId)
                ? current.filter((id) => id !== optionId)
                : [...current, optionId]
        })
    }

    return (
        <DialogShell
            onClose={onClose}
            title={copy.title}
            description="Nothing runs on it until you answer."
            term="approval"
            aside={
                waiting > 1 ? (
                    <span className="micro-label shrink-0">
                        {pluralize(waiting, 'agent')} waiting
                    </span>
                ) : undefined
            }
            footer={
                <>
                    <Button variant="ghost" size="sm" onClick={onClose}>
                        Not now
                    </Button>
                    <span className="flex-1" />
                    <Button variant="destructive" size="sm" disabled={busy} onClick={reject}>
                        {decision.labelOf(false, 'Reject', 'Rejecting…')}
                    </Button>
                    <Button size="sm" disabled={busy || picked.length === 0} onClick={approve}>
                        {decision.labelOf(true, 'Approve', 'Approving…')}
                    </Button>
                </>
            }
        >
            <div className="flex flex-col gap-6 p-4">
                <div className="flex flex-col gap-2 rounded-lg bg-state-approval-tint p-3">
                    <div className="flex items-center justify-between gap-3">
                        <Badge className="bg-state-approval text-white">{copy.badge}</Badge>
                        {blocked && (
                            <span className="flex items-baseline gap-2">
                                <span className="micro-label">Blocked</span>
                                <span className="font-mono text-lg tabular-nums text-state-approval">
                                    {blocked}
                                </span>
                            </span>
                        )}
                    </div>
                    <p className="text-sm text-muted-foreground">{copy.blurb}</p>
                </div>

                <p className="text-xl leading-[1.5]">{approval.question}</p>

                {approval.detail && (
                    <section className="flex flex-col gap-2">
                        <span className="micro-label">Context</span>
                        <p className="border-l border-border pl-3 text-base leading-[1.7] whitespace-pre-wrap">
                            {approval.detail}
                        </p>
                    </section>
                )}

                <section className="flex flex-col gap-3">
                    <div className="flex flex-col gap-2">
                        <span className="micro-label">
                            {approval.multiSelect ? 'Pick any that apply' : 'Pick one'}
                        </span>
                        <span className="text-sm text-muted-foreground">
                            Approve needs at least one. Reject does not.
                        </span>
                    </div>

                    {approval.options.length === 0 ? (
                        <p className="text-base text-muted-foreground">
                            This request arrived with no options, so it can only be rejected. Say
                            why below and the agent picks it up from there.
                        </p>
                    ) : (
                        <div className="flex flex-col gap-2">
                            {approval.options.map((option) => (
                                <OptionRow
                                    key={option.id}
                                    option={option}
                                    selected={picked.includes(option.id)}
                                    multiSelect={approval.multiSelect}
                                    disabled={busy}
                                    onPick={() => pick(option.id)}
                                />
                            ))}
                        </div>
                    )}
                </section>

                <section className="flex flex-col gap-2">
                    <span className="micro-label">Comment</span>
                    <Textarea
                        value={guidance}
                        disabled={busy}
                        placeholder="Anything the agent should know before it carries on."
                        onChange={(event) => setGuidance(event.target.value)}
                    />
                    <span className="text-sm text-muted-foreground">
                        Sent either way — as guidance when you approve, as the reason when you
                        reject.
                    </span>
                </section>
            </div>
        </DialogShell>
    )
}
