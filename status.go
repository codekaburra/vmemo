package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type itemStatus struct {
	rawPath  string
	stem     string
	category string
	models   []string
}

func scanStatus(dir string) ([]itemStatus, []string, error) {
	rawFiles, err := findRawFiles(dir)
	if err != nil {
		return nil, nil, err
	}

	modelSet := make(map[string]bool)
	var items []itemStatus

	for _, raw := range rawFiles {
		fileDir := filepath.Dir(raw)
		base := filepath.Base(raw)
		stem := stemFromRaw(base)

		rel, _ := filepath.Rel(dir, fileDir)
		parts := strings.SplitN(rel, string(os.PathSeparator), 2)
		category := parts[0]

		entries, err := os.ReadDir(fileDir)
		if err != nil {
			continue
		}

		var models []string
		for _, e := range entries {
			name := e.Name()
			if name == base || !strings.HasPrefix(name, stem+"_") {
				continue
			}
			suffix := strings.TrimPrefix(name, stem+"_")
			suffix = strings.TrimSuffix(suffix, ".txt")
			suffix = strings.TrimPrefix(suffix, "clean_")
			models = append(models, suffix)
			modelSet[suffix] = true
		}

		items = append(items, itemStatus{
			rawPath:  raw,
			stem:     stem,
			category: category,
			models:   models,
		})
	}

	var allModels []string
	for m := range modelSet {
		allModels = append(allModels, m)
	}
	sort.Strings(allModels)

	return items, allModels, nil
}

type categoryStats struct {
	total    int
	byModel map[string]int
}

func printStatus(dir string) error {
	items, allModels, err := scanStatus(dir)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Println("no *_raw.txt files in", dir)
		return nil
	}

	cats := make(map[string]*categoryStats)
	var catOrder []string

	for _, it := range items {
		cs, ok := cats[it.category]
		if !ok {
			cs = &categoryStats{byModel: make(map[string]int)}
			cats[it.category] = cs
			catOrder = append(catOrder, it.category)
		}
		cs.total++
		for _, m := range it.models {
			cs.byModel[m]++
		}
	}
	sort.Strings(catOrder)

	fmt.Printf("\n  vtidy status — %s\n", dir)
	fmt.Printf("  %d transcripts, %d AI models detected\n\n", len(items), len(allModels))

	colW := 12
	for _, m := range allModels {
		if len(m)+2 > colW {
			colW = len(m) + 2
		}
	}

	fmt.Printf("  %-12s %6s", "Category", "Files")
	for _, m := range allModels {
		fmt.Printf("  %-*s", colW, m)
	}
	fmt.Println()
	fmt.Printf("  %-12s %6s", "--------", "-----")
	for range allModels {
		fmt.Printf("  %-*s", colW, strings.Repeat("-", colW-2))
	}
	fmt.Println()

	totalFiles := 0
	totalByModel := make(map[string]int)

	for _, cat := range catOrder {
		cs := cats[cat]
		totalFiles += cs.total

		display := cat
		if len(display) > 10 {
			display = display[:10]
		}
		fmt.Printf("  %-12s %4d  ", display, cs.total)
		for _, m := range allModels {
			count := cs.byModel[m]
			totalByModel[m] += count
			if count == cs.total {
				fmt.Printf("  \033[32m%d/%d ✓\033[0m", count, cs.total)
				pad := colW - len(fmt.Sprintf("%d/%d ✓", count, cs.total))
				fmt.Printf("%-*s", pad, "")
			} else if count > 0 {
				fmt.Printf("  \033[33m%d/%d\033[0m", count, cs.total)
				pad := colW - len(fmt.Sprintf("%d/%d", count, cs.total))
				fmt.Printf("%-*s", pad, "")
			} else {
				fmt.Printf("  \033[31m0/%d\033[0m", cs.total)
				pad := colW - len(fmt.Sprintf("0/%d", cs.total))
				fmt.Printf("%-*s", pad, "")
			}
		}
		fmt.Println()
	}

	fmt.Printf("  %-12s %6s", "--------", "-----")
	for range allModels {
		fmt.Printf("  %-*s", colW, strings.Repeat("-", colW-2))
	}
	fmt.Println()
	fmt.Printf("  %-12s %4d  ", "Total", totalFiles)
	for _, m := range allModels {
		fmt.Printf("  %d/%d", totalByModel[m], totalFiles)
		pad := colW - len(fmt.Sprintf("%d/%d", totalByModel[m], totalFiles))
		fmt.Printf("%-*s", pad, "")
	}
	fmt.Println()
	fmt.Println()

	return nil
}
