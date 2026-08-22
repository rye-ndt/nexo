import type {Catalog} from '@/shared/lib/i18n/messages/catalog'

export const AGENT_MESSAGES = {
    'agent.panel.title': {
        en: 'Agents that run your steps',
        vi: 'Agent chạy các bước của bạn',
    },
    'agent.panel.hint': {
        en: 'An agent has to be installed and logged in before a workflow can assign work to it.',
        vi: 'Agent phải được cài đặt và đăng nhập trước khi nhận việc.',
    },
    'agent.panel.empty': {
        en: 'No agents configured. Add one to config.yaml to see it here.',
        vi: 'Chưa có agent nào được hỗ trợ.',
    },
    'agent.panel.notInstalled': {en: 'Not installed', vi: 'Chưa cài đặt'},
    'agent.panel.meta': {en: 'v{version} · {count} running', vi: 'v{version} · {count} đang chạy'},

    'agent.install.downloading': {en: 'Downloading {percent}%', vi: 'Đang tải {percent}%'},

    'agent.login.body': {
        en: 'A login page opened in your browser. Most logins finish on their own once you approve, so paste a code below only if the page shows you one.',
        vi: 'Nếu đăng nhập không tự chuyển hướng, dán code trên màn hình vào đây.',
    },
    'agent.login.reopen': {en: 'Open the login page again', vi: 'Mở lại trang đăng nhập'},
    'agent.login.code': {en: 'Authorization code', vi: 'Mã ủy quyền'},
    'agent.login.submit': {en: 'Submit', vi: 'Gửi mã'},

    'agent.error.load': {en: 'Could not load the agents', vi: 'Không tải được danh sách agent'},
    'agent.error.codeIncomplete': {
        en: 'That code looks incomplete. Copy the whole value from the login page.',
        vi: 'Mã không đúng.',
    },
} as const satisfies Catalog
