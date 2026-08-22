import {t} from '@/shared/lib/i18n'

export function lockedHint() {
    return t('workflow.hint.locked')
}

export function cancelledHint() {
    return t('workflow.hint.cancelled')
}

export function pausedHint() {
    return t('workflow.hint.paused')
}

export function lockConfirm() {
    return {
        description: t('workflow.lock.description'),
        confirmLabel: t('workflow.lock.confirm'),
    }
}

export function unlockConfirm() {
    return {
        title: t('workflow.unlock.title'),
        description: t('workflow.unlock.description'),
        confirmLabel: t('workflow.unlock.confirm'),
        dismissLabel: t('workflow.unlock.dismiss'),
    }
}

export function pauseConfirm() {
    return {
        title: t('workflow.pause.title'),
        description: t('workflow.pause.description'),
        confirmLabel: t('workflow.pause.confirm'),
        dismissLabel: t('workflow.pause.dismiss'),
    }
}

export function cancelConfirm() {
    return {
        title: t('workflow.cancel.title'),
        description: t('workflow.cancel.description'),
        dismissLabel: t('workflow.cancel.dismiss'),
    }
}
