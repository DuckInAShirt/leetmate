package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/DuckInAShirt/leetmate/internal/domain"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestProblemMetaRoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	in := domain.ProblemMeta{
		FrontendID: "1", Slug: "two-sum", Title: "Two Sum",
		Difficulty: domain.DifficultyEasy, Tags: []string{"Array", "Hash Table"},
	}
	if err := s.UpsertProblemMeta(ctx, in); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	got, ok, err := s.GetProblemMeta(ctx, "two-sum")
	if err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if got.Title != "Two Sum" || got.Difficulty != domain.DifficultyEasy || len(got.Tags) != 2 {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestAttemptInsertAndList(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	a := domain.Attempt{Slug: "x", FinishedAt: time.Now(), AC: true, RuntimeMS: 5, Rating: domain.RatingGood}
	id, err := s.InsertAttempt(ctx, a)
	if err != nil || id == 0 {
		t.Fatalf("Insert: id=%d err=%v", id, err)
	}
	list, err := s.ListAttempts(ctx, "x")
	if err != nil || len(list) != 1 {
		t.Fatalf("List: %d err=%v", len(list), err)
	}
	if !list[0].AC || list[0].RuntimeMS != 5 {
		t.Errorf("attempt mismatch: %+v", list[0])
	}
}

func TestStudyPlanProgress(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.MarkStudyPlanDone(ctx, "hot100", "1"); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
	if err := s.MarkStudyPlanDone(ctx, "hot100", "49"); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
	done, err := s.StudyPlanDoneSet(ctx, "hot100")
	if err != nil {
		t.Fatalf("DoneSet: %v", err)
	}
	if !done["1"] || !done["49"] || len(done) != 2 {
		t.Errorf("done set = %v", done)
	}
	// Re-marking is idempotent.
	if err := s.MarkStudyPlanDone(ctx, "hot100", "1"); err != nil {
		t.Fatalf("re-MarkDone: %v", err)
	}
	done, _ = s.StudyPlanDoneSet(ctx, "hot100")
	if len(done) != 2 {
		t.Errorf("MarkDone not idempotent: %d rows", len(done))
	}
	// Different plans are independent.
	if err := s.MarkStudyPlanDone(ctx, "interview150", "88"); err != nil {
		t.Fatalf("MarkDone other plan: %v", err)
	}
	done, _ = s.StudyPlanDoneSet(ctx, "interview150")
	if len(done) != 1 || !done["88"] {
		t.Errorf("plans should be independent: %v", done)
	}
}

func TestCardUpsertAndDue(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	due := now.Add(-time.Hour)       // due 1h ago
	future := now.Add(24 * time.Hour) // not due

	for slug, when := range map[string]time.Time{"due1": due, "later": future} {
		if err := s.UpsertCard(ctx, domain.Card{Slug: slug, DueAt: when, Reps: 1}); err != nil {
			t.Fatalf("Upsert %s: %v", slug, err)
		}
	}

	got, err := s.DueCards(ctx, now, 10)
	if err != nil {
		t.Fatalf("DueCards: %v", err)
	}
	if len(got) != 1 || got[0].Slug != "due1" {
		t.Errorf("expected only due1 due, got %+v", got)
	}

	// Re-scheduling due1 into the future removes it from the due set.
	if err := s.UpsertCard(ctx, domain.Card{Slug: "due1", DueAt: future, Reps: 2}); err != nil {
		t.Fatalf("re-Upsert: %v", err)
	}
	got, _ = s.DueCards(ctx, now, 10)
	if len(got) != 0 {
		t.Errorf("expected no due cards after re-schedule, got %+v", got)
	}
}
