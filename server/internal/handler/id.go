package handler

import (
	"crypto/rand"
	"encoding/hex"
)

// newID returns a 16-byte random hex identifier (32 chars) used for generated
// workspace ids.
func newID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand never errors on supported platforms; a zero value fallback
		// keeps the id format stable without crashing.
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(buf)
}