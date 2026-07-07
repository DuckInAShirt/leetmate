package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DuckInAShirt/leetmate/internal/coach"
	"github.com/DuckInAShirt/leetmate/internal/domain"
	"github.com/DuckInAShirt/leetmate/internal/llm"
	"github.com/DuckInAShirt/leetmate/internal/review"
)

func timeNow() time.Time { return time.Now() }

// --- leetgo result messages ---

type pickResultMsg struct {
	problem domain.Problem
	err     error
	planCtx *planCtx // non-nil when the pick was launched from a study plan
}

type submitResultMsg struct {
	result     domain.SubmitResult
	err        error
	persistErr error
}

type testResultMsg struct {
	result domain.TestResult
	err    error
}

type editorDoneMsg struct{ err error }

type editorSavedMsg struct {
	content string
	err     error
}

type reviewEmptyMsg struct{}

// --- coaching messages ---

// coachStartedMsg carries the live stream channel back to the model so it can
// attach a listener. Building the connection happens inside the command's
// goroutine so the UI never blocks on the HTTP handshake.
type coachStartedMsg struct {
	stream <-chan llm.Chunk
	tier   domain.Tier
}

type coachChunkMsg struct {
	text string
	kind llm.ChunkKind
}
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
		now := timeNow()
		rating := review.RatingForSubmit(res.Accepted, gaveUp)
		var persistErr error
		if _, persistErr = deps.Store.InsertAttempt(ctx, domain.Attempt{
			Slug: slug, FinishedAt: now, AC: res.Accepted,
			RuntimeMS: res.RuntimeMS, MemoryKB: res.MemoryKB, Rating: rating, GaveUp: gaveUp,
		}); persistErr == nil {
			card, _, err := deps.Store.GetCard(ctx, slug)
			if err != nil {
				persistErr = err
			} else {
				card.Slug = slug
				persistErr = deps.Store.UpsertCard(ctx, review.Rate(card, rating, now))
			}
		}
		return submitResultMsg{result: res, persistErr: persistErr}
	}
}

func testCmd(deps Deps, qid string) tea.Cmd {
	return func() tea.Msg {
		res, err := deps.Leetgo.Test(context.Background(), qid)
		return testResultMsg{result: res, err: err}
	}
}

func reviewPickCmd(deps Deps) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		cards, err := deps.Store.DueCards(ctx, timeNow(), 1)
		if err != nil {
			return pickResultMsg{err: err}
		}
		if len(cards) == 0 {
			return reviewEmptyMsg{}
		}
		card := cards[0]
		qid := card.Slug
		if meta, ok, err := deps.Store.GetProblemMeta(ctx, card.Slug); err != nil {
			return pickResultMsg{err: err}
		} else if ok && meta.FrontendID != "" {
			qid = meta.FrontendID
		}
		p, err := deps.Leetgo.Pick(ctx, qid)
		if err != nil {
			return pickResultMsg{err: fmt.Errorf("review %s: %w", card.Slug, err)}
		}
		_ = deps.Store.UpsertProblemMeta(ctx, p.ProblemMeta)
		return pickResultMsg{problem: p}
	}
}

func saveCodeCmd(path, content string) tea.Cmd {
	return func() tea.Msg {
		if path == "" {
			return editorSavedMsg{content: content}
		}
		err := os.WriteFile(path, []byte(content), 0o644)
		return editorSavedMsg{content: content, err: err}
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
		return coachChunkMsg{text: chunk.Text, kind: chunk.Kind}
	}
}
