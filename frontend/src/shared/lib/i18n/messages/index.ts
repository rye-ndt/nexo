import {AGENT_MESSAGES} from '@/shared/lib/i18n/messages/agents'
import {APPROVAL_MESSAGES} from '@/shared/lib/i18n/messages/approvals'
import {ONBOARDING_MESSAGES} from '@/shared/lib/i18n/messages/onboarding'
import {ROLE_MESSAGES} from '@/shared/lib/i18n/messages/roles'
import {SETTINGS_MESSAGES} from '@/shared/lib/i18n/messages/settings'
import {SHARED_MESSAGES} from '@/shared/lib/i18n/messages/shared'
import {WORKFLOW_CANVAS_MESSAGES} from '@/shared/lib/i18n/messages/workflows-canvas'
import {WORKFLOW_MESSAGES} from '@/shared/lib/i18n/messages/workflows'

export const MESSAGES = {
    ...AGENT_MESSAGES,
    ...APPROVAL_MESSAGES,
    ...ONBOARDING_MESSAGES,
    ...ROLE_MESSAGES,
    ...SETTINGS_MESSAGES,
    ...SHARED_MESSAGES,
    ...WORKFLOW_MESSAGES,
    ...WORKFLOW_CANVAS_MESSAGES,
} as const

export type MessageKey = keyof typeof MESSAGES
