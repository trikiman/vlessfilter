.PHONY: build test lint smoke fmt clean tidy

BIN := bin/vlessfilter
PKG := ./cmd/vlessfilter

build:
	@mkdir -p bin
	go build -o $(BIN) $(PKG)
	@echo "built $(BIN)"

test:
	# ./... not ./internal/... — the latter silently skipped every test under
	# cmd/, so the dedup-endpoints tests did not run under `make test`.
	go test ./... -count=1

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

# Runs in a throwaway dir. This used to `rm -rf subs/ README.md` in the repo
# root, which deletes 319 TRACKED files — subs/ is gitignored but published, so
# the ignore rule made it look disposable. It also regenerated README.md via
# the deprecated single-protocol path, so committing afterwards would replace
# the multi-protocol README with the old format.
SMOKE_DIR ?= /tmp/vf-smoke

smoke: build
	@rm -rf $(SMOKE_DIR) && mkdir -p $(SMOKE_DIR)
	./$(BIN) run --sources sources.yaml --out $(SMOKE_DIR) --threads2 5 --limit 50
	@test -d $(SMOKE_DIR)/subs || (echo "FAIL: subs/ not created" && exit 1)
	@test -f $(SMOKE_DIR)/README.md || (echo "FAIL: README.md not created" && exit 1)
	@grep -q "| Country |" $(SMOKE_DIR)/README.md || (echo "FAIL: README.md missing table header" && exit 1)
	@echo "✓ smoke passed ($(SMOKE_DIR))"

# Build artifacts only. Never the published data.
clean:
	rm -rf bin $(SMOKE_DIR)
