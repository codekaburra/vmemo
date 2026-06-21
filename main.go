package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const usage = `vtidy — tidy and analyze voice-to-text transcripts using local LLMs

Usage:
  vtidy <command> [flags]

Commands:
  tidy      Clean up transcripts  → <stem>_clean_<model>.txt
  analyze   Analyze transcripts   → <stem>_analysis_<model>.txt
  run       Full pipeline: tidy then analyze
  watch     Live monitor (fsnotify + 30s sweep)
  add       Ingest from clipboard (pbpaste) or blob file

Flags (shared across commands):
  --models   Comma-separated model list  (default "mistral:7b,phi4")
  --dir      Resources directory         (default "resources")
  --sweep    Watcher safety sweep interval (default "30s")
  --sep      Blob separator string       (default "---")

Examples:
  vtidy run
  vtidy tidy --models phi4
  vtidy watch
  vtidy add
`

// shared flags
var (
	flagModels = flag.String("models", "mistral:7b,phi4", "comma-separated model list")
	flagDir    = flag.String("dir", "resources", "resources directory")
	flagSweep  = flag.String("sweep", "30s", "watcher safety sweep interval")
	flagSep    = flag.String("sep", "---", "blob separator")
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(0)
	}

	cmd := os.Args[1]
	if cmd == "--help" || cmd == "-h" || cmd == "help" {
		fmt.Print(usage)
		os.Exit(0)
	}

	// parse flags after the subcommand
	flag.CommandLine.Parse(os.Args[2:])

	switch cmd {
	case "tidy":
		cmdTidy()
	case "analyze":
		cmdAnalyze()
	case "run":
		cmdRun()
	case "watch":
		cmdWatch()
	case "add":
		cmdAdd()
	case "ask":
		cmdAsk()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func parseModels(s string) []string {
	var out []string
	for _, m := range strings.Split(s, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			out = append(out, m)
		}
	}
	return out
}

func cmdTidy() {
	models := parseModels(*flagModels)
	if err := tidy(*flagDir, models); err != nil {
		fmt.Fprintf(os.Stderr, "tidy: %v\n", err)
		os.Exit(1)
	}
}

func cmdAnalyze() {
	fmt.Printf("analyze  models=%s  dir=%s\n", *flagModels, *flagDir)
	fmt.Println("(not yet implemented — Stage 3)")
}

func cmdRun() {
	fmt.Printf("run  models=%s  dir=%s\n", *flagModels, *flagDir)
	fmt.Println("(not yet implemented — Stage 3)")
}

func cmdWatch() {
	fmt.Printf("watch  models=%s  dir=%s  sweep=%s\n", *flagModels, *flagDir, *flagSweep)
	fmt.Println("(not yet implemented — Stage 7)")
}

func cmdAdd() {
	fmt.Printf("add  models=%s  dir=%s  sep=%s\n", *flagModels, *flagDir, *flagSep)
	fmt.Println("(not yet implemented — Stage 6)")
}

// temporary command to verify Ollama round-trip (Stage 1)
func cmdAsk() {
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: vtidy ask \"your question\"")
		os.Exit(1)
	}
	question := strings.Join(args, " ")
	fmt.Printf("asking %s: %s\n", *flagModels, question)

	reply, err := Chat(*flagModels, "You are a helpful assistant. Be concise.", question)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(reply)
}
