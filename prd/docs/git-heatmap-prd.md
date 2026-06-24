# PRD — `git-heat`: Terminal Git Contribution Heatmap

> **Status:** Draft v1.0  
> **Author:** Shashank  
> **Target:** Portfolio project for backend/DevOps internship applications

---

## 1. Overview

`git-heat` is a CLI tool that visualizes code ownership and contribution patterns directly in the terminal, built on top of git internals (`git log`, `git blame`, commit graph traversal). It answers the question: **"Who actually wrote what in this codebase, and when?"** — without leaving your terminal.

---

## 2. Problem Statement

Code ownership is invisible until something breaks. `git log` gives you commits. `git blame` gives you line-level history. But neither tells you the full picture:

- Which engineer owns which **module or directory** over time?
- Which files are **hot zones** — touched most recently or most frequently?
- Where is **ownership concentrated** (single point of failure)?

Existing tools are either GUI-only (GitLens), paid (GitHub Insights), or require uploading your repo to a third-party service. `git-heat` gives you this as a fast, offline, terminal-first tool.

---

## 3. Goals

- Parse a git repo's history using git plumbing commands
- Render a heatmap in the terminal (color-coded by author / recency / churn)
- Support multiple views: per-file, per-directory, per-author
- Work on any git repo, any size, offline, zero signup
- Complete in < 3 seconds on repos with up to 10,000 commits

---

## 4. Non-Goals (Hard NOs)

These are explicitly out of scope. Do not add them, even if they seem useful.

| What | Why Not |
|------|---------|
| GUI / web UI | Terminal-first is the identity. A web UI is a different product. |
| GitHub/GitLab API integration | Must work fully offline on any repo |
| Real-time file watching | Out of scope; this is a snapshot tool, not a daemon |
| Blame diffing across branches | Too complex for v1; confuses the core use case |
| Rewriting git history or mutating the repo | Non-negotiable safety boundary |
| Authentication / tokens for private repos | Tool reads local `.git/` only — auth is handled by git itself |
| Windows support (v1) | Terminal rendering with ANSI colors is unreliable on CMD/PowerShell in v1 |
| Plugin architecture / extensibility API | Premature in v1 |
| Exporting to PDF | Not the format engineers want |

---

## 5. Scope

### 5.1 Must Have (v1)

- **`git-heat files`** — Heatmap of all files in the repo, color-coded by last-touched date (green = recent, red = stale)
- **`git-heat dir [path]`** — Heatmap of subdirectories; defaults to repo root
- **`git-heat author [name]`** — Show which files/dirs a specific author owns (% of lines)
- **`git-heat summary`** — Top 10 hottest files (most commits in last 90 days) + ownership concentration score
- Author filtering via `--author` flag on any command
- Date range filtering: `--since` / `--until` (accepts natural format: `30d`, `6m`, `1y`)
- Output modes: `--color` (default, ANSI 256), `--no-color`, `--json` (structured output for piping)
- Respects `.gitignore` automatically
- Works on any local git repo; point at it with `--repo <path>` or run from inside

### 5.2 Should Have (v1.5)

- `--json` output piped to a file, consumable by Prometheus pushgateway or a Grafana JSON datasource
- `--top N` flag to limit output rows
- Config file (`.githeatrc`) for default flags per project
- Support for multiple authors in `--author` (comma-separated)

### 5.3 Won't Have (v1)

Everything listed in section 4.

---

## 6. Core Data Model

```
Commit
  └── sha, author, email, timestamp, files_changed[]

FileEntry
  └── path, last_commit_sha, last_author, commit_count_90d, line_ownership{}

LineOwnership
  └── author → line_count (from git blame)

HeatScore
  └── recency_score (0–100), churn_score (0–100), ownership_concentration (0–1)
```

---

## 7. Technical Design

### Stack

| Layer | Choice | Reason |
|-------|--------|--------|
| Language | **Go** | Single binary, fast startup, excellent `os/exec` for git plumbing |
| Terminal rendering | **`github.com/charmbracelet/bubbletea`** + `lipgloss` | Best-in-class terminal UI for Go |
| Git parsing | `git log --format`, `git blame --porcelain` via subprocess | Avoids libgit2 dependency; simpler |
| Output | ANSI 256-color + plain JSON | Covers both human and machine consumers |

