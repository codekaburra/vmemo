package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

var requiredModels = []string{"mistral:7b", "phi4"}

func TestOllamaRunning(t *testing.T) {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		t.Fatalf("Ollama not reachable at localhost:11434: %v", err)
	}
	resp.Body.Close()
}

func TestRequiredModelsInstalled(t *testing.T) {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		t.Skipf("Ollama not reachable, skipping: %v", err)
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
