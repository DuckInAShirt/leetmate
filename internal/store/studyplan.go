package store

import (
	"context"
	"time"
)

// MarkStudyPlanDone records that a problem is finished within a study plan.
// Re-marking is idempotent.
func (s *Store) MarkStudyPlanDone(ctx context.Context, planID, frontendID string) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO studyplan_progress(plan_id, frontend_id, updated_at)
VALUES(?,?,?)
ON CONFLICT(plan_id, frontend_id) DO UPDATE SET updated_at=excluded.updated_at`,
		planID, frontendID, time.Now().UTC().Format(time.RFC3339))
	return err
}

// StudyPlanDoneSet returns the set of finished frontend ids for a plan.
func (s *Store) StudyPlanDoneSet(ctx context.Context, planID string) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT frontend_id FROM studyplan_progress WHERE plan_id=?`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]bool)
	for rows.Next() {
		var fid string
		if err := rows.Scan(&fid); err != nil {
			return nil, err
		}
		out[fid] = true
	}
	return out, rows.Err()
}
