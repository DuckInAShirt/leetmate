# LeetMate

<p align="center">
  <strong>LeetMate, your LeetCode practice companion.</strong><br/>
  Hints, not handouts — then review problems before you forget them.
</p>

<p align="center">
  <strong>English</strong> · <a href="README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <a href="https://github.com/DuckInAShirt/leetmate/releases"><img alt="Latest release" src="https://img.shields.io/github/v/release/DuckInAShirt/leetmate?style=for-the-badge&label=version" /></a>
  <a href="https://github.com/DuckInAShirt/leetmate/actions/workflows/auto-release.yml"><img alt="Build status" src="https://img.shields.io/github/actions/workflow/status/DuckInAShirt/leetmate/auto-release.yml?branch=main&style=for-the-badge&label=build" /></a>
  <a href="https://go.dev/"><img alt="Go 1.26+" src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white" /></a>
  <a href="LICENSE"><img alt="MIT License" src="https://img.shields.io/github/license/DuckInAShirt/leetmate?style=for-the-badge" /></a>
  <a href="https://github.com/DuckInAShirt/homebrew-tap"><img alt="Homebrew tap" src="https://img.shields.io/badge/Homebrew-DuckInAShirt%2Ftap-FBB040?style=for-the-badge&logo=homebrew&logoColor=white" /></a>
  <a href="https://cloud.siliconflow.cn/i/hoNe8cdD"><img alt="SiliconFlow invite code hoNe8cdD" src="https://img.shields.io/badge/SiliconFlow-hoNe8cdD-7C3AED?style=for-the-badge" /></a>
</p>

<p align="center">
  <img src="docs/demo.gif" width="820" alt="LeetMate demo" />
</p>

LeetMate is a terminal-based **LeetCode coaching TUI** built for interview prep. It does not solve problems for you: when you are stuck, it gives a **hint** instead of an answer. Practiced problems are queued for spaced review so you revisit them before they fade.

## Why

Most LeetCode tools either hand you problems or hand you full solutions. LeetMate sits in the middle: press `1` for a high-level Hint, `2` for a targeted Nudge, or `3` for a Review of your code. The first three tiers **never output a complete solution**. Only `4` (Answer), with a second confirmation, can reveal a full implementation. Practiced problems are scheduled into a lightweight FSRS-style review queue; full FSRS integration remains planned.

## Features

