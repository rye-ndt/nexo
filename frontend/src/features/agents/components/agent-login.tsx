import {useState, type ChangeEvent, type KeyboardEvent} from 'react'

import {Button} from '@/shared/ui/button'
import {Input} from '@/shared/ui/input'
import {AGENT_ACTION_BUSY_LABELS, AgentAction} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import type {AgentControls} from '@/features/agents/controls'

export function AgentLogin({
    agentId,
    authUrl,
    controls,
}: {
    agentId: string
    authUrl: string
    controls: AgentControls
}) {
    const [code, setCode] = useState('')

    const trimmed = code.trim()
    const canSubmit = trimmed.length > 0 && !controls.busy
    const verifying = controls.actionOf(agentId) === AgentAction.Verify

    const submit = () => {
        if (!canSubmit) return

        controls.submitAuthCode(agentId, trimmed)
        setCode('')
    }

    const openUrl = () => controls.openAuthUrl(authUrl)

    const changeCode = (event: ChangeEvent<HTMLInputElement>) => setCode(event.target.value)

    const submitOnEnter = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key === 'Enter') submit()
    }

    return (
        <div className="flex flex-col gap-3 rounded-xl bg-muted/40 p-3 ring-1 ring-border">
            <p className="text-sm text-muted-foreground">
                {t('agent.login.body')}{' '}
                <button
                    type="button"
                    className="underline underline-offset-2 hover:text-foreground"
                    onClick={openUrl}
                >
                    {t('agent.login.reopen')}
                </button>
            </p>

            <div className="flex gap-2">
                <Input
                    value={code}
                    placeholder={t('agent.login.code')}
                    aria-label={t('agent.login.code')}
                    className="bg-background font-mono"
                    onChange={changeCode}
                    onKeyDown={submitOnEnter}
                />
                <Button size="sm" disabled={!canSubmit} onClick={submit}>
                    {verifying
                        ? t(AGENT_ACTION_BUSY_LABELS[AgentAction.Verify])
                        : t('agent.login.submit')}
                </Button>
            </div>
        </div>
    )
}
