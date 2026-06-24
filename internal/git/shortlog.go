package git

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/Shashank0701-byte/git-heat/internal/model"
)

// ListAuthors runs `git shortlog -sne` to get all authors with commit counts.
func ListAuthors(repoRoot string) ([]model.AuthorStat, error) {
	out, err := RunGit(repoRoot, "shortlog", "-sne", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to list authors: %w", err)
	}

	return parseShortlogOutput(out)
}

// parseShortlogOutput parses the output of `git shortlog -sne`.
// Each line has the format:  "  <count>\t<name> <email>"
func parseShortlogOutput(raw string) ([]model.AuthorStat, error) {
	var authors []model.AuthorStat

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Split on tab: "  123\tJohn Doe <john@example.com>"
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}

		count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}

		nameEmail := strings.TrimSpace(parts[1])
		name, email := parseNameEmail(nameEmail)

		authors = append(authors, model.AuthorStat{
			Name:        name,
			Email:       email,
			CommitCount: count,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning shortlog output: %w", err)
	}

	return authors, nil
}

// parseNameEmail splits "John Doe <john@example.com>" into name and email.
func parseNameEmail(s string) (string, string) {
	idx := strings.LastIndex(s, " <")
	if idx == -1 {
		return s, ""
	}

	name := s[:idx]
	email := strings.TrimSuffix(strings.TrimPrefix(s[idx+2:], "<"), ">")
	return name, email
}
