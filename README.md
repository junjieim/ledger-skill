# ledger-skill

`ledger-skill` 是一个面向 Agent 的本地记账技能仓库，采用 `Agent first` 和 `CLI first` 的设计思路：把记账能力收敛为一个窄而稳定的命令行接口，再通过 skill 模板暴露给 Codex 或其他支持本地技能的 Agent 使用。

当前仓库包含两部分：

- `ledger-cli/`：Go 实现的本地 CLI，负责数据校验、命令分发和 SQLite 持久化。
- `skill-template/`：技能模板，包含 `SKILL.md`、CLI 参考文档、场景示例和脚本入口。

日常开发和构建默认通过仓库根目录的 `Makefile` 统一入口完成，底层实现仍然位于 `ledger-cli/`。

## 设计目标

- 对 Agent 友好：除 `help` 外，所有数据命令都输出稳定的 JSON。
- 对脚本友好：命令参数简单、字段明确、错误码固定。
- 本地优先：使用 SQLite，无需外部服务。
- 依赖克制：核心只依赖 Go 标准库和 `modernc.org/sqlite`。

## 当前能力

- `add`：新增账目
- `list`：按 `currency`、`category`、时间范围做精确过滤
- `search`：仅针对 `note` 字段做不区分大小写搜索
- `get`：按 `id` 获取单条记录
- `update`：按 `id` 局部更新记录
- `delete`：按 `id` 删除记录
- `help`：输出文本帮助

此外，CLI 还会自动：

- 初始化 SQLite schema
- 将所有时间规范化为 UTC 的 RFC3339 字符串
- 将空 `note` 规范为 `none`
- 统一返回 `invalid_argument`、`not_found`、`internal` 三类错误码

## 架构概览

运行路径如下：

```text
Agent / User
    |
    v
skill-template/scripts/ledger.sh
    |
    v
compiled ledger binary
    |
    v
ledger-cli/cmd
    |
    v
ledger-cli/internal/app
    |
    v
ledger-cli/internal/store
    |
    v
SQLite database (../data/ledger.db relative to the binary)
```

分层职责：

- `cmd/`：命令分发、flag 解析、帮助文本、JSON 输出
- `internal/app.go`：应用层，负责校验、规范化、错误映射、更新时间戳
- `internal/entry.go`：数据模型与输入规范化逻辑
- `internal/store.go`：SQLite 存储实现和 schema 初始化
- `internal/output.go`：统一响应结构

这套架构的核心约束是：CLI 层尽量薄，业务规则集中在应用层，持久化细节收敛在 store 层。

## 仓库结构

```text
ledger-skill/
├─ architecture.md
├─ Makefile
├─ ledger-cli/
│  ├─ go.mod
│  ├─ main.go
│  ├─ cmd/
│  └─ internal/
└─ skill-template/
   ├─ SKILL.md
   ├─ agents/
   ├─ examples/
   ├─ references/
   └─ scripts/
```

补充说明：

- Go module 不在仓库根目录，而在 `ledger-cli/` 下；`Makefile` 负责把根目录命令映射到该模块。
- `skill-template/scripts/ledger.sh` 是面向 Unix-like 环境的薄包装脚本。
- 数据库目录不会预先提交，程序首次运行时会自动创建 `data/ledger.db`。

## 数据模型

每条账目最终都会以如下字段存储：

- `id`
- `datetime`
- `amount`
- `currency`
- `category`
- `note`
- `created_at`
- `updated_at`

字段规则：

- `datetime`：必填，输入必须是 RFC3339，存储时统一转成 UTC
- `amount`：必填，使用十进制字符串，如 `10`、`10.50`、`-5.25`
- `currency`：必填，仅支持 `RMB`、`HKD`、`USD`、`EUR`、`JPY`、`GBP`、`AUD`、`CAD`、`SGD`、`TWD`
- `category`：必填，精确匹配字段
- `note`：存储层视为必填；如果输入为空，会自动写成 `none`

## 命令接口

命令概览：

| 命令 | 用途 | 说明 |
| --- | --- | --- |
| `add` | 新增记录 | 返回单条 entry |
| `list` | 精确过滤列表 | 按字段硬过滤 |
| `search` | 按备注搜索 | 只搜索 `note` |
| `get` | 获取单条记录 | 通过 `id` |
| `update` | 更新单条记录 | 至少传一个待更新字段 |
| `delete` | 删除单条记录 | 返回被删除的 `id` |
| `help` | 查看帮助 | 唯一的纯文本命令 |

CLI 用法：

