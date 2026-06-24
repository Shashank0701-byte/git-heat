#  git-heat

> Terminal-native Git contribution heatmap — see who owns what, instantly.

`git-heat` is a CLI tool that visualizes **code ownership** and **contribution patterns** directly in your terminal. It answers: _"Who actually wrote what in this codebase, and when?"_ — without leaving your terminal.

-  **Offline & fast** — reads your local `.git/` directory, no API keys or uploads
-  **Color-coded heatmaps** — ANSI 256-color terminal output
-  **Multiple views** — files, directories, per-author ownership, summary
-  **Machine-readable** — `--json` output for piping to other tools

---

## Quick Start

```bash
# Install globally (requires Go 1.21+)
go install github.com/Shashank0701-byte/git-heat/cmd/git-heat@latest

# Run from inside any git repo
cd your-project/
git-heat summary
git-heat files --top 10
git-heat author "YourName"
```

That's it — one command to install, works on any repo, fully offline.

### Prerequisites

- **Go 1.21+** — [install Go](https://go.dev/dl/)
- **Git** — must be available in your `PATH`

### Build from Source

```bash
git clone https://github.com/Shashank0701-byte/git-heat.git
cd git-heat
go build ./cmd/git-heat/
```

---

## Usage

Run from inside any git repo, or use `--repo <path>` to point at one:

### `files` — File-level heatmap

```bash
git-heat files                    # all files
git-heat files --top 20           # top 20 hottest
git-heat files --since 30d        # last 30 days only
```

```
git-heat files — repo: SystemCraft  (last 90 days)

  ██████  .github/workflows/ci.yml        Shashank       1d ago   23 commits
  ██████  components/canvas/Design...     coderabbit…    6d ago   22 commits
  █████░  package-lock.json               Shashank       1d ago   14 commits
  █████░  app/interview/[id]/page.tsx     Shashank      1mo ago   13 commits
  ████░░  components/canvas/Whiteb...     Shashank       7d ago    7 commits
  ███░░░  src/lib/ai/geminiClient.ts      Shashank       3d ago    6 commits

Legend:  ██ Hot   █░ Warm   ░░ Mild   ░░ Cold
```

### `dir` — Directory-level heatmap

```bash
git-heat dir                      # top-level directories
git-heat dir src/                 # subdirectories of src/
```

```
git-heat dir — repo: SystemCraft  (last 90 days)

  ██████  app/              Shashank       1d ago   35 files
  ██████  components/       coderabbit…    6d ago   32 files
  █████░  src/              Shashank       1d ago   35 files
  █████░  helm/             Shashank       1d ago   15 files
  ██░░░░  public/           Shashank      4mo ago    7 files
```

### `author` — Ownership by author

```bash
git-heat author "Shashank"        # files owned by Shashank
git-heat author "Shashank" --top 5
```

```
git-heat author — Shashank in SystemCraft

  100.0%  .github/workflows/ci.yml          225/ 225 lines  1d ago
  100.0%  components/AuthCard.tsx            407/ 407 lines  3mo ago
   99.8%  app/interview/[id]/page.tsx        470/ 471 lines  1mo ago
   98.0%  components/canvas/DesignCanvas.tsx 1339/1366 lines  6d ago
   51.1%  package.json                        23/  45 lines  1d ago
```

### `summary` — Quick overview

```bash
git-heat summary                  # terminal output
git-heat summary --json           # machine-readable JSON
```

```
git-heat summary — repo: SystemCraft  (last 90 days)

  Total files:       186
  Total authors:     5
  Total commits:     329
  Most active:       Shashank
  Avg concentration: 0.79

  Top 10 hottest files:

   1. ██████  .github/workflows/ci.yml          1d ago   23 commits
   2. ██████  components/canvas/DesignCanvas...  6d ago   22 commits
   3. █████░  package-lock.json                  1d ago   14 commits
```

---

## Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--repo <path>` | Path to git repo (default: `.`) | `--repo ~/projects/myapp` |
| `--since <dur>` | Show commits after this duration | `--since 30d`, `--since 6m` |
| `--until <dur>` | Show commits before this duration | `--until 1y` |
| `--author <name>` | Filter by author name | `--author "Jane"` |
| `--top <N>` | Limit output rows | `--top 20` |
| `--json` | Output as structured JSON | `git-heat summary --json` |
| `--no-color` | Plain text (no ANSI colors) | For piping to files |

Duration format: `30d` (days), `6m` (months), `1y` (years)

---

## How It Works

`git-heat` reads your local git history using plumbing commands:

1. **`git log`** — extracts commits, timestamps, authors, and changed files
2. **`git blame --porcelain`** — determines line-level ownership per file
3. **`git shortlog -sne`** — lists all contributors

### Heat Score Formula

```
recency_score = 100 × (1 - days_since_last_commit / 365)
churn_score   = min(100, commit_count_90d × 5)
heat          = 0.6 × recency + 0.4 × churn
```

| Heat Score | Color | Meaning |
|------------|-------|---------|
| ≥ 80 | 🔴 Red | Hot — actively changing |
| 60–79 | 🟠 Orange | Warm — recent activity |
| 40–59 | 🟡 Yellow | Mild — moderate activity |
| < 40 | 🟢 Green | Cold — stable/stale |

**Ownership Concentration** uses the Herfindahl-Hirschman Index (HHI):
- `1.0` = single author owns everything (risk!)
- `< 0.5` = well-distributed ownership

---

## JSON Output

The `--json` flag produces structured output for piping to other tools:

```json
{
  "generated_at": "2026-06-25T10:00:00Z",
  "repo": "SystemCraft",
  "window_days": 90,
  "files": [
    {
      "path": "src/server/index.ts",
      "heat_score": 94.5,
      "last_author": "Shashank",
      "days_since_last_commit": 3,
      "commit_count": 42,
      "ownership": { "Shashank": 0.87, "Riya": 0.13 }
    }
  ]
}
```

---

## Tech Stack

| Layer | Choice | Why |
|-------|--------|-----|
| Language | Go | Single binary, fast startup, excellent subprocess handling |
| CLI framework | [cobra](https://github.com/spf13/cobra) | Industry-standard Go CLI framework |
| Terminal styling | [lipgloss v2](https://github.com/charmbracelet/lipgloss) | Best-in-class terminal rendering |
| Git data | `git log`, `git blame`, `git shortlog` | No libgit2 dependency; uses the git you already have |

---

## Performance

| Repo size | Target | Method |
|-----------|--------|--------|
| < 1,000 commits | < 500ms | Full analysis |
| 1,000–10,000 commits | < 3s | Blame limited to top 50 files |
| > 10,000 commits | Warning shown | Use `--since 90d` to limit scope |

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

1. Fork the repo
2. Create your feature branch (`git checkout -b feat/awesome-feature`)
3. Commit your changes (`git commit -m 'feat: add awesome feature'`)
4. Push to the branch (`git push origin feat/awesome-feature`)
5. Open a Pull Request

---

Built by [Shashank](https://github.com/Shashank0701-byte) — a portfolio project demonstrating git internals, Go CLI development, and terminal UI design.
