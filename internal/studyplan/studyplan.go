// Package studyplan provides built-in and user-defined problem lists (Hot 100,
// Interview 150, …) plus progress tracking on top of the store.
package studyplan

import (
	"encoding/json"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Plan is a named, ordered list of problem frontend ids.
type Plan struct {
	ID          string   `json:"id" yaml:"id"`
	Title       string   `json:"title" yaml:"title"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Source      string   `json:"source,omitempty" yaml:"source,omitempty"`
	Items       []string `json:"items" yaml:"items"` // frontend ids, leetcode.cn
	Builtin     bool     `json:"-" yaml:"-"`
}

//go:embed data/hot100.json
var hot100JSON []byte

//go:embed data/interview150.json
var interview150JSON []byte

// Builtin returns the plans shipped with leetmate.
func Builtin() ([]*Plan, error) {
	var out []*Plan
	for _, b := range [][]byte{hot100JSON, interview150JSON} {
		var p Plan
		if err := json.Unmarshal(b, &p); err != nil {
			return nil, err
		}
		p.Builtin = true
		out = append(out, &p)
	}
	return out, nil
}

// LoadCustom reads user plans (*.yaml/*.yml/*.json) from dir. A missing dir is
// not an error. Plans without an explicit id derive it from the filename.
func LoadCustom(dir string) ([]*Plan, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	var out []*Plan
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var p Plan
		switch ext {
		case ".json":
			err = json.Unmarshal(b, &p)
		default:
			err = yaml.Unmarshal(b, &p)
		}
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		if p.ID == "" {
			p.ID = strings.TrimSuffix(e.Name(), ext)
		}
		out = append(out, &p)
	}
	return out, nil
}

// All returns builtin plans followed by user plans.
func All(customDir string) ([]*Plan, error) {
	b, err := Builtin()
	if err != nil {
		return nil, err
	}
	c, err := LoadCustom(customDir)
	if err != nil {
		return nil, err
	}
	return append(b, c...), nil
}

// Find returns the plan with the given id, or nil.
func Find(plans []*Plan, id string) *Plan {
	for _, p := range plans {
		if p.ID == id {
			return p
		}
	}
	return nil
}
