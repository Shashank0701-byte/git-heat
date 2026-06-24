// Package render provides terminal and JSON output formatters for git-heat.
package render

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/Shashank0701-byte/git-heat/internal/model"
)

// Color palette for heat levels
var (
	hotStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")) // bright red
	warmStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")) // orange
	mildStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")) // yellow
	coldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4CAF50")) // green
	dimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")) // dim gray

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	legendStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	authorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88AAFF"))

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999"))
)

const barBlocks = 6

// heatBar generates a colored heat bar like "██████" or "███░░░"
func heatBar(heat float64) string {
	filled := int(math.Round(heat / 100.0 * float64(barBlocks)))
	if filled > barBlocks {
		filled = barBlocks
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barBlocks-filled)

	style := coldStyle
	switch {
	case heat >= 80:
		style = hotStyle
	case heat >= 60:
		style = warmStyle
	case heat >= 40:
		style = mildStyle
	}

	return style.Render(bar)
}

// timeAgo formats a duration since a timestamp as a human-readable string.
func timeAgo(t time.Time) string {
	d := time.Since(t)
	days := int(d.Hours() / 24)

	switch {
	case days == 0:
		return "today"
	case days == 1:
		return "1d ago"
	case days < 30:
		return fmt.Sprintf("%dd ago", days)
	case days < 365:
		months := days / 30
		return fmt.Sprintf("%dmo ago", months)
	default:
		years := days / 365
		return fmt.Sprintf("%dy ago", years)
	}
}

// truncatePath shortens a file path to fit in maxLen characters.
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "…" + path[len(path)-maxLen+1:]
}

// RenderFileHeatmap renders the heatmap table for a list of FileHeat entries.
func RenderFileHeatmap(files []model.FileHeat, repoName string, windowDays int, noColor bool) {
	if len(files) == 0 {
		fmt.Println(dimStyle.Render("No files found matching the criteria."))
		return
	}

	// Sort by heat descending
	sort.Slice(files, func(i, j int) bool {
		return files[i].Heat > files[j].Heat
	})

	// Header
	header := fmt.Sprintf("git-heat files — repo: %s  (last %d days)", repoName, windowDays)
	if noColor {
		fmt.Println(header)
	} else {
		fmt.Println(headerStyle.Render(header))
	}
	fmt.Println()

	// Calculate column widths
	maxPathLen := 40
	maxAuthorLen := 15

	for _, f := range files {
		bar := heatBar(f.Heat)
		path := truncatePath(f.Path, maxPathLen)
		author := f.LastAuthor
		if len(author) > maxAuthorLen {
			author = author[:maxAuthorLen-1] + "…"
		}
		ago := timeAgo(f.LastCommitTime)

		if noColor {
			fmt.Printf("  %-6s  %-*s  %-*s  %8s  %3d commits\n",
				heatBarPlain(f.Heat), maxPathLen, path, maxAuthorLen, author, ago, f.CommitCount90d)
		} else {
			fmt.Printf("  %s  %s  %s  %s  %s\n",
				bar,
				pathStyle.Render(fmt.Sprintf("%-*s", maxPathLen, path)),
				authorStyle.Render(fmt.Sprintf("%-*s", maxAuthorLen, author)),
				metaStyle.Render(fmt.Sprintf("%8s", ago)),
				metaStyle.Render(fmt.Sprintf("%3d commits", f.CommitCount90d)),
			)
		}
	}

	// Legend
	fmt.Println()
	if noColor {
		fmt.Println("Legend:  ██ Hot   █░ Warm   ░░ Cold")
	} else {
		legend := fmt.Sprintf("Legend:  %s Hot   %s Warm   %s Mild   %s Cold",
			hotStyle.Render("██"),
			warmStyle.Render("█░"),
			mildStyle.Render("░░"),
			coldStyle.Render("░░"),
		)
		fmt.Println(legendStyle.Render(legend))
	}
}

