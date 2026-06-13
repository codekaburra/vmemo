# vmemo — Project Plan & Requirements

> Local tool that takes voice→text transcripts and uses **local LLMs (Ollama)** to
> **tidy** and **analyze** them. Batch pipeline, not a web app.

- **Project / module / binary name:** `vmemo`
- **Status:** planning complete, not yet built
- **Folder:** `vmemo` (renamed from `vocie-memo-handler`, fixing the "vocie" typo)
- **Platform:** macOS (darwin), local-only, no cloud

---

## 1. Goal

The user has many voice→text transcripts. Use local LLMs to:
1. **Tidy** each transcript (fix grammar/punctuation/filler, keep meaning).
2. **Analyze** each transcript (summary · themes · tags · action items/todos).

Decided shape: a **Go CLI** (not a web app), because the requirements are a
run-and-done batch pipeline with three logical outputs per item. A web UI can
wrap the same core later with no rework if browsing/search is ever wanted.

---

## 2. Input sources

Two kinds, both supported:
- **Text files** — dropped into `inbox/` anytime. The original file *is* the raw;
  no duplicate raw copy is written.
- **Clipboard / big blob** — via `vmemo add`: reads `pbpaste`, or splits
  `inbox/_blob.txt` on `---` lines into separate items. For these,
  a `.raw.txt` **is** written (to preserve the original, since there's no source file).

Avoid pasting directly into the terminal as a prompt — `pbpaste`/files instead.

---

## 3. Models (two tiers, different clocks)

| Model | Tier | Role | When it runs |
|---|---|---|---|
| `mistral:7b` | **fast** | tidy + analyze | immediately on demand, live monitor, or frequent poll |
| `phi4` | **slow** | tidy + analyze | **daily 3:00am only** (too slow for interactive use) |
| ~~`starcoder2:7b`~~ | — | — | **unused** (code model, not for prose) |

Both models run **each stage so their outputs can be compared** side by side.
mistral lands within seconds; phi4 provides a slower overnight second opinion.

Models overridable via `--models` flag. Defaults: interactive = `mistral:7b`,
the 3am job passes `--models phi4`.

---

## 4. File naming pattern

Shared group **id** as prefix, **stage** + **model** as suffix, so all files for
one memo sort together and provenance is obvious.

```
<id>.raw.txt                 # only for clipboard/blob sources
<id>.clean_mistral7b.md      # mistral tidy
<id>.clean_phi4.md           # phi4 tidy (overnight)
<id>.analysis_mistral7b.md   # mistral analysis
<id>.analysis_phi4.md        # phi4 analysis (overnight)
```

- `id = YYYY-MM-DD_<slug>`
- `<slug>` = original filename stem (file source) or first non-empty line / short
  title (blob source), sanitized + lowercased.

Example:
```
2026-06-13_grocery-startup-idea.clean_mistral7b.md
2026-06-13_grocery-startup-idea.analysis_phi4.md
```

**Analysis input:** each model analyzes its *own* cleaned version
(`clean_mistral7b → analysis_mistral7b`, `clean_phi4 → analysis_phi4`),
giving two complete end-to-end pipelines to compare.

---

## 5. Pipeline & state machine

State = **which output files exist** (no database). For **each model independently**:

```
clean missing                 → next step: tidy
clean exists, analysis missing → next step: analyze
all 4 generated files present  → DONE
```

Each pass advances every item one step per model. Properties:
- **Resumable** — skip-if-exists; drop new files anytime.
- **Atomic writes** — temp file + rename, so a crash can't leave a half-written output.
- **Full auto** — watcher takes new drops all the way to (fast-tier) done with no prompting.
- **Tiered** — `runPass(models)`: watch/run use `["mistral:7b"]`; 3am job uses `["phi4"]`.

An item is **"fast-done"** once mistral's 2 files exist, **"fully done"** after the 3am phi4 pass.

---

## 6. Commands

| Command | Does |
|---|---|
| `vmemo tidy` | tidy stage only → `*.clean_<model>.md` |
| `vmemo analyze` | analysis stage only → `*.analysis_<model>.md` (skips + warns if matching clean file missing) |
| `vmemo run` | full pipeline: tidy → analyze, to done |
| `vmemo watch` | live monitor (fsnotify + 30s safety sweep, debounced); full pipeline, mistral |
| `vmemo add` | clipboard (`pbpaste`) / blob → raw items, then process |

**Shared flags:** `--models <list>` (default `mistral:7b`), `--inbox <path>`,
`--out <path>`, `--stage`, `--sweep <duration>`, `--sep <string>`.

---

## 7. Trigger / scheduling map

| Trigger | Command |
|---|---|
| Tidy only, on demand | `vmemo tidy` |
| Analysis only, on demand | `vmemo analyze` |
| Full run, on demand | `vmemo run` |
| Live monitor | `vmemo watch` |
| Frequent poll (optional) | `vmemo run` via launchd every ~10 min |
| Nightly slow tier | `vmemo run --models phi4` via **launchd at 3:00am** |

Scheduling via **launchd** (macOS-native, survives reboots/login); cron as fallback.
Deliverable includes two `.plist` files: `watch`-at-login and phi4-at-3am.

---

## 8. Proposed layout

```
vmemo/
  inbox/                 # drop .txt anytime
  out/                   # the 4 (+raw) files per item
  prompts/
    tidy.txt             # editable cleanup prompt
    analyze.txt          # summary · themes · tags · todos
  main.go                # run / tidy / analyze / watch / add
  ollama.go              # Chat(model, system, user) → POST localhost:11434/api/chat
  pipeline.go            # tidy(model) / analyze(model)
  state.go               # per-model: which files exist → next steps
  slug.go                # id = YYYY-MM-DD_<slug>
  scheduler/             # launchd plists: watch-at-login + phi4-at-3am
```

Prompts live as editable text files so behavior can be tuned without recompiling.

---

## 9. Risks / notes (all handled)

- **phi4 is 14B → slow** → confined to the 3am job; interactive runs never block on it.
- **fsnotify can fire mid-copy** → debounce until file size is stable before processing.
- **Mac asleep at 3am** → launchd runs the phi4 job at next wake (catch-up overnight). Acceptable.
- Uses only models already installed (`mistral:7b`, `phi4`) — no `ollama pull` needed.

---

## 10. Build stages (vertical slices — each one builds & is demoable)

Ordered so every stage ends in something runnable, not internal scaffolding.

**Stage 0 — Skeleton**
`go mod init vmemo`, `main.go` command routing, `vmemo --help`.
✅ See: binary builds, prints its command list.

**Stage 1 — Ollama round-trip**
`ollama.go` `Chat(model, system, user)`. Temporary `vmemo ask "..."` to prove it talks to `mistral:7b` on `localhost:11434`.
✅ See: type a question, get a model reply.

**Stage 2 — Tidy (mistral, files only)** ← *first real value*
`vmemo tidy`: `inbox/*.txt` → mistral → `<id>.clean_mistral7b.md`. Includes `slug.go`.
✅ See: drop a messy transcript, get a cleaned `.md`.

**Stage 3 — Analyze + full run** ← *first milestone, stop & use it*
`vmemo analyze` → `<id>.analysis_mistral7b.md`. `vmemo run` = tidy then analyze.
✅ See: one command turns a raw transcript into clean + analysis. Now usable.

**Stage 4 — State machine hardening**
`state.go`: state from existing files, skip-if-exists, atomic writes (temp+rename).
✅ See: re-running does no duplicate work; deleting one output regenerates just it.

**Stage 5 — Second model / compare**
`--models` flag; produce `*_phi4.md` pair alongside mistral's.
✅ See: two clean + two analysis files per item, diffable side by side.

**Stage 6 — Clipboard / blob input**
`vmemo add`: `pbpaste` + `inbox/_blob.txt` split on `---`, writes `.raw.txt`.
✅ See: copy text, run `add`, it gets ingested + processed.

**Stage 7 — Watcher (full auto)**
`vmemo watch`: fsnotify + debounce + 30s sweep, fast tier.
✅ See: leave it running, drop a file, output appears unprompted.

**Stage 8 — Scheduling**
launchd plists: `watch`-at-login + `phi4` at 3am. (§11 decision applies here.)
✅ See: hands-off; phi4 files appear overnight.

### How they stack
```
0 → 1 → 2 → 3   = usable single-model tool (aim here first)
        4       = robust/resumable
        5       = adds phi4 compare
        6,7,8   = automation & convenience (can reorder/defer)
```
Recommendation: treat **Stage 3** as the first milestone — use it on real transcripts
and let that feedback tune the prompts before automating. Stages 6–8 are pure
convenience and can be deferred.

---

## 11. Open decision (last item before build)

**What should the nightly 3am phi4 job run?**
- `vmemo run --models phi4` (full) — phi4 does its own tidy + analysis → complete compare pair. Slower.
- `vmemo analyze --models phi4` (analyze-only) — phi4 only adds `analysis_phi4`, reusing mistral's clean text. Faster, no phi4 clean version.

_To be decided._
