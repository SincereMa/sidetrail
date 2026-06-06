// Package schema is the test surface for the schema package.
package schema

import (
	_ "embed"
	"testing"
)

//go:embed testdata/valid_decision.json
var validDecision []byte

//go:embed testdata/valid_constraint.json
var validConstraint []byte

//go:embed testdata/invalid_missing_reason.json
var invalidMissingReason []byte

//go:embed testdata/invalid_wrong_kind.json
var invalidWrongKind []byte

//go:embed testdata/invalid_decision_no_decided_at.json
var invalidDecisionNoDate []byte

// TestValidate_Valid checks that the valid fixtures pass.
func TestValidate_Valid(t *testing.T) {
	for name, data := range map[string][]byte{
		"decision":   validDecision,
		"constraint": validConstraint,
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateRecord(data); err != nil {
				t.Fatalf("expected valid, got: %v", err)
			}
		})
	}
}

// TestValidate_Invalid checks that the invalid fixtures fail.
func TestValidate_Invalid(t *testing.T) {
	for name, data := range map[string][]byte{
		"missing_reason":      invalidMissingReason,
		"wrong_kind":          invalidWrongKind,
		"decision_no_decided": invalidDecisionNoDate,
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateRecord(data); err == nil {
				t.Fatalf("expected invalid, got nil")
			}
		})
	}
}
