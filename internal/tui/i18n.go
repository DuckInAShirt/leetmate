package tui

import "fmt"

// Lightweight i18n. All user-facing strings live in `messages` and are looked
// up by key + language. Add a language by adding a column; add a string by
// adding a key. Missing translations fall back to English.

// Lang is a UI language code.
type Lang string

const (
	LangZH Lang = "zh"
	LangEN Lang = "en"
)

type dict struct{ lang Lang }

func loadDict(language string) dict {
	switch Lang(language) {
	case LangEN:
		return dict{lang: LangEN}
	default:
		return dict{lang: LangZH}
	}
}

// Text returns one translated user-facing string for CLI and TUI callers.
func Text(language, key string) string { return loadDict(language).t(key) }

// Textf formats one translated string with the provided arguments.
func Textf(language, key string, args ...any) string {
	return fmt.Sprintf(Text(language, key), args...)
}

func (d dict) t(key string) string {
	if m, ok := messages[key]; ok {
		if s, ok := m[d.lang]; ok && s != "" {
			return s
		}
		if s, ok := m[LangEN]; ok && s != "" {
			return s
		}
	}
	return key
}

var messages = map[string]map[Lang]string{
	"brand.subtitle": {
		LangZH: "· 辅导，不给答案",
		LangEN: "— coach, don't cheat",
	},
	"menu.busy": {
		LangZH: "正在生成题目骨架…",
		LangEN: "Generating problem skeleton…",
	},
	"menu.today":       {LangZH: "今日题目", LangEN: "Today's problem"},
	"menu.today.desc":  {LangZH: "LeetCode 每日一题", LangEN: "the LeetCode daily"},
	"menu.review":      {LangZH: "待复习", LangEN: "Due for review"},
	"menu.review.desc": {LangZH: "打开最早到期题", LangEN: "Open earliest due problem"},
	"menu.quit":        {LangZH: "退出", LangEN: "Quit"},
	"menu.hint": {
		LangZH: "↑/↓ 选择 · 回车打开 · q 退出",
		LangEN: "↑/↓ select · enter open · q quit",
	},
	"menu.reviewEmpty": {
		LangZH: "暂无到期复习。",
		LangEN: "No reviews due.",
	},
	"home.tips.title": {
		LangZH: "上手提示：",
		LangEN: "Tips:",
	},
	"home.tip.1": {
		LangZH: "先自己想，再要提示。",
		LangEN: "Think first, then ask for hints.",
	},
	"home.tip.2": {
		LangZH: "前三档只点拨，不给完整答案。",
		LangEN: "The first three tiers never reveal full code.",
	},
	"home.tip.3": {
		LangZH: "Answer 需确认，适合复盘。",
		LangEN: "Answer needs confirmation; use it for review.",
	},
	"home.prompt": {
		LangZH: "选择训练入口",
		LangEN: "Choose a training entry",
	},
	"home.status.ready": {
		LangZH: "ready · leetgo + LLM preset",
		LangEN: "ready · leetgo + LLM preset",
	},
	"home.status.leetgoOnly": {
		LangZH: "ready · 仅 leetgo 模式 · Coach 未启用",
		LangEN: "ready · leetgo-only mode · Coach disabled",
	},
	"home.status.busy": {
		LangZH: "working · 准备题目中",
		LangEN: "working · preparing problem",
	},

	"practice.testing":     {LangZH: "正在运行测试…", LangEN: "Running tests…"},
	"practice.submitting":  {LangZH: "正在提交…", LangEN: "Submitting…"},
	"practice.testPassed":  {LangZH: "✓ 全部测试用例通过", LangEN: "✓ All test cases passed"},
	"practice.testFailed":  {LangZH: "✗ 存在未通过的用例", LangEN: "✗ Some test cases failed"},
	"practice.testError":   {LangZH: "⚠ 测试失败：", LangEN: "⚠ test failed: "},
	"practice.submitError": {LangZH: "⚠ 提交失败：", LangEN: "⚠ submit failed: "},
	"practice.accepted":    {LangZH: "✓ 通过（Accepted）", LangEN: "✓ Accepted"},
	"practice.notAccepted": {LangZH: "未通过", LangEN: "Not accepted"},
	"practice.hint": {
		LangZH: "e 外部编辑 · i 内置编辑 · t 测试 · s 提交 · 1-4 辅导 · o 展开 · b 返回 · q 退出",
		LangEN: "e external editor · i edit here · t test · s submit · 1-4 coach · o expand · b back · q quit",
	},
	"practice.editorHint": {
		LangZH: "编辑中：esc 保存退出 · ctrl+s 保存 · 括号/引号自动补全",
		LangEN: "editing: esc save & exit · ctrl+s save · brackets/quotes auto-pair",
	},
	"practice.saved": {
		LangZH: "✓ 已保存",
		LangEN: "✓ Saved",
	},
	"practice.saveError": {
		LangZH: "⚠ 保存失败：",
		LangEN: "⚠ save failed: ",
	},
	"practice.reviewSaveError": {
		LangZH: "复习记录保存失败",
		LangEN: "review record save failed",
	},
	"expand.hint": {
		LangZH: "Tab/←/→ 切换 · o/esc 收起 · ↑/↓ 或 j/k 滚动",
		LangEN: "Tab/←/→ switch · o/esc collapse · ↑/↓ or j/k scroll",
	},
	"expand.statement": {LangZH: "完整题面", LangEN: "Full statement"},
	"expand.error":     {LangZH: "完整错误", LangEN: "Full error"},
	"expand.coach":     {LangZH: "辅导全文", LangEN: "Full coach reply"},
	"section.code":     {LangZH: "代码", LangEN: "Code"},
	"section.coach":    {LangZH: "辅导", LangEN: "Coach"},
	"menu.plans":       {LangZH: "题单", LangEN: "Study plans"},
	"menu.plans.desc":  {LangZH: "Hot 100 · 面试 150", LangEN: "Hot 100 · Interview 150"},
	"plan.hint.list": {
		LangZH: "↑/↓ 选择 · 回车打开 · b 返回 · q 退出",
		LangEN: "↑/↓ select · enter open · b back · q quit",
	},
	"plan.hint.items": {
		LangZH: "↑/↓ 选择 · 回车开始做题 · b 返回题单 · q 退出",
		LangEN: "↑/↓ select · enter start · b back · q quit",
	},
	"plan.complete": {
		LangZH: "🎉 全部完成！",
		LangEN: "🎉 All done!",
	},
	"coach.section": {
		LangZH: "── 辅导 ──",
		LangEN: "── Coach ──",
	},
	"coach.thinking": {
		LangZH: "思考中…",
		LangEN: "Thinking…",
	},
	"coach.waiting": {
		LangZH: "已连接，等待模型输出…",
		LangEN: "Connected, waiting for model output…",
	},
	"coach.unavailable": {
		LangZH: "LLM 未配置：请运行 leetmate config 查看当前 preset，并在配置目录的 .env 中设置对应 API key",
		LangEN: "LLM is not configured: run leetmate config to check the preset, then set its API key in the config directory's .env file",
	},
	"coach.reasoning": {
		LangZH: "模型正在推理，已隐藏推理内容…",
		LangEN: "Model is reasoning; hidden reasoning content…",
	},
	"coach.confirm": {
		LangZH: "确定要看完整答案吗？将标记为「放弃独立完成」。  [y] 确认 / [n] 取消",
		LangEN: "See the full answer? This marks the problem as 'gave up'.  [y] yes / [n] no",
	},
	"coach.empty": {
		LangZH: "按 1 提示 · 2 点拨 · 3 审查 · 4 答案",
		LangEN: "Press 1 hint · 2 nudge · 3 review · 4 answer",
	},
	"coach.gaveup": {
		LangZH: "（已查看答案，本次计为放弃独立完成）",
		LangEN: "(answer revealed — counted as gave up)",
	},

	"doctor.title":                      {LangZH: "LeetMate 环境检查", LangEN: "LeetMate environment check"},
	"doctor.next.fix":                   {LangZH: "\n下一步：修复第一个失败项，然后重新运行 `leetmate doctor`。", LangEN: "\nNext: fix the first failed check, then run `leetmate doctor` again."},
	"doctor.ready":                      {LangZH: "\n必需环境已就绪。认证状态仅做本地检查，首次 test/submit 时仍会由 leetgo 验证。", LangEN: "\nRequired setup is ready. Authentication was checked locally; leetgo verifies it on the first test or submit."},
	"doctor.label.config":               {LangZH: "配置", LangEN: "config"},
	"doctor.label.leetgo":               {LangZH: "leetgo", LangEN: "leetgo"},
	"doctor.label.workspace":            {LangZH: "工作区", LangEN: "workspace"},
	"doctor.label.auth":                 {LangZH: "认证", LangEN: "auth"},
	"doctor.label.llm":                  {LangZH: "LLM", LangEN: "LLM"},
	"doctor.label.config_dir":           {LangZH: "配置目录", LangEN: "config dir"},
	"doctor.label.data":                 {LangZH: "数据目录", LangEN: "data"},
	"doctor.config.found":               {LangZH: "已读取 %s", LangEN: "loaded %s"},
	"doctor.config.missing":             {LangZH: "缺少配置；运行 `leetmate init`", LangEN: "missing; run `leetmate init`"},
	"doctor.config.unreadable":          {LangZH: "配置不可读：%s", LangEN: "unreadable: %s"},
	"doctor.leetgo.found":               {LangZH: "已找到 %s", LangEN: "found %s"},
	"doctor.leetgo.missing":             {LangZH: "未找到；运行 `brew install j178/tap/leetgo` 或 `go install github.com/j178/leetgo@latest`", LangEN: "not found; run `brew install j178/tap/leetgo` or `go install github.com/j178/leetgo@latest`"},
	"doctor.workspace.ready":            {LangZH: "已找到 %s", LangEN: "found %s"},
	"doctor.workspace.missing":          {LangZH: "未找到 leetgo.yaml；先运行 `leetgo init`，再执行 `leetmate config set leetgo.workspace /path/to/workspace`", LangEN: "no leetgo.yaml found; run `leetgo init`, then `leetmate config set leetgo.workspace /path/to/workspace`"},
	"doctor.workspace.not_directory":    {LangZH: "配置的路径不是目录：%s", LangEN: "configured path is not a directory: %s"},
	"doctor.workspace.no_config":        {LangZH: "目录中缺少 leetgo.yaml：%s", LangEN: "leetgo.yaml is missing from %s"},
	"doctor.workspace.invalid_config":   {LangZH: "leetgo.yaml 无效：%s", LangEN: "invalid leetgo.yaml: %s"},
	"doctor.workspace.missing_language": {LangZH: "leetgo.yaml 缺少 code.lang；运行 `leetmate config set code.lang go` 或使用其他语言", LangEN: "leetgo.yaml is missing code.lang; run `leetmate config set code.lang go` or choose another language"},
	"doctor.auth.cookies_ready":         {LangZH: "cookies 已在环境或 workspace .env 中配置（未联网验证）", LangEN: "cookies are configured in the environment or workspace .env (not verified online)"},
	"doctor.auth.cookies_missing":       {LangZH: "cookies 未完整配置；在工作区 .env 设置 LEETCODE_SESSION 和 LEETCODE_CSRFTOKEN", LangEN: "cookies are incomplete; set LEETCODE_SESSION and LEETCODE_CSRFTOKEN in the workspace .env"},
	"doctor.auth.runtime_unverified":    {LangZH: "使用 %s，只能在运行时由 leetgo 验证", LangEN: "using %s; only leetgo can verify it at runtime"},
	"doctor.auth.missing":               {LangZH: "未声明认证方式；test/submit 可能不可用", LangEN: "no authentication method declared; test/submit may be unavailable"},
	"doctor.llm.found":                  {LangZH: "已检测到 %s", LangEN: "detected %s"},
	"doctor.llm.missing":                {LangZH: "未设置 %s；可继续刷题，但 Coach 不可用", LangEN: "%s is unset; practice still works, but Coach is unavailable"},
	"doctor.llm.invalid_provider":       {LangZH: "未知 provider：%s；运行 `leetmate config --presets` 查看可用配置", LangEN: "unknown provider: %s; run `leetmate config --presets` to list supported presets"},
	"doctor.path.writable":              {LangZH: "可写 %s", LangEN: "writable %s"},
	"doctor.path.unwritable":            {LangZH: "不可写 %s：%s", LangEN: "not writable %s: %s"},
	"onboarding.not_configured":         {LangZH: "LeetMate 尚未配置。", LangEN: "LeetMate is not configured."},
	"onboarding.run_init":               {LangZH: "运行：leetmate init --workspace %q", LangEN: "Run: leetmate init --workspace %q"},
	"onboarding.run_leetgo_init":        {LangZH: "进入工作区运行 `leetgo init`，然后运行 `leetmate init`。", LangEN: "Run `leetgo init` in a workspace, then run `leetmate init`."},
	"onboarding.title":                  {LangZH: "LeetMate 首次设置", LangEN: "LeetMate first-time setup"},
	"onboarding.language":               {LangZH: "界面语言 [zh/en] (zh)：", LangEN: "UI language [zh/en] (en): "},
	"onboarding.found_workspace":        {LangZH: "已发现 leetgo workspace：%s", LangEN: "Found leetgo workspace: %s"},
	"onboarding.invalid_choice":         {LangZH: "选项无效", LangEN: "Invalid choice"},
	"onboarding.workspace":              {LangZH: "leetgo workspace（%s）：", LangEN: "leetgo workspace (%s): "},
	"onboarding.invalid_path":           {LangZH: "路径无效", LangEN: "Invalid path"},
	"onboarding.no_workspace":           {LangZH: "未找到 leetgo.yaml，请先在该目录运行 `leetgo init`。", LangEN: "No leetgo.yaml found. Run `leetgo init` there first."},
	"onboarding.environment_check":      {LangZH: "\n环境检查", LangEN: "\nEnvironment check"},
	"onboarding.continue":               {LangZH: "\n按回车进入 LeetMate；LLM key 可稍后配置。", LangEN: "\nPress Enter to continue; the LLM key is optional for now."},
	"init.wrote":                        {LangZH: "✓ 已写入 %s", LangEN: "✓ wrote %s"},
	"init.next":                         {LangZH: "下一步：", LangEN: "next:"},
	"init.set_workspace":                {LangZH: "  1. 设置 workspace：leetmate config set leetgo.workspace /path/to/workspace", LangEN: "  1. set the workspace: leetmate config set leetgo.workspace /path/to/workspace"},
	"init.run_doctor.first":             {LangZH: "  1. 运行 leetmate doctor", LangEN: "  1. run leetmate doctor"},
	"init.run_doctor.second":            {LangZH: "  2. 运行 leetmate doctor", LangEN: "  2. run leetmate doctor"},
	"init.run_app":                      {LangZH: "  2. 运行 leetmate", LangEN: "  2. run leetmate"},
	"init.optional_key":                 {LangZH: "可选：编辑 %s 以启用 Coach", LangEN: "optional: edit %s to enable Coach"},

	"difficulty.easy":   {LangZH: "简单", LangEN: "Easy"},
	"difficulty.medium": {LangZH: "中等", LangEN: "Medium"},
	"difficulty.hard":   {LangZH: "困难", LangEN: "Hard"},
}

func (d dict) difficultyLabel(diff string) string {
	switch diff {
	case "Easy":
		return d.t("difficulty.easy")
	case "Medium":
		return d.t("difficulty.medium")
	case "Hard":
		return d.t("difficulty.hard")
	default:
		return diff
	}
}
