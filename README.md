# vmemo

A local CLI tool that takes voice-to-text transcripts and uses **local LLMs via [Ollama](https://ollama.com)** to tidy and analyze them. Runs entirely on-device — no cloud, no API keys.

## What it does

For each transcript it produces up to four output files:

```
<YYYY-MM-DD_slug>.clean_mistral7b.md      # cleaned-up transcript
<YYYY-MM-DD_slug>.analysis_mistral7b.md   # summary · themes · tags · todos
<YYYY-MM-DD_slug>.clean_phi4.md           # overnight second opinion (tidy)
<YYYY-MM-DD_slug>.analysis_phi4.md        # overnight second opinion (analysis)
```

## Models

| Model | Tier | When |
|---|---|---|
| `mistral:7b` | fast | on demand, live watcher, frequent poll |
| `phi4` | slow | nightly at 3:00am via launchd |

## Commands

```
vmemo tidy       # tidy stage only
vmemo analyze    # analysis stage only (skips if clean file missing)
vmemo run        # full pipeline: tidy → analyze
vmemo watch      # live monitor — drop a file, output appears automatically
vmemo add        # ingest from clipboard (pbpaste) or inbox/_blob.txt
```

Common flags: `--models <list>`, `--inbox <path>`, `--out <path>`, `--sweep <duration>`.

## Input

- **Text files** — drop `.txt` files into `inbox/` anytime.
- **Clipboard / blob** — run `vmemo add` to pull from `pbpaste` or split a multi-entry `inbox/_blob.txt` on `---` lines.

## Requirements

- Go 1.22+
- [Ollama](https://ollama.com) running locally with `mistral:7b` and `phi4` pulled

## Status

Planning complete. See [PLAN.md](PLAN.md) for full requirements and build stages.
