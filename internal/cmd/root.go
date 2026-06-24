// Package cmd implements the cobra CLI commands for git-heat.
package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
)

// Global flags shared across all commands.
var (
	flagRepo    string
	flagSince   string
	flagUntil   string
	flagAuthor  string
	flagNoColor bool
	flagJSON    bool
	flagTop     int
)

// rootCmd is the base command for git-heat.
var rootCmd = &cobra.Command{
	Use:   "git-heat",
	Short: "Terminal Git Contribution Heatmap",
	Long: `git-heat visualizes code ownership and contribution patterns
directly in your terminal. It answers: "Who actually wrote what
in this codebase, and when?" — without leaving your terminal.

Run on any local git repo, fully offline, zero signup.`,
}

// Execute runs the root command. Called from main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagRepo, "repo", ".", "Path to the git repository")
	rootCmd.PersistentFlags().StringVar(&flagSince, "since", "", "Show commits after date (e.g. 30d, 6m, 1y)")
	rootCmd.PersistentFlags().StringVar(&flagUntil, "until", "", "Show commits before date (e.g. 30d, 6m, 1y)")
	rootCmd.PersistentFlags().StringVar(&flagAuthor, "author", "", "Filter by author name")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().IntVar(&flagTop, "top", 0, "Limit output to top N entries")
}

// naturalDurationRegex matches patterns like "30d", "6m", "1y".
var naturalDurationRegex = regexp.MustCompile(`^(\d+)([dDmMyY])$`)

// parseDuration converts a natural duration string (30d, 6m, 1y) to
// a format git understands ("30 days ago", "6 months ago", "1 year ago").
// If the input doesn't match, it's passed through as-is (for git-native dates).
func parseDuration(s string) string {
	if s == "" {
		return ""
	}

	matches := naturalDurationRegex.FindStringSubmatch(s)
	if matches == nil {
		// Not our format — pass through to git (could be ISO date, etc.)
		return s
	}

	num, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch unit {
	case "d", "D":
		return fmt.Sprintf("%d days ago", num)
	case "m", "M":
		return fmt.Sprintf("%d months ago", num)
	case "y", "Y":
		return fmt.Sprintf("%d years ago", num)
	default:
		return s
	}
}
