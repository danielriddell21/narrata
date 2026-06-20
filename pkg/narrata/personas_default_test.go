package narrata

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRootDefaultPersonasMatchEmbedded guards against the documented root
// personas.default.json drifting from the embedded source of truth.
func TestRootDefaultPersonasMatchEmbedded(t *testing.T) {
	// Test runs from the package directory; the repo root is two levels up.
	rootPath := filepath.Join("..", "..", "personas.default.json")
	rootBytes, err := os.ReadFile(rootPath)
	if err != nil {
		t.Skipf("root personas.default.json not found: %v", err)
	}
	if string(rootBytes) != string(defaultPersonasJSON) {
		t.Fatal("root personas.default.json differs from embedded pkg/narrata/personas.default.json")
	}
}
