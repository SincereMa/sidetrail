// Package schema loads the record JSON Schema and exposes
// ValidateRecord for the rest of the sidecar to call.
package schema

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// embeddedRecordSchema is the record schema source compiled into
// the binary. It is loaded once at init and reused for every
// validation call.
//
//go:embed record.schema.json
var embeddedRecordSchema []byte

// recordSchema is the compiled record schema. It is set by init
// and read by ValidateRecord. Storing it in a package-level
// variable avoids recompiling the schema on every call.
var recordSchema *jsonschema.Schema

func init() {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	if err := compiler.AddResource("record.json", bytes.NewReader(embeddedRecordSchema)); err != nil {
		panic(fmt.Sprintf("schema: add resource: %v", err))
	}
	s, err := compiler.Compile("record.json")
	if err != nil {
		panic(fmt.Sprintf("schema: compile: %v", err))
	}
	recordSchema = s
}

// ValidateRecord reports whether data is a JSON object that
// conforms to the record schema. It returns a descriptive error
// suitable for surfacing to humans and to `--json`-mode agents.
func ValidateRecord(data []byte) error {
	var v any
	if err := jsonUnmarshal(data, &v); err != nil {
		return fmt.Errorf("not valid json: %w", err)
	}
	return recordSchema.Validate(v)
}
