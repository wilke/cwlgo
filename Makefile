# Makefile for CWLGo

.PHONY: build test clean example

# Build the library
build:
	go build ./...

# Run tests
test:
	go test -v ./...

# Run the echo example
echo-example:
	cd examples/echo && go run simple.go echo.cwl message="Hello, CWL!"

# Run the grep example
grep-example:
	cd examples/grep && go run grep_example.go

# Clean up
clean:
	rm -rf output/
	find . -name "*.test" -delete

# Install the library
install:
	go install ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	go vet ./...

# Generate documentation
doc:
	godoc -http=:6060

# Run all checks
check: fmt lint test