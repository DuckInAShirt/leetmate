package tui

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

// t returns the translated string for key, falling back to English then the key.
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
	"menu.review.desc": {LangZH: "FSRS 复习队列（M3）", LangEN: "FSRS review queue (M3)"},
	"menu.quit":        {LangZH: "退出", LangEN: "Quit"},
	"menu.hint": {
		LangZH: "↑/↓ 选择 · 回车打开 · q 退出",
		LangEN: "↑/↓ select · enter open · q quit",
	},
	"menu.reviewNotice": {
		LangZH: "复习队列将在 M3 推出。",
		LangEN: "Review queue arrives in M3.",
	},

	"practice.testing":    {LangZH: "正在运行测试…", LangEN: "Running tests…"},
	"practice.submitting": {LangZH: "正在提交…", LangEN: "Submitting…"},
	"practice.testPassed": {LangZH: "✓ 全部测试用例通过", LangEN: "✓ All test cases passed"},
	"practice.testFailed": {LangZH: "✗ 存在未通过的用例", LangEN: "✗ Some test cases failed"},
	"practice.testError":  {LangZH: "⚠ 测试失败：", LangEN: "⚠ test failed: "},
	"practice.submitError": {LangZH: "⚠ 提交失败：", LangEN: "⚠ submit failed: "},
	"practice.accepted":    {LangZH: "✓ 通过（Accepted）", LangEN: "✓ Accepted"},
	"practice.notAccepted": {LangZH: "未通过", LangEN: "Not accepted"},
	"practice.hint": {
		LangZH: "e 编辑 · t 测试 · s 提交 · 1-4 辅导 · Tab 切区 · o 展开详情 · b 返回 · q 退出",
		LangEN: "e edit · t test · s submit · 1-4 coach · Tab pane · o expand · b back · q quit",
	},
	"expand.hint": {
		LangZH: "o / esc 收起 · ↑/↓ 或 j/k 滚动",
		LangEN: "o / esc collapse · ↑/↓ or j/k scroll",
	},
	"expand.error":  {LangZH: "完整错误", LangEN: "Full error"},
	"expand.coach":  {LangZH: "辅导全文", LangEN: "Full coach reply"},
	"section.code":  {LangZH: "代码", LangEN: "Code"},
	"section.coach": {LangZH: "辅导", LangEN: "Coach"},
	"menu.plans":      {LangZH: "题单", LangEN: "Study plans"},
	"menu.plans.desc": {LangZH: "Hot 100 · 面试 150", LangEN: "Hot 100 · Interview 150"},
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
