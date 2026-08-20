import type {Catalog} from '@/shared/lib/i18n/messages/catalog'

export const APPROVAL_MESSAGES = {
    'approval.kind.decision.badge': {en: 'Decision', vi: 'Quyết định'},
    'approval.kind.decision.title': {
        en: 'The agent needs a decision',
        vi: 'Agent cần bạn quyết định',
    },
    'approval.kind.decision.blurb': {
        en: 'It reached a fork it should not pick on its own.',
        vi: 'Nó không biết nên quyết định thế nào.',
    },
    'approval.kind.permission.badge': {en: 'Permission', vi: 'Quyền'},
    'approval.kind.permission.title': {
        en: 'The agent needs permission',
        vi: 'Agent cần bạn cấp quyền',
    },
    'approval.kind.permission.blurb': {
        en: 'It wants to do something its step does not already cover.',
        vi: 'Nó muốn làm một việc nằm ngoài phạm vi của nó.',
    },

    'approval.dialog.description': {
        en: 'Nothing runs on it until you answer.',
        vi: 'Agent dừng ở đây cho đến khi bạn trả lời.',
    },
    'approval.dialog.waiting.one': {en: '{count} agent waiting', vi: '{count} agent đang chờ'},
    'approval.dialog.waiting.other': {en: '{count} agents waiting', vi: '{count} agent đang chờ'},
    'approval.dialog.blocked': {en: 'Blocked', vi: 'Bị chặn'},
    'approval.dialog.context': {en: 'Context', vi: 'Bối cảnh'},
    'approval.dialog.pickAny': {en: 'Pick any that apply', vi: 'Chọn các phương án phù hợp'},
    'approval.dialog.pickOne': {en: 'Pick one', vi: 'Chọn một phương án'},
    'approval.dialog.pickHint': {
        en: 'Approve needs at least one. Reject does not.',
        vi: 'Phê duyệt cần ít nhất một phương án. Từ chối thì không.',
    },
    'approval.dialog.noOptions': {
        en: 'This request arrived with no options, so it can only be rejected. Say why below and the agent picks it up from there.',
        vi: 'Yêu cầu này không kèm phương án nào nên chỉ có thể từ chối. Nêu lý do bên dưới để agent làm tiếp từ đó.',
    },
    'approval.dialog.comment': {en: 'Comment', vi: 'Ghi chú'},
    'approval.dialog.commentPlaceholder': {
        en: 'Anything the agent should know before it carries on.',
        vi: 'Điều agent cần biết trước khi làm tiếp.',
    },
    'approval.dialog.commentHint': {
        en: 'Sent either way - as guidance when you approve, as the reason when you reject.',
        vi: 'Luôn được gửi cho agent - kèm theo phê duyệt, hoặc làm lý do từ chối.',
    },
    'approval.dialog.notNow': {en: 'Not now', vi: 'Để sau'},
    'approval.dialog.reject': {en: 'Reject', vi: 'Từ chối'},
    'approval.dialog.rejecting': {en: 'Rejecting…', vi: 'Đang từ chối…'},
    'approval.dialog.approve': {en: 'Approve', vi: 'Phê duyệt'},
    'approval.dialog.approving': {en: 'Approving…', vi: 'Đang phê duyệt…'},

    'approval.error.load': {
        en: 'Could not load the waiting approvals',
        vi: 'Không tải được các phê duyệt đang chờ',
    },
    'approval.error.answer': {en: 'Could not send your answer', vi: 'Không gửi được câu trả lời'},
    'approval.error.gone': {
        en: 'That request is no longer waiting for an answer.',
        vi: 'Yêu cầu này không còn chờ trả lời nữa.',
    },

    'app.workflow.defaultName': {en: 'Workflow {count}', vi: 'Workflow {count}'},
    'app.transfer.importing.title': {en: 'Importing workflow', vi: 'Đang nhập workflow'},
    'app.transfer.importing.description': {
        en: 'Reading the file. Hold on.',
        vi: 'Đang đọc file.',
    },
    'app.transfer.exporting.title': {en: 'Exporting workflow', vi: 'Đang lưu workflow'},
    'app.transfer.exporting.description': {
        en: 'Writing the file. Hold on.',
        vi: 'Đang ghi file.',
    },

    'app.crash.title': {en: 'The app stopped', vi: 'Ứng dụng đã dừng'},
    'app.crash.hint': {
        en: 'Reload to start over. Your workflows are saved.',
        vi: 'Workflow của bạn đã được lưu.',
    },
    'app.crash.reload': {en: 'Reload', vi: 'Tải lại'},
} as const satisfies Catalog
