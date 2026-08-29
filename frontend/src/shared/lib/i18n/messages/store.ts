import type {Catalog} from '@/shared/lib/i18n/messages/catalog'

export const STORE_MESSAGES = {
    'store.nav.workflows': {en: 'Workflows', vi: 'Workflow'},
    'store.nav.store': {en: 'Store', vi: 'Store'},

    'store.rail.title': {en: 'Store', vi: 'Store'},
    'store.rail.workflows': {en: 'Workflows', vi: 'Workflow'},
    'store.rail.roles': {en: 'Roles', vi: 'Vai trò'},

    'store.header.search': {en: 'Search the store', vi: 'Tìm trong store'},
    'store.header.clearSearch': {en: 'Clear the search', vi: 'Xóa tìm kiếm'},

    'store.workflows.title': {en: 'Ready-made work', vi: 'Việc có sẵn'},
    'store.workflows.subtitle': {
        en: 'Workflows you can point at a folder and run today. All free.',
        vi: 'Workflow chỉ cần trỏ vào một thư mục là chạy được. Tất cả đều miễn phí.',
    },
    'store.roles.title': {en: 'Ready-made roles', vi: 'Vai trò có sẵn'},
    'store.roles.subtitle': {
        en: 'Kinds of worker your steps can start from. All free.',
        vi: 'Các kiểu nhân sự mà bước của bạn có thể khởi đi. Tất cả đều miễn phí.',
    },

    'store.empty.title': {en: 'Nothing matches “{query}”', vi: 'Không có gì khớp “{query}”'},
    'store.empty.body': {en: 'Try a shorter word.', vi: 'Thử một từ ngắn hơn.'},

    'store.card.steps.one': {en: '{count} step', vi: '{count} bước'},
    'store.card.steps.other': {en: '{count} steps', vi: '{count} bước'},
    'store.card.usesRoles.one': {en: '{count} role', vi: '{count} vai trò'},
    'store.card.usesRoles.other': {en: '{count} roles', vi: '{count} vai trò'},
    'store.card.roles': {en: 'Roles it brings', vi: 'Vai trò đi kèm'},
    'store.card.roleOwned': {en: '{name}, already yours', vi: '{name}, bạn đã có'},
    'store.card.inputs.one': {en: '{count} input', vi: '{count} đầu vào'},
    'store.card.inputs.other': {en: '{count} inputs', vi: '{count} đầu vào'},
    'store.card.noInputs': {en: 'No inputs', vi: 'Không cần đầu vào'},
    'store.card.freeform': {en: 'Reports in its own words', vi: 'Báo cáo theo cách của nó'},
    'store.card.add': {en: 'Add', vi: 'Thêm'},
    'store.card.added': {en: 'Added', vi: 'Đã thêm'},
    'store.card.adding': {en: 'Adding…', vi: 'Đang thêm…'},
    'store.card.addWorkflow': {en: 'Add {name} to your rail', vi: 'Thêm {name} vào thanh bên'},
    'store.card.addRole': {en: 'Add {name} to your roles', vi: 'Thêm {name} vào vai trò của bạn'},
    'store.card.openWorkflow': {en: 'Preview {name}', vi: 'Xem trước {name}'},
    'store.card.openRole': {en: 'Preview {name}', vi: 'Xem trước {name}'},
    'store.card.openStep': {en: 'Preview {name}', vi: 'Xem trước {name}'},

    'store.preview.close': {en: 'Close', vi: 'Đóng'},
    'store.preview.role': {en: 'Role', vi: 'Vai trò'},
    'store.preview.prompt': {en: 'Prompt', vi: 'Prompt'},
    'store.preview.inputs': {en: 'What it asks you for', vi: 'Nó cần bạn điền gì'},
    'store.preview.noInputs': {en: 'Nothing to fill in.', vi: 'Không cần điền gì.'},
    'store.preview.instructions': {en: 'What it is told', vi: 'Nó được dặn những gì'},
    'store.preview.report': {en: 'What it reports', vi: 'Nó báo cáo những gì'},
    'store.preview.pauses': {en: 'pauses for review', vi: 'dừng chờ kiểm duyệt'},
    'store.preview.noRetry': {en: 'no retry', vi: 'không thử lại'},
    'store.preview.default': {en: 'Default: {value}', vi: 'Mặc định: {value}'},
    'store.preview.first': {en: 'Runs first.', vi: 'Chạy đầu tiên.'},
    'store.preview.waitsFor': {
        en: 'Starts from what {names} hands over.',
        vi: 'Bắt đầu từ những gì {names} bàn giao.',
    },

    'store.already.title': {en: 'Already in your library', vi: 'Đã có trong thư viện'},
    'store.already.body': {
        en: '“{name}” is already one of your roles. Adding it again would overwrite the copy you have, edits and all.',
        vi: '“{name}” đã là một vai trò của bạn. Thêm lại sẽ ghi đè bản bạn đang có, kể cả những sửa đổi.',
    },
    'store.already.detail': {en: 'Settings › Roles', vi: 'Cài đặt › Vai trò'},

    'store.add.description': {
        en: 'It lands in your rail as a draft you can edit before you run it.',
        vi: 'Nó sẽ vào thanh bên dưới dạng bản nháp, sửa được trước khi chạy.',
    },
    'store.add.steps': {en: 'What you get', vi: 'Bạn sẽ nhận được'},
    'store.add.rolesToo': {
        en: 'The roles these steps run on come with it. Roles you already have are left alone.',
        vi: 'Các vai trò mà những bước này dùng sẽ đi kèm. Vai trò bạn đã có sẽ được giữ nguyên.',
    },
    'store.add.confirm': {en: 'Add to my rail', vi: 'Thêm vào thanh bên'},
    'store.add.cancel': {en: 'Cancel', vi: 'Hủy'},

    'store.error.addWorkflow': {
        en: 'The workflow was not added.',
        vi: 'Chưa thêm được workflow.',
    },
} as const satisfies Catalog
