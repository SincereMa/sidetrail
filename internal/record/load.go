// Package record: file-level loading and validation helpers.
//
// LoadFile is the entry point for turning a record JSON file on
// disk into a validated *Record. The schema is the source of truth
// for what a record is; this layer only adds I/O, parse, and a
// single, descriptive error.
package record

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/SincereMa/sidetrail/internal/schema"
)

// LoadFile reads path, fills in any missing auto-generated fields,
// validates the bytes against the record schema, and returns the
// parsed *Record. The schema is enforced after defaults are applied,
// so a *Record returned from this function is known-valid.
func LoadFile(path string) (*Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	data, err = fillDefaults(data)
	if err != nil {
		return nil, fmt.Errorf("fill defaults for %q: %w", path, err)
	}
	if err := schema.ValidateRecord(data); err != nil {
		return nil, fmt.Errorf("validate %q: %w", path, err)
	}
	var r Record
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse %q: %w", path, err)
	}
	return &r, nil
}

// fillDefaults unmarshals data into a map, fills in missing
// auto-generated fields (id, created_at, last_verified_at, and
// decided_at for decisions), and re-marshals the result.
func fillDefaults(data []byte) ([]byte, error) {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	now := time.Now().UTC()

	if id, _ := m["id"].(string); id == "" {
		newID, err := NewID()
		if err != nil {
			return nil, fmt.Errorf("generate id: %w", err)
		}
		m["id"] = newID
	}

	if _, ok := m["created_at"]; !ok {
		m["created_at"] = now.Format(time.RFC3339)
	}

	if _, ok := m["last_verified_at"]; !ok {
		m["last_verified_at"] = now.Format(time.RFC3339)
	}

	if kind, _ := m["kind"].(string); kind == "decision" {
		if _, ok := m["decided_at"]; !ok {
			m["decided_at"] = now.Format(time.RFC3339)
		}
	}

	return json.Marshal(m)
}
