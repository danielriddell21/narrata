# Narrata developer tasks. Run `just` to list recipes.

# Default pure-Go build (no cgo, no model files).
[group('build')]
build:
	go build ./...

[group('dev')]
vet:
	go vet ./...

[group('test')]
test:
	go test ./...

[group('test')]
test-race:
	go test -race ./...

# Run the bundled examples end-to-end.
[group('run')]
examples:
	go run ./examples/monitoring
	go run ./examples/game
	go run ./examples/home_automation

[group('dev')]
tidy:
	go mod tidy

# Build the experimental cgo backends (require local libs; see backend/README.md).
[group('build')]
build-llama:
	go build -tags llama ./...

[group('build')]
build-kokoro:
	go build -tags kokoro ./...

# Run the standard checks.
[group('dev')]
all: build vet test
