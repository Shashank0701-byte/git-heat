package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/Shashank0701-byte/git-heat/internal/git"
	"github.com/Shashank0701-byte/git-heat/internal/model"
	"github.com/Shashank0701-byte/git-heat/internal/render"
	"github.com/Shashank0701-byte/git-heat/internal/score"
)

const defaultSummaryTop = 10

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Quick overview of the repo's hottest files and ownership",
	Long: `Show a summary including the top 10 hottest files (most commits
in the last 90 days), overall ownership concentration, and contributor stats.`,
	RunE: runSummary,
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}

func runSummary(cmd *cobra.Command, args []string) error {
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

	// Get author stats
	authors, err := git.ListAuthors(repoRoot)
	if err != nil {
		return err
	}

	// Parse git log
	commits, err := git.ParseLog(repoRoot, since, until, flagAuthor)
	if err != nil {
		return err
	}

	if len(commits) == 0 {
		fmt.Fprintln(os.Stderr, "No commits found matching the criteria.")
		return nil
	}

	// Build file entries
	fileMap := git.BuildFileEntries(commits)

	// Sort by commit count to get hottest files
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

	// Determine how many top files to show
	topN := defaultSummaryTop
	if flagTop > 0 {
		topN = flagTop
	}
	if topN > len(sorted) {
		topN = len(sorted)
	}

	// Run blame on top files
	for i := 0; i < topN; i++ {
		ownership, totalLines, err := git.ParseBlame(repoRoot, sorted[i].path)
		if err != nil {
			continue
		}
		sorted[i].entry.LineOwnership = ownership
		sorted[i].entry.TotalLines = totalLines
	}

	// Score top files
	var topFiles []model.FileHeat
	totalConcentration := 0.0
	scoredCount := 0

	for i := 0; i < topN; i++ {
		hs := score.ScoreFile(*sorted[i].entry)
		topFiles = append(topFiles, model.FileHeat{
			FileEntry: *sorted[i].entry,
			HeatScore: hs,
		})
		totalConcentration += hs.OwnershipConcentration
		scoredCount++
	}

	// Sort top files by heat
	sort.Slice(topFiles, func(i, j int) bool {
		return topFiles[i].Heat > topFiles[j].Heat
	})

	// Find most active author
	mostActive := "unknown"
	totalCommits := 0
	if len(authors) > 0 {
		mostActive = authors[0].Name
		for _, a := range authors {
			totalCommits += a.CommitCount
		}
	}

	avgConcentration := 0.0
	if scoredCount > 0 {
		avgConcentration = totalConcentration / float64(scoredCount)
	}

	windowDays := 90
	if flagSince != "" {
		windowDays = estimateWindowDays(flagSince)
	}

	report := model.SummaryReport{
		RepoName:         repoName,
		WindowDays:       windowDays,
		TotalFiles:       len(fileMap),
		TotalAuthors:     len(authors),
		TotalCommits:     totalCommits,
		MostActiveAuthor: mostActive,
		TopFiles:         topFiles,
		AvgConcentration: avgConcentration,
	}

	if flagJSON {
		return render.RenderSummaryJSON(report)
	}
	render.RenderSummary(report, flagNoColor)
	return nil
}
