package leetgo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"leetmate/internal/domain"
	"gopkg.in/yaml.v3"
)

// --- leetgo.yaml parsing ----------------------------------------------------

type leetgoConfigFile struct {
	Code struct {
		Lang string `yaml:"lang"`
	} `yaml:"code"`
}

// readLang reads the configured code language from leetgo.yaml in the workspace.
func readLang(workspace string) (string, error) {
	b, err := os.ReadFile(filepath.Join(workspace, "leetgo.yaml"))
	if err != nil {
		return "", err
	}
	var f leetgoConfigFile
	if err := yaml.Unmarshal(b, &f); err != nil {
		return "", err
	}
	return f.Code.Lang, nil
}

// langExt maps a leetgo language to its source file extension.
func langExt(lang string) string {
	switch strings.ToLower(lang) {
	case "go":
		return ".go"
	case "python", "python3":
		return ".py"
	case "cpp", "c++":
		return ".cpp"
	case "c":
		return ".c"
	case "java":
		return ".java"
	case "rust", "rs":
		return ".rs"
	case "javascript", "js":
		return ".js"
	case "typescript", "ts":
		return ".ts"
	case "kotlin":
		return ".kt"
	case "ruby":
		return ".rb"
	case "swift":
		return ".swift"
	case "csharp", "c#":
		return ".cs"
	default:
		return ".txt"
	}
}

// isTestFile reports whether a generated file is a test/stub rather than the
// learner's solution file, per language convention.
func isTestFile(name, lang string) bool {
	switch strings.ToLower(lang) {
	case "go":
		return strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, ".test.go")
	case "python", "python3":
		return strings.HasPrefix(name, "test_") || strings.HasSuffix(name, "_test.py")
	default:
		return strings.Contains(strings.ToLower(name), "test")
	}
}

// readStatement returns the problem statement text from the generated directory.
func readStatement(dir, slug string) string {
	if md, err := os.ReadFile(filepath.Join(dir, "question.md")); err == nil {
		return string(md)
	}
	if html, err := os.ReadFile(filepath.Join(dir, "question.html")); err == nil {
		return string(html)
	}
	return ""
}

// --- info --format json parsing ---------------------------------------------

// parseMeta is defensive: leetgo's `info --format json` output is observed to
// be a JSON **array** of question objects (even for a single id), but we also
// tolerate a bare object. Field names are not documented as stable, so we read
// several candidate keys per field.
func parseMeta(out []byte) (domain.ProblemMeta, error) {
	// Try array-of-objects first (the shape leetgo currently emits).
	var arr []map[string]any
	if err := json.Unmarshal(out, &arr); err == nil {
		if len(arr) == 0 {
			return domain.ProblemMeta{}, fmt.Errorf("leetgo info returned empty list")
		}
		return metaFromMap(arr[0]), nil
	}
	// Fall back to a single object.
	var obj map[string]any
	if err := json.Unmarshal(out, &obj); err != nil {
		return domain.ProblemMeta{}, err
	}
	return metaFromMap(obj), nil
}

// metaFromMap reads problem metadata from a single leetgo question object,
// tolerant of several field-name spellings.
func metaFromMap(raw map[string]any) domain.ProblemMeta {
	m := domain.ProblemMeta{
		FrontendID: firstString(raw, "frontend_id", "QuestionFrontendId", "questionFrontendId", "FrontendId"),
		Slug:       firstString(raw, "slug", "TitleSlug", "titleSlug"),
		Title:      firstString(raw, "title", "Title", "TitleCN", "translatedTitle"),
		Difficulty: domain.Difficulty(firstString(raw, "difficulty", "Difficulty")),
		IsPaidOnly: firstBool(raw, "is_paid_only", "IsPaidOnly", "isPaidOnly"),
		Tags:       topicTags(raw),
	}
	if m.Title == "" {
		m.Title = m.Slug
	}
	return m
}

func topicTags(raw map[string]any) []string {
	for _, key := range []string{"tags", "Tags", "TopicTags", "topicTags"} {
		v, ok := raw[key]
		if !ok {
			continue
		}
		switch t := v.(type) {
		case []string:
			return t
		case []any:
			out := make([]string, 0, len(t))
			for _, item := range t {
				switch it := item.(type) {
				case string:
					out = append(out, it)
				case map[string]any:
					if name := firstString(it, "name", "Name", "slug", "Slug"); name != "" {
						out = append(out, name)
					}
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return nil
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func firstBool(m map[string]any, keys ...string) bool {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}
	return false
}

// --- test output parsing ----------------------------------------------------

func parseTestOutput(out string) domain.TestResult {
	res := domain.TestResult{Passed: strings.Contains(out, "PASS") || strings.Contains(out, "Accepted")}
	res.Passed = res.Passed && !strings.Contains(out, "FAIL")
	return res
}

// --- submit output parsing --------------------------------------------------

var (
	runtimeRe = regexp.MustCompile(`(?i)runtime[:\s]*([0-9,]+)\s*ms`)
	memoryRe  = regexp.MustCompile(`(?i)memory[:\s]*([0-9.,]+)\s*(mb|kb)`)
	numberRe  = regexp.MustCompile(`[0-9]`)
)

// parseSubmitOutput extracts the verdict from `leetgo submit` stdout. This is
// inherently fragile (leetgo's human-readable format), so the raw text is
// preserved on the result for debugging and the coach's context.
func parseSubmitOutput(out string) domain.SubmitResult {
	res := domain.SubmitResult{Raw: out}
	upper := strings.ToUpper(out)

	switch {
	case strings.Contains(upper, "ACCEPTED"):
		res.Accepted = true
	case strings.Contains(upper, "WRONG ANSWER"), strings.Contains(upper, "WRONG_ANSWER"):
		res.Accepted = false
	default:
		// Unknown verdict — leave Accepted false, keep Raw.
	}

	if m := runtimeRe.FindStringSubmatch(out); len(m) > 1 {
		if n, err := strconv.Atoi(strings.ReplaceAll(m[1], ",", "")); err == nil {
			res.RuntimeMS = n
		}
	}
	if m := memoryRe.FindStringSubmatch(out); len(m) > 1 {
		val, err := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", ""), 64)
		if err == nil {
			if strings.EqualFold(m[2], "mb") {
				res.MemoryKB = int(val * 1024)
			} else {
				res.MemoryKB = int(val)
			}
		}
	}
	_ = numberRe
	return res
}
