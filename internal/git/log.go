package git

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Shashank0701-byte/git-heat/internal/model"
)

const logSeparator = "|"

// ParseLog runs `git log` with the specified filters and returns parsed commits.
// It uses a custom format to extract SHA, author email, timestamp, and subject,
// along with the list of changed files per commit.
func ParseLog(repoRoot string, since, until, author string) ([]model.Commit, error) {
	args := []string{
		"log",
		"--format=" + "%H" + logSeparator + "%ae" + logSeparator + "%at" + logSeparator + "%an" + logSeparator + "%s",
		"--name-only",
	}

	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}
	if author != "" {
		args = append(args, "--author="+author)
	}

	out, err := RunGit(repoRoot, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse git log: %w", err)
	}

	return parseLogOutput(out)
}

// parseLogOutput parses the raw output of git log into Commit structs.
// The format alternates between commit header lines and file name lines,
// separated by blank lines.
func parseLogOutput(raw string) ([]model.Commit, error) {
	var commits []model.Commit
	var current *model.Commit

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines — git log --name-only inserts blank lines
		// between the commit header and the file list, so we cannot
		// use blank lines as commit boundaries.
		if line == "" {
			continue
		}

		// Try to parse as a commit header line (contains our separator)
		parts := strings.SplitN(line, logSeparator, 5)
		if len(parts) == 5 && len(parts[0]) == 40 {
			// Finalize previous commit before starting a new one
			if current != nil && current.SHA != "" {
				commits = append(commits, *current)
			}

			ts, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid timestamp %q in commit %s: %w", parts[2], parts[0], err)
			}

			current = &model.Commit{
				SHA:       parts[0],
				Email:     parts[1],
				Timestamp: time.Unix(ts, 0),
				Author:    parts[3],
				Message:   parts[4],
			}
			continue
		}

		// Otherwise it's a file path belonging to the current commit
		if current != nil {
			current.FilesChanged = append(current.FilesChanged, line)
		}
	}

	// Don't forget the last commit
	if current != nil && current.SHA != "" {
		commits = append(commits, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning git log output: %w", err)
	}

	return commits, nil
}

// BuildFileEntries processes a list of commits into a map of FileEntry structs.
// It aggregates commit counts (within the last 90 days) and tracks the most
// recent commit metadata per file.
func BuildFileEntries(commits []model.Commit) map[string]*model.FileEntry {
	files := make(map[string]*model.FileEntry)
	now := time.Now()
	ninetyDaysAgo := now.AddDate(0, 0, -90)

	for _, c := range commits {
		for _, path := range c.FilesChanged {
			entry, exists := files[path]
			if !exists {
				entry = &model.FileEntry{
					Path:          path,
					LineOwnership: make(map[string]int),
				}
				files[path] = entry
			}

			// Track the most recent commit for this file
			if c.Timestamp.After(entry.LastCommitTime) {
				entry.LastCommitTime = c.Timestamp
				entry.LastCommitSHA = c.SHA
				entry.LastAuthor = c.Author
			}

			// Count commits within the 90-day window
			if c.Timestamp.After(ninetyDaysAgo) {
				entry.CommitCount90d++
			}
		}
	}

	return files
}
