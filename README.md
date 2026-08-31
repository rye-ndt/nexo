# Nexo - Nexo EXecution Orchestrator

**Streamlined, debuggable, retryable agentic workflow orchestration - that you can set up yourself.**

[![License](https://img.shields.io/github/license/rye-ndt/nexo)](LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/rye-ndt/nexo)](go.mod)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)](#requirements)
[![Stars](https://img.shields.io/github/stars/rye-ndt/nexo?style=social)](https://github.com/rye-ndt/nexo/stargazers)

![Nexo running a graph of steps](docs/demo.gif)

- **[English](#english)**
- **[Tiếng Việt](#tiếng-việt)**

---

## English

You draw the work as a graph. Each step is one scoped unit of work: a role, a
model, the files it's allowed to touch, plus whatever guidance it needs. Nexo
runs agents through the graph, watches them, captures the diff, and feeds what
each step learned into the next step's prompt.

**Contents**

- [Why Nexo exists](#why-nexo-exists)
- [Requirements](#requirements)
- [Quick start](#quick-start)
- [MCP server](#mcp-server)
- [Support](#support)

### Why Nexo exists

The limitation of LLMs on long, multi-step work isn't the model, it's the
context. An agent that starts a step knowing only that step will improvise the
rest. So Nexo makes the handoff between steps the thing you engineer.

Claude Code, OpenCode and Codex are great at one step, and Nexo runs on top of
them anyway. But:

- can you delegate a 10-step job, walk away, and trust it without reading every line?
- can you make it follow _your_ procedure instead of improvising one?
- can you see why it chose what it chose?
- can you undo step 4 and keep steps 1 through 3?

Four things separate it from a wrapper:

**Steps are a DAG, not a queue.** Chaining is the entire point.

**The handoff is the product.** A finished step writes down what it did, what
blocked it, what got approved, what got _rejected_, and what's still broken.
That handoff becomes the next step's context. Telling the next agent what
**not** to do matters as much as telling it what happened.

**It assumes agents fail.** Every step write commits to SQLite as it happens, so
a crash costs you nothing already reported. Agents send heartbeats, and one that
goes quiet drops its step back into the pool. Every file change is stored as a
unified diff, so you can revert without depending on git.

**Agents never see your secrets.** Nexo does the OAuth once, encrypts the token,
and hands agents a placeholder. The proxy swaps in the real key on the way out,
and refuses anything aimed at an auth endpoint.

### Requirements

- **macOS** 11+ on Apple Silicon - where Nexo is developed and tested
- **Windows** 10+ on x64 - builds ship, lightly tested
- **Linux** x64, Ubuntu 24.04+ / Debian 13+ - builds ship, lightly tested. Older
  releases ship webkit2gtk 4.0 and cannot satisfy the dependency.

The coding agents themselves - Claude Code, OpenCode, Codex - are installed from
inside the app, so you don't need them beforehand.

### Quick start

Grab the installer for your platform from the
[Releases page](https://github.com/rye-ndt/nexo/releases):

| Platform | File | Install |
| --- | --- | --- |
| macOS | `Nexo-macos-arm64.dmg` | open it, drag Nexo to Applications |
| Windows | `Nexo-windows-amd64-setup.exe` | run it |
| Linux | `nexo-linux-amd64.deb` | `sudo apt install ./nexo-linux-amd64.deb` |

The builds are unsigned, so your OS will say so on first launch. On macOS,
right-click the app and choose Open, or run
`xattr -dr com.apple.quarantine /Applications/Nexo.app`. On Windows, SmartScreen
shows "unrecognised app": click More info, then Run anyway.

Then, in the app:

1. install and sign in to your agents - Claude Code, OpenCode, Codex - in settings
2. authorize any MCP servers you need
3. start a workflow and point it at a project folder
4. make roles, each one a kind of worker that does a kind of work
5. drop steps on the canvas and wire them together, mixing agents freely
6. lock the workflow, run it
7. go play pickleball
8. come back, review, revert what you don't like
9. commit

Steps 3-5 have a shortcut. Open the **Store** in the sidebar and add a ready-made
workflow: it arrives with its steps already wired and the roles they run on
already in your library, so you go straight to locking and running it. Open any
card first to read what you are about to take: every step's prompt, and every
role's instructions and report format. The store also lists the roles on their
own, to start your own graph from. Everything in it is free.

### MCP server

Nexo serves an MCP endpoint on a random port behind a token that changes every
launch, so a static HTTP registration would go stale daily. `cmd/nexo-mcp` is a
stdio server that reads the current port and token from the file the app writes
and forwards each call there. Register it once and it keeps working across
restarts.

Building the shim needs the repo and Go 1.25+:

```sh
git clone https://github.com/rye-ndt/nexo.git
cd nexo
make mcp-shim
claude mcp add nexo -- /full/path/to/nexo/build/bin/nexo-mcp
```

Use the absolute path, since Claude Code does not resolve it against your
project. Eight tools come across: `list_roles`, `create_workflow`,
`start_workflow`, `pause_workflow`, `cancel_workflow`, `workflow_status`,
`list_workflows` and `answer_review`.

The catch: **none of it works while the Nexo app is closed.** The shim only
forwards. Add `-launch` to have it open Nexo and wait, rather than failing the
call.

### Support

Questions and bugs go to [GitHub Issues](https://github.com/rye-ndt/nexo/issues).
Want help setting this up for your team, custom roles, or tuning it for your
company? Contact nduytung.1611@gmail.com.

Not vibe-coded: I built the backend and the architecture by hand. Claude helped
with the frontend.

---

## Tiếng Việt

Bạn vẽ công việc thành một đồ thị. Mỗi step là một đơn vị công việc có phạm vi
rõ ràng: một role, một model, những file nó được phép sửa, cùng với hướng dẫn nó
cần. Nexo cho agent chạy qua đồ thị đó, giám sát chúng, lưu lại diff, và đưa
những gì mỗi step học được vào prompt của step kế tiếp.

**Mục lục**

- [Vì sao có Nexo](#vì-sao-có-nexo)
- [Yêu cầu](#yêu-cầu)
- [Bắt đầu nhanh](#bắt-đầu-nhanh)
- [MCP server](#mcp-server-1)
- [Hỗ trợ](#hỗ-trợ)

### Vì sao có Nexo

Giới hạn của LLM trong những công việc dài, nhiều bước không nằm ở model mà nằm
ở context. Một agent bắt đầu một step mà chỉ biết mỗi step đó sẽ tự bịa ra phần
còn lại. Vì vậy Nexo biến handoff giữa các step thành thứ bạn chủ động thiết kế.

Claude Code, OpenCode và Codex làm rất tốt một step, và Nexo cũng chạy trên
chính chúng. Nhưng:

- bạn có dám giao một việc 10 bước, bỏ đi, rồi tin nó mà không đọc từng dòng không?
- bạn có bắt nó theo đúng _quy trình của bạn_ thay vì tự chế ra một quy trình không?
- bạn có thấy được vì sao nó chọn như vậy không?
- bạn có hoàn tác được step 4 mà vẫn giữ nguyên step 1 đến 3 không?

Bốn điểm khiến nó không chỉ là một lớp bọc:

**Các step là một DAG, không phải hàng đợi.** Việc nối chuỗi chính là toàn bộ
mục đích.

**Handoff mới là sản phẩm.** Một step hoàn tất sẽ ghi lại nó đã làm gì, cái gì
chặn nó, cái gì được duyệt, cái gì bị _từ chối_, và cái gì vẫn còn hỏng. Handoff
đó trở thành context của step kế tiếp. Nói cho agent tiếp theo biết điều **không
nên** làm cũng quan trọng như kể cho nó chuyện đã xảy ra.

**Nó mặc định agent sẽ hỏng.** Mọi thay đổi của step được ghi thẳng vào SQLite
ngay khi xảy ra, nên một cú crash không làm mất những gì đã báo cáo. Agent gửi
heartbeat, và agent nào im lặng quá hạn sẽ bị trả step về hàng chờ. Mọi thay đổi
file được lưu dưới dạng unified diff, nên bạn revert được mà không cần git.

**Agent không bao giờ thấy secret của bạn.** Nexo làm OAuth một lần, mã hoá
token, và đưa cho agent một placeholder. Proxy thay bằng key thật ở lượt gửi đi,
và chặn mọi request nhắm vào endpoint xác thực.

### Yêu cầu

- **macOS** 11+ trên Apple Silicon - nơi Nexo được phát triển và kiểm thử
- **Windows** 10+ trên x64 - có bản build, kiểm thử còn ít
- **Linux** x64, Ubuntu 24.04+ / Debian 13+ - có bản build, kiểm thử còn ít. Các
  bản phát hành cũ hơn chỉ có webkit2gtk 4.0 nên không đáp ứng được phụ thuộc.

Bản thân các coding agent - Claude Code, OpenCode, Codex - được cài từ trong ứng
dụng, nên bạn không cần chuẩn bị trước.

### Bắt đầu nhanh

Tải file cài đặt cho nền tảng của bạn ở
[trang Releases](https://github.com/rye-ndt/nexo/releases):

| Nền tảng | File | Cách cài |
| --- | --- | --- |
| macOS | `Nexo-macos-arm64.dmg` | mở lên, kéo Nexo vào Applications |
| Windows | `Nexo-windows-amd64-setup.exe` | chạy file |
| Linux | `nexo-linux-amd64.deb` | `sudo apt install ./nexo-linux-amd64.deb` |

Các bản build chưa được ký, nên hệ điều hành sẽ cảnh báo ở lần mở đầu tiên. Trên
macOS, chuột phải vào ứng dụng rồi chọn Open, hoặc chạy
`xattr -dr com.apple.quarantine /Applications/Nexo.app`. Trên Windows,
SmartScreen báo "unrecognised app": bấm More info, rồi Run anyway.

Sau đó, trong ứng dụng:

1. cài và đăng nhập các agent - Claude Code, OpenCode, Codex - trong phần settings
2. cấp quyền cho những MCP server bạn cần
3. tạo một workflow và trỏ nó tới thư mục dự án
4. tạo các role, mỗi role là một kiểu thợ làm một kiểu việc
5. thả các step lên canvas và nối chúng lại, trộn nhiều agent thoải mái
6. khoá workflow lại, chạy
7. đi chơi pickleball
8. quay lại, xem kết quả, revert những gì bạn không ưng
9. commit

Bước 3-5 có lối tắt. Mở **Store** ở thanh bên rồi thêm một workflow có sẵn: nó về
với các step đã nối sẵn và các role đi kèm đã nằm trong thư viện của bạn, nên bạn
đi thẳng tới khoá và chạy. Bấm vào một thẻ trước để đọc kỹ thứ mình sắp lấy: prompt
của từng step, chỉ dẫn và định dạng báo cáo của từng role. Store cũng liệt kê riêng
các role, để bạn tự dựng graph của mình. Mọi thứ trong đó đều miễn phí.

### MCP server

Nexo phục vụ một endpoint MCP trên cổng ngẫu nhiên sau một token đổi mỗi lần
khởi động, nên một đăng ký HTTP tĩnh sẽ hỏng sau mỗi ngày. `cmd/nexo-mcp` là một
stdio server đọc cổng và token hiện tại từ file mà ứng dụng ghi ra rồi chuyển
tiếp từng lời gọi tới đó. Đăng ký một lần là dùng được qua mọi lần khởi động lại.

Build shim này cần repo và Go 1.25+:

```sh
git clone https://github.com/rye-ndt/nexo.git
cd nexo
make mcp-shim
claude mcp add nexo -- /full/path/to/nexo/build/bin/nexo-mcp
```

Dùng đường dẫn tuyệt đối, vì Claude Code không phân giải đường dẫn theo dự án
của bạn. Tám tool được đưa sang: `list_roles`, `create_workflow`,
`start_workflow`, `pause_workflow`, `cancel_workflow`, `workflow_status`,
`list_workflows` và `answer_review`.

Điểm cần lưu ý: **không thứ gì hoạt động khi ứng dụng Nexo đang đóng.** Shim chỉ
chuyển tiếp. Thêm `-launch` để nó mở Nexo và chờ, thay vì làm hỏng lời gọi.

### Hỗ trợ

Câu hỏi và lỗi xin gửi vào
[GitHub Issues](https://github.com/rye-ndt/nexo/issues). Cần hỗ trợ triển khai
cho đội của bạn, viết role riêng, hay tinh chỉnh cho công ty? Liên hệ
nduytung.1611@gmail.com.

Không phải vibe-code: tôi tự viết tay phần backend và kiến trúc. Claude phụ giúp
phần frontend.

## License

[MIT](LICENSE), v1.2.0
