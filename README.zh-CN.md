# LeetMate

<p align="center">
  <strong>LeetMate，你的力扣刷题伙伴。</strong><br/>
  不给答案，只给提示——然后在你忘记之前，把题排进复习。
</p>

<p align="center">
  <a href="README.md">English</a> · <strong>简体中文</strong>
</p>

<p align="center">
  <a href="https://github.com/DuckInAShirt/leetmate/releases"><img alt="最新版本" src="https://img.shields.io/github/v/release/DuckInAShirt/leetmate?style=for-the-badge&label=version" /></a>
  <a href="https://github.com/DuckInAShirt/leetmate/actions/workflows/auto-release.yml"><img alt="构建状态" src="https://img.shields.io/github/actions/workflow/status/DuckInAShirt/leetmate/auto-release.yml?branch=main&style=for-the-badge&label=build" /></a>
  <a href="https://go.dev/"><img alt="Go 1.26+" src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white" /></a>
  <a href="LICENSE"><img alt="MIT License" src="https://img.shields.io/github/license/DuckInAShirt/leetmate?style=for-the-badge" /></a>
  <a href="https://github.com/DuckInAShirt/homebrew-tap"><img alt="Homebrew tap" src="https://img.shields.io/badge/Homebrew-DuckInAShirt%2Ftap-FBB040?style=for-the-badge&logo=homebrew&logoColor=white" /></a>
</p>

<p align="center">
  <img src="docs/demo.gif" width="820" alt="LeetMate demo" />
</p>

LeetMate 是一个跑在终端里的 **LeetCode 刷题辅导工具**，围绕「面试备战」设计。它不替你解题——你卡住时给一个**提示**而不是答案；练过的题会被排进**间隔复习队列**，赶在你忘掉之前回来。

## 为什么造它

市面上的 LeetCode 工具要么只给题、要么直接给答案，都不利于真正学会。LeetMate 介于两者之间：卡住时按 `1`（Hint）只拿到算法方向、按 `2`（Nudge）拿到卡点提示、按 `3`（Review）让它挑你代码的 bug——**前三级绝不输出完整代码**；只有按 `4`（Answer）并二次确认才给完整解法。每一道练过的题自动进 FSRS 复习队列（开发中）。

## 功能

- 🧠 **苏格拉底式辅导** — Hint / Nudge / Review / Answer 四级，防代答 system prompt 守住前三级不泄答案
- 📋 **题单** — 内置「热题 100」「面试经典 150」，进度跟踪 + 自动跳到下一题；支持自定义题单
- 🧩 **基于 [leetgo](https://github.com/j178/leetgo)** — 代码骨架、本地测试、提交全交给 leetgo
- 🔁 **间隔复习** — FSRS 调度（开发中，M3）
- 🎛️ **自带模型** — 一行 `preset` 切换 Gemini / 硅基流动 / Groq / DeepSeek，只填 key 即可
- 🗳️ **本地优先** — 所有练习记录、对话、进度都在本地 SQLite，数据不离机
- 🌐 **中英文界面** — config 一行切换
- 📜 **流式输出 + 可展开详情** — 辅导打字机式流式；默认折叠预览，`o` 展开全文 / 完整错误

## 状态

🧪 **Alpha**。辅导 + 题单闭环已可用，FSRS 复习队列开发中。

| 模块 | 状态 |
|------|------|
| leetgo 集成（pick/test/submit）| ✅ |
| LLM 辅导（四级 + 防代答 + 流式） | ✅ |
| 题单 + 进度 | ✅ |
| preset 多模型 | ✅ |
| FSRS 间隔复习 | 🚧 进行中 |
| `leetmate init` 配置生成 | ✅ |

## 前置依赖

- Go 1.26+
- [leetgo](https://github.com/j178/leetgo)：`brew install leetgo` 或 `go install github.com/j178/leetgo@latest`，并 `leetgo init` 配好 LeetCode 认证
- 一个 LLM API key（Gemini / 硅基流动 / Groq / DeepSeek 任一，均有免费额度）

## 安装

**Homebrew**（macOS / Linux）：

```bash
brew install DuckInAShirt/tap/leetmate
```

**go install**（需已安装 Go）：

```bash
go install github.com/DuckInAShirt/leetmate/cmd/leetmate@latest
```

**预编译二进制**：从 [Releases](https://github.com/DuckInAShirt/leetmate/releases) 下载对应平台的压缩包，解压后把 `leetmate` 放进 `PATH`。

**从源码构建**：

```bash
git clone https://github.com/DuckInAShirt/leetmate.git
cd leetmate
go build -o leetmate ./cmd/leetmate
```

## 配置

推荐先生成配置模板：

```bash
leetmate init --preset siliconflow --workspace /path/to/your/leetgo/workspace
```

然后把对应平台的 key 填进 `~/.config/leetmate/.env`：

```dotenv
SILICONFLOW_API_KEY=sk-...   # 或 GEMINI_API_KEY / GROQ_API_KEY / DEEPSEEK_API_KEY
```

查看或更新当前配置：

```bash
leetmate config
leetmate config --presets
leetmate config set language zh
leetmate config set leetgo.workspace /path/to/your/leetgo/workspace
leetmate config set llm.preset siliconflow
leetmate config set code.lang go
```

支持写入的 key：`language`、`editor`、`leetgo.workspace`、`leetgo.binary`、`code.lang`、`llm.preset`、`llm.model`、`llm.max_history`、`db.path`。

手动配置也可以，`~/.config/leetmate/config.yaml`：

```yaml
language: zh          # 或 en

leetgo:
  workspace: /path/to/your/leetgo/workspace

# 选一个 preset，对应平台的 key 放进 .env 即可
llm:
  preset: siliconflow  # gemini | siliconflow | groq | deepseek
```

刷题代码语言由 leetgo 管，不在 LeetMate 配置里保存。`leetmate config set code.lang <lang>` 会更新 leetgo workspace 里的 `leetgo.yaml`，也可以手动编辑：

```yaml
code:
  lang: go             # 或 python3 / cpp / java / ...
```

可选 preset：

| preset | 平台 | 默认模型 | 备注 |
|--------|------|---------|------|
| `gemini`（默认） | Google | gemini-2.0-flash | 全球，免费 tier |
| `siliconflow` | 硅基流动 | GLM-4-9B（可改 DeepSeek-V3 等） | 国内访问稳，需实名 |
| `groq` | Groq | llama-3.3-70b | 免费、极快，海外网络 |
| `deepseek` | DeepSeek 官方 | deepseek-v4-flash | 极便宜、快、指令遵循强 |

## 用法

```bash
leetmate
```

进 TUI 后：

- **今日题目 / 题单** — 选题开始
- `e` 编辑代码 · `t` 本地测试 · `s` 提交
- `1` Hint · `2` Nudge · `3` Review · `4` Answer（二次确认）
- `Tab` 切代码/辅导区 · `o` 展开详情 · `↑/↓` 滚动

## 自定义题单

`~/.config/leetmate/studyplans/my-plan.yaml`：

```yaml
id: my-plan
title: 我的薄弱题
items: ["5", "53", "200"]   # leetcode 题号
```

## 发布

代码推送到 `main` 后会自动创建下一个 patch tag（`vX.Y.Z`）并运行 GoReleaser，发布 GitHub Release 产物并更新 Homebrew tap。手动推送 `v*` tag 仍会触发常规 release workflow。

## 技术栈

Go · [bubbletea](https://github.com/charmbracelet/bubbletea) · [leetgo](https://github.com/j178/leetgo) · SQLite (modernc) · FSRS

## License

MIT
