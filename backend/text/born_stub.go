//go:build !born

package text

import "fmt"

// newBorn is the stub used in the default, zero-dependency build. The real,
// pure-Go Born backend is compiled only with the "born" build tag (see
// born.go), which also adds the github.com/born-ml/born module dependency.
func newBorn(_ Options) (Backend, error) {
	return nil, fmt.Errorf("text: born backend not built; rebuild with -tags born (adds github.com/born-ml/born)")
}