- 🧠 **Socratic coaching** — Hint / Nudge / Review / Answer tiers, with guardrails that keep the first three tiers from leaking full code
- 📋 **Study plans** — Built-in Hot 100 and Interview 150, progress tracking, auto-next flow, and custom YAML plans
- 🧩 **Powered by [leetgo](https://github.com/j178/leetgo)** — Uses leetgo for code scaffolding, local tests, and submissions
- 🔁 **Spaced review** — Lightweight FSRS-style scheduling queues practiced problems for review
- 🎛️ **Model presets** — Switch Gemini / SiliconFlow / Groq / DeepSeek with one `preset`; put secrets in `.env`
- 🗳️ **Local-first data** — Attempts, conversations, and progress live in local SQLite
- 🌐 **Chinese and English UI** — One config value switches the interface language
- 📜 **Streaming + expandable details** — Coach replies stream in; press `o` to expand the full reply or full error output

## Status

🧪 **Alpha**. Coaching, study-plan flows, and a lightweight FSRS-style review queue are usable; full FSRS tuning is still planned.

| Module | Status |
|--------|--------|
| leetgo integration (pick/test/submit) | ✅ |
| LLM coaching (four tiers + guardrails + streaming) | ✅ |
| Study plans + progress | ✅ |
| Model presets | ✅ |
| FSRS-style spaced review MVP | ✅ |
| `leetmate init` config generator | ✅ |

## Requirements

- Go 1.26+
- [leetgo](https://github.com/j178/leetgo): `brew install leetgo` or `go install github.com/j178/leetgo@latest`, then run `leetgo init` and configure LeetCode authentication
- One LLM API key: Gemini, SiliconFlow, Groq, or DeepSeek all have free or low-cost options

## Installation

**Homebrew** (macOS / Linux):

```bash
brew install DuckInAShirt/tap/leetmate
```

**go install**:

```bash
go install github.com/DuckInAShirt/leetmate/cmd/leetmate@latest
```

**Prebuilt binaries**: download the archive for your platform from [Releases](https://github.com/DuckInAShirt/leetmate/releases), extract it, and put `leetmate` in your `PATH`.

**Build from source**:

```bash
git clone https://github.com/DuckInAShirt/leetmate.git
cd leetmate
go build -o leetmate ./cmd/leetmate
```

## Configuration

Generate a starter config first:

```bash
leetmate init --preset siliconflow --workspace /path/to/your/leetgo/workspace
```

Then put the corresponding API key in `~/.config/leetmate/.env`:

```dotenv
SILICONFLOW_API_KEY=sk-...   # or GEMINI_API_KEY / GROQ_API_KEY / DEEPSEEK_API_KEY
```

Need a SiliconFlow key? You can register with invite code [`hoNe8cdD`](https://cloud.siliconflow.cn/i/hoNe8cdD) to try the recommended China-friendly preset.

Inspect or update the resolved config:

```bash
leetmate config
leetmate config --presets
leetmate config set language en
leetmate config set leetgo.workspace /path/to/your/leetgo/workspace
leetmate config set llm.preset siliconflow
leetmate config set code.lang go
```

Supported writable keys: `language`, `editor`, `leetgo.workspace`, `leetgo.binary`, `code.lang`, `llm.preset`, `llm.model`, `llm.max_history`, `db.path`.

Manual config lives at `~/.config/leetmate/config.yaml`:

```yaml
language: en          # or zh

leetgo:
  workspace: /path/to/your/leetgo/workspace

# Pick one preset; put the matching API key in .env.
llm:
  preset: siliconflow  # gemini | siliconflow | groq | deepseek
```

The coding language is configured by leetgo, not LeetMate. `leetmate config set code.lang <lang>` updates `leetgo.yaml` inside your leetgo workspace, or you can edit it manually:

```yaml
code:
  lang: go             # or python3 / cpp / java / ...
```

Available presets:

| preset | Provider | Default model | Notes |
|--------|----------|---------------|-------|
| `gemini` (default) | Google | gemini-2.0-flash | Global access, free tier |
| `siliconflow` | SiliconFlow | GLM-4-9B (can override to DeepSeek-V3, etc.) | Reliable from China, requires real-name verification |
| `groq` | Groq | llama-3.3-70b | Free and very fast, best with overseas network |
| `deepseek` | DeepSeek official | deepseek-v4-flash | Cheap, fast, strong instruction following |

## Usage

```bash
leetmate
```

Inside the TUI:

- **Today's problem / Study plans** — Pick a problem and start practicing
- `e` edit code · `t` run local tests · `s` submit
- `1` Hint · `2` Nudge · `3` Review · `4` Answer (requires confirmation)
- `Tab` switch between code and coach panes · `o` expand details · `↑/↓` scroll

## Custom Study Plans

`~/.config/leetmate/studyplans/my-plan.yaml`:

```yaml
id: my-plan
title: My weak spots
items: ["5", "53", "200"]   # LeetCode problem IDs
```

## Release

Pushes to `main` automatically create the next patch tag (`vX.Y.Z`) and run GoReleaser, publishing GitHub Release assets and updating the Homebrew tap. Manual `v*` tag pushes still trigger the regular release workflow.

## Tech Stack

Go · [bubbletea](https://github.com/charmbracelet/bubbletea) · [leetgo](https://github.com/j178/leetgo) · SQLite (modernc) · FSRS

## License

MIT
