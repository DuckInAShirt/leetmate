# LeetMate

<p align="center">
  <strong>LeetMate，你的力扣刷题伙伴。</strong><br/>
  不给答案，只给提示——然后在你忘记之前，把题排进复习。
</p>

<p align="center">
  <a href="README.en.md">English</a> · <strong>简体中文</strong>
</p>

<p align="center">
  <a href="https://github.com/DuckInAShirt/leetmate/releases"><img alt="最新版本" src="https://img.shields.io/github/v/release/DuckInAShirt/leetmate?style=for-the-badge&label=version" /></a>
  <a href="https://github.com/DuckInAShirt/leetmate/actions/workflows/auto-release.yml"><img alt="构建状态" src="https://img.shields.io/github/actions/workflow/status/DuckInAShirt/leetmate/auto-release.yml?branch=main&style=for-the-badge&label=build" /></a>
  <a href="https://go.dev/"><img alt="Go 1.26+" src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white" /></a>
  <a href="LICENSE"><img alt="MIT License" src="https://img.shields.io/github/license/DuckInAShirt/leetmate?style=for-the-badge" /></a>
  <a href="https://github.com/DuckInAShirt/homebrew-tap"><img alt="Homebrew tap" src="https://img.shields.io/badge/Homebrew-DuckInAShirt%2Ftap-FBB040?style=for-the-badge&logo=homebrew&logoColor=white" /></a>
  <a href="https://cloud.siliconflow.cn/i/hoNe8cdD"><img alt="硅基流动邀请码 hoNe8cdD" src="https://img.shields.io/badge/SiliconFlow-hoNe8cdD-7C3AED?style=for-the-badge" /></a>
</p>

<p align="center">
  <a href="docs/demo.mp4">
    <img src="docs/demo.gif" width="820" alt="LeetMate demo" />
  </a>
</p>

LeetMate 是一个跑在终端里的 **LeetCode 刷题辅导工具**，围绕「面试备战」设计。它不替你解题——你卡住时给一个**提示**而不是答案；练过的题会被排进**间隔复习队列**，赶在你忘掉之前回来。

## 为什么造它

市面上的 LeetCode 工具要么只给题、要么直接给答案，都不利于真正学会。LeetMate 介于两者之间：卡住时按 `1`（Hint）只拿到算法方向、按 `2`（Nudge）拿到卡点提示、按 `3`（Review）让它挑你代码的 bug——**前三级绝不输出完整代码**；只有按 `4`（Answer）并二次确认才给完整解法。每一道练过的题自动进轻量 FSRS-style 复习队列；完整 FSRS 算法后续接入。

## 功能

- 🧠 **苏格拉底式辅导** — Hint / Nudge / Review / Answer 四级，防代答 system prompt 守住前三级不泄答案
- 📋 **题单** — 内置「热题 100」「面试经典 150」，进度跟踪 + 自动跳到下一题；支持自定义题单
- 🧩 **基于 [leetgo](https://github.com/j178/leetgo)** — 代码骨架、本地测试、提交全交给 leetgo
- 🔁 **间隔复习** — 轻量 FSRS-style 调度，练过的题会进入复习队列
- 🎛️ **自带模型** — 一行 `preset` 切换 Gemini / 硅基流动 / Groq / DeepSeek，只填 key 即可
- 🗳️ **本地优先** — 所有练习记录、对话、进度都在本地 SQLite，数据不离机
- 🌐 **中英文界面** — config 一行切换
- 📜 **流式输出 + 可展开详情** — 辅导打字机式流式；默认折叠预览，`o` 展开全文 / 完整错误

## 状态

🧪 **Alpha**。辅导、题单闭环和轻量 FSRS-style 复习队列已可用；完整 FSRS 调参后续接入。

| 模块 | 状态 |
|------|------|
| leetgo 集成（pick/test/submit）| ✅ |
| LLM 辅导（四级 + 防代答 + 流式） | ✅ |
| 题单 + 进度 | ✅ |
| preset 多模型 | ✅ |
| FSRS-style 间隔复习 MVP | ✅ |
| 首跑引导 + `leetmate doctor` | ✅ |
| `leetmate init` 配置生成 | ✅ |

## 前置依赖

- [leetgo](https://github.com/j178/leetgo)：`brew install j178/tap/leetgo` 或 `go install github.com/j178/leetgo@latest`，然后运行 `leetgo init`
- 用于 test/submit 的 LeetCode cookies（按下文配置到 leetgo workspace）
- 可选的 LLM API key，用于 Coach：Gemini / 硅基流动 / Groq / DeepSeek
- 只有从源码构建或通过 `go install` 安装 leetgo 时才需要 Go 1.26+

## 安装

**Homebrew**（macOS / Linux）：

```bash
brew install DuckInAShirt/tap/leetmate
```

**npm**（会安装对应平台的 GitHub Release 二进制）：

```bash
npm install -g @zxr55555/leetmate
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

## 首次运行

安装后直接运行 `leetmate`。在交互终端中，首跑引导会从当前目录向上发现 `leetgo.yaml`，生成 LeetMate 配置，执行本地环境检查，然后启动 TUI。已有配置的用户不会多一步提示。

非交互环境或需要显式配置时：

```bash
leetmate init --preset siliconflow --workspace /path/to/your/leetgo/workspace --lang zh
leetmate doctor
leetmate
```

`leetmate doctor` 会检查配置、leetgo binary、workspace、本地认证配置、LLM key 和数据目录写权限。它不会发送凭据，也不会请求 LeetCode。脚本可使用 `leetmate doctor --json`。

### LeetCode 认证

在 workspace 的 `leetgo.yaml` 中配置：

```yaml
leetcode:
  site: https://leetcode.cn # 或 https://leetcode.com
  credentials:
    from: cookies
```

在同目录的 `.env` 中配置：

```dotenv
LEETCODE_SESSION=<LEETCODE_SESSION cookie>
LEETCODE_CSRFTOKEN=<csrftoken cookie>
# LEETCODE_CFCLEARANCE=<cf_clearance cookie> # leetcode.com 某些环境还需要
```

`doctor` 只确认必需 cookie 已填写。会话是否仍有效，会在首次 test/submit 时由 leetgo 验证。

### 可选 Coach

没有 LLM key 时 LeetMate 会进入“仅 leetgo 模式”：选题、编辑、测试、提交和间隔复习仍可使用，AI 辅导不可用。要启用 Coach，把 preset 对应的 key 填进 LeetMate 配置目录的 `.env`（通常是 `~/.config/leetmate/.env`）：

```dotenv
SILICONFLOW_API_KEY=... # 或 GEMINI_API_KEY / GROQ_API_KEY / DEEPSEEK_API_KEY
```

还没有硅基流动 key 的话，可以用邀请码 [`hoNe8cdD`](https://cloud.siliconflow.cn/i/hoNe8cdD) 注册，体验推荐的国内友好 preset。

## 配置

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

代码推送到 `main` 后会运行 GoReleaser，发布 GitHub Release、npm 和 Homebrew tap。`VERSION` 可指定下一个 minor 或 major 版本（当前分支设置为 `0.3.0`）；最新 tag 达到该版本后，后续 push 恢复自动递增 patch。同一 commit 的 workflow 重跑会复用已有 release tag。手动推送 `v*` tag 仍会触发常规 release workflow。

## 技术栈

Go · [bubbletea](https://github.com/charmbracelet/bubbletea) · [leetgo](https://github.com/j178/leetgo) · SQLite (modernc) · FSRS

## License

MIT
