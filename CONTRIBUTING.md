# Contributing to narrata

## Requirements

* [Go](https://go.dev) (stable — version from `go.mod`)
* [just](https://github.com/casey/just)
* [golangci-lint](https://golangci-lint.run/welcome/install/) (for linting)
* [gremlins](https://github.com/go-gremlins/gremlins) (for mutation testing)

## Development workflow

```
just build         # build the narrata CLI
just test          # run unit tests
just test-race     # run tests with the race detector
just vet           # go vet
just examples      # run the bundled examples
just tidy          # go mod tidy
just build-llama   # build with the llama.cpp text backend
just build-kokoro  # build with the kokoro TTS backend
just all           # vet + test + build
```

Run `just --list` to see every recipe. Lint with `golangci-lint run --config .golangci.yml` and run the tests before each commit. CI runs lint + test + build on every push to `trunk` and every pull request targeting `trunk`.

## Project layout

narrata is both an embeddable library and a developer CLI. The public library
package lives at the module root (`import "github.com/danielriddell21/narrata"`).

```
narrata package    public library API (module root)
backend/           pluggable text and TTS backends
cmd/narrata/       developer CLI (validate, gen, generate, personas)
internal/          implementation packages (not part of the public API)
examples/          runnable usage examples
docs/              documentation
```

## Commit style

```
type(scope): short imperative description
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`. No period at the end of the subject line; keep it under 72 characters.

## Releases

Releases are triggered by pushing a semver tag — maintainers only. A GitHub Actions workflow runs GoReleaser to build the CLI binaries and update the Homebrew tap; it requires the tap app credentials configured as repository secrets. Library consumers pin a version with `go get`.
