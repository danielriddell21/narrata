# list available recipes
default:
    @just --list

# default pure-Go build (no cgo, no model files)
[group('build')]
build:
    go build ./...

# run the tests
[group('test')]
test:
    go test ./...

# run the linter
[group('dev')]
lint:
    golangci-lint run

# format the code
[group('dev')]
fmt:
    golangci-lint fmt

# tidy module dependencies
[group('dev')]
tidy:
    go mod tidy

# full gate: lint + test + build. all must pass before committing
[group('dev')]
ci: lint test build

# run the tests with the race detector
[group('test')]
test-race:
    go test -race ./...

# run the bundled examples end-to-end
[group('run')]
examples:
    go run ./examples/monitoring
    go run ./examples/game
    go run ./examples/home_automation
