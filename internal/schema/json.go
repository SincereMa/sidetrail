package schema

import "encoding/json"

// jsonUnmarshal is a thin wrapper around encoding/json kept in
// its own file so the schema package's only call to encoding/json
// is easy to audit and to mock in future tests.
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
