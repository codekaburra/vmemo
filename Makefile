.PHONY: build clean test preflight smoke

build:
	go build -o vmemo .

clean:
	rm -f vmemo

test:
	go test ./...

preflight:
	go test -v -run TestOllama -run TestRequired ./...

smoke: build
	@echo "=== smoke: help ==="
	./vmemo --help | head -1
	@echo "=== smoke: unknown command ==="
	./vmemo badcmd 2>&1 | grep -q 'unknown command' && echo "PASS" || echo "FAIL"
	@echo "=== smoke: ask no args ==="
	./vmemo ask 2>&1 | grep -q 'usage:' && echo "PASS" || echo "FAIL"
	@echo "=== smoke: ask round-trip (needs Ollama) ==="
	./vmemo ask "Reply with only the word OK" | grep -qi 'ok' && echo "PASS" || echo "FAIL"
	@echo "=== all smoke tests done ==="
