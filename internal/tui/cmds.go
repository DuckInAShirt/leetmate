package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DuckInAShirt/leetmate/internal/coach"
	"github.com/DuckInAShirt/leetmate/internal/domain"
	"github.com/DuckInAShirt/leetmate/internal/llm"
)

func timeNow() time.Time { return time.Now() }

// --- leetgo result messages ---

type pickResultMsg struct {
	problem domain.Problem
	err     error
	planCtx *planCtx // non-nil when the pick was launched from a study plan
}

type submitResultMsg struct {
	result domain.SubmitResult
	err    error
}

type testResultMsg struct {
	result domain.TestResult
	err    error
}

type editorDoneMsg struct{ err error }

// --- coaching messages ---

// coachStartedMsg carries the live stream channel back to the model so it can
// attach a listener. Building the connection happens inside the command's
// goroutine so the UI never blocks on the HTTP handshake.
type coachStartedMsg struct {
	stream <-chan llm.Chunk
	tier   domain.Tier
}

type coachChunkMsg struct{ text string }
type coachDoneMsg struct{}
type coachErrMsg struct{ err error }

// planCtx ties a practice session to the study plan + item it came from, so an
// accepted submission can mark that item done.
type planCtx struct {
	planID string
	fid    string
}

// planMarkedMsg is emitted after a plan item is recorded done.
type planMarkedMsg struct {
	planID string
	fid    string
}

// --- leetgo commands ---

func pickCmd(deps Deps, qid string, pc *planCtx) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		p, err := deps.Leetgo.Pick(ctx, qid)
		if err == nil {
			_ = deps.Store.UpsertProblemMeta(ctx, p.ProblemMeta)
		}
		return pickResultMsg{problem: p, err: err, planCtx: pc}
	}
}

func submitCmd(deps Deps, slug, qid string, gaveUp bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		res, err := deps.Leetgo.Submit(ctx, qid)
		if err != nil {
			return submitResultMsg{err: err}
		}
		_, _ = deps.Store.InsertAttempt(ctx, domain.Attempt{
			Slug: slug, FinishedAt: timeNow(), AC: res.Accepted,
			RuntimeMS: res.RuntimeMS, MemoryKB: res.MemoryKB, GaveUp: gaveUp,
		})
		return submitResultMsg{result: res}
	}
}

func testCmd(deps Deps, qid string) tea.Cmd {
	return func() tea.Msg {
		res, err := deps.Leetgo.Test(context.Background(), qid)
		return testResultMsg{result: res, err: err}
	}
}

// --- coaching commands ---

// coachStartCmd opens the streaming connection in a goroutine and returns the
// stream channel via coachStartedMsg.
func coachStartCmd(deps Deps, req coach.Request) tea.Cmd {
	return func() tea.Msg {
		stream, err := deps.Coach.Stream(context.Background(), req)
		if err != nil {
			return coachErrMsg{err: err}
		}
		return coachStartedMsg{stream: stream, tier: req.Tier}
	}
}

// markDoneCmd records a study-plan item as finished (fire-and-forget).
func markDoneCmd(deps Deps, pc planCtx) tea.Cmd {
	return func() tea.Msg {
		_ = deps.Plans.MarkDone(context.Background(), pc.planID, pc.fid)
		return planMarkedMsg{planID: pc.planID, fid: pc.fid}
	}
}

// listenCoach reads one chunk from the stream and turns it into a message; the
// model re-issues this command after each chunk to keep draining the stream.
func listenCoach(stream <-chan llm.Chunk) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-stream
		if !ok {
			return coachDoneMsg{}
		}
		if chunk.Err != nil {
			return coachErrMsg{err: chunk.Err}
		}
		return coachChunkMsg{text: chunk.Text}
	}
}
