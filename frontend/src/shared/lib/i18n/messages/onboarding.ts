import type {Catalog} from '@/shared/lib/i18n/messages/catalog'

export const ONBOARDING_MESSAGES = {
    'onboarding.welcome.eyebrow': {en: 'Welcome', vi: 'Chào mừng'},
    'onboarding.welcome.title': {
        en: 'Nexo runs a fleet of coding agents',
        vi: 'Nexo chạy workflow của bạn, từng bước một',
    },
    'onboarding.welcome.body': {
        en: 'You describe work as a graph of steps. Each step runs on one agent and hands what it learned to the next.',
        vi: 'Mô tả công việc của bạn thành nhiều bước. Mỗi bước chạy trên một agent, tuần tự đến hết.',
    },
    'onboarding.welcome.language': {en: 'Language', vi: 'Ngôn ngữ'},
    'onboarding.welcome.next': {en: 'Next', vi: 'Tiếp tục'},

    'onboarding.agent.eyebrow': {en: 'Your agent', vi: 'Agent của bạn'},
    'onboarding.agent.title': {
        en: 'Pick the CLI that runs your steps',
        vi: 'Chọn CLI bạn quen thuộc',
    },
    'onboarding.agent.body': {
        en: 'Nexo installs just this one. Add the others any time in Settings › Agents.',
        vi: 'Bạn có thể thêm những Agent CLI còn lại bất cứ lúc nào trong Cài đặt › Agent.',
    },
    'onboarding.agent.recommended': {en: 'Recommended', vi: 'Nên chọn'},
    'onboarding.agent.installed': {en: 'Already installed', vi: 'Đã cài đặt'},
    'onboarding.agent.install': {en: 'Install {name}', vi: 'Cài {name}'},
    'onboarding.agent.continue': {en: 'Continue with {name}', vi: 'Tiếp tục với {name}'},
    'onboarding.agent.claude_code.blurb': {
        en: 'By Anthropic. The best-supported agent in Nexo.',
        vi: 'Của Anthropic. Agent được hỗ trợ tốt nhất.',
    },
    'onboarding.agent.codex.blurb': {en: 'By OpenAI.', vi: 'Của OpenAI.'},
    'onboarding.agent.open_code.blurb': {
        en: 'Open source. Bring your own model key.',
        vi: 'Mã nguồn mở. Dùng key của riêng bạn.',
    },

    'onboarding.install.eyebrow': {en: 'Setting up', vi: 'Đang thiết lập'},
    'onboarding.install.title': {en: 'Getting {name} ready', vi: 'Đang chuẩn bị {name}'},
    'onboarding.install.busy': {en: 'Installing {name}…', vi: 'Đang cài {name}…'},
    'onboarding.install.failed': {en: '{name} didn’t install.', vi: 'Không cài được {name}.'},
    'onboarding.install.retry': {en: 'Try again', vi: 'Thử lại'},
    'onboarding.install.pickAnother': {en: 'Pick a different agent', vi: 'Chọn agent khác'},

    'onboarding.login.eyebrow': {en: 'Sign in', vi: 'Đăng nhập'},
    'onboarding.login.title': {en: 'Sign in to {name}', vi: 'Đăng nhập {name}'},
    'onboarding.login.body': {
        en: 'Nexo drives {name} with your own account. Nothing runs until you are signed in.',
        vi: 'Nexo điều khiển {name} bằng tài khoản của bạn. Yêu cầu đăng nhập.',
    },
    'onboarding.login.ready': {en: 'Ready', vi: 'Sẵn sàng'},
    'onboarding.login.signedIn': {en: 'Signed in', vi: 'Đã đăng nhập'},
    'onboarding.login.notSignedIn': {en: 'Not signed in', vi: 'Chưa đăng nhập'},
    'onboarding.login.hint': {en: 'Sign in to continue.', vi: 'Đăng nhập để tiếp tục.'},
    'onboarding.login.start': {en: 'Start the tour', vi: 'Bắt đầu hướng dẫn'},

    'onboarding.tour.next': {en: 'Next', vi: 'Tiếp'},
    'onboarding.tour.skip': {en: 'Skip tour', vi: 'Bỏ qua'},
    'onboarding.tour.done': {en: 'Start building', vi: 'Bắt đầu'},
    'onboarding.tour.role.title': {en: 'Roles', vi: 'Vai trò'},
    'onboarding.tour.role.body': {
        en: 'A role is a kind of worker you write once - what it does, how hard it tries, and the inputs it needs. Every step starts from one.',
        vi: 'Vai trò là một kiểu nhân sự bạn định nghĩa một lần - nó làm gì, cố gắng tới đâu, cần đầu vào nào. Mọi bước ứng với một vai trò.',
    },
    'onboarding.tour.workflow.title': {en: 'Workflows', vi: 'Workflow'},
    'onboarding.tour.workflow.body': {
        en: 'A workflow is a graph of steps pointed at one project folder. This one ships with the app — pick a folder and run it.',
        vi: 'Luồng công việc là một sơ đồ các bước trỏ vào một thư mục dự án. Luồng này có sẵn trong ứng dụng - chọn thư mục rồi chạy thử.',
    },
    'onboarding.tour.step.title': {en: 'Steps', vi: 'Bước'},
    'onboarding.tour.step.body': {
        en: 'Each step runs on one agent, then hands what it learned to the steps wired after it. Click one to change its prompt or read its result.',
        vi: 'Mỗi bước chạy trên một Agent, rồi handoff kết quả cho bước tiếp theo. Bấm vào một bước để sửa prompt hoặc đọc kết quả.',
    },
} as const satisfies Catalog
