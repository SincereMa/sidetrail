package record

import "crypto/rand"

// randReader adapts crypto/rand to the io.Reader shape that
// oklog/ulid's Monotonic entropy source expects.
type randReader struct{}

func (randReader) Read(p []byte) (int, error) {
	return rand.Read(p)
}
