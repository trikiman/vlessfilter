.PHONY: build test lint smoke fmt clean tidy

BIN := bin/vlessfilter
PKG := ./cmd/vlessfilter

build:
	@mkdir -p bin
	go build -o $(BIN) $(PKG)
	@echo "built $(BIN)"

test:
	go test ./internal/... -count=1

lint:
	go vet ./...
	@unformatted=$$(gofmt -l . | grep -v '^\.planning/' || true); \
		if [ -n "$$unformatted" ]; then \
			echo "gofmt failures:"; echo "$$unformatted"; exit 1; \
		fi

fmt:
	gofmt -w .

tidy:
	go mod tidy

smoke: build
	@rm -rf subs/ README.md
	./$(BIN) run --sources sources.yaml --threads2 5 --limit 50
	@test -d subs/ || (echo "FAIL: subs/ not created" && exit 1)
	@test -f README.md || (echo "FAIL: README.md not created" && exit 1)
	@grep -q "| Country |" README.md || (echo "FAIL: README.md missing table header" && exit 1)
	@echo "✓ smoke passed"

clean:
	rm -rf bin subs README.md all-results.csv raw
