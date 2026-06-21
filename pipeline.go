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

// stemFromRaw extracts the base stem from a _raw.txt filename.
// e.g. "grocery-idea_raw.txt" → "grocery-idea"
func stemFromRaw(filename string) string {
	return strings.TrimSuffix(filename, "_raw.txt")
}

func findRawFiles(dir string) ([]string, error) {
	var files []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_raw.txt") {
			files = append(files, path)
		}
		return nil
	})
	return files, nil
}

func tidy(dir string, models []string) error {
	prompt, err := loadPrompt("tidy")
	if err != nil {
		return err
	}

	files, err := findRawFiles(dir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("no *_raw.txt files in", dir)
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

		stem := stemFromRaw(filepath.Base(f))
		fileDir := filepath.Dir(f)

		for _, model := range models {
			outFile := filepath.Join(fileDir, stem+"_clean_"+modelSlug(model)+".txt")
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