> Alternative: Python with `rich` library — simpler to write, acceptable performance for v1.

### Git Commands Used (Plumbing Layer)

```bash
# Commit graph
git log --format="%H|%ae|%at|%s" --name-only

# File ownership (line-level)
git blame --porcelain <file>

# Active files in last N days
git log --since="90 days ago" --name-only --format="" | sort | uniq -c

# Repo-wide author list
git shortlog -sne
```

### Heat Score Formula

```
recency_score  = 100 * (1 - days_since_last_commit / 365)   # capped at 0
churn_score    = min(100, commit_count_90d * 5)
heat           = 0.6 * recency_score + 0.4 * churn_score
```

Color mapping: heat ≥ 80 → bright red, 60–79 → orange, 40–59 → yellow, < 40 → green/dim

---

## 8. CLI Interface

```bash
# Installation
go install github.com/shashank/git-heat@latest

# Commands
git-heat files                        # heatmap of all files
git-heat files --top 20               # limit to top 20 hottest
git-heat dir src/                     # directory view
git-heat author "Shashank"            # files owned by author
git-heat summary                      # quick overview
git-heat summary --json               # machine-readable output
git-heat summary --since 90d          # last 90 days only
git-heat --repo /path/to/other/repo files
```

---

## 9. Output Design

### Terminal (default)

```
git-heat files — repo: SystemCraft  (last 90 days)

██████  src/server/index.ts          Shashank      3d ago    42 commits
█████░  src/db/schema.ts             Shashank      5d ago    31 commits
████░░  src/workers/queue.ts         Riya          12d ago   18 commits
░░░░░░  docs/old-api.md              Unknown       8mo ago    2 commits

Legend:  ██ Hot   █░ Warm   ░░ Cold
```

### JSON (--json)

```json
{
  "generated_at": "2026-06-25T10:00:00Z",
  "repo": "SystemCraft",
  "window_days": 90,
  "files": [
    {
      "path": "src/server/index.ts",
      "heat_score": 94,
      "last_author": "Shashank",
      "days_since_last_commit": 3,
      "commit_count": 42,
      "ownership": { "Shashank": 0.87, "Riya": 0.13 }
    }
  ]
}
```

---

## 10. Performance Constraints

| Repo size | Max runtime |
|-----------|------------|
| < 1,000 commits | < 500ms |
| 1,000–10,000 commits | < 3s |
| > 10,000 commits | Warn user; offer `--since 90d` to limit scope |

Large repos: stream git output instead of buffering; blame only the top-N files by commit count.

---

## 11. Error Handling

| Scenario | Behavior |
|----------|----------|
| Not a git repo | `Error: no git repository found at <path>` |
| Author not found | `Warning: no commits found for author "<name>"` |
| Binary files | Skip silently (detect via `git blame` exit code) |
| Repo with 0 commits | `Error: repository has no commit history` |
| Git not installed | `Error: git not found in PATH` |

---

## 12. Success Metrics

This is a portfolio project. Success looks like:

- **Functional:** All v1 commands work correctly on repos including SystemCraft and DSCE Vertex Club backend
- **Demo-ready:** A single `git-heat summary` on SystemCraft produces a compelling terminal screenshot for the README
- **Installable:** Single binary, `go install` works, tested on macOS + Linux
- **Resume bullet:** "Built `git-heat`, a CLI tool that parses git internals to render terminal heatmaps of code ownership and commit churn — used to analyze SystemCraft's codebase"
- **Stretch:** `--json` output wired into Grafana via JSON datasource panel

---

## 13. Milestones

| Week | Deliverable |
|------|-------------|
| 1 | Git plumbing layer: parse `git log` + `git blame`, build data model |
| 2 | Terminal renderer: ANSI colors, bar visualization, `files` + `dir` commands |
| 3 | `author` command, `--json` output, error handling, `--since`/`--until` |
| 4 | README with demo GIF, `go install` tested, run against SystemCraft repo |

---

## 14. Open Questions

- Go vs Python for v1? Go = better binary distribution + resume signal. Python = faster to build. **Lean Go unless stuck.**
- Should `git-heat author` use line count or commit count as the ownership metric? Default: line count (via blame), with `--by-commits` flag as override.
- Add a `--metric prometheus` output mode in v1.5 that emits Prometheus exposition format directly?
