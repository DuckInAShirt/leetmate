# LeetMate

<p align="center">
  <strong>LeetMate, your LeetCode practice companion.</strong><br/>
  Hints, not handouts — then review problems before you forget them.
</p>

<p align="center">
  <strong>English</strong> · <a href="README.md">简体中文</a>
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
  <a href="docs/demo.mp4">
    <img src="docs/demo.gif" width="820" alt="LeetMate demo" />
  </a>
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
- ⌨️ **ACM mode (experimental)** — Practice writing input/output yourself: a blank slate for `import` + IO + algorithm; `o` expands, `tab` cycles statement/code/input/output, `r` runs and jumps to output

## Status

🧪 **Alpha**. Coaching, study-plan flows, and a lightweight FSRS-style review queue are usable; full FSRS tuning is still planned.

| Module | Status |
|--------|--------|
| leetgo integration (pick/test/submit) | ✅ |
| LLM coaching (four tiers + guardrails + streaming) | ✅ |
| Study plans + progress | ✅ |
| Model presets | ✅ |
| FSRS-style spaced review MVP | ✅ |
| First-run guide + `leetmate doctor` | ✅ |
| `leetmate init` config generator | ✅ |
| ACM mode (write your own IO) | 🧪 |

## Requirements

- [leetgo](https://github.com/j178/leetgo): `brew install j178/tap/leetgo` or `go install github.com/j178/leetgo@latest`, then run `leetgo init`
- LeetCode cookies for test/submit (configured in the leetgo workspace as shown below)
- Optional LLM API key for Coach: Gemini, SiliconFlow, Groq, or DeepSeek
- Go 1.26+ only when building from source or installing leetgo with `go install`

## Installation

**Homebrew** (macOS / Linux):

```bash
brew install DuckInAShirt/tap/leetmate
```

**npm** (installs the matching GitHub Release binary):

```bash
npm install -g @zxr55555/leetmate
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

## First Run

Run `leetmate` after installation. On an interactive terminal, the first-run guide discovers a `leetgo.yaml` in the current directory or its parents, writes the LeetMate config, runs a local environment check, and starts the TUI. Existing users start normally without an extra prompt.

For a non-interactive setup, or to configure explicitly:

```bash
leetmate init --preset gemini --workspace /path/to/your/leetgo/workspace --lang en
leetmate doctor
leetmate
```

`leetmate doctor` checks the config, leetgo binary, workspace, local authentication setup, LLM key, and writable data paths. It does not send credentials or make a LeetCode request. Use `leetmate doctor --json` for machine-readable output.

### LeetCode authentication

In the workspace's `leetgo.yaml`:

```yaml
leetcode:
  site: https://leetcode.com # or https://leetcode.cn
  credentials:
    from: cookies
```

In `.env` next to that `leetgo.yaml`:

```dotenv
LEETCODE_SESSION=<LEETCODE_SESSION cookie>
LEETCODE_CSRFTOKEN=<csrftoken cookie>
# LEETCODE_CFCLEARANCE=<cf_clearance cookie> # sometimes required by leetcode.com
```

`doctor` only confirms that the required cookie values are present. leetgo validates whether the session is still accepted when you first test or submit.

### Optional Coach

LeetMate works in leetgo-only mode without an LLM key: pick, edit, test, submit, and spaced review remain available, while AI coaching is disabled. To enable Coach, put the selected preset's key in the LeetMate config directory's `.env` (normally `~/.config/leetmate/.env`):

```dotenv
GEMINI_API_KEY=... # or SILICONFLOW_API_KEY / GROQ_API_KEY / DEEPSEEK_API_KEY
```

Need a SiliconFlow key? You can register with invite code [`hoNe8cdD`](https://cloud.siliconflow.cn/i/hoNe8cdD) to try the recommended China-friendly preset.

## Configuration

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

Pushes to `main` run GoReleaser and publish GitHub Release assets, npm, and the Homebrew tap. `VERSION` can request the next minor or major release (this branch sets `0.3.0`); once the latest tag reaches that value, later pushes resume automatic patch bumps. Reruns reuse a release tag already pointing at the same commit. Manual `v*` tag pushes still trigger the regular release workflow.

## Tech Stack

Go · [bubbletea](https://github.com/charmbracelet/bubbletea) · [leetgo](https://github.com/j178/leetgo) · SQLite (modernc) · FSRS

## License

MIT
