# Narrata developer tasks. Run `just` to list recipes.

# Default pure-Go build (no cgo, no model files).
build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

test-race:
	go test -race ./...

# Run the bundled examples end-to-end.
examples:
	go run ./examples/monitoring
	go run ./examples/game
	go run ./examples/home_automation

tidy:
	go mod tidy

# Build the experimental cgo backends (require local libs; see backend/README.md).
build-llama:
	go build -tags llama ./...

build-kokoro:
	go build -tags kokoro ./...

# Run the standard checks.
all: build vet test
