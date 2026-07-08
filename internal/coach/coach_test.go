package coach

import (
	"context"
	"strings"
	"testing"

	"github.com/DuckInAShirt/leetmate/internal/domain"
	"github.com/DuckInAShirt/leetmate/internal/llm"
)

// fakeProvider captures the messages sent and replays canned chunks.
type fakeProvider struct {
	chunks []string
	msgs   []llm.Message
	opts   llm.Options
}

func (f *fakeProvider) Chat(_ context.Context, msgs []llm.Message, opts llm.Options) (<-chan llm.Chunk, error) {
	f.msgs = msgs
	f.opts = opts
	ch := make(chan llm.Chunk, len(f.chunks))
	for _, c := range f.chunks {
		ch <- llm.Chunk{Text: c}
	}
	close(ch)
	return ch, nil
}

func TestSystemPromptHasGuardrailsForHint(t *testing.T) {
	f := &fakeProvider{}
	c := New(f)
	req := Request{
		Tier: domain.TierHint,
		Problem: domain.Problem{
			ProblemMeta: domain.ProblemMeta{FrontendID: "1", Title: "Two Sum", Difficulty: domain.DifficultyEasy},
			Content:     "给定一个整数数组 nums",
		},
		Code: "func twoSum() {}",
		Lang: "go",
	}
	if _, err := c.Stream(context.Background(), req); err != nil {
		t.Fatalf("Stream: %v", err)
	}
	if len(f.msgs) < 2 {
		t.Fatalf("expected system + user messages, got %d", len(f.msgs))
	}
	sys := f.msgs[0].Content
	// The anti-cheat guardrail and the tier instruction must both be injected.
	if !strings.Contains(sys, "绝不") {
		t.Errorf("system prompt missing anti-cheat guardrail:\n%s", sys)
	}
	if !strings.Contains(sys, "Hint") {
		t.Errorf("system prompt missing tier instruction:\n%s", sys)
	}
	user := f.msgs[len(f.msgs)-1].Content
	if !strings.Contains(user, "Two Sum") || !strings.Contains(user, "给定一个整数数组") {
		t.Errorf("user context missing problem statement:\n%s", user)
	}
}

func TestReviewTierAsksForHighConfidenceFindings(t *testing.T) {
	f := &fakeProvider{}
	c := New(f)
	_, _ = c.Stream(context.Background(), Request{
		Tier:    domain.TierReview,
		Problem: domain.Problem{ProblemMeta: domain.ProblemMeta{FrontendID: "49", Title: "Group Anagrams", Difficulty: domain.DifficultyMedium}},
		Code:    "func groupAnagrams(strs []string) [][]string { return nil }",
		Lang:    "go",
	})
	sys := f.msgs[0].Content
	for _, want := range []string{"高置信", "只报告", "可优化", "不要夸大"} {
		if !strings.Contains(sys, want) {
			t.Errorf("Review tier system prompt missing %q:\n%s", want, sys)
		}
	}
}

func TestAnswerTierGetsAnswerInstruction(t *testing.T) {
	f := &fakeProvider{}
	c := New(f)
	_, _ = c.Stream(context.Background(), Request{
		Tier:    domain.TierAnswer,
		Problem: domain.Problem{ProblemMeta: domain.ProblemMeta{FrontendID: "1", Title: "X", Difficulty: domain.DifficultyEasy}},
	})
	if !strings.Contains(f.msgs[0].Content, "完整解答") {
		t.Errorf("Answer tier system prompt should permit full solution:\n%s", f.msgs[0].Content)
	}
}

func TestStreamAssemblesChunks(t *testing.T) {
	f := &fakeProvider{chunks: []string{"Hello", " ", "world"}}
	c := New(f)
	stream, err := c.Stream(context.Background(), Request{
		Tier:    domain.TierHint,
		Problem: domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "x"}},
	})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	var sb strings.Builder
	for ch := range stream {
		if ch.Err != nil {
			t.Fatal(ch.Err)
		}
		sb.WriteString(ch.Text)
	}
	if sb.String() != "Hello world" {
		t.Errorf("assembled text = %q, want %q", sb.String(), "Hello world")
	}
}

