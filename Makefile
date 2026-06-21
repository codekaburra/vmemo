.PHONY: build clean test preflight smoke

build:
	go build -o vtidy .

clean:
	rm -f vtidy

test:
	go test ./...

preflight:
	OLLAMA_PREFLIGHT=1 go test -v -run 'TestOllama|TestRequired' ./...

smoke: build
	@echo "=== smoke: help ==="
	./vtidy --help | head -1
	@echo "=== smoke: unknown command ==="
	./vtidy badcmd 2>&1 | grep -q 'unknown command' && echo "PASS" || echo "FAIL"
	@echo "=== smoke: ask no args ==="
	./vtidy ask 2>&1 | grep -q 'usage:' && echo "PASS" || echo "FAIL"
	@echo "=== smoke: ask round-trip (needs Ollama) ==="
	./vtidy ask "Reply with only the word OK" | grep -qiw 'ok' && echo "PASS" || echo "FAIL"
	@echo "=== all smoke tests done ==="
