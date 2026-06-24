package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/Shashank0701-byte/git-heat/internal/git"
	"github.com/Shashank0701-byte/git-heat/internal/model"
	"github.com/Shashank0701-byte/git-heat/internal/render"
)

var authorCmd = &cobra.Command{
	Use:   "author [name]",
	Short: "Show files owned by a specific author",
	Long: `Show which files and directories a specific author owns,
ranked by percentage of lines attributed to them via git blame.`,
	Args: cobra.ExactArgs(1),
	RunE: runAuthor,
}

func init() {
	rootCmd.AddCommand(authorCmd)
}

func runAuthor(cmd *cobra.Command, args []string) error {
	authorName := args[0]

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

	// Parse git log filtered by this author
	commits, err := git.ParseLog(repoRoot, since, until, authorName)
	if err != nil {
		return err
	}

	if len(commits) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: no commits found for author %q\n", authorName)
		return nil
	}

	// Collect unique files touched by this author
	fileSet := make(map[string]bool)
	fileCommitCount := make(map[string]int)
	fileLastCommit := make(map[string]model.Commit)

	for _, c := range commits {
		for _, path := range c.FilesChanged {
			fileSet[path] = true
			fileCommitCount[path]++
			if existing, ok := fileLastCommit[path]; !ok || c.Timestamp.After(existing.Timestamp) {
				fileLastCommit[path] = c
			}
		}
	}

	// Run blame on each file to get actual line ownership
	var results []model.AuthorFileOwnership
	blamedCount := 0
	blameLimit := defaultBlameLimit

	// Sort files by commit count for prioritization
	type fileRank struct {
		path  string
		count int
	}
	var ranked []fileRank
	for path := range fileSet {
		ranked = append(ranked, fileRank{path, fileCommitCount[path]})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].count > ranked[j].count
	})

	for _, fr := range ranked {
		if blamedCount >= blameLimit {
			break
		}

		ownership, totalLines, err := git.ParseBlame(repoRoot, fr.path)
		if err != nil || totalLines == 0 {
			continue
		}

		blamedCount++
		linesOwned := 0

		// Find this author's lines (case-insensitive partial match)
		for owner, count := range ownership {
			if containsIgnoreCase(owner, authorName) {
				linesOwned += count
			}
		}

		if linesOwned == 0 {
			continue
		}

		lc := fileLastCommit[fr.path]
		results = append(results, model.AuthorFileOwnership{
			Path:           fr.path,
			LinesOwned:     linesOwned,
			TotalLines:     totalLines,
			OwnershipPct:   float64(linesOwned) / float64(totalLines),
			CommitCount:    fileCommitCount[fr.path],
			LastCommitTime: lc.Timestamp,
		})
	}

	// Apply --top limit
	if flagTop > 0 && flagTop < len(results) {
		results = results[:flagTop]
	}

	render.RenderAuthorOwnership(authorName, results, repoName, flagNoColor)
	return nil
}

// containsIgnoreCase checks if s contains substr (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		len(substr) > 0 &&
		(s == substr ||
			len(s) > 0 && containsLower(toLower(s), toLower(substr)))
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
