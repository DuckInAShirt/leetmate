# LeetMate

<p align="center">
  <strong>LeetMate, your LeetCode practice companion.</strong><br/>
  Hints, not handouts — then review problems before you forget them.
</p>

<p align="center">
  <strong>English</strong> · <a href="README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <img src="docs/demo.gif" width="820" alt="LeetMate demo" />
</p>

LeetMate is a terminal-based **LeetCode coaching TUI** built for interview prep. It does not solve problems for you: when you are stuck, it gives a **hint** instead of an answer. Practiced problems are queued for spaced review so you revisit them before they fade.

## Why

Most LeetCode tools either hand you problems or hand you full solutions. LeetMate sits in the middle: press `1` for a high-level Hint, `2` for a targeted Nudge, or `3` for a Review of your code. The first three tiers **never output a complete solution**. Only `4` (Answer), with a second confirmation, can reveal a full implementation. Practiced problems are scheduled into an FSRS-style review queue (in progress).

## Features

- 🧠 **Socratic coaching** — Hint / Nudge / Review / Answer tiers, with guardrails that keep the first three tiers from leaking full code
- 📋 **Study plans** — Built-in Hot 100 and Interview 150, progress tracking, auto-next flow, and custom YAML plans
- 🧩 **Powered by [leetgo](https://github.com/j178/leetgo)** — Uses leetgo for code scaffolding, local tests, and submissions
- 🔁 **Spaced review** — FSRS scheduling is in progress
- 🎛️ **Model presets** — Switch Gemini / SiliconFlow / Groq / DeepSeek with one `preset`; put secrets in `.env`
- 🗳️ **Local-first data** — Attempts, conversations, and progress live in local SQLite
- 🌐 **Chinese and English UI** — One config value switches the interface language
- 📜 **Streaming + expandable details** — Coach replies stream in; press `o` to expand the full reply or full error output

## Status

🧪 **Alpha**. Coaching and study-plan flows are usable; the FSRS review queue is still in progress.

| Module | Status |
|--------|--------|
| leetgo integration (pick/test/submit) | ✅ |
| LLM coaching (four tiers + guardrails + streaming) | ✅ |
| Study plans + progress | ✅ |
| Model presets | ✅ |
| FSRS spaced review | 🚧 In progress |
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

Inspect the resolved config:

```bash
leetmate config
leetmate config --presets
```

Manual config lives at `~/.config/leetmate/config.yaml`:

```yaml
language: en          # or zh

leetgo:
  workspace: /path/to/your/leetgo/workspace

# Pick one preset; put the matching API key in .env.
llm:
  preset: siliconflow  # gemini | siliconflow | groq | deepseek
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

## Tech Stack

Go · [bubbletea](https://github.com/charmbracelet/bubbletea) · [leetgo](https://github.com/j178/leetgo) · SQLite (modernc) · FSRS

## License

MIT
