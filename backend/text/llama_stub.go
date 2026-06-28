//go:build !llama

package text

import "fmt"

// newLlama is the stub used in the default pure-Go build. The real llama.cpp
// backend is compiled only with the "llama" build tag (see llama_cpp.go), so
// CI and tests never require a model file or cgo toolchain.
func newLlama(_ Options) (Backend, error) {
	return nil, fmt.Errorf("text: llama.cpp backend not built; rebuild with -tags llama")
}
