import type {Catalog} from '@/shared/lib/i18n/messages/catalog'

export const SETTINGS_MESSAGES = {
    'settings.language.title': {en: 'Language', vi: 'Ngôn ngữ'},
    'settings.language.hint': {
        en: 'What Nexo speaks. Agents keep speaking whatever your prompts do.',
        vi: 'Ngôn ngữ hiển thị. Agent vẫn nói theo ngôn ngữ trong prompt.',
    },

    'settings.dialog.title': {en: 'Settings', vi: 'Cài đặt'},
    'settings.dialog.description': {
        en: 'Settings shared by every workflow.',
        vi: 'Cài đặt chung cho mọi workflow.',
    },
    'settings.tab.preferences': {en: 'Preferences', vi: 'Tùy chọn'},
    'settings.tab.roles': {en: 'Roles', vi: 'Vai trò'},
    'settings.tab.mcp': {en: 'MCP', vi: 'MCP'},
    'settings.tab.agents': {en: 'Agents', vi: 'Agent'},

    'settings.autopilot.title': {en: 'Autopilot', vi: 'Autopilot'},
    'settings.autopilot.hint': {
        en: 'No manual approval needed.',
        vi: 'Không cần phê duyệt thủ công.',
    },

    'settings.maxAgents.title': {en: 'Max running agents', vi: 'Số agent chạy cùng lúc'},
    'settings.maxAgents.hint': {
        en: 'How many agents may run at the same time across every workflow.',
        vi: 'Số agent được phép chạy cùng lúc trên toàn bộ workflow.',
    },
    'settings.maxAgents.fewer': {en: 'Fewer agents at once', vi: 'Giảm số agent chạy cùng lúc'},
    'settings.maxAgents.more': {en: 'More agents at once', vi: 'Tăng số agent chạy cùng lúc'},

    'settings.defaults.title': {en: 'Model per effort level', vi: 'Model cho từng mức nỗ lực'},
    'settings.defaults.hint': {
        en: 'Simplified model selections.',
        vi: 'Chọn model một cách đơn giản.',
    },
    'settings.defaults.effortColumn': {en: 'Effort', vi: 'Mức nỗ lực'},
    'settings.defaults.modelColumn': {en: 'Model', vi: 'Model'},
    'settings.defaults.thinkingColumn': {en: 'Thinking', vi: 'Suy luận'},
    'settings.defaults.loading': {en: 'Loading preferences…', vi: 'Đang tải tùy chọn…'},
    'settings.defaults.modelFor': {en: '{effort} model', vi: 'Model cho mức {effort}'},
    'settings.defaults.thinkingFor': {en: '{effort} thinking', vi: 'Suy luận cho mức {effort}'},
    'settings.defaults.emptyTitle': {
        en: 'No agent is logged in',
        vi: 'Chưa có agent nào đăng nhập',
    },
    'settings.defaults.emptyBody': {
        en: 'Nexo fills the four levels from the first agent it finds logged in — Claude Code, then Codex, then Open Code. Log one in and the rows appear here, yours to change.',
        vi: 'Nexo điền cả bốn mức từ agent đầu tiên đang đăng nhập — Claude Code, rồi Codex, rồi Open Code. Đăng nhập một agent là các dòng hiện ra ở đây để bạn sửa.',
    },
    'settings.defaults.openAgents': {en: 'Open Agents', vi: 'Mở mục Agent'},

    'settings.prices.title': {en: 'What each model costs', vi: 'Chi phí của từng model'},
    'settings.prices.hint': {
        en: 'Optional, to calculate cost per workflow (API Pricing).',
        vi: 'Tùy chọn, để tính chi phí mỗi workflow (theo giá API).',
    },
    'settings.prices.modelColumn': {en: 'Model', vi: 'Model'},
    'settings.prices.band': {en: 'US$ per million tokens', vi: 'US$ mỗi triệu token'},
    'settings.prices.loading': {en: 'Loading prices…', vi: 'Đang tải bảng giá…'},
    'settings.prices.input': {en: 'in', vi: 'vào'},
    'settings.prices.cachedInput': {en: 'cache', vi: 'cache'},
    'settings.prices.output': {en: 'out', vi: 'ra'},
    'settings.prices.inputFor': {
        en: '{model} input price per million tokens',
        vi: 'Giá token vào của {model} cho mỗi triệu token',
    },
    'settings.prices.cachedInputFor': {
        en: '{model} cached input price per million tokens',
        vi: 'Giá token vào đã cache của {model} cho mỗi triệu token',
    },
    'settings.prices.outputFor': {
        en: '{model} output price per million tokens',
        vi: 'Giá token ra của {model} cho mỗi triệu token',
    },

    'settings.mcp.title': {en: 'Servers agents can call', vi: 'Server mà agent gọi được'},
    'settings.mcp.hint': {
        en: 'The list comes from config.yaml. Authorize a server once and every agent reaches it through the proxy.',
        vi: 'Danh sách lấy từ config.yaml. Cấp quyền một lần là mọi agent đều gọi được qua proxy.',
    },
    'settings.mcp.loading': {en: 'Loading MCP servers…', vi: 'Đang tải danh sách MCP server…'},
    'settings.mcp.empty': {
        en: 'No MCP servers in config.yaml. Add one there and reopen this panel.',
        vi: 'Chưa có MCP server nào trong config.yaml. Thêm vào đó rồi mở lại mục này.',
    },
    'settings.mcp.tokenPlaceholder': {en: 'Paste access token', vi: 'Dán access token'},
    'settings.mcp.authorize': {en: 'Authorize', vi: 'Cấp quyền'},
    'settings.mcp.authorizing': {en: 'Authorizing', vi: 'Đang cấp quyền'},
    'settings.mcp.revoke': {en: 'Revoke', vi: 'Thu hồi'},
    'settings.mcp.revoking': {en: 'Revoking', vi: 'Đang thu hồi'},
    'settings.mcp.saveToken': {en: 'Save token', vi: 'Lưu token'},
    'settings.mcp.savingToken': {en: 'Saving', vi: 'Đang lưu'},
    'settings.mcp.enable': {en: 'Enable', vi: 'Bật'},
    'settings.mcp.enabling': {en: 'Enabling', vi: 'Đang bật'},
    'settings.mcp.disable': {en: 'Disable', vi: 'Tắt'},
    'settings.mcp.disabling': {en: 'Disabling', vi: 'Đang tắt'},
    'settings.mcp.revokeTitle': {
        en: 'Revoke {name} access',
        vi: 'Thu hồi quyền truy cập {name}',
    },
    'settings.mcp.revokeAuthAccount': {
        en: 'Agents lose access to {name} until you authorize it again. This deletes the credential stored for {account}.',
        vi: 'Agent sẽ mất quyền gọi {name} cho tới khi bạn cấp quyền lại. Thao tác này xóa thông tin đăng nhập đã lưu cho {account}.',
    },
    'settings.mcp.revokeAuth': {
        en: 'Agents lose access to {name} until you authorize it again. This deletes the stored credential.',
        vi: 'Agent sẽ mất quyền gọi {name} cho tới khi bạn cấp quyền lại. Thao tác này xóa thông tin đăng nhập đã lưu.',
    },
    'settings.mcp.revokeTokenAccount': {
        en: 'Agents lose access to {name} until you save a new token. This deletes the credential stored for {account}.',
        vi: 'Agent sẽ mất quyền gọi {name} cho tới khi bạn lưu token mới. Thao tác này xóa thông tin đăng nhập đã lưu cho {account}.',
    },
    'settings.mcp.revokeToken': {
        en: 'Agents lose access to {name} until you save a new token. This deletes the stored credential.',
        vi: 'Agent sẽ mất quyền gọi {name} cho tới khi bạn lưu token mới. Thao tác này xóa thông tin đăng nhập đã lưu.',
    },
    'settings.mcp.disableTitle': {en: 'Disable {name}', vi: 'Tắt {name}'},
    'settings.mcp.disableDescription': {
        en: 'Nexo quits the Chrome it started, and agents stop driving your browser. You stay signed in to every site, and turning it back on takes one click.',
        vi: 'Nexo sẽ đóng cửa sổ Chrome mà nó mở, và agent ngừng điều khiển trình duyệt. Bạn vẫn đăng nhập ở mọi trang, bật lại chỉ mất một cú bấm.',
    },

    'settings.error.setLanguage': {
        en: 'Could not change the language',
        vi: 'Không đổi được ngôn ngữ',
    },
    'settings.error.loadPreferences': {
        en: 'Could not load your preferences',
        vi: 'Không tải được tùy chọn của bạn',
    },
    'settings.error.loadModels': {
        en: 'Could not load the models this app can run',
        vi: 'Không tải được danh sách model chạy được',
    },
    'settings.error.saveDefault': {
        en: 'Could not save that default',
        vi: 'Không lưu được mặc định đó',
    },
    'settings.error.loadAutopilot': {
        en: 'Could not load the autopilot setting',
        vi: 'Không tải được cài đặt Autopilot',
    },
    'settings.error.saveAutopilot': {
        en: 'Could not change autopilot',
        vi: 'Không đổi được Autopilot',
    },
    'settings.error.loadMaxRunningAgents': {
        en: 'Could not load the running agent limit',
        vi: 'Không tải được giới hạn agent chạy cùng lúc',
    },
    'settings.error.saveMaxRunningAgents': {
        en: 'Could not change the running agent limit',
        vi: 'Không đổi được giới hạn agent chạy cùng lúc',
    },
    'settings.error.loadPrices': {
        en: 'Could not load what your models cost',
        vi: 'Không tải được giá model của bạn',
    },
    'settings.error.savePrice': {en: 'Could not save that price', vi: 'Không lưu được giá đó'},
    'settings.error.loadServers': {
        en: 'Could not load the MCP servers',
        vi: 'Không tải được danh sách MCP server',
    },
    'settings.error.authorizeServer': {
        en: 'Could not authorize the server',
        vi: 'Không cấp quyền được cho server',
    },
    'settings.error.saveToken': {en: 'Could not save the token', vi: 'Không lưu được token'},
    'settings.error.revokeServer': {
        en: 'Could not revoke the authorization',
        vi: 'Không thu hồi được quyền truy cập',
    },
    'settings.error.serverGone': {
        en: 'That MCP server is no longer configured.',
        vi: 'MCP server đó không còn trong cấu hình.',
    },
    'settings.error.serverTakesNoToken': {
        en: 'That MCP server does not take a pasted token.',
        vi: 'MCP server đó không nhận token dán vào.',
    },
    'settings.error.unknownModel': {
        en: '{model} is not a model this agent can run.',
        vi: '{model} không phải model mà agent này chạy được.',
    },
    'settings.error.unknownEffort': {
        en: '{effort} is not an effort level.',
        vi: '{effort} không phải một mức nỗ lực.',
    },
    'settings.error.unknownThinkingLevel': {
        en: '{level} is not a thinking level.',
        vi: '{level} không phải một mức suy luận.',
    },
} as const satisfies Catalog
