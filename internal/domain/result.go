package domain

// TestCaseResult is the outcome of a single test case.
type TestCaseResult struct {
	Passed   bool
	Input    string
	Expected string
	Actual   string
	StdError string
}

// TestResult is the outcome of `leetgo test` (local run).
type TestResult struct {
	Passed bool
	Cases  []TestCaseResult
	// Raw holds the full leetgo output (stdout+stderr) for expand-to-detail.
	Raw string
}

// SubmitResult is the outcome of `leetgo submit`. When Accepted is false,
// the first failing case is captured for the coaching context.
type SubmitResult struct {
	Accepted    bool
	RuntimeMS   int
	MemoryKB    int
	// FailedInput / FailedExpected / FailedActual describe the first failing case, if any.
	FailedInput    string
	FailedExpected string
	FailedActual   string
	StdError       string
	// Raw holds the raw stdout for debugging / edge cases we don't parse yet.
	Raw string
}
