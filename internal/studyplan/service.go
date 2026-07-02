package studyplan

import (
	"context"

	"leetmate/internal/store"
)

// Service wraps the plan list with progress tracking on top of the store.
type Service struct {
	st    *store.Store
	plans []*Plan
}

// NewService builds a Service over the given plans and store.
func NewService(st *store.Store, plans []*Plan) *Service {
	return &Service{st: st, plans: plans}
}

// Plans returns all loaded plans (builtin + custom).
func (s *Service) Plans() []*Plan { return s.plans }

// Plan returns the plan with the given id, or nil.
func (s *Service) Plan(id string) *Plan { return Find(s.plans, id) }

// Progress returns (done, total) for a plan.
func (s *Service) Progress(ctx context.Context, id string) (int, int) {
	p := Find(s.plans, id)
	if p == nil {
		return 0, 0
	}
	done, err := s.st.StudyPlanDoneSet(ctx, id)
	if err != nil {
		return 0, len(p.Items)
	}
	n := 0
	for _, fid := range p.Items {
		if done[fid] {
			n++
		}
	}
	return n, len(p.Items)
}

// NextTodo returns the first not-yet-done item's frontend id and its 1-based
// position, or ok=false if the plan is complete.
func (s *Service) NextTodo(ctx context.Context, id string) (fid string, position int, ok bool) {
	p := Find(s.plans, id)
	if p == nil {
		return "", 0, false
	}
	done, err := s.st.StudyPlanDoneSet(ctx, id)
	if err != nil {
		done = nil
	}
	for i, fid := range p.Items {
		if !done[fid] {
			return fid, i + 1, true
		}
	}
	return "", 0, false
}

// MarkDone records a problem as finished in a plan.
func (s *Service) MarkDone(ctx context.Context, id, fid string) error {
	return s.st.MarkStudyPlanDone(ctx, id, fid)
}

// DoneSet returns the finished frontend-id set for a plan.
func (s *Service) DoneSet(ctx context.Context, id string) map[string]bool {
	done, err := s.st.StudyPlanDoneSet(ctx, id)
	if err != nil {
		return map[string]bool{}
	}
	return done
}