// heatBarPlain generates an uncolored heat bar for --no-color mode.
func heatBarPlain(heat float64) string {
	filled := int(math.Round(heat / 100.0 * float64(barBlocks)))
	if filled > barBlocks {
		filled = barBlocks
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("#", filled) + strings.Repeat(".", barBlocks-filled)
}

// RenderDirHeatmap renders a directory-level heatmap.
func RenderDirHeatmap(dirs []model.DirHeat, repoName string, windowDays int, noColor bool) {
	if len(dirs) == 0 {
		fmt.Println(dimStyle.Render("No directories found matching the criteria."))
		return
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Heat > dirs[j].Heat
	})

	header := fmt.Sprintf("git-heat dir — repo: %s  (last %d days)", repoName, windowDays)
	if noColor {
		fmt.Println(header)
	} else {
		fmt.Println(headerStyle.Render(header))
	}
	fmt.Println()

	maxPathLen := 40
	maxAuthorLen := 15

	for _, d := range dirs {
		bar := heatBar(d.Heat)
		path := truncatePath(d.Path+"/", maxPathLen)
		author := d.LastAuthor
		if len(author) > maxAuthorLen {
			author = author[:maxAuthorLen-1] + "…"
		}
		ago := timeAgo(d.LastCommitTime)

		if noColor {
			fmt.Printf("  %-6s  %-*s  %-*s  %8s  %3d files\n",
				heatBarPlain(d.Heat), maxPathLen, path, maxAuthorLen, author, ago, d.FileCount)
		} else {
			fmt.Printf("  %s  %s  %s  %s  %s\n",
				bar,
				pathStyle.Render(fmt.Sprintf("%-*s", maxPathLen, path)),
				authorStyle.Render(fmt.Sprintf("%-*s", maxAuthorLen, author)),
				metaStyle.Render(fmt.Sprintf("%8s", ago)),
				metaStyle.Render(fmt.Sprintf("%3d files", d.FileCount)),
			)
		}
	}

	fmt.Println()
	if noColor {
		fmt.Println("Legend:  ## Hot   #. Warm   .. Cold")
	} else {
		legend := fmt.Sprintf("Legend:  %s Hot   %s Warm   %s Mild   %s Cold",
			hotStyle.Render("██"),
			warmStyle.Render("█░"),
			mildStyle.Render("░░"),
			coldStyle.Render("░░"),
		)
		fmt.Println(legendStyle.Render(legend))
	}
}

// RenderAuthorOwnership renders the author ownership view.
func RenderAuthorOwnership(authorName string, files []model.AuthorFileOwnership, repoName string, noColor bool) {
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: no commits found for author %q\n", authorName)
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].OwnershipPct > files[j].OwnershipPct
	})

	header := fmt.Sprintf("git-heat author — %s in %s", authorName, repoName)
	if noColor {
		fmt.Println(header)
	} else {
		fmt.Println(headerStyle.Render(header))
	}
	fmt.Println()

	maxPathLen := 45

	for _, f := range files {
		path := truncatePath(f.Path, maxPathLen)
		pct := f.OwnershipPct * 100
		pctBar := ownershipBar(pct)

		if noColor {
			fmt.Printf("  %5.1f%%  %-*s  %4d/%4d lines  %s\n",
				pct, maxPathLen, path, f.LinesOwned, f.TotalLines, timeAgo(f.LastCommitTime))
		} else {
			fmt.Printf("  %s  %s  %s  %s\n",
				pctBar,
				pathStyle.Render(fmt.Sprintf("%-*s", maxPathLen, path)),
				metaStyle.Render(fmt.Sprintf("%4d/%4d lines", f.LinesOwned, f.TotalLines)),
				metaStyle.Render(timeAgo(f.LastCommitTime)),
			)
		}
	}
}

// ownershipBar renders a percentage as a colored bar.
func ownershipBar(pct float64) string {
	bar := fmt.Sprintf("%5.1f%%", pct)
	switch {
	case pct >= 80:
		return hotStyle.Render(bar)
	case pct >= 50:
		return warmStyle.Render(bar)
	case pct >= 20:
		return mildStyle.Render(bar)
	default:
		return coldStyle.Render(bar)
	}
}

// RenderSummary renders the summary overview.
func RenderSummary(report model.SummaryReport, noColor bool) {
	header := fmt.Sprintf("git-heat summary — repo: %s  (last %d days)", report.RepoName, report.WindowDays)
	if noColor {
		fmt.Println(header)
	} else {
		fmt.Println(headerStyle.Render(header))
	}
	fmt.Println()

	// Stats
	statsLines := []string{
		fmt.Sprintf("  Total files:       %d", report.TotalFiles),
		fmt.Sprintf("  Total authors:     %d", report.TotalAuthors),
		fmt.Sprintf("  Total commits:     %d", report.TotalCommits),
		fmt.Sprintf("  Most active:       %s", report.MostActiveAuthor),
		fmt.Sprintf("  Avg concentration: %.2f", report.AvgConcentration),
	}

	for _, line := range statsLines {
		if noColor {
			fmt.Println(line)
		} else {
			fmt.Println(metaStyle.Render(line))
		}
	}

	fmt.Println()

	// Top hottest files
	topLabel := fmt.Sprintf("  Top %d hottest files:", len(report.TopFiles))
	if noColor {
		fmt.Println(topLabel)
	} else {
		fmt.Println(authorStyle.Render(topLabel))
	}
	fmt.Println()

	for i, f := range report.TopFiles {
		bar := heatBar(f.Heat)
		path := truncatePath(f.Path, 40)

		if noColor {
			fmt.Printf("  %2d. %-6s  %-40s  %s  %3d commits\n",
				i+1, heatBarPlain(f.Heat), path, timeAgo(f.LastCommitTime), f.CommitCount90d)
		} else {
			rank := metaStyle.Render(fmt.Sprintf("%2d.", i+1))
			fmt.Printf("  %s %s  %s  %s  %s\n",
				rank,
				bar,
				pathStyle.Render(fmt.Sprintf("%-40s", path)),
				metaStyle.Render(fmt.Sprintf("%8s", timeAgo(f.LastCommitTime))),
				metaStyle.Render(fmt.Sprintf("%3d commits", f.CommitCount90d)),
			)
		}
	}
}
