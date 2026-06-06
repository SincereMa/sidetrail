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

	"github.com/SincereMa/cortex-sidemark/internal/schema"
)

// LoadFile reads path, validates the bytes against the record
// schema, and returns the parsed *Record. The schema is enforced
// before the JSON is unmarshalled, so a *Record returned from this
// function is known-valid.
func LoadFile(path string) (*Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
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
