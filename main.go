package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const usage = `vmemo — tidy and analyze voice-to-text transcripts using local LLMs

Usage:
  vmemo <command> [flags]

Commands:
  tidy      Clean up transcripts  → <id>.clean_<model>.md
  analyze   Analyze transcripts   → <id>.analysis_<model>.md
  run       Full pipeline: tidy then analyze
  watch     Live monitor inbox/ (fsnotify + 30s sweep)
  add       Ingest from clipboard (pbpaste) or inbox/_blob.txt

Flags (shared across commands):
  --models   Comma-separated model list  (default "mistral:7b")
  --inbox    Inbox directory             (default "inbox")
  --out      Output directory            (default "out")
  --sweep    Watcher safety sweep interval (default "30s")
  --sep      Blob separator string       (default "---")

Examples:
  vmemo run
  vmemo tidy --models phi4
  vmemo watch
  vmemo add
`

// shared flags
var (
	flagModels = flag.String("models", "mistral:7b", "comma-separated model list")
	flagInbox  = flag.String("inbox", "inbox", "inbox directory")
	flagOut    = flag.String("out", "out", "output directory")
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

func cmdTidy() {
	fmt.Printf("tidy  models=%s  inbox=%s  out=%s\n", *flagModels, *flagInbox, *flagOut)
	fmt.Println("(not yet implemented — Stage 2)")
}

func cmdAnalyze() {
	fmt.Printf("analyze  models=%s  inbox=%s  out=%s\n", *flagModels, *flagInbox, *flagOut)
	fmt.Println("(not yet implemented — Stage 3)")
}

func cmdRun() {
	fmt.Printf("run  models=%s  inbox=%s  out=%s\n", *flagModels, *flagInbox, *flagOut)
	fmt.Println("(not yet implemented — Stage 3)")
}

func cmdWatch() {
	fmt.Printf("watch  models=%s  inbox=%s  out=%s  sweep=%s\n", *flagModels, *flagInbox, *flagOut, *flagSweep)
	fmt.Println("(not yet implemented — Stage 7)")
}

func cmdAdd() {
	fmt.Printf("add  models=%s  inbox=%s  out=%s  sep=%s\n", *flagModels, *flagInbox, *flagOut, *flagSep)
	fmt.Println("(not yet implemented — Stage 6)")
}

// temporary command to verify Ollama round-trip (Stage 1)
func cmdAsk() {
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: vmemo ask \"your question\"")
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
