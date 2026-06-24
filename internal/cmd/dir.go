package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Shashank0701-byte/git-heat/internal/git"
	"github.com/Shashank0701-byte/git-heat/internal/model"
	"github.com/Shashank0701-byte/git-heat/internal/render"
	"github.com/Shashank0701-byte/git-heat/internal/score"
)

var dirCmd = &cobra.Command{
	Use:   "dir [path]",
	Short: "Heatmap of subdirectories",
	Long:  `Show a heatmap aggregated at the directory level. Defaults to repo root.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDir,
}

func init() {
	rootCmd.AddCommand(dirCmd)
}

func runDir(cmd *cobra.Command, args []string) error {
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

	// Target directory (relative to repo root)
	targetDir := ""
	if len(args) > 0 {
		targetDir = filepath.Clean(args[0])
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

	// Aggregate by directory
	dirMap := make(map[string]*model.DirHeat)

	for path, entry := range fileMap {
		// Filter by target directory if specified
		if targetDir != "" && !strings.HasPrefix(path, targetDir+"/") && path != targetDir {
			continue
		}

		// Get the immediate directory
		dir := filepath.Dir(path)
		if targetDir != "" {
			// Show one level below target
			rel, err := filepath.Rel(targetDir, dir)
			if err != nil {
				continue
			}
			parts := strings.SplitN(filepath.ToSlash(rel), "/", 2)
			if parts[0] == "." {
				dir = targetDir
			} else {
				dir = filepath.Join(targetDir, parts[0])
			}
		} else {
			// Show top-level directories
			parts := strings.SplitN(filepath.ToSlash(path), "/", 2)
			if len(parts) > 1 {
				dir = parts[0]
			} else {
				dir = "."
			}
		}

		dh, exists := dirMap[dir]
		if !exists {
			dh = &model.DirHeat{
				Path:          dir,
				LineOwnership: make(map[string]int),
			}
			dirMap[dir] = dh
		}

		dh.FileCount++
		dh.TotalCommits += entry.CommitCount90d

		if entry.LastCommitTime.After(dh.LastCommitTime) {
			dh.LastCommitTime = entry.LastCommitTime
			dh.LastAuthor = entry.LastAuthor
		}

		for author, lines := range entry.LineOwnership {
			dh.LineOwnership[author] += lines
			dh.TotalLines += lines
		}
	}

	if len(dirMap) == 0 {
		fmt.Fprintln(os.Stderr, "No directories found matching the criteria.")
		return nil
	}

	// Score each directory
	var results []model.DirHeat
	for _, dh := range dirMap {
		// Create a synthetic FileEntry for scoring
		fe := model.FileEntry{
			LastCommitTime: dh.LastCommitTime,
			CommitCount90d: dh.TotalCommits / max(dh.FileCount, 1), // average per file
			LineOwnership:  dh.LineOwnership,
			TotalLines:     dh.TotalLines,
		}
		hs := score.ScoreFile(fe)
		// Override churn to use total commits for directory
		hs.ChurnScore = score.CalcChurn(dh.TotalCommits)
		hs.Heat = score.CalcHeat(hs.RecencyScore, hs.ChurnScore)
		dh.HeatScore = hs
		results = append(results, *dh)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Heat > results[j].Heat
	})

	if flagTop > 0 && flagTop < len(results) {
		results = results[:flagTop]
	}

	windowDays := 90
	if flagSince != "" {
		windowDays = estimateWindowDays(flagSince)
	}

	render.RenderDirHeatmap(results, repoName, windowDays, flagNoColor)
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
