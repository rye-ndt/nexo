import type {Catalog} from '@/shared/lib/i18n/messages/catalog'

export const WORKFLOW_CANVAS_MESSAGES = {
    'canvas.empty.title': {en: 'No steps yet.', vi: 'Chưa có bước nào.'},
    'canvas.empty.lockedTitle': {
        en: 'This workflow has no steps.',
        vi: 'Workflow này không có bước nào.',
    },
    'canvas.empty.newStep': {en: 'New step', vi: 'Thêm bước'},

    'canvas.menu.pane': {en: 'Canvas', vi: 'Canvas'},
    'canvas.menu.newStepHere': {en: 'New step here', vi: 'Thêm bước tại đây'},
    'canvas.menu.fitView': {en: 'Fit view', vi: 'Vừa khung hình'},
    'canvas.menu.openStep': {en: 'Open step', vi: 'Mở bước'},
    'canvas.menu.duplicateStep': {en: 'Duplicate step', vi: 'Nhân bản bước'},
    'canvas.menu.deleteStep': {en: 'Delete step', vi: 'Xóa bước'},
    'canvas.menu.unlink': {en: 'Unlink', vi: 'Bỏ liên kết'},
    'canvas.menu.edgeSubject': {en: '{source} → {target}', vi: '{source} → {target}'},

    'canvas.edge.unlink': {
        en: 'Unlink {source} from {target}',
        vi: 'Bỏ liên kết {source} khỏi {target}',
    },

    'canvas.card.inputs': {en: 'Inputs', vi: 'Đầu vào'},
    'canvas.card.runsAfter': {en: 'Runs after {name}', vi: 'Chạy sau {name}'},
    'canvas.card.runsAfterAll': {
        en: 'Runs after all {count} upstream steps',
        vi: 'Chạy sau cả {count} bước phía trước',
    },

    'canvas.activity.label': {
        en: 'What this step is doing now',
        vi: 'Bước này đang làm gì',
    },
    'canvas.failure.label': {en: 'Why this step stopped: ', vi: 'Lý do bước này dừng: '},

    'canvas.context.reading': {
        en: '{used} of {total} context used',
        vi: 'Đã dùng {used} trên {total} context',
    },
    'canvas.context.share': {
        en: '{percent}% of the context window used',
        vi: 'Đã dùng {percent}% context window',
    },

    'canvas.zoom.in': {en: 'Zoom in', vi: 'Phóng to'},
    'canvas.zoom.out': {en: 'Zoom out', vi: 'Thu nhỏ'},
    'canvas.zoom.fit': {en: 'Fit view', vi: 'Vừa khung hình'},

    'canvas.header.settings': {en: 'Settings', vi: 'Cài đặt'},
    'canvas.header.lockTitle': {en: 'Lock “{name}”?', vi: 'Khóa “{name}”?'},
    'canvas.header.pausing': {en: 'Pausing…', vi: 'Đang tạm dừng…'},
    'canvas.header.cancelling': {en: 'Cancelling…', vi: 'Đang hủy…'},
    'canvas.header.cancelRun': {en: 'Cancel run', vi: 'Hủy lần chạy'},

    'canvas.rail.hide': {en: 'Hide workflows', vi: 'Ẩn danh sách workflow'},
    'canvas.rail.show': {en: 'Show workflows', vi: 'Hiện danh sách workflow'},

    'canvas.sheet.open': {en: 'Workflow menu', vi: 'Menu workflow'},
    'canvas.sheet.title': {en: 'Menu', vi: 'Menu'},
    'canvas.sheet.description': {
        en: 'Directories, progress and actions for this workflow.',
        vi: 'Thư mục, tiến độ và thao tác của workflow này.',
    },
    'canvas.sheet.settings': {en: 'Settings', vi: 'Cài đặt'},
    'canvas.sheet.projectDir': {en: 'Project folder', vi: 'Thư mục dự án'},
    'canvas.sheet.noProjectDir': {en: 'Not set', vi: 'Chưa chọn'},
    'canvas.sheet.progress': {en: 'Progress', vi: 'Tiến độ'},

    'canvas.dirs.none': {en: 'No project folder', vi: 'Chưa có thư mục dự án'},
    'canvas.dirs.label': {en: 'Workflow directories', vi: 'Thư mục của workflow'},
    'canvas.dirs.change': {en: 'Click to change it.', vi: 'Bấm để đổi.'},

    'canvas.name.label': {en: 'Workflow name', vi: 'Tên workflow'},
    'canvas.name.locked': {en: 'This workflow is locked', vi: 'Workflow này đã khóa'},

    'canvas.cost.open': {en: 'Run cost', vi: 'Chi phí lần chạy'},
    'canvas.cost.title': {en: 'Run cost', vi: 'Chi phí lần chạy'},
    'canvas.cost.ledger': {
        en: '{input} in · {cached} cached · {output} out',
        vi: '{input} vào · {cached} cached · {output} ra',
    },
    'canvas.cost.cost': {en: 'Cost', vi: 'Chi phí'},
    'canvas.cost.unpriced': {
        en: 'Add prices in Settings → Preferences to see this run in dollars.',
        vi: 'Thêm đơn giá trong Cài đặt → Tùy chọn để xem lần chạy này bằng đô la.',
    },
    'canvas.cost.elapsed': {en: 'Elapsed', vi: 'Thời gian chạy'},
    'canvas.cost.finishedAt': {en: 'Finished {moment}', vi: 'Kết thúc {moment}'},
    'canvas.cost.startedAt': {en: 'Started {moment}', vi: 'Bắt đầu {moment}'},
    'canvas.cost.breakdown': {en: 'Cost breakdown', vi: 'Phân tích chi phí'},
    'canvas.cost.byStep': {en: 'By step', vi: 'Theo bước'},
    'canvas.cost.byModel': {en: 'By model', vi: 'Theo model'},
    'canvas.cost.unknownModel': {en: 'Unknown model', vi: 'Model không rõ'},
    'canvas.cost.stepCount.one': {en: '{count} step', vi: '{count} bước'},
    'canvas.cost.stepCount.other': {en: '{count} steps', vi: '{count} bước'},
    'canvas.cost.everyAttempt': {en: 'Every attempt', vi: 'Gồm mọi lần thử'},
    'canvas.cost.nothingSpent': {
        en: 'No step has spent anything yet.',
        vi: 'Chưa bước nào tiêu tốn gì.',
    },

    'inspector.context.share': {
        en: '{percent}% of the context window used',
        vi: 'Đã dùng {percent}% context window',
    },

    'inspector.files.label': {en: 'Files changed', vi: 'File đã đổi'},
    'inspector.files.empty': {
        en: 'This step left the project folder as it found it.',
        vi: 'Bước này không đổi gì trong thư mục dự án.',
    },
    'inspector.files.renamedPath': {en: '{from} → {to}', vi: '{from} → {to}'},
    'inspector.files.created': {en: 'Created', vi: 'Đã tạo'},
    'inspector.files.modified': {en: 'Modified', vi: 'Đã sửa'},
    'inspector.files.deleted': {en: 'Deleted', vi: 'Đã xóa'},
    'inspector.files.renamed': {en: 'Renamed', vi: 'Đã đổi tên'},

    'inspector.diff.loading': {
        en: 'Reading what this step changed…',
        vi: 'Đang đọc thay đổi của bước này…',
    },
    'inspector.diff.failed': {
        en: 'Could not read what this step changed.',
        vi: 'Không đọc được thay đổi của bước này.',
    },
    'inspector.diff.retry': {en: 'Try again', vi: 'Thử lại'},

    'inspector.handoff.label': {en: 'Handoff', vi: 'Bàn giao'},
    'inspector.handoff.empty': {
        en: 'None. Nothing was carried forward.',
        vi: 'Không có. Không có gì được chuyển tiếp.',
    },
    'inspector.handoff.goesNext': {en: 'Goes to the next step', vi: 'Chuyển sang bước sau'},
    'inspector.handoff.step': {en: 'Step', vi: 'Bước'},
    'inspector.handoff.outcome': {en: 'Outcome', vi: 'Kết quả'},
    'inspector.handoff.blockers': {en: 'Blockers', vi: 'Vướng mắc'},
    'inspector.handoff.approvedDecisions': {en: 'Approved decisions', vi: 'Quyết định đã duyệt'},
    'inspector.handoff.rejectedDecisions': {
        en: 'Rejected decisions',
        vi: 'Quyết định bị từ chối',
    },
    'inspector.handoff.currentBehaviors': {en: 'Current behaviors', vi: 'Hành vi hiện tại'},
    'inspector.handoff.changedBehaviors': {en: 'Changed behaviors', vi: 'Hành vi đã đổi'},
    'inspector.handoff.mustAvoid': {en: 'Must avoid', vi: 'Cần tránh'},
    'inspector.handoff.nuances': {en: 'Nuances', vi: 'Điểm cần lưu ý'},
    'inspector.handoff.knownGaps': {en: 'Known gaps', vi: 'Phần còn thiếu'},

    'inspector.revert.title': {en: 'Revert to “{name}”?', vi: 'Hoàn tác về “{name}”?'},
    'inspector.revert.confirm': {en: 'Revert to this step', vi: 'Hoàn tác về bước này'},
    'inspector.revert.none': {
        en: 'Files in {dir} go back to how this step left them, and anything written after it is lost. No step after it has run, so nothing else is undone. The run stops, and you start it again when you are ready.',
        vi: 'File trong {dir} quay lại đúng như bước này để lại, mọi thứ ghi sau đó sẽ mất. Chưa bước nào sau nó chạy nên không có gì khác bị hoàn tác. Lần chạy sẽ dừng, bạn chạy lại khi sẵn sàng.',
    },
    'inspector.revert.one': {
        en: 'Files in {dir} go back to how this step left them, and anything written after it is lost. The one step after it is undone and goes back to not started. The run stops, and you start it again when you are ready.',
        vi: 'File trong {dir} quay lại đúng như bước này để lại, mọi thứ ghi sau đó sẽ mất. Một bước sau nó bị hoàn tác và trở về trạng thái chưa chạy. Lần chạy sẽ dừng, bạn chạy lại khi sẵn sàng.',
    },
    'inspector.revert.many': {
        en: 'Files in {dir} go back to how this step left them, and anything written after it is lost. The {count} steps after it are undone and go back to not started. The run stops, and you start it again when you are ready.',
        vi: 'File trong {dir} quay lại đúng như bước này để lại, mọi thứ ghi sau đó sẽ mất. {count} bước sau nó bị hoàn tác và trở về trạng thái chưa chạy. Lần chạy sẽ dừng, bạn chạy lại khi sẵn sàng.',
    },

    'inspector.status.emptyCancelled': {
        en: 'This step never started. The run was cancelled before its turn came.',
        vi: 'Bước này chưa từng bắt đầu. Lần chạy bị hủy trước khi tới lượt nó.',
    },
    'inspector.status.emptyBlocked': {
        en: 'This step has not run. It is waiting on the steps upstream of it.',
        vi: 'Bước này chưa chạy. Nó đang chờ các bước phía trước.',
    },
    'inspector.status.empty': {en: 'This step has not run yet.', vi: 'Bước này chưa chạy.'},
    'inspector.status.pendingCancelled': {
        en: 'You cancelled the run while this step was working. Its work was discarded, so there is no result.',
        vi: 'Bạn đã hủy lần chạy khi bước này đang làm việc. Phần việc đó bị bỏ nên không có kết quả.',
    },
    'inspector.status.pending': {
        en: 'The result lands when this step stops.',
        vi: 'Kết quả sẽ có khi bước này dừng.',
    },
    'inspector.status.model': {en: '{model} · {thinking}', vi: '{model} · {thinking}'},
    'inspector.status.inputs': {en: 'Inputs', vi: 'Đầu vào'},
    'inspector.status.started': {en: 'Started', vi: 'Bắt đầu'},
    'inspector.status.stopped': {en: 'Stopped', vi: 'Đã dừng'},
    'inspector.status.finished': {en: 'Finished', vi: 'Kết thúc'},
    'inspector.status.elapsed': {en: 'Elapsed', vi: 'Thời gian chạy'},
    'inspector.status.retries': {en: 'Retries', vi: 'Số lần thử lại'},
    'inspector.status.context': {en: 'Context', vi: 'Context'},
    'inspector.status.contextUsed': {en: '{percent}% used', vi: 'Đã dùng {percent}%'},
    'inspector.status.reverting': {en: 'Reverting…', vi: 'Đang hoàn tác…'},
    'inspector.status.revert': {en: 'Revert to this step', vi: 'Hoàn tác về bước này'},
    'inspector.status.close': {en: 'Close', vi: 'Đóng'},

    'step.untitled': {en: 'Untitled step', vi: 'Bước chưa đặt tên'},

    'step.edit.deleteStep': {en: 'Delete step', vi: 'Xóa bước'},
    'step.edit.cancel': {en: 'Cancel', vi: 'Hủy'},
    'step.edit.save': {en: 'Save', vi: 'Lưu'},
    'step.edit.role': {en: 'Role', vi: 'Vai trò'},
    'step.edit.roleGone': {en: 'No longer available', vi: 'Không còn nữa'},
    'step.edit.roleFixed': {
        en: 'Fixed when the step was created.',
        vi: 'Cố định từ lúc tạo bước.',
    },
    'step.edit.deleteTitle': {en: 'Delete “{name}”?', vi: 'Xóa “{name}”?'},
    'step.edit.deleteThisStep': {en: 'this step', vi: 'bước này'},
    'step.edit.deleteBody': {
        en: 'Its prompt, inputs and any result it produced go with it. This cannot be undone.',
        vi: 'Prompt, đầu vào và mọi kết quả của nó cũng mất theo. Không hoàn tác được.',
    },

    'step.agent.effort': {en: 'Effort', vi: 'Mức nỗ lực'},
    'step.agent.notSet': {en: 'Not set', vi: 'Chưa đặt'},
    'step.agent.noRole': {
        en: 'Set once this step has a role.',
        vi: 'Sẽ có khi bước này có vai trò.',
    },
    'step.agent.fromRole': {
        en: 'Inherited from the role. Which model runs each effort level is set under Settings → Preferences.',
        vi: 'Kế thừa từ vai trò. Model chạy từng mức nỗ lực được đặt trong Cài đặt → Tùy chọn.',
    },
    'step.agent.fromExport': {
        en: 'Set when this workflow was exported. Which model runs each effort level is set under Settings → Preferences.',
        vi: 'Đặt lúc workflow này được export. Model chạy từng mức nỗ lực được đặt trong Cài đặt → Tùy chọn.',
    },

    'step.inputs.missing.one': {
        en: '{count} required input left',
        vi: 'Còn {count} đầu vào bắt buộc',
    },
    'step.inputs.missing.other': {
        en: '{count} required inputs left',
        vi: 'Còn {count} đầu vào bắt buộc',
    },
    'step.inputs.empty.one': {
        en: '{count} input still empty — the run goes ahead with the prompt as written.',
        vi: 'Còn {count} đầu vào trống — lần chạy vẫn dùng prompt đúng như đang viết.',
    },
    'step.inputs.empty.other': {
        en: '{count} inputs still empty — the run goes ahead with the prompt as written.',
        vi: 'Còn {count} đầu vào trống — lần chạy vẫn dùng prompt đúng như đang viết.',
    },

    'step.missing.title': {en: 'Fill in the empty inputs', vi: 'Điền các đầu vào còn trống'},
    'step.missing.asking.one': {
        en: '{count} step still asking. Anything left empty reaches the agent as written.',
        vi: '{count} bước còn thiếu. Chỗ nào để trống thì agent nhận đúng như đang viết.',
    },
    'step.missing.asking.other': {
        en: '{count} steps still asking. Anything left empty reaches the agent as written.',
        vi: '{count} bước còn thiếu. Chỗ nào để trống thì agent nhận đúng như đang viết.',
    },
    'step.missing.count.one': {en: '{count} input', vi: '{count} đầu vào'},
    'step.missing.count.other': {en: '{count} inputs', vi: '{count} đầu vào'},
    'step.missing.cancel': {en: 'Cancel', vi: 'Hủy'},
    'step.missing.run': {en: 'Save and run', vi: 'Lưu và chạy'},

    'step.new.title': {en: 'New step', vi: 'Bước mới'},
    'step.new.roles': {en: 'Roles', vi: 'Vai trò'},
    'step.new.create': {en: 'Create step', vi: 'Tạo bước'},
    'step.new.cancel': {en: 'Cancel', vi: 'Hủy'},

    'step.form.title': {en: 'Title', vi: 'Tiêu đề'},
    'step.form.titlePlaceholder': {
        en: 'Review the auth change',
        vi: 'Rà soát thay đổi phần auth',
    },
    'step.form.prompt': {en: 'Prompt', vi: 'Prompt'},
    'step.form.promptHint': {
        en: 'Starts as the role’s instructions. Edit it to change this step only. Wrap an input key in {{ }} to drop its value into the text.',
        vi: 'Bắt đầu từ chỉ dẫn của vai trò. Sửa ở đây chỉ đổi riêng bước này. Bọc tên đầu vào trong {{ }} để chèn giá trị của nó vào chỗ đó.',
    },
    'step.form.inputs': {en: 'Inputs', vi: 'Đầu vào'},

    'step.locked.description': {
        en: 'The graph is locked. Inputs stay open until the run starts.',
        vi: 'Sơ đồ đã khóa. Đầu vào vẫn sửa được cho tới khi bắt đầu chạy.',
    },
    'step.locked.roleGone': {
        en: 'This step’s role is gone, so its inputs cannot be read. Unlock the workflow and rebuild the step.',
        vi: 'Vai trò của bước này không còn nên không đọc được đầu vào. Mở khóa workflow rồi dựng lại bước.',
    },
    'step.locked.promptPreview': {
        en: 'Prompt the agent receives',
        vi: 'Prompt mà agent sẽ nhận',
    },
    'step.locked.cancel': {en: 'Cancel', vi: 'Hủy'},
    'step.locked.saving': {en: 'Saving…', vi: 'Đang lưu…'},
    'step.locked.save': {en: 'Save inputs', vi: 'Lưu đầu vào'},
} as const satisfies Catalog
