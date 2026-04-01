# Summary

一个 Agent skill 项目——Agent first 的记账程序。

# 特点

- Agent first
- cli first
- 接口简单，便于 Agent 使用

# 主要功能

- 基础的增删改查
- 使用 sqlite，数据存放在 `/data/`
- 完善的 `SKILL.md` 
- 暴露接口
	- `ledger add`：添加账目
	- `ledger list`：硬边界列举，针对字段过滤
	- `ledger search` ：软边界搜索，主要针对 `note` 字段搜索
	- `ledger get`：根据 id 获取特定账目
	- `ledger update`：根据 id 更新特定账目
	- `ledger delete`：根据 id 删除特定账目
	- `ledger help`: 解释命令使用

# 数据库 schema

所有都是必填
- id
- datetime
- amount
- currency: 货币类型充当账户作用
- category
- note：没有合适的话也要填 none
- created_at / updated_at

# 代码组织

```
ledger-skill/
├─ ledger-cli/               # Go 源码项目
│  ├─ go.mod
│  ├─ go.sum
│  ├─ main.go
│  ├─ cmd/                   # 定义命令和 flag，接收输入参数，调用 `internal/app`，输出结果
|  │  ├─ root.go             # 根命令
|  │  ├─ add.go
|  │  ├─ list.go             # 硬边界列举，针对字段过滤
|  │  ├─ search.go           # 软边界搜索，主要针对 `note` 字段搜索
|  │  ├─ update.go
|  │  └─ delete.go
|  │  └─ help.go             # 对各个命令的使用方法的解释，和一般的 cli 程序一致 
│  └─ internal/
|     ├─ app.go              # 应用层逻辑
|     ├─ store.go            # 存储实现
|     ├─ entry.go            # 统一输出 JSON，用来将结果返回给 Agent
|     └─ output.go
├─ skill-template/           # skill 模板目录
│  ├─ SKILL.md
│  ├─ scripts/               # 空文件夹，理论上在 dist/ 里面才有内容
│  │  └─ ledger              # 编译好的二进制
│  │  └─ ledger.sh           # 很薄的包装脚本
│  ├─ references/
│  └─ examples/              # 不是对每个命令的解释，而是对使用场景的解释，列举场景，以及解决该场景的命令组合
├─ dist/                     # 构建产物，按 Linux, macOS 生成，Github release 目录
│  ├─ ledger-skill-linux-amd64/
│  └─ ledger-skill-darwin-arm64/
├─ Makefile
└─ README.md
```
