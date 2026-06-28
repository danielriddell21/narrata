//go:build !kokoro

package tts

import "fmt"

// newKokoro is the stub used in the default build. The real Kokoro/ONNX
// backend is compiled only with the "kokoro" build tag (see kokoro.go).
func newKokoro(_ Options) (Backend, error) {
	return nil, fmt.Errorf("tts: kokoro backend not built; rebuild with -tags kokoro")
}
