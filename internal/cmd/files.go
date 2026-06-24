package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/Shashank0701-byte/git-heat/internal/git"
	"github.com/Shashank0701-byte/git-heat/internal/render"
	"github.com/Shashank0701-byte/git-heat/internal/score"
	"github.com/Shashank0701-byte/git-heat/internal/model"
)

const defaultBlameLimit = 50

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Heatmap of all files in the repo",
	Long:  `Show a heatmap of all tracked files, color-coded by recency and churn.`,
	RunE:  runFiles,
}

func init() {
	rootCmd.AddCommand(filesCmd)
}

func runFiles(cmd *cobra.Command, args []string) error {
	// Validate repository
	repoRoot, err := git.ValidateRepo(flagRepo)
	if err != nil {
		return err
	}

	if err := git.HasCommits(repoRoot); err != nil {
		return err
	}

	repoName := git.GetRepoName(repoRoot)
	since := parseDuration(flagSince)
	until := parseDuration(flagUntil)

	// Parse git log
	commits, err := git.ParseLog(repoRoot, since, until, flagAuthor)
	if err != nil {
		return err
	}

	if len(commits) == 0 {
		fmt.Fprintln(os.Stderr, "No commits found matching the criteria.")
		return nil
	}

	// Build file entries from commits
	fileMap := git.BuildFileEntries(commits)

	// Sort files by commit count to prioritize blame
	type fileSort struct {
		path  string
		entry *model.FileEntry
	}
	var sorted []fileSort
	for path, entry := range fileMap {
		sorted = append(sorted, fileSort{path, entry})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].entry.CommitCount90d > sorted[j].entry.CommitCount90d
	})

	// Run blame on top files for ownership data
	blameLimit := defaultBlameLimit
	if blameLimit > len(sorted) {
		blameLimit = len(sorted)
	}

	for i := 0; i < blameLimit; i++ {
		ownership, totalLines, err := git.ParseBlame(repoRoot, sorted[i].path)
		if err != nil {
			continue // skip files that can't be blamed
		}
		sorted[i].entry.LineOwnership = ownership
		sorted[i].entry.TotalLines = totalLines
	}

	if blameLimit < len(sorted) {
		fmt.Fprintf(os.Stderr, "Note: blame data limited to top %d files (use --top to adjust)\n\n", blameLimit)
	}

	// Calculate heat scores
	var results []model.FileHeat
	for _, fs := range sorted {
		hs := score.ScoreFile(*fs.entry)
		results = append(results, model.FileHeat{
			FileEntry: *fs.entry,
			HeatScore: hs,
		})
	}

	// Sort by heat
	sort.Slice(results, func(i, j int) bool {
		return results[i].Heat > results[j].Heat
	})

	// Apply --top limit
	if flagTop > 0 && flagTop < len(results) {
		results = results[:flagTop]
	}

	// Determine window days for display
	windowDays := 90
	if flagSince != "" {
		windowDays = estimateWindowDays(flagSince)
	}

	// Render
	if flagJSON {
		return render.RenderJSON(results, repoName, windowDays)
	}
	render.RenderFileHeatmap(results, repoName, windowDays, flagNoColor)
	return nil
}

// estimateWindowDays converts the since flag to an approximate day count.
func estimateWindowDays(since string) int {
	matches := naturalDurationRegex.FindStringSubmatch(since)
	if matches == nil {
		return 90 // default
	}

	num := 0
	fmt.Sscanf(matches[1], "%d", &num)

	switch matches[2] {
	case "d", "D":
		return num
	case "m", "M":
		return num * 30
	case "y", "Y":
		return num * 365
	}
	return 90
}
