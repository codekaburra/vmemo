package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
)

var requiredModels = []string{"mistral:7b", "phi4"}

// preflight tests check the local environment (Ollama running, models
// installed) rather than code. They are gated behind OLLAMA_PREFLIGHT so a
// plain `go test ./...` stays hermetic; `make preflight` sets the env var.
func requirePreflight(t *testing.T) {
	if os.Getenv("OLLAMA_PREFLIGHT") == "" {
		t.Skip("set OLLAMA_PREFLIGHT=1 to run preflight checks")
	}
}

func TestOllamaRunning(t *testing.T) {
	requirePreflight(t)
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		t.Fatalf("Ollama not reachable at localhost:11434: %v", err)
	}
	resp.Body.Close()
}

func TestRequiredModelsInstalled(t *testing.T) {
	requirePreflight(t)
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		t.Fatalf("Ollama not reachable at localhost:11434: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode Ollama response: %v", err)
	}

	installed := make(map[string]bool)
	for _, m := range result.Models {
		installed[m.Name] = true
	}

	for _, want := range requiredModels {
		t.Run(want, func(t *testing.T) {
			if !installed[want] && !installed[want+":latest"] {
				t.Errorf("model %q not installed — run: ollama pull %s", want, want)
			} else {
				fmt.Printf("  ✓ %s installed\n", want)
			}
		})
	}
}
