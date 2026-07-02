# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概要

LeetMate 是跑在终端里的 LeetCode 刷题辅导 TUI（Go + bubbletea）。核心产品理念：**不给答案，只给提示**——卡住时按级给提示（Hint/Nudge/Review），只有 Answer 级（二次确认）才输出完整代码。练过的题进 FSRS 间隔复习队列（开发中）。

## 常用命令

```bash
# 构建（产物为 ./leetmate）
go build -o leetmate ./cmd/leetmate

# 运行（需先配好 leetgo 工作区 + 至少一个 LLM key；缺 key 时退化为「仅 leetgo」模式）
./leetmate

# 跑全部单测（默认排除 integration 构建标签）
go test ./...

# 跑单个包 / 单个测试
go test ./internal/coach/...
go test ./internal/store -run TestStoreRoundtrip

# 带 leetgo 真实工作区的集成测试（默认关闭，会真的调用 leetgo CLI）
go test -tags=integration ./internal/leetgo/...

# vet
go vet ./...
```

外部依赖：[leetgo](https://github.com/j178/leetgo) CLI 必须在 PATH 中且工作区已 `leetgo init`；LLM key 通过 preset 机制提供（见下）。

## 架构

严格分层、依赖单向流动，`internal/domain` 是所有层都能依赖的纯数据层：

```
cmd/leetmate (main: 组装依赖 → 启动 bubbletea)
  │
  ├── config      配置加载（yaml + .env + LEETMATE_* 环境变量）
  ├── domain      纯实体：Problem / Attempt / Card / Conversation / Result（无 SQL、无 LLM SDK）
  │
  ├── leetgo      adapter：shell out 调外部 leetgo CLI（pick/test/submit），解析其结构化输出
  ├── llm         provider 路由 + 流式聊天（gemini.go 原生；openai.go 走 OpenAI 兼容协议）
  ├── store       SQLite 持久化（modernc 纯 Go 驱动，无 CGO，schema.sql 嵌入）
  │
  ├── coach       产品灵魂：拼装分级 prompt + 流式返回 LLM 回复，系统提示词守住「前三级不泄答案」
  ├── studyplan   内置题单（热题100/面试150，data/*.json embed）+ 用户自定义 yaml + 进度跟踪
  │
  └── tui         bubbletea 根 Model 状态机 + 各视图
```

### TUI 是状态机驱动
`tui/app.go` 的根 `Model` 用 `view` 常量（`viewMenu`/`viewPractice`/`viewPlanList`/`viewPlanItems`）切视图。所有副作用（leetgo 命令、LLM 流、写库）都封装成 `tea.Cmd`，结果以 `*Msg` 回流到 `Update`——`cmds.go` 是这些 Cmd/Msg 的集中定义处，是理解异步数据流（pick→test→submit、coach 流式）的入口。

### LLM provider 路由
`llm.New(cfg)` 按 `cfg.Provider` 选 `gemini.go`（Google 原生 API）或 `openai.go`（任意 OpenAI 兼容端点）。两者都实现 `Provider.Chat()`，返回 `<-chan Chunk` 流式通道；`Chunk.Err` 非空表示中途失败，通道关闭表示正常结束。**新增 provider 时实现该接口即可，coach/TUI 无需改动。**

### Preset 机制（用户配置的核心）
`config/presets.go` 里 `Presets` 是内置 profile 列表。用户在 `config.yaml` 设 `llm.preset: siliconflow`，provider/base_url/model/api_key_env 自动填好，只需把对应 key 放进 `.env`。`gemini` 是全局默认，`siliconflow` 是国内推荐。**新增平台：往 `Presets` 加一项即可。**

### Coach 的分级提示词（核心产品逻辑）
`coach/prompts/*.md` 用 `//go:embed` 编进二进制。`system.md` 是防代答护栏（前三级绝不输出完整代码），`hint/nudge/review/answer.md` 是各级追加指令。改提示词行为只动这些 markdown + `coach.go` 的 `buildMessages`，不要在别处硬编码提示词。LLM 是**可选依赖**——无 key 时 `Coach` 为 nil，main.go 已处理降级。

### leetgo 集成约定
**绝不 import leetgo 的内部包**（非稳定 API），只 shell out 到 CLI 并解析输出（`leetgo/parse.go`）。所有命令在配置的 leetgo workspace（含 `leetgo.yaml` 的目录）下执行。

### 持久化
`store/` 用 modernc 纯 Go SQLite 驱动（便于交叉编译，无 CGO）。`schema.sql` 嵌入并在 `Open` 时执行。表：`problems` / `attempts` / `cards`（FSRS 调度状态，WIP）/ `conversations` / `weakness_tags` / `studyplan_progress`。

### i18n
`tui/i18n.go` 的轻量字典：所有面向用户的字符串按 key + language 查表，缺翻译回退英文。加语言加列、加字符串加 key。

## 约定

- 用户可见字符串必须走 `tui/i18n.go` 的 `dict.t()`，不要散落硬编码。
- 新的异步副作用：在 `cmds.go` 里定义 Cmd + 对应 `*Msg`，在 `app.go`/`practice.go` 的 `Update` 里处理结果，保持 bubbletea 的单向数据流。
- 涉及 LLM 提示词的改动，从产品语义上确认是否触及「防代答」护栏——前三级（Hint/Nudge/Review）禁止输出完整可提交代码。
- 配置目录解析见 `config.ConfigDir()`：`$LEETMATE_CONFIG_DIR` > `$XDG_CONFIG_HOME/leetmate` > `~/.config/leetmate`。
