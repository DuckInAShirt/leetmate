package leetgo

import (
	"testing"

	"github.com/DuckInAShirt/leetmate/internal/domain"
)

func TestParseMeta(t *testing.T) {
	// Simulates leetgo's `info --format json` shape; the parser should be
	// resilient to whichever casing leetgo happens to use.
	in := []byte(`{
		"TitleSlug": "two-sum",
		"QuestionFrontendId": "1",
		"Title": "Two Sum",
		"Difficulty": "Easy",
		"IsPaidOnly": false,
		"TopicTags": [{"Name": "Array"}, {"Name": "Hash Table"}]
	}`)
	got, err := parseMeta(in)
	if err != nil {
		t.Fatalf("parseMeta: %v", err)
	}
	if got.Slug != "two-sum" || got.FrontendID != "1" || got.Title != "Two Sum" {
		t.Errorf("unexpected meta: %+v", got)
	}
	if got.Difficulty != domain.DifficultyEasy {
		t.Errorf("difficulty = %q, want Easy", got.Difficulty)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "Array" {
		t.Errorf("tags = %v", got.Tags)
	}
}

func TestParseMetaArrayShape(t *testing.T) {
	// leetgo info --format json actually emits a JSON array, not a bare object.
	in := []byte(`[{
		"TitleSlug": "two-sum",
		"QuestionFrontendId": "1",
		"Title": "Two Sum",
		"Difficulty": "Easy",
		"TopicTags": [{"Name": "Array"}]
	}]`)
	got, err := parseMeta(in)
	if err != nil {
		t.Fatalf("parseMeta: %v", err)
	}
	if got.Slug != "two-sum" || got.FrontendID != "1" {
		t.Errorf("unexpected meta from array shape: %+v", got)
	}
}

func TestParseTestOutputLeetgoPassedCases(t *testing.T) {
	in := `● Case 1:    Passed
● Case 2:    Passed
● Case 3:    Passed

● running test locally question=longest-consecutive-sequence
● building file=go/0128.longest-consecutive-sequence/solution.go`
	r := parseTestOutput(in)
	if !r.Passed {
		t.Fatal("expected all passed leetgo cases to pass")
	}
}

func TestParseTestOutputLeetgoFailedCase(t *testing.T) {
	in := `● Case 1:    Passed
● Case 2:    Failed
● running test locally question=longest-consecutive-sequence`
	r := parseTestOutput(in)
	if r.Passed {
		t.Fatal("expected failed leetgo case not to pass")
	}
}

func TestParseSubmitOutputAccepted(t *testing.T) {
	in := "Accepted\nRuntime: 3 ms, faster than 95%\nMemory: 4.2 mb"
	r := parseSubmitOutput(in)
	if !r.Accepted {
		t.Error("expected Accepted")
	}
	if r.RuntimeMS != 3 {
		t.Errorf("runtime = %d, want 3", r.RuntimeMS)
	}
	if r.MemoryKB <= 0 {
		t.Errorf("memory = %d, want >0", r.MemoryKB)
	}
}

func TestParseSubmitOutputWrongAnswer(t *testing.T) {
	in := "Wrong Answer\ninput: [2,7]\nOutput: 0\nExpected: [0,1]"
	r := parseSubmitOutput(in)
	if r.Accepted {
		t.Error("expected not accepted")
	}
	if r.Raw == "" {
		t.Error("expected raw output preserved")
	}
}
