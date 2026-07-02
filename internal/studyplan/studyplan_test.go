package studyplan

import "testing"

func TestBuiltinPlansHaveExpectedCounts(t *testing.T) {
	plans, err := Builtin()
	if err != nil {
		t.Fatalf("Builtin: %v", err)
	}
	if len(plans) != 2 {
		t.Fatalf("expected 2 builtin plans, got %d", len(plans))
	}
	want := map[string]int{"hot100": 100, "interview150": 150}
	for _, p := range plans {
		n := len(p.Items)
		if n != want[p.ID] {
			t.Errorf("%s: %d items, want %d", p.ID, n, want[p.ID])
		}
		// Items must be unique within a plan.
		seen := make(map[string]bool, n)
		for _, fid := range p.Items {
			if fid == "" {
				t.Errorf("%s: empty frontend id", p.ID)
			}
			if seen[fid] {
				t.Errorf("%s: duplicate frontend id %s", p.ID, fid)
			}
			seen[fid] = true
		}
		if !p.Builtin {
			t.Errorf("%s: Builtin flag not set", p.ID)
		}
	}
}
