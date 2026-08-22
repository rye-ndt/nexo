import type {Catalog} from '@/shared/lib/i18n/messages/catalog'

export const WORKFLOW_MESSAGES = {
    'workflow.rail.title': {en: 'Workflows', vi: 'Workflow'},
    'workflow.rail.import': {en: 'Import workflow', vi: 'Nhập workflow'},
    'workflow.rail.new': {en: 'New workflow', vi: 'Workflow mới'},
    'workflow.rail.empty': {en: 'No workflows yet.', vi: 'Chưa có workflow nào.'},
    'workflow.rail.createFirst': {en: 'Create your first workflow', vi: 'Tạo workflow đầu tiên'},

    'workflow.row.options': {en: 'Options for {name}', vi: 'Tùy chọn cho {name}'},
    'workflow.row.duplicate': {en: 'Duplicate', vi: 'Nhân bản'},
    'workflow.row.export': {en: 'Export', vi: 'Xuất'},
    'workflow.row.delete': {en: 'Delete', vi: 'Xóa'},
    'workflow.row.cancelled': {en: 'Cancelled', vi: 'Đã hủy'},
    'workflow.row.paused': {en: 'Paused', vi: 'Đang tạm dừng'},
    'workflow.row.locked': {en: 'Locked', vi: 'Đã khóa'},
    'workflow.row.notRun': {en: 'Not run yet', vi: 'Chưa chạy'},
    'workflow.row.tokens': {en: '{tokens} tokens', vi: '{tokens} token'},
    'workflow.row.ranFor': {en: 'ran for {elapsed}', vi: 'chạy trong {elapsed}'},
    'workflow.row.noRunTime': {en: 'run time unknown', vi: 'không rõ thời gian chạy'},
    'workflow.row.cost': {en: 'cost', vi: 'chi phí'},
    'workflow.row.notPriced': {en: 'not priced', vi: 'chưa tính giá'},

    'workflow.hint.locked': {
        en: 'Locked — duplicate to make changes.',
        vi: 'Đã khóa — nhân bản để chỉnh sửa.',
    },
    'workflow.hint.cancelled': {
        en: 'Run cancelled — duplicate to start over.',
        vi: 'Đã hủy lần chạy — nhân bản để chạy lại từ đầu.',
    },
    'workflow.hint.paused': {
        en: 'Paused — resume when you are ready.',
        vi: 'Đang tạm dừng — chạy tiếp khi bạn sẵn sàng.',
    },

    'workflow.action.newStep': {en: 'New step', vi: 'Bước mới'},
    'workflow.action.duplicate': {en: 'Duplicate', vi: 'Nhân bản'},
    'workflow.action.lock': {en: 'Lock', vi: 'Khóa'},
    'workflow.action.run': {en: 'Run', vi: 'Chạy'},
    'workflow.action.pause': {en: 'Pause run', vi: 'Tạm dừng'},
    'workflow.action.resume': {en: 'Resume run', vi: 'Chạy tiếp'},
    'workflow.action.cancel': {en: 'Cancel run', vi: 'Hủy lần chạy'},

    'workflow.lock.description': {
        en: 'The graph and its project folder lock for good. To change anything after this you have to duplicate the workflow. Nothing runs until you press Run.',
        vi: 'Sơ đồ và thư mục dự án sẽ khóa vĩnh viễn. Sau đó muốn sửa gì bạn phải nhân bản workflow. Chưa có gì chạy cho tới khi bạn bấm Chạy.',
    },
    'workflow.lock.confirm': {en: 'Lock workflow', vi: 'Khóa workflow'},

    'workflow.pause.title': {en: 'Pause the run?', vi: 'Tạm dừng lần chạy?'},
    'workflow.pause.description': {
        en: 'Steps running right now lose their work and start over when you resume — there will be no result for this attempt. Finished steps keep theirs, and you can close the app or shut the machine down; the run picks up where it stopped.',
        vi: 'Các bước đang chạy sẽ mất phần việc dở dang và chạy lại từ đầu khi bạn tiếp tục — lần thử này không để lại kết quả. Các bước đã xong vẫn giữ kết quả, và bạn có thể đóng ứng dụng hoặc tắt máy; lần chạy sẽ tiếp tục từ chỗ đã dừng.',
    },
    'workflow.pause.confirm': {en: 'Pause run', vi: 'Tạm dừng'},
    'workflow.pause.dismiss': {en: 'Keep running', vi: 'Chạy tiếp'},

    'workflow.cancel.title': {en: 'Cancel the run?', vi: 'Hủy lần chạy?'},
    'workflow.cancel.description': {
        en: 'The step running right now loses its work — there will be no result for it. Finished steps keep theirs, and the workflow stays on the rail. You cannot resume a cancelled run; duplicate it to start over.',
        vi: 'Bước đang chạy sẽ mất phần việc dở dang — bước đó không để lại kết quả. Các bước đã xong vẫn giữ kết quả, và workflow vẫn nằm trong danh sách. Lần chạy đã hủy thì không chạy tiếp được; nhân bản nó để chạy lại từ đầu.',
    },
    'workflow.cancel.dismiss': {en: 'Keep running', vi: 'Chạy tiếp'},

    'workflow.dialog.cancel': {en: 'Cancel', vi: 'Hủy'},
    'workflow.dialog.save': {en: 'Save', vi: 'Lưu'},

    'workflow.new.title': {en: 'New workflow', vi: 'Workflow mới'},
    'workflow.new.description': {
        en: 'Where this workflow runs, and the agents it needs.',
        vi: 'Nơi workflow này chạy và những agent nó cần.',
    },
    'workflow.new.create': {en: 'Create workflow', vi: 'Tạo workflow'},

    'workflow.locations.title': {en: 'Project folder', vi: 'Thư mục dự án'},
    'workflow.locations.description': {
        en: 'Where this workflow runs. Locking fixes it for good.',
        vi: 'Nơi workflow này chạy. Khóa lại là cố định vĩnh viễn.',
    },
    'workflow.locations.label': {en: 'Project folder', vi: 'Thư mục dự án'},
    'workflow.locations.hint': {
        en: 'Pick it now — it is fixed once the workflow is locked.',
        vi: 'Chọn ngay — khóa workflow rồi thì không đổi được nữa.',
    },
    'workflow.locations.picker': {en: 'Choose the project folder', vi: 'Chọn thư mục dự án'},
    'workflow.locations.moveTitle': {
        en: 'Move this workflow?',
        vi: 'Chuyển workflow sang thư mục khác?',
    },
    'workflow.locations.moveDescription': {
        en: 'Steps that already ran keep their results, but every step from here on runs against the new folder.',
        vi: 'Các bước đã chạy vẫn giữ kết quả, nhưng từ đây trở đi mọi bước đều chạy trên thư mục mới.',
    },
    'workflow.locations.moveConfirm': {en: 'Save folder', vi: 'Lưu thư mục'},

    'workflow.import.title': {en: 'Import workflow', vi: 'Nhập workflow'},
    'workflow.import.description': {
        en: 'Where this workflow runs on this machine.',
        vi: 'Nơi workflow này chạy trên máy này.',
    },
    'workflow.import.confirm': {en: 'Import workflow', vi: 'Nhập workflow'},
    'workflow.import.fromFile': {en: 'From the file', vi: 'Từ tệp'},
    'workflow.import.steps.one': {en: '{count} step', vi: '{count} bước'},
    'workflow.import.steps.other': {en: '{count} steps', vi: '{count} bước'},
    'workflow.import.rolesStayBehind': {
        en: 'These steps run on their own: the roles they were built from stay on the machine that exported them.',
        vi: 'Các bước này chạy độc lập: vai trò tạo ra chúng vẫn nằm lại trên máy đã xuất tệp.',
    },

    'workflow.step.untitled': {en: 'Untitled step', vi: 'Bước chưa đặt tên'},

    'workflow.delete.title': {en: 'Delete “{name}”?', vi: 'Xóa “{name}”?'},
    'workflow.delete.noSteps': {
        en: 'It has no steps yet. This cannot be undone.',
        vi: 'Workflow này chưa có bước nào. Không hoàn tác được.',
    },
    'workflow.delete.steps.one': {
        en: 'Its {count} step goes with it. This cannot be undone.',
        vi: '{count} bước trong đó cũng mất theo. Không hoàn tác được.',
    },
    'workflow.delete.steps.other': {
        en: 'Its {count} steps go with it. This cannot be undone.',
        vi: '{count} bước trong đó cũng mất theo. Không hoàn tác được.',
    },
    'workflow.delete.finished': {
        en: 'Its {total} steps go with it, including {done} finished and their results. This cannot be undone.',
        vi: '{total} bước trong đó cũng mất theo, gồm {done} bước đã xong và kết quả của chúng. Không hoàn tác được.',
    },
    'workflow.delete.confirm': {en: 'Delete workflow', vi: 'Xóa workflow'},

    'workflow.workspace.empty': {
        en: 'No workflow open. Create one to start a chain of steps.',
        vi: 'Chưa mở workflow nào. Tạo một workflow để bắt đầu chuỗi các bước.',
    },
    'workflow.deleteStep.title': {en: 'Delete “{title}”?', vi: 'Xóa “{title}”?'},
    'workflow.deleteStep.untitled': {en: 'Delete this step?', vi: 'Xóa bước này?'},
    'workflow.deleteStep.description': {
        en: 'Its prompt, inputs and any result it produced go with it. This cannot be undone.',
        vi: 'Prompt, đầu vào và mọi kết quả của bước đó cũng mất theo. Không hoàn tác được.',
    },
    'workflow.deleteStep.confirm': {en: 'Delete step', vi: 'Xóa bước'},

    'workflow.review.description': {
        en: 'Finished. Nothing downstream runs until you decide.',
        vi: 'Đã xong. Các bước phía sau chưa chạy cho tới khi bạn quyết định.',
    },
    'workflow.review.waiting.one': {en: '{count} step waiting', vi: '{count} bước đang chờ'},
    'workflow.review.waiting.other': {en: '{count} steps waiting', vi: '{count} bước đang chờ'},
    'workflow.review.notNow': {en: 'Not now', vi: 'Để sau'},
    'workflow.review.reject': {en: 'Reject and stop', vi: 'Từ chối và dừng'},
    'workflow.review.rejecting': {en: 'Rejecting…', vi: 'Đang từ chối…'},
    'workflow.review.confirm': {en: 'Confirm and continue', vi: 'Chấp nhận và chạy tiếp'},
    'workflow.review.confirming': {en: 'Confirming…', vi: 'Đang chấp nhận…'},
    'workflow.review.empty': {
        en: 'This step carried nothing forward. Confirm only if you have checked its work yourself.',
        vi: 'Bước này không bàn giao gì cho bước sau. Chỉ chấp nhận nếu bạn đã tự kiểm tra phần việc của nó.',
    },
    'workflow.review.handoffIndex': {
        en: 'Handoff {index} of {total}',
        vi: 'Bàn giao {index}/{total}',
    },
    'workflow.review.inShort': {en: 'In short', vi: 'Tóm tắt'},
    'workflow.review.step': {en: 'Step', vi: 'Bước'},
    'workflow.review.outcome': {en: 'Outcome', vi: 'Kết quả'},
    'workflow.review.blockers': {en: 'Blockers', vi: 'Vướng mắc'},
    'workflow.review.knownGaps': {en: 'Known gaps', vi: 'Phần còn thiếu'},
    'workflow.review.mustAvoid': {en: 'Must avoid', vi: 'Cần tránh'},
    'workflow.review.rejectedDecisions': {en: 'Rejected decisions', vi: 'Quyết định bị bác'},
    'workflow.review.approvedDecisions': {en: 'Approved decisions', vi: 'Quyết định đã duyệt'},
    'workflow.review.changedBehaviors': {en: 'Changed behaviors', vi: 'Hành vi đã thay đổi'},
    'workflow.review.currentBehaviors': {en: 'Current behaviors', vi: 'Hành vi hiện tại'},
    'workflow.review.nuances': {en: 'Nuances', vi: 'Điểm cần lưu ý'},

    'workflow.agents.intro': {
        en: 'Your preferences run on these agents. Each one has to be logged in before the workflow can start.',
        vi: 'Tùy chọn của bạn chạy trên những agent này. Mỗi agent phải đăng nhập trước khi workflow chạy được.',
    },
    'workflow.agents.checking': {en: 'Checking agents…', vi: 'Đang kiểm tra agent…'},
    'workflow.agents.noneTitle': {
        en: 'No agent is logged in',
        vi: 'Chưa có agent nào đăng nhập',
    },
    'workflow.agents.noneBody': {
        en: 'A workflow runs on whichever agent your effort levels resolve to, and none of them resolve yet. Install Claude Code, Codex or Open Code under Settings › Agents and log in; the levels fill in and this workflow can be created.',
        vi: 'Workflow chạy trên agent mà các mức nỗ lực của bạn trỏ tới, và hiện chưa mức nào trỏ tới đâu cả. Cài Claude Code, Codex hoặc Open Code trong Cài đặt › Agent rồi đăng nhập; các mức nỗ lực sẽ được điền và bạn tạo được workflow này.',
    },
    'workflow.agents.notAvailable': {
        en: 'Not available — add it to config.yaml',
        vi: 'Không có sẵn — thêm nó vào config.yaml',
    },
    'workflow.agents.notInstalled': {en: 'Not installed', vi: 'Chưa cài đặt'},
    'workflow.agents.notLoggedIn': {en: 'Not logged in', vi: 'Chưa đăng nhập'},
    'workflow.agents.ready': {en: 'v{version} · Ready', vi: 'v{version} · Sẵn sàng'},
    'workflow.agents.models': {
        en: 'Runs {models} for your effort levels.',
        vi: 'Chạy {models} cho các mức nỗ lực của bạn.',
    },

    'workflow.transfer.importPicker': {en: 'Import workflow', vi: 'Nhập workflow'},
    'workflow.transfer.exportPicker': {en: 'Export workflow', vi: 'Xuất workflow'},
    'workflow.transfer.imported.title': {en: 'Imported {name}', vi: 'Đã nhập {name}'},
    'workflow.transfer.imported.description': {
        en: 'It is a draft: nothing runs until you lock and start it.',
        vi: 'Đây là bản nháp: chưa có gì chạy cho tới khi bạn khóa và bắt đầu.',
    },
    'workflow.transfer.exported.title': {en: 'Exported {name}', vi: 'Đã xuất {name}'},
    'workflow.transfer.exported.description': {
        en: 'The file carries the graph and what each step runs as. Roles stay here.',
        vi: 'Tệp mang theo sơ đồ và những gì mỗi bước sẽ chạy. Vai trò vẫn ở lại máy này.',
    },

    'workflow.duplicate.name': {en: '{name} copy', vi: '{name} - bản sao'},
    'workflow.duplicate.title': {en: 'Duplicate “{name}”?', vi: 'Nhân bản “{name}”?'},
    'workflow.duplicate.description': {
        en: 'The copy is a draft with the same graph and the same steps; its project folder starts empty and it has no run history.',
        vi: 'Bản sao là một bản nháp với cùng sơ đồ và cùng các bước; thư mục dự án của nó để trống và chưa có lịch sử chạy.',
    },
    'workflow.duplicate.copyInputs': {
        en: 'Copy input values',
        vi: 'Chép giá trị đầu vào',
    },

    'workflow.error.load': {
        en: 'Could not load your workflows',
        vi: 'Không tải được danh sách workflow',
    },
    'workflow.error.create': {en: 'Could not create the workflow', vi: 'Không tạo được workflow'},
    'workflow.error.duplicate': {
        en: 'Could not duplicate the workflow',
        vi: 'Không nhân bản được workflow',
    },
    'workflow.error.import': {en: 'Could not import the workflow', vi: 'Không nhập được workflow'},
    'workflow.error.export': {en: 'Could not export the workflow', vi: 'Không xuất được workflow'},
    'workflow.error.readFile': {en: 'Could not read that file', vi: 'Không đọc được tệp đó'},
    'workflow.error.filePicker': {
        en: 'Could not open the file picker',
        vi: 'Không mở được hộp thoại chọn tệp',
    },
    'workflow.error.rename': {
        en: 'Could not rename the workflow',
        vi: 'Không đổi được tên workflow',
    },
    'workflow.error.locations': {
        en: 'Could not save the workflow directories',
        vi: 'Không lưu được thư mục của workflow',
    },
    'workflow.error.lock': {en: 'Could not lock the workflow', vi: 'Không khóa được workflow'},
    'workflow.error.start': {en: 'Could not start the run', vi: 'Không bắt đầu được lần chạy'},
    'workflow.error.pause': {en: 'Could not pause the run', vi: 'Không tạm dừng được lần chạy'},
    'workflow.error.resume': {en: 'Could not resume the run', vi: 'Không chạy tiếp được'},
    'workflow.error.cancel': {en: 'Could not cancel the run', vi: 'Không hủy được lần chạy'},
    'workflow.error.reorder': {
        en: 'Could not reorder the workflows',
        vi: 'Không sắp xếp lại được danh sách workflow',
    },
    'workflow.error.delete': {en: 'Could not delete the workflow', vi: 'Không xóa được workflow'},
    'workflow.error.addStep': {en: 'Could not add the step', vi: 'Không thêm được bước'},
    'workflow.error.duplicateStep': {
        en: 'Could not duplicate the step',
        vi: 'Không nhân bản được bước',
    },
    'workflow.error.saveStep': {en: 'Could not save the step', vi: 'Không lưu được bước'},
    'workflow.error.moveStep': {en: 'Could not move the step', vi: 'Không di chuyển được bước'},
    'workflow.error.saveInputs': {en: 'Could not save the inputs', vi: 'Không lưu được đầu vào'},
    'workflow.error.decision': {
        en: 'Could not record your decision',
        vi: 'Không ghi nhận được quyết định của bạn',
    },
    'workflow.error.revert': {
        en: 'Could not revert to this step',
        vi: 'Không hoàn tác về bước này được',
    },
    'workflow.error.deleteStep': {en: 'Could not delete the step', vi: 'Không xóa được bước'},
    'workflow.error.connect': {en: 'Could not link the steps', vi: 'Không nối được các bước'},
    'workflow.error.disconnect': {
        en: 'Could not unlink the steps',
        vi: 'Không bỏ nối được các bước',
    },
    'workflow.error.activity': {
        en: 'Could not read what this step is doing',
        vi: 'Không đọc được bước này đang làm gì',
    },
    'workflow.error.diff': {
        en: 'Could not read what this step changed',
        vi: 'Không đọc được bước này đã thay đổi gì',
    },

    'workflow.api.gone': {
        en: 'That workflow is gone. Pick another one from the rail.',
        vi: 'Workflow đó không còn nữa. Chọn workflow khác trong danh sách.',
    },
    'workflow.api.locked': {
        en: 'This workflow is locked. Duplicate it to make changes.',
        vi: 'Workflow này đã khóa. Nhân bản nó để chỉnh sửa.',
    },
    'workflow.api.stepGone': {
        en: 'That step is gone. Pick another one on the canvas.',
        vi: 'Bước đó không còn nữa. Chọn bước khác trên sơ đồ.',
    },
    'workflow.api.needsProjectDir': {
        en: 'A workflow needs a project folder.',
        vi: 'Workflow cần một thư mục dự án.',
    },
    'workflow.api.noUnlock': {
        en: 'A locked workflow cannot go back to a draft. Duplicate it instead.',
        vi: 'Workflow đã khóa không quay lại bản nháp được. Hãy nhân bản nó.',
    },
    'workflow.api.folderBeforeLock': {
        en: 'Choose a project folder before locking this workflow.',
        vi: 'Chọn thư mục dự án trước khi khóa workflow này.',
    },
    'workflow.api.lockBeforeRun': {
        en: 'Lock the workflow before running it.',
        vi: 'Khóa workflow trước khi chạy.',
    },
    'workflow.api.cancelled': {
        en: 'This run was cancelled. Duplicate the workflow to run it again.',
        vi: 'Lần chạy này đã bị hủy. Nhân bản workflow để chạy lại.',
    },
    'workflow.api.inputsLocked': {
        en: 'This workflow is running. Its inputs are locked.',
        vi: 'Workflow này đang chạy. Đầu vào của nó đã khóa.',
    },
    'workflow.api.notAwaitingReview': {
        en: 'This step is not waiting to be accepted. It may already have been answered.',
        vi: 'Bước này không chờ kiểm duyệt. Có thể nó đã được trả lời rồi.',
    },
    'workflow.api.notPaused': {
        en: 'This workflow is not paused.',
        vi: 'Workflow này không ở trạng thái tạm dừng.',
    },
    'workflow.api.cycle': {
        en: 'That link would loop back on itself. Point it at another step.',
        vi: 'Liên kết đó tạo thành vòng lặp. Hãy trỏ nó sang bước khác.',
    },
    'workflow.api.stepNotRun': {
        en: 'This step has not run yet, so there is nothing to inspect.',
        vi: 'Bước này chưa chạy nên chưa có gì để xem.',
    },
    'workflow.api.setProjectDir': {
        en: 'Set a project folder before running this workflow.',
        vi: 'Đặt thư mục dự án trước khi chạy workflow này.',
    },
    'workflow.api.noWorkflowId': {
        en: 'The run started without a workflow id. Check the app log and try again.',
        vi: 'Lần chạy bắt đầu mà không có workflow id. Kiểm tra log của ứng dụng rồi thử lại.',
    },
    'workflow.api.fileInvalid': {
        en: 'That file is not a workflow export.',
        vi: 'Tệp đó không phải tệp workflow đã xuất.',
    },
} as const satisfies Catalog
