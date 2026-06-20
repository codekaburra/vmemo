package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadPrompt(name string) (string, error) {
	path := filepath.Join("prompts", name+".txt")
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("load prompt %s: %w", path, err)
	}
	return strings.TrimSpace(string(b)), nil
}

func writeAtomic(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".vtidy-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

func tidy(inboxDir, outDir string, models []string) error {
	prompt, err := loadPrompt("tidy")
	if err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(inboxDir, "*.txt"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("no .txt files in", inboxDir)
		return nil
	}

	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", f, err)
			continue
		}
		content := strings.TrimSpace(string(raw))
		if content == "" {
			fmt.Fprintf(os.Stderr, "skip %s: empty\n", f)
			continue
		}

		id := makeID(filepath.Base(f))

		for _, model := range models {
			outFile := filepath.Join(outDir, id+".clean_"+modelSlug(model)+".md")
			if _, err := os.Stat(outFile); err == nil {
				fmt.Printf("skip %s (exists)\n", outFile)
				continue
			}

			fmt.Printf("tidy  %s  model=%s\n", filepath.Base(f), model)
			reply, err := Chat(model, prompt, content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error %s [%s]: %v\n", filepath.Base(f), model, err)
				continue
			}

			if err := writeAtomic(outFile, reply); err != nil {
				fmt.Fprintf(os.Stderr, "write %s: %v\n", outFile, err)
				continue
			}
			fmt.Printf("wrote %s\n", outFile)
		}
	}
	return nil
}