func TestHistoryIsCarried(t *testing.T) {
	f := &fakeProvider{}
	c := New(f)
	_, _ = c.Stream(context.Background(), Request{
		Tier:    domain.TierNudge,
		Problem: domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "x"}},
		History: []domain.Conversation{
			{Role: domain.RoleAssistant, Content: "previous hint"},
		},
	})
	// Expect system, then the history assistant turn, then the new user turn.
	if len(f.msgs) != 3 {
		t.Fatalf("expected 3 messages (system+history+user), got %d", len(f.msgs))
	}
	if f.msgs[1].Content != "previous hint" || f.msgs[1].Role != llm.RoleAssistant {
		t.Errorf("history not carried correctly: %+v", f.msgs[1])
	}
}

func TestReviewTierSkipsHistory(t *testing.T) {
	f := &fakeProvider{}
	c := New(f)
	_, _ = c.Stream(context.Background(), Request{
		Tier: domain.TierReview,
		Problem: domain.Problem{ProblemMeta: domain.ProblemMeta{
			FrontendID: "1", Title: "Two Sum", Difficulty: domain.DifficultyEasy,
		}},
		Code: "func twoSum(nums []int, target int) []int { return []int{} }",
		Lang: "go",
		History: []domain.Conversation{
			{Role: domain.RoleAssistant, Content: "old review about stale code"},
		},
	})
	if len(f.msgs) != 2 {
		t.Fatalf("expected review to send only system+current user context, got %d", len(f.msgs))
	}
	for _, msg := range f.msgs {
		if strings.Contains(msg.Content, "old review about stale code") {
			t.Fatalf("review should not carry stale history: %+v", f.msgs)
		}
	}
	if !strings.Contains(f.msgs[1].Content, "func twoSum") {
		t.Fatalf("review should include current code: %s", f.msgs[1].Content)
	}
}

func TestCoachUsesConciseTokenBudgets(t *testing.T) {
	cases := []struct {
		tier domain.Tier
		want int
	}{
		{domain.TierHint, 80},
		{domain.TierNudge, 180},
		{domain.TierReview, 420},
		{domain.TierAnswer, 900},
	}
	for _, tc := range cases {
		t.Run(string(tc.tier), func(t *testing.T) {
			f := &fakeProvider{}
			c := New(f)
			_, _ = c.Stream(context.Background(), Request{
				Tier:    tc.tier,
				Problem: domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s", Title: "Sample"}},
			})
			if f.opts.MaxTokens != tc.want {
				t.Fatalf("MaxTokens = %d, want %d", f.opts.MaxTokens, tc.want)
			}
			if f.opts.Temperature != 0.3 {
				t.Fatalf("Temperature = %v, want 0.3", f.opts.Temperature)
			}
		})
	}
}

func TestReviewPromptAsksForConciseDiagnosis(t *testing.T) {
	f := &fakeProvider{}
	c := New(f)
	_, _ = c.Stream(context.Background(), Request{
		Tier:    domain.TierReview,
		Problem: domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s", Title: "Sample"}},
		Code:    "func trap(height []int) int { return 0 }",
	})
	if len(f.msgs) == 0 {
		t.Fatal("expected messages")
	}
	sys := f.msgs[0].Content
	for _, want := range []string{"默认短答", "输出最多 3 条", "不要长篇模拟样例"} {
		if !strings.Contains(sys, want) {
			t.Fatalf("system/review prompt missing %q:\n%s", want, sys)
		}
	}
}

func TestReviewPromptIncludesPassedTestContext(t *testing.T) {
	f := &fakeProvider{}
	c := New(f)
	_, _ = c.Stream(context.Background(), Request{
		Tier:        domain.TierReview,
		Problem:     domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s", Title: "Sample"}},
		Code:        "func trap(height []int) int { return 6 }",
		TestContext: "最近一次测试集已通过；这不等于逻辑已被证明正确。",
	})
	if len(f.msgs) < 2 {
		t.Fatal("expected system and user messages")
	}
	for _, want := range []string{"最近测试结果", "当前测试集通过", "不代表逻辑已被证明正确", "最小反例"} {
		if !strings.Contains(f.msgs[0].Content, want) {
			t.Fatalf("review system prompt missing %q:\n%s", want, f.msgs[0].Content)
		}
	}
	if !strings.Contains(f.msgs[1].Content, "【最近测试结果】") || !strings.Contains(f.msgs[1].Content, "最近一次测试集已通过") || !strings.Contains(f.msgs[1].Content, "不等于逻辑已被证明正确") {
		t.Fatalf("user prompt missing test context:\n%s", f.msgs[1].Content)
	}
}