```bash
ledger add --datetime <RFC3339> --amount <decimal> --currency <currency_code> --category <text> [--note <text>]
ledger list [--currency <currency_code>] [--category <text>] [--from <RFC3339>] [--to <RFC3339>] [--limit <n>]
ledger search --query <text> [--limit <n>]
ledger get <id>
ledger update <id> [--datetime <RFC3339>] [--amount <decimal>] [--currency <currency_code>] [--category <text>] [--note <text>]
ledger delete <id>
ledger help [command]
```

行为约束：

- `list` 使用精确字段过滤，不做模糊搜索。
- `currency` 仅允许上述 10 个预设值；未显式指定时，建议 Agent 默认使用 `RMB`。
- `search` 仅对 `note` 做不区分大小写匹配。
- `list` 和 `search` 都按 `datetime DESC, created_at DESC` 排序。
- `limit=0` 表示不限制结果数量。

## JSON 输出契约

除 `help` 外，所有数据命令都输出以下结构：

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

失败时：

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "invalid_argument",
    "message": "amount is required"
  }
}
```

已实现错误码：

- `invalid_argument`
- `not_found`
- `internal`

`data` 的形态取决于命令：

- `add`、`get`、`update`：单条 entry
- `list`、`search`：entry 数组
- `delete`：`{"id":"..."}` 对象

## 快速开始

下面的运行示例默认从仓库根目录执行。

### 1. 构建 CLI

通过 `Makefile` 构建：

```bash
make build
```

当前 `make build` 会在 `ledger-cli/` 目录下生成 `ledger` 二进制（即 `ledger-cli/ledger`）。

如果你想绕过 `Makefile` 直接执行底层 Go 命令，等价方式是：

```bash
go -C ledger-cli build -o ledger .
```

说明：

- 直接运行 `ledger-cli/ledger` 时，默认数据库路径会按二进制所在目录解析
- `OpenSQLiteStore` 会自动创建缺失的 `data/` 目录和表结构

### 2. 查看帮助

Unix-like 环境：

```bash
skill-template/scripts/ledger.sh help
```

Windows 本地：

```powershell
.\skill-template\scripts\ledger.exe help
```

### 3. 添加并查询一条记录

```bash
skill-template/scripts/ledger.sh add \
  --datetime 2026-04-01T12:30:00+08:00 \
  --amount 68.00 \
  --currency HKD \
  --category food \
  --note "team lunch"

skill-template/scripts/ledger.sh list --category food --limit 5
```

### 4. 根据备注搜索并修正记录

```bash
skill-template/scripts/ledger.sh search --query taxi --limit 10
skill-template/scripts/ledger.sh update <id> --amount 85.00
skill-template/scripts/ledger.sh get <id>
```

## Skill 模板说明

`skill-template/` 已经包含可交付 skill 的骨架：

- `SKILL.md`：告诉 Agent 何时使用这个技能，以及如何选择命令
- `references/cli-reference.md`：CLI 契约、字段规则、错误码和输出结构
- `examples/`：面向场景的命令组合示例
- `agents/openai.yaml`：技能显示名称和默认提示词
- `scripts/ledger.sh`：执行本地二进制的薄包装脚本

这意味着仓库不只是一个 CLI 程序，也是在为“可被 Agent 调用的本地技能”做打包准备。

## 常用命令

默认通过根目录 `Makefile`：

```bash
make build
make fmt
make test
make vet
```

这几个目标对应的底层动作为：

- `make build`：执行 `go -C ledger-cli build -o ledger .`
- `make fmt`：执行 `go -C ledger-cli fmt ./...`
- `make test`：执行 `go -C ledger-cli test ./...`
- `make vet`：执行 `go -C ledger-cli vet ./...`

## 开发与验证

优先使用根目录 `Makefile`：

```bash
make fmt
make test
make vet
```

如果需要直接执行底层 Go 命令，再进入 `ledger-cli/`：

```bash
cd ledger-cli
gofmt -w <changed-files>
go fmt ./...
go test ./...
go vet ./...
```

如果未来有依赖调整，再补充执行：

```bash
cd ledger-cli
go mod tidy
```

## 适用场景

这个项目适合以下场景：

- 让 Agent 记录轻量级支出或收支流水
- 用本地 SQLite 保存记账数据，避免外部服务依赖
- 用稳定 JSON 接口把 CLI 接入自动化流程
- 为 Codex/OpenAI 风格的本地 skill 提供一个最小但完整的记账示例

如果目标是多人协作记账、复杂报表、账户体系或远程同步，这个仓库目前还不是那个层级的系统；它更偏向一个面向 Agent 的本地原子能力。
