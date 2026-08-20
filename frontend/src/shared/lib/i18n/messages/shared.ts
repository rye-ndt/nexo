import type {Catalog} from '@/shared/lib/i18n/messages/catalog'

export const SHARED_MESSAGES = {
    'enum.stepState.idle': {en: 'Idle', vi: 'Chưa chạy'},
    'enum.stepState.blocked': {en: 'Blocked', vi: 'Bị chặn'},
    'enum.stepState.queued': {en: 'Ready', vi: 'Sẵn sàng'},
    'enum.stepState.running': {en: 'Running', vi: 'Đang chạy'},
    'enum.stepState.awaitingApproval': {en: 'Needs approval', vi: 'Cần phê duyệt'},
    'enum.stepState.awaitingReview': {en: 'Needs review', vi: 'Cần kiểm duyệt'},
    'enum.stepState.done': {en: 'Done', vi: 'Xong'},
    'enum.stepState.failed': {en: 'Failed', vi: 'Lỗi'},
    'enum.stepState.cancelled': {en: 'Cancelled', vi: 'Đã hủy'},

    'enum.workflowStatus.empty': {en: 'Empty', vi: 'Trống'},
    'enum.workflowStatus.draft': {en: 'Draft', vi: 'Bản nháp'},
    'enum.workflowStatus.ready': {en: 'Ready to run', vi: 'Sẵn sàng chạy'},
    'enum.workflowStatus.running': {en: 'Running', vi: 'Đang chạy'},
    'enum.workflowStatus.paused': {en: 'Paused', vi: 'Tạm dừng'},
    'enum.workflowStatus.done': {en: 'Done', vi: 'Xong'},
    'enum.workflowStatus.failed': {en: 'Failed', vi: 'Lỗi'},
    'enum.workflowStatus.cancelled': {en: 'Cancelled', vi: 'Đã hủy'},

    'enum.effort.quick': {en: 'Quick', vi: 'Nhanh'},
    'enum.effort.standard': {en: 'Standard', vi: 'Tiêu chuẩn'},
    'enum.effort.deep': {en: 'Deep', vi: 'Sâu'},
    'enum.effort.exhaustive': {en: 'Exhaustive', vi: 'Tối đa'},

    'enum.thinking.low': {en: 'Low', vi: 'Thấp'},
    'enum.thinking.medium': {en: 'Medium', vi: 'Trung bình'},
    'enum.thinking.high': {en: 'High', vi: 'Cao'},
    'enum.thinking.xhigh': {en: 'Very high', vi: 'Rất cao'},
    'enum.thinking.max': {en: 'Maximum', vi: 'Cao nhất'},

    'enum.installStage.queued': {en: 'Waiting', vi: 'Đang chờ'},
    'enum.installStage.resolve': {en: 'Finding release', vi: 'Đang tìm bản phát hành'},
    'enum.installStage.download': {en: 'Downloading', vi: 'Đang tải về'},
    'enum.installStage.extract': {en: 'Unpacking', vi: 'Đang giải nén'},
    'enum.installStage.done': {en: 'Ready', vi: 'Sẵn sàng'},

    'enum.agentAction.install': {en: 'Install', vi: 'Cài đặt'},
    'enum.agentAction.uninstall': {en: 'Uninstall', vi: 'Gỡ cài đặt'},
    'enum.agentAction.logIn': {en: 'Log in', vi: 'Đăng nhập'},
    'enum.agentAction.logOut': {en: 'Log out', vi: 'Đăng xuất'},
    'enum.agentAction.verify': {en: 'Verify', vi: 'Xác nhận'},

    'enum.agentActionBusy.install': {en: 'Installing', vi: 'Đang cài'},
    'enum.agentActionBusy.uninstall': {en: 'Uninstalling', vi: 'Đang gỡ'},
    'enum.agentActionBusy.logIn': {en: 'Logging in', vi: 'Đang đăng nhập'},
    'enum.agentActionBusy.logOut': {en: 'Logging out', vi: 'Đang đăng xuất'},
    'enum.agentActionBusy.verify': {en: 'Verifying', vi: 'Đang xác nhận'},

    'enum.agentActionFailure.install': {
        en: 'Could not install {name}',
        vi: 'Không cài được {name}',
    },
    'enum.agentActionFailure.uninstall': {
        en: 'Could not uninstall {name}',
        vi: 'Không gỡ được {name}',
    },
    'enum.agentActionFailure.logIn': {
        en: 'Could not start the login for {name}',
        vi: 'Không bắt đầu được lượt đăng nhập {name}',
    },
    'enum.agentActionFailure.logOut': {
        en: 'Could not log out of {name}',
        vi: 'Không đăng xuất được khỏi {name}',
    },
    'enum.agentActionFailure.verify': {
        en: 'Could not verify that code for {name}',
        vi: 'Không xác nhận được mã đó cho {name}',
    },

    'enum.inputType.text': {en: 'Text', vi: 'Văn bản'},
    'enum.inputType.textarea': {en: 'Long text', vi: 'Văn bản dài'},
    'enum.inputType.number': {en: 'Number', vi: 'Số'},
    'enum.inputType.boolean': {en: 'Boolean', vi: 'Đúng/sai'},
    'enum.inputType.select': {en: 'Choice', vi: 'Chọn một'},
    'enum.inputType.multiselect': {en: 'Multiple choices', vi: 'Chọn nhiều'},
    'enum.inputType.file': {en: 'File', vi: 'Tệp'},

    'glossary.workflow.title': {en: 'Workflow', vi: 'Workflow'},
    'glossary.workflow.what': {
        en: 'A graph of steps bound to one project folder.',
        vi: 'Sơ đồ các bước gắn với một thư mục dự án.',
    },
    'glossary.workflow.why': {
        en: 'You build it, lock it, run it, and review what it changed.',
        vi: 'Bạn dựng nó, khóa lại, chạy, rồi xem nó đã đổi những gì.',
    },
    'glossary.step.title': {en: 'Step', vi: 'Bước'},
    'glossary.step.what': {
        en: 'One scoped piece of work, run by one agent.',
        vi: 'Một phần việc có phạm vi rõ ràng, do một agent chạy.',
    },
    'glossary.step.why': {
        en: 'Steps chain, so each one starts with what the ones before it learned.',
        vi: 'Các bước nối tiếp nhau, nên mỗi bước bắt đầu với những gì bước trước đã biết.',
    },
    'glossary.role.title': {en: 'Role', vi: 'Vai trò'},
    'glossary.role.what': {
        en: 'A reusable kind of worker: what it does, what it needs, how hard it tries.',
        vi: 'Một kiểu nhân sự dùng lại được: làm gì, cần gì, cố gắng tới đâu.',
    },
    'glossary.role.why': {
        en: 'Define a reviewer once and every step that reviews starts from it.',
        vi: 'Định nghĩa người kiểm duyệt một lần, mọi bước kiểm duyệt đều bắt đầu từ đó.',
    },
    'glossary.agent.title': {en: 'Agent', vi: 'Agent'},
    'glossary.agent.what': {
        en: 'The coding CLI that runs a step — Claude Code, OpenCode, or Codex.',
        vi: 'CLI lập trình chạy một bước — Claude Code, OpenCode hoặc Codex.',
    },
    'glossary.agent.why': {
        en: 'Install and sign in once under Settings; steps pick which one they use.',
        vi: 'Cài và đăng nhập một lần trong Cài đặt; mỗi bước tự chọn agent để dùng.',
    },
    'glossary.handoff.title': {en: 'Handoff', vi: 'Bàn giao'},
    'glossary.handoff.what': {
        en: 'What a finished step writes down for the steps after it.',
        vi: 'Những gì một bước đã xong ghi lại cho các bước sau.',
    },
    'glossary.handoff.why': {
        en: 'The next agent reads it instead of guessing, including what not to do.',
        vi: 'Agent tiếp theo đọc nó thay vì đoán, kể cả những việc không nên làm.',
    },
    'glossary.result.title': {en: 'Result', vi: 'Kết quả'},
    'glossary.result.what': {
        en: 'Everything a finished step produced: its handoffs, files, and cost.',
        vi: 'Toàn bộ những gì một bước đã xong tạo ra: bàn giao, tệp và chi phí.',
    },
    'glossary.result.why': {
        en: 'It is the record you check before letting the run continue.',
        vi: 'Đây là bản ghi bạn xem trước khi cho lượt chạy đi tiếp.',
    },
    'glossary.lock.title': {en: 'Locking', vi: 'Khóa'},
    'glossary.lock.what': {
        en: 'Freezes the graph and the project folder so the workflow can run.',
        vi: 'Đóng băng sơ đồ và thư mục dự án để workflow có thể chạy.',
    },
    'glossary.lock.why': {
        en: 'A run that could be edited underneath itself could not be trusted or replayed.',
        vi: 'Một lượt chạy bị sửa ngay lúc đang chạy thì không đáng tin và không chạy lại được.',
    },
    'glossary.approval.title': {en: 'Approval', vi: 'Phê duyệt'},
    'glossary.approval.what': {
        en: 'An agent stopped mid-step to ask you something it should not decide alone.',
        vi: 'Agent dừng giữa bước để hỏi bạn điều nó không nên tự quyết.',
    },
    'glossary.approval.why': {
        en: 'It waits for your answer, then keeps going.',
        vi: 'Nó chờ câu trả lời của bạn rồi chạy tiếp.',
    },
    'glossary.review.title': {en: 'Review', vi: 'Kiểm duyệt'},
    'glossary.review.what': {
        en: 'A step finished and is holding everything downstream until you accept it.',
        vi: 'Một bước đã xong và giữ lại mọi bước phía sau cho tới khi bạn chấp nhận.',
    },
    'glossary.review.why': {
        en: 'Catches a bad handoff before it becomes the next three steps’ context.',
        vi: 'Chặn một bàn giao sai trước khi nó thành context của ba bước kế tiếp.',
    },
    'glossary.effort.title': {en: 'Effort', vi: 'Mức nỗ lực'},
    'glossary.effort.what': {
        en: 'How hard the agent tries: Quick, Standard, Deep, or Exhaustive.',
        vi: 'Agent cố gắng tới đâu: Nhanh, Tiêu chuẩn, Sâu hoặc Tối đa.',
    },
    'glossary.effort.why': {
        en: 'Sets the model and thinking budget. Higher costs more and takes longer.',
        vi: 'Quyết định model và ngân sách suy luận. Mức cao hơn tốn hơn và lâu hơn.',
    },
    'glossary.input.title': {en: 'Inputs', vi: 'Đầu vào'},
    'glossary.input.what': {
        en: 'Named values a step fills in before it runs.',
        vi: 'Các giá trị có tên mà một bước điền trước khi chạy.',
    },
    'glossary.input.why': {
        en: 'The role declares them once so each step only supplies what changes.',
        vi: 'Vai trò khai báo một lần, mỗi bước chỉ điền phần thay đổi.',
    },
    'glossary.instructions.title': {en: 'Instructions', vi: 'Chỉ dẫn'},
    'glossary.instructions.what': {
        en: 'The role’s prompt blocks, composed into what the agent receives.',
        vi: 'Các khối prompt của vai trò, ghép lại thành nội dung agent nhận được.',
    },
    'glossary.instructions.why': {
        en: 'Keeps standing rules in the role instead of retyping them per step.',
        vi: 'Giữ các quy tắc cố định trong vai trò thay vì gõ lại ở từng bước.',
    },
    'glossary.prompt.title': {en: 'Prompt', vi: 'Prompt'},
    'glossary.prompt.what': {
        en: 'This step’s own wording, on top of its role’s instructions.',
        vi: 'Câu chữ của riêng bước này, đặt trên chỉ dẫn của vai trò.',
    },
    'glossary.prompt.why': {
        en: 'Change it to steer one step without touching every step sharing the role.',
        vi: 'Sửa nó để lái một bước mà không đụng tới mọi bước dùng chung vai trò.',
    },
    'glossary.projectFolder.title': {en: 'Project folder', vi: 'Thư mục dự án'},
    'glossary.projectFolder.what': {
        en: 'The checkout agents read and change.',
        vi: 'Bản checkout mà agent đọc và sửa.',
    },
    'glossary.projectFolder.why': {
        en: 'Everything a run touches lives here, so a revert never leaves the folder.',
        vi: 'Mọi thứ một lượt chạy đụng tới đều nằm ở đây, nên hoàn tác không bao giờ ra khỏi thư mục.',
    },
    'glossary.reportFormat.title': {en: 'Report format', vi: 'Định dạng báo cáo'},
    'glossary.reportFormat.what': {
        en: 'The shape you want the step to report back in.',
        vi: 'Khuôn bạn muốn bước này báo cáo lại.',
    },
    'glossary.reportFormat.why': {
        en: 'A fixed shape is readable by the next step. Leave it empty for prose.',
        vi: 'Khuôn cố định giúp bước sau đọc được. Để trống nếu muốn văn xuôi.',
    },
    'glossary.retry.title': {en: 'Retry on failure', vi: 'Thử lại khi lỗi'},
    'glossary.retry.what': {
        en: 'A step that fails is put back in the pool and tried again.',
        vi: 'Bước bị lỗi được đưa lại vào hàng đợi và chạy lại.',
    },
    'glossary.retry.why': {
        en: 'Turn it off for anything that should not run twice.',
        vi: 'Tắt nó với những việc không nên chạy hai lần.',
    },
    'glossary.revert.title': {en: 'Revert', vi: 'Hoàn tác'},
    'glossary.revert.what': {
        en: 'Undoes every file change from this step onward.',
        vi: 'Gỡ bỏ mọi thay đổi tệp từ bước này trở đi.',
    },
    'glossary.revert.why': {
        en: 'Changes are stored as diffs, so this works without git.',
        vi: 'Thay đổi được lưu dưới dạng diff, nên không cần git vẫn hoàn tác được.',
    },
    'glossary.duplicate.title': {en: 'Duplicate', vi: 'Nhân bản'},
    'glossary.duplicate.what': {
        en: 'Copies the workflow with fresh ids and no run history.',
        vi: 'Chép workflow với id mới và không kèm lịch sử chạy.',
    },
    'glossary.duplicate.why': {
        en: 'The only way to change a locked workflow.',
        vi: 'Cách duy nhất để sửa một workflow đã khóa.',
    },
    'glossary.context.title': {en: 'Context', vi: 'Context'},
    'glossary.context.what': {
        en: 'How much of its context window the agent has used.',
        vi: 'Agent đã dùng bao nhiêu phần cửa sổ context.',
    },
    'glossary.context.why': {
        en: 'An agent that runs out mid-step loses the thread. Watch it on long steps.',
        vi: 'Agent hết context giữa bước sẽ mất mạch. Để mắt tới nó ở những bước dài.',
    },
    'glossary.autopilot.title': {en: 'Autopilot', vi: 'Autopilot'},
    'glossary.autopilot.what': {
        en: 'Answers routine approvals for you.',
        vi: 'Tự trả lời các phê duyệt thông thường thay bạn.',
    },
    'glossary.autopilot.why': {
        en: 'Lets a run finish unattended. Reviews still stop and wait.',
        vi: 'Giúp lượt chạy tự hoàn tất. Kiểm duyệt vẫn dừng lại và chờ.',
    },
    'glossary.mcp.title': {en: 'MCP', vi: 'MCP'},
    'glossary.mcp.what': {
        en: 'A standard way for agents to call outside services.',
        vi: 'Cách chuẩn để agent gọi các dịch vụ bên ngoài.',
    },
    'glossary.mcp.why': {
        en: 'Authorize a server once here; agents get a placeholder, never your key.',
        vi: 'Cấp quyền cho server một lần ở đây; agent chỉ nhận placeholder, không bao giờ thấy key của bạn.',
    },
    'glossary.thinking.title': {en: 'Thinking', vi: 'Suy luận'},
    'glossary.thinking.what': {
        en: 'How much reasoning the model does before it answers.',
        vi: 'Model suy luận bao nhiêu trước khi trả lời.',
    },
    'glossary.thinking.why': {
        en: 'Set per effort level so you tune it once, not per step.',
        vi: 'Đặt theo từng mức nỗ lực để chỉnh một lần, không phải từng bước.',
    },

    'error.fallbackTitle': {en: 'Something went wrong', vi: 'Đã xảy ra lỗi'},
    'error.cannotGetAuthInfo.title': {
        en: 'Stored credentials could not be read',
        vi: 'Không đọc được thông tin đăng nhập đã lưu',
    },
    'error.cannotGetAuthInfo.hint': {
        en: 'Authorize the server again in Settings › MCP.',
        vi: 'Cấp quyền lại cho server trong Cài đặt › MCP.',
    },
    'error.mcpNotFound.title': {
        en: 'That server is not configured',
        vi: 'Server đó chưa được cấu hình',
    },
    'error.mcpNotFound.hint': {
        en: 'The roster comes from config.yaml. Add the server there and restart the app.',
        vi: 'Danh sách server lấy từ config.yaml. Thêm server vào đó rồi khởi động lại ứng dụng.',
    },
    'error.mcpDiscoveryFailed.title': {
        en: 'The server did not say how to authorize',
        vi: 'Server không cho biết cách cấp quyền',
    },
    'error.mcpDiscoveryFailed.hint': {
        en: 'Check the server URL in config.yaml, then try again.',
        vi: 'Kiểm tra URL server trong config.yaml rồi thử lại.',
    },
    'error.mcpRegistrationFailed.title': {
        en: 'The server refused to register this app',
        vi: 'Server từ chối đăng ký ứng dụng này',
    },
    'error.mcpRegistrationFailed.hint': {
        en: 'Check the server URL in config.yaml, then try again.',
        vi: 'Kiểm tra URL server trong config.yaml rồi thử lại.',
    },
    'error.mcpAuthorizeFailed.title': {
        en: 'Authorization did not finish',
        vi: 'Việc cấp quyền chưa hoàn tất',
    },
    'error.mcpAuthorizeFailed.hint': {
        en: 'Start it again and approve in the browser tab that opens.',
        vi: 'Bắt đầu lại và chấp nhận ở tab trình duyệt vừa mở.',
    },
    'error.mcpAuthorizeTimeout.title': {
        en: 'Authorization timed out',
        vi: 'Việc cấp quyền đã quá hạn',
    },
    'error.mcpAuthorizeTimeout.hint': {
        en: 'The approval took too long. Start it again and finish in the browser tab that opens.',
        vi: 'Bước chấp nhận mất quá lâu. Bắt đầu lại và hoàn tất ở tab trình duyệt vừa mở.',
    },
    'error.mcpTokenExchange.title': {
        en: 'The server rejected the login',
        vi: 'Server từ chối lần đăng nhập này',
    },
    'error.mcpTokenExchange.hint': {
        en: 'Start the authorization again. If it keeps failing, the client id in config.yaml is wrong.',
        vi: 'Cấp quyền lại từ đầu. Nếu vẫn lỗi thì client id trong config.yaml bị sai.',
    },
    'error.mcpStoreCredentials.title': {
        en: 'The credentials could not be saved',
        vi: 'Không lưu được thông tin đăng nhập',
    },
    'error.mcpStoreCredentials.hint': {
        en: 'Check that the app data directory is writable, then authorize again.',
        vi: 'Kiểm tra thư mục dữ liệu của ứng dụng có ghi được không, rồi cấp quyền lại.',
    },
    'error.mcpNotAuthenticated.title': {
        en: 'That server is not authorized yet',
        vi: 'Server đó chưa được cấp quyền',
    },
    'error.mcpNotAuthenticated.hint': {
        en: 'Authorize it in Settings › MCP before an agent can call it.',
        vi: 'Cấp quyền trong Cài đặt › MCP trước khi agent có thể gọi nó.',
    },
    'error.mcpCredentialsExpired.title': {
        en: 'The credentials expired',
        vi: 'Thông tin đăng nhập đã hết hạn',
    },
    'error.mcpCredentialsExpired.hint': {
        en: 'Authorize the server again in Settings › MCP.',
        vi: 'Cấp quyền lại cho server trong Cài đặt › MCP.',
    },
    'error.mcpForbiddenRequest.title': {
        en: 'The proxy blocked that request',
        vi: 'Proxy đã chặn yêu cầu đó',
    },
    'error.mcpForbiddenRequest.hint': {
        en: 'An agent asked for a URL the proxy does not forward. Nothing was sent.',
        vi: 'Một agent xin một URL mà proxy không chuyển tiếp. Không có gì được gửi đi.',
    },
    'error.mcpRequestFailed.title': {en: 'The server did not answer', vi: 'Server không phản hồi'},
    'error.mcpRequestFailed.hint': {
        en: 'Check your connection, then try again.',
        vi: 'Kiểm tra kết nối mạng rồi thử lại.',
    },
    'error.roleFileInvalid.title': {
        en: 'That file is not a role export',
        vi: 'Tệp đó không phải bản xuất vai trò',
    },
    'error.roleFileInvalid.hint': {
        en: 'Nothing was imported. Export a fresh file from the app that has the roles.',
        vi: 'Không có gì được nhập. Hãy xuất tệp mới từ ứng dụng đang giữ các vai trò đó.',
    },
    'error.workflowFileInvalid.title': {
        en: 'That file is not a workflow export',
        vi: 'Tệp đó không phải bản xuất workflow',
    },
    'error.workflowFileInvalid.hint': {
        en: 'Nothing was imported. Export a fresh file from the app that has the workflow.',
        vi: 'Không có gì được nhập. Hãy xuất tệp mới từ ứng dụng đang giữ workflow đó.',
    },
    'error.roleConflict.title': {
        en: 'Those roles clash with the ones you have',
        vi: 'Các vai trò đó trùng với vai trò bạn đang có',
    },
    'error.roleConflict.hint': {
        en: 'Nothing was imported. Delete the ones named here, then import again — an imported role keeps the id it was exported with, so renaming yours is not enough.',
        vi: 'Không có gì được nhập. Xóa những vai trò được nêu ở đây rồi nhập lại — vai trò nhập vào giữ nguyên id lúc xuất, nên chỉ đổi tên là chưa đủ.',
    },
    'error.mcpTokenRequired.title': {
        en: 'That server needs an access token',
        vi: 'Server đó cần một access token',
    },
    'error.mcpTokenRequired.hint': {
        en: 'Paste a token for it in Settings › MCP.',
        vi: 'Dán token cho nó trong Cài đặt › MCP.',
    },
    'error.chromeNotFound.title': {
        en: 'Google Chrome is not installed',
        vi: 'Chưa cài Google Chrome',
    },
    'error.chromeNotFound.hint': {
        en: 'Nexo drives the Chrome in your applications folder. Install it there, then enable chrome-devtools again.',
        vi: 'Nexo điều khiển bản Chrome trong thư mục ứng dụng. Cài Chrome vào đó rồi bật lại chrome-devtools.',
    },
    'error.chromeLaunchFailed.title': {
        en: 'Chrome would not start with remote debugging on',
        vi: 'Chrome không khởi động được ở chế độ remote debugging',
    },
    'error.chromeLaunchFailed.hint': {
        en: 'Quit the Chrome window Nexo opened, then enable it again. Nexo drives its own copy of your profile on its own port, so your everyday Chrome can stay open.',
        vi: 'Đóng cửa sổ Chrome mà Nexo đã mở rồi bật lại. Nexo dùng bản sao profile của riêng nó trên một cổng riêng, nên Chrome hằng ngày của bạn vẫn mở được.',
    },
    'error.chromeNotConnected.title': {en: 'Chrome is not running', vi: 'Chrome chưa chạy'},
    'error.chromeNotConnected.hint': {
        en: 'Enable chrome-devtools in Settings › MCP, then run the step again.',
        vi: 'Bật chrome-devtools trong Cài đặt › MCP rồi chạy lại bước đó.',
    },

    'format.justNow': {en: 'just now', vi: 'vừa xong'},
    'format.minutesAgo': {en: '{count}m ago', vi: '{count} phút trước'},
    'format.hoursAgo': {en: '{count}h ago', vi: '{count} giờ trước'},
    'format.daysAgo': {en: '{count}d ago', vi: '{count} ngày trước'},
    'format.yearsAgo': {en: '{count}y ago', vi: '{count} năm trước'},

    'shared.helpTip.aria': {en: 'What is {term}?', vi: '{term} là gì?'},

    'shared.confirm.dismiss': {en: 'Cancel', vi: 'Hủy'},

    'shared.dialog.close': {en: 'Close', vi: 'Đóng'},

    'shared.directory.empty': {en: 'No folder chosen', vi: 'Chưa chọn thư mục'},
    'shared.directory.change': {en: 'Change', vi: 'Đổi'},
    'shared.directory.choose': {en: 'Choose', vi: 'Chọn'},
    'shared.directory.pickerFailed': {
        en: 'Could not open the folder picker',
        vi: 'Không mở được cửa sổ chọn thư mục',
    },

    'shared.error.details': {en: 'Details', vi: 'Chi tiết'},
    'shared.error.copy': {en: 'Copy', vi: 'Sao chép'},
    'shared.error.copied': {en: 'Copied', vi: 'Đã sao chép'},
    'shared.error.code': {en: 'Code', vi: 'Mã'},
    'shared.error.level': {en: 'Level', vi: 'Mức'},
    'shared.error.time': {en: 'Time', vi: 'Lúc'},
    'shared.error.queuedOne': {en: '{count} more error', vi: 'còn {count} lỗi nữa'},
    'shared.error.queuedOther': {en: '{count} more errors', vi: 'còn {count} lỗi nữa'},
    'shared.error.dismiss': {en: 'Dismiss', vi: 'Đóng'},

    'shared.notice.done': {en: 'Done', vi: 'Xong'},

    'shared.pathPicker.fileName': {en: 'File name', vi: 'Tên tệp'},
    'shared.pathPicker.cancel': {en: 'Cancel', vi: 'Hủy'},
    'shared.pathPicker.saveHere': {en: 'Save here', vi: 'Lưu vào đây'},
    'shared.pathPicker.chooseFolder': {en: 'Choose folder', vi: 'Chọn thư mục'},
    'shared.pathPicker.replaces': {en: 'replaces this', vi: 'sẽ ghi đè tệp này'},
    'shared.pathPicker.emptyFile': {
        en: 'Nothing in this folder. Go back up and try another one.',
        vi: 'Thư mục này trống. Quay lên trên và thử thư mục khác.',
    },
    'shared.pathPicker.emptySave': {
        en: 'Nothing here yet. Name the file below, or go somewhere else.',
        vi: 'Chưa có gì ở đây. Đặt tên tệp bên dưới, hoặc chọn nơi khác.',
    },
    'shared.pathPicker.emptyDirectory': {
        en: 'No folders inside this one. Choose it, make one, or go back up.',
        vi: 'Không có thư mục con nào. Chọn thư mục này, tạo một thư mục mới, hoặc quay lên trên.',
    },
    'shared.pathPicker.newFolder': {en: 'New folder', vi: 'Thư mục mới'},
    'shared.pathPicker.folderName': {en: 'Folder name', vi: 'Tên thư mục'},
    'shared.pathPicker.newFolderName': {en: 'New folder name', vi: 'Tên thư mục mới'},
    'shared.pathPicker.create': {en: 'Create', vi: 'Tạo'},
    'shared.pathPicker.homeFailed': {
        en: 'Could not find your home folder',
        vi: 'Không tìm được thư mục home của bạn',
    },
    'shared.pathPicker.readFailed': {
        en: 'Could not read that folder',
        vi: 'Không đọc được thư mục đó',
    },
    'shared.pathPicker.createFailed': {
        en: 'Could not create the folder',
        vi: 'Không tạo được thư mục',
    },
    'shared.pathPicker.unavailable': {
        en: 'No path picker is available right now.',
        vi: 'Hiện không có cửa sổ chọn đường dẫn nào dùng được.',
    },

    'shared.mockFs.unreadable': {en: '{path} could not be read.', vi: 'Không đọc được {path}.'},
    'shared.mockFs.fileNameRequired': {en: 'Give the file a name.', vi: 'Hãy đặt tên cho tệp.'},
    'shared.mockFs.folderNameRequired': {
        en: 'Give the folder a name.',
        vi: 'Hãy đặt tên cho thư mục.',
    },
    'shared.mockFs.folderSlash': {
        en: 'A folder name cannot contain a slash.',
        vi: 'Tên thư mục không được chứa dấu gạch chéo.',
    },
    'shared.mockFs.folderExists': {
        en: '{name} already exists here.',
        vi: '{name} đã tồn tại ở đây.',
    },
} as const satisfies Catalog
