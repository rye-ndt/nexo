import type {Catalog} from '@/shared/lib/i18n/messages/catalog'

export const ROLE_MESSAGES = {
    'role.panel.title': {en: 'Roles your steps start from', vi: 'Vai trò để dựng các bước'},
    'role.panel.pickTitle': {en: 'Pick a role', vi: 'Chọn một vai trò'},
    'role.panel.subtitle': {
        en: 'One kind of work: the agent’s role, its inputs, and how hard to try.',
        vi: 'Một loại công việc: agent đóng vai gì, cần đầu vào nào, và cố gắng tới đâu.',
    },
    'role.panel.transfer': {
        en: 'Roles travel as a .json file.',
        vi: 'Vai trò đi kèm nhau trong một file .json.',
    },

    'role.action.new': {en: 'New role', vi: 'Vai trò mới'},
    'role.action.import': {en: 'Import', vi: 'Nhập file'},
    'role.action.export': {en: 'Export', vi: 'Xuất file'},

    'role.list.loading': {en: 'Loading roles…', vi: 'Đang tải vai trò…'},
    'role.list.empty': {
        en: 'No roles yet. A role describes one kind of work — the agent’s role, the inputs it needs, and how hard it should try.',
        vi: 'Chưa có vai trò nào. Một vai trò mô tả một loại công việc — agent đóng vai gì, cần đầu vào nào, và cố gắng tới đâu.',
    },

    'role.card.structured': {en: 'Structured', vi: 'Có cấu trúc'},
    'role.card.noRetry': {en: 'no retry', vi: 'không thử lại'},
    'role.card.noDescription': {en: 'No description set.', vi: 'Chưa có mô tả.'},
    'role.card.noInputs': {en: 'No inputs', vi: 'Không có đầu vào'},
    'role.card.edit': {en: 'Edit {name}', vi: 'Sửa {name}'},
    'role.card.remove': {en: 'Delete {name}', vi: 'Xóa {name}'},
    'role.card.removeTitle': {en: 'Delete “{name}”?', vi: 'Xóa “{name}”?'},
    'role.card.removeBody': {
        en: 'Steps already built from it keep their prompt. This cannot be undone.',
        vi: 'Các bước đã dựng từ nó vẫn giữ prompt của mình. Không hoàn tác được.',
    },
    'role.card.removeConfirm': {en: 'Delete role', vi: 'Xóa vai trò'},

    'role.form.titleNew': {en: 'New role', vi: 'Vai trò mới'},
    'role.form.titleEdit': {en: 'Edit role', vi: 'Sửa vai trò'},
    'role.form.create': {en: 'Create role', vi: 'Tạo vai trò'},
    'role.form.save': {en: 'Save changes', vi: 'Lưu thay đổi'},
    'role.form.saving': {en: 'Saving', vi: 'Đang lưu'},
    'role.form.cancel': {en: 'Cancel', vi: 'Hủy'},
    'role.form.undo': {en: 'Undo', vi: 'Hoàn tác'},
    'role.form.filling': {
        en: 'An agent is reading the project and writing this. It can take a few minutes.',
        vi: 'Một agent đang đọc dự án và viết phần này. Có thể mất vài phút.',
    },
    'role.form.saveTitle': {en: 'Save changes to “{name}”?', vi: 'Lưu thay đổi cho “{name}”?'},
    'role.form.saveBody': {
        en: 'The role is overwritten. Steps already built from it keep their prompt.',
        vi: 'Vai trò bị ghi đè. Các bước đã dựng từ nó vẫn giữ prompt của mình.',
    },
    'role.form.discardEditTitle': {
        en: 'Discard changes to “{name}”?',
        vi: 'Bỏ thay đổi cho “{name}”?',
    },
    'role.form.discardNewTitle': {en: 'Discard this role?', vi: 'Bỏ vai trò này?'},
    'role.form.discardEditBody': {
        en: 'The edits made here are lost and the saved role stays as it was.',
        vi: 'Những sửa đổi ở đây sẽ mất, vai trò đã lưu giữ nguyên.',
    },
    'role.form.discardNewBody': {
        en: 'Nothing is saved, and everything filled in here is lost.',
        vi: 'Không có gì được lưu, mọi thứ ở đây sẽ mất.',
    },
    'role.form.discard': {en: 'Discard', vi: 'Bỏ'},
    'role.form.keepEditing': {en: 'Keep editing', vi: 'Tiếp tục sửa'},

    'role.form.name': {en: 'Name', vi: 'Tên'},
    'role.form.namePlaceholder': {en: 'Code reviewer', vi: 'Người review code'},
    'role.form.description': {en: 'What it does', vi: 'Vai trò này làm gì'},
    'role.form.descriptionPlaceholder': {
        en: 'Reads a diff and reports the defects it can prove.',
        vi: 'Đọc diff và báo cáo những lỗi nó chứng minh được.',
    },
    'role.form.effort': {en: 'Effort', vi: 'Mức nỗ lực'},
    'role.form.retryable': {en: 'Retry on failure', vi: 'Thử lại khi lỗi'},
    'role.form.pauseForReview': {en: 'Pause for my review', vi: 'Dừng chờ tôi kiểm duyệt'},
    'role.form.inputs': {en: 'Inputs', vi: 'Đầu vào'},
    'role.form.inputsHint': {en: 'What a step fills in.', vi: 'Tham số đầu vào.'},
    'role.form.addInput': {en: 'Add input', vi: 'Thêm đầu vào'},
    'role.form.inputRefHint': {
        en: 'Wrap an input key in {{ }} to drop its value into the text.',
        vi: 'Bọc key của đầu vào trong {{ }} để chèn giá trị vào văn bản.',
    },
    'role.form.reportFormat': {en: 'Report format', vi: 'Định dạng báo cáo'},
    'role.form.reportFormatHint': {
        en: 'One field per line, written as name: what goes in it. Indent two spaces to nest, and start a line with a dash to describe one element of a list. Leave this empty and the step reports in its own words.',
        vi: 'Mỗi dòng một trường, viết theo dạng tên: nội dung. Thụt vào hai dấu cách để lồng cấp, và bắt đầu dòng bằng dấu gạch ngang để mô tả một phần tử. Để trống thì bước sẽ báo cáo tự do.',
    },
    'role.form.advanced': {en: 'Advanced', vi: 'Nâng cao'},
    'role.form.advancedHint': {
        en: 'Inputs, effort, review, report format',
        vi: 'Đầu vào, nỗ lực, kiểm duyệt, định dạng báo cáo',
    },
    'role.form.inputCountOne': {en: '{count} input', vi: '{count} đầu vào'},
    'role.form.inputCountOther': {en: '{count} inputs', vi: '{count} đầu vào'},
    'role.form.reviewTag': {en: 'pauses for review', vi: 'dừng chờ kiểm duyệt'},
    'role.form.conflictAware': {en: 'Conflict aware', vi: 'Tránh xung đột'},
    'role.form.instructions': {en: 'Instructions', vi: 'Chỉ dẫn'},
    'role.form.addInstruction': {en: 'Add instruction', vi: 'Thêm chỉ dẫn'},

    'role.input.key': {en: 'Input key', vi: 'Key của đầu vào'},
    'role.input.type': {en: 'Input type', vi: 'Kiểu đầu vào'},
    'role.input.remove': {en: 'Remove input', vi: 'Xóa đầu vào'},
    'role.input.label': {en: 'Label shown on the step', vi: 'Nhãn hiện trên bước'},
    'role.input.labelPlaceholder': {en: 'Directory to review', vi: 'Thư mục cần review'},
    'role.input.default': {en: 'Default value', vi: 'Giá trị mặc định'},
    'role.input.multiDefault': {
        en: 'Default picks, separated by commas',
        vi: 'Lựa chọn mặc định, cách nhau bằng dấu phẩy',
    },
    'role.input.options': {
        en: 'Options, separated by commas',
        vi: 'Các lựa chọn, cách nhau bằng dấu phẩy',
    },
    'role.input.required': {en: 'Required', vi: 'Bắt buộc'},

    'role.instruction.key': {en: 'Instruction key', vi: 'Key của chỉ dẫn'},
    'role.instruction.remove': {en: 'Remove instruction', vi: 'Xóa chỉ dẫn'},
    'role.instruction.placeholder': {
        en: 'Fetch the ticket at https://jira/browse/{{ticket_id}} and summarise it.',
        vi: 'Lấy ticket tại https://jira/browse/{{ticket_id}} rồi tóm tắt lại.',
    },
    'role.instruction.text': {en: 'Instruction text', vi: 'Nội dung chỉ dẫn'},

    'role.field.pickOne': {en: 'Pick one', vi: 'Chọn một'},
    'role.field.pickAny': {en: 'Pick any', vi: 'Chọn một hoặc nhiều'},
    'role.field.noFile': {en: 'No file chosen', vi: 'Chưa chọn file'},
    'role.field.changeFile': {en: 'Change', vi: 'Đổi'},
    'role.field.chooseFile': {en: 'Choose', vi: 'Chọn'},
    'role.field.filePicker': {en: 'Choose a file for {label}', vi: 'Chọn file cho {label}'},

    'role.prompt.unknownOne': {
        en: '{refs} is not an input on this role, so it reaches the agent as written.',
        vi: '{refs} không phải tên tham số hợp lệ.',
    },
    'role.prompt.unknownOther': {
        en: '{refs} are not inputs on this role, so they reach the agent as written.',
        vi: '{refs} không phải tên tham số hợp lệ.',
    },

    'role.helper.fill': {en: 'Fill in the rest', vi: 'Tự điền chỗ trống'},
    'role.helper.filling': {en: 'Filling in', vi: 'Đang điền'},
    'role.helper.needsInstall': {
        en: 'Install Claude Code to use this.',
        vi: 'Cài Claude Code để dùng chức năng này.',
    },
    'role.helper.needsLogin': {
        en: 'Log in to Claude Code to use this.',
        vi: 'Đăng nhập Claude Code để dùng chức năng này.',
    },

    'role.export.title': {en: 'Export roles', vi: 'Xuất vai trò'},
    'role.export.description': {
        en: 'They go into one .json file you choose the place for.',
        vi: 'Tất cả nằm trong một file .json, bạn chọn nơi lưu.',
    },
    'role.export.selectAll': {en: 'Select all', vi: 'Chọn tất cả'},
    'role.export.clear': {en: 'Clear', vi: 'Bỏ chọn'},
    'role.export.hint': {
        en: 'Tick the ones to take with you.',
        vi: 'Đánh dấu những vai trò bạn muốn mang theo.',
    },
    'role.export.selected': {en: '{count} of {total} selected', vi: 'Đã chọn {count} trên {total}'},
    'role.export.confirmOne': {en: 'Export {count} role', vi: 'Lưu {count} vai trò'},
    'role.export.confirmOther': {en: 'Export {count} roles', vi: 'Lưu {count} vai trò'},
    'role.export.working': {en: 'Exporting roles', vi: 'Đang lưu vai trò'},
    'role.export.workingBody': {
        en: 'Writing the file. Hold on.',
        vi: 'Đang ghi file.',
    },
    'role.export.doneOne': {en: 'Exported {count} role', vi: 'Đã lưu {count} vai trò'},
    'role.export.doneOther': {en: 'Exported {count} roles', vi: 'Đã lưu {count} vai trò'},
    'role.export.doneBody': {
        en: 'The file is yours to keep, move, or hand to someone else.',
        vi: 'File là của bạn: giữ lại, chuyển đi, hoặc đưa cho người khác.',
    },

    'role.import.title': {en: 'Import roles', vi: 'Nhập vai trò'},
    'role.import.working': {en: 'Importing roles', vi: 'Đang nhập vai trò'},
    'role.import.workingBody': {
        en: 'Reading the file. Hold on.',
        vi: 'Đang đọc file.',
    },
    'role.import.doneOne': {en: 'Imported {count} role', vi: 'Đã nhập {count} vai trò'},
    'role.import.doneOther': {en: 'Imported {count} roles', vi: 'Đã nhập {count} vai trò'},
    'role.import.doneBody': {
        en: 'They are ready to build steps from.',
        vi: 'Đã sẵn sàng để dựng bước từ chúng.',
    },

    'role.issue.name': {en: 'Give the role a name.', vi: 'Đặt tên cho vai trò.'},
    'role.issue.inputKey': {
        en: 'Every input needs a key the agent can read.',
        vi: 'Mỗi đầu vào cần một key agent đọc được.',
    },
    'role.issue.duplicateInput': {
        en: 'Two inputs share the same key.',
        vi: 'Hai đầu vào trùng key.',
    },
    'role.issue.options': {
        en: 'A choice input needs at least one option.',
        vi: 'Đầu vào dạng lựa chọn cần ít nhất một lựa chọn.',
    },
    'role.issue.noInstructions': {
        en: 'A role needs at least one instruction.',
        vi: 'Vai trò cần ít nhất một chỉ dẫn.',
    },
    'role.issue.instructionKey': {
        en: 'Every instruction needs a key.',
        vi: 'Mỗi chỉ dẫn cần một key.',
    },

    'role.structure.atLine': {en: 'Line {line}: {message}', vi: 'Dòng {line}: {message}'},
    'role.structure.empty': {en: 'Describe at least one field.', vi: 'Mô tả ít nhất một trường.'},
    'role.structure.tabs': {
        en: 'Indent with spaces, not tabs.',
        vi: 'Thụt lề bằng dấu cách, không dùng tab.',
    },
    'role.structure.bulletEmpty': {
        en: 'A list item needs a field after the dash.',
        vi: 'Một mục danh sách cần một trường sau dấu gạch ngang.',
    },
    'role.structure.noColon': {
        en: 'A field needs a colon between its name and what goes in it.',
        vi: 'Một trường cần dấu hai chấm giữa tên và nội dung của nó.',
    },
    'role.structure.badName': {
        en: 'A field name can only use letters, numbers and underscores, and cannot start with a number.',
        vi: 'Tên trường chỉ dùng chữ cái, chữ số và dấu gạch dưới, và không bắt đầu bằng chữ số.',
    },
    'role.structure.duplicate': {
        en: 'Two fields at this level are both called "{key}".',
        vi: 'Hai trường cùng cấp đều tên là "{key}".',
    },
    'role.structure.listOneItem': {
        en: 'A list describes one element and the agent repeats it as often as the work needs, so it can only hold one item.',
        vi: 'Danh sách mô tả một phần tử và agent lặp lại bao nhiêu lần tùy công việc, nên chỉ được có một mục.',
    },
    'role.structure.listNeedsOwner': {
        en: 'A list needs a field above it to name what the list holds.',
        vi: 'Danh sách cần một trường phía trên để gọi tên thứ nó chứa.',
    },
    'role.structure.indentStep': {
        en: 'Indent two spaces for each level.',
        vi: 'Thụt vào hai dấu cách cho mỗi cấp.',
    },
    'role.structure.indentDeep': {
        en: 'This field is indented too far for what it sits under.',
        vi: 'Trường này thụt sâu quá so với trường nó nằm dưới.',
    },
    'role.structure.indentOff': {
        en: 'This field is not lined up with the fields it sits beside.',
        vi: 'Trường này không thẳng hàng với các trường cùng cấp.',
    },
    'role.structure.groupEmpty': {
        en: '"{key}" needs either a description or fields indented under it.',
        vi: '"{key}" cần một mô tả, hoặc các trường thụt vào bên dưới.',
    },

    'role.error.filePicker': {
        en: 'Could not open the file picker',
        vi: 'Không mở được hộp thoại chọn file',
    },
    'role.error.nameRequired': {
        en: 'A role needs a name before it can be saved.',
        vi: 'Vai trò cần có tên trước khi lưu.',
    },
    'role.error.pickToExport': {
        en: 'Pick at least one role to export.',
        vi: 'Chọn ít nhất một vai trò để xuất.',
    },
    'role.error.roleGone': {
        en: 'One of those roles is no longer here.',
        vi: 'Một trong những vai trò đó không còn nữa.',
    },
    'role.error.load': {en: 'Could not load your roles', vi: 'Không tải được vai trò của bạn'},
    'role.error.save': {en: 'Could not save the role', vi: 'Không lưu được vai trò'},
    'role.error.remove': {en: 'Could not delete the role', vi: 'Không xóa được vai trò'},
    'role.error.export': {en: 'Could not export your roles', vi: 'Không xuất được vai trò của bạn'},
    'role.error.import': {en: 'Could not import that file', vi: 'Không nhập được file đó'},
    'role.error.fill': {en: 'Could not fill this role in', vi: 'Không điền được vai trò này'},
} as const satisfies Catalog
