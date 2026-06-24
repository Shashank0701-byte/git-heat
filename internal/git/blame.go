package git

import (
	"bufio"
	"fmt"
	"strings"
)

// ParseBlame runs `git blame --porcelain` on a file and returns line ownership
// as a map of author name → line count, plus the total line count.
// Returns an empty map for binary files or files that can't be blamed.
func ParseBlame(repoRoot, filePath string) (map[string]int, int, error) {
	out, err := RunGit(repoRoot, "blame", "--porcelain", filePath)
	if err != nil {
		// Binary files or deleted files will fail — return empty gracefully
		return make(map[string]int), 0, nil
	}

	return parseBlameOutput(out)
}

// parseBlameOutput parses git blame --porcelain output.
//
// The porcelain format outputs blocks like:
//
//	<sha> <orig-line> <final-line> <num-lines>
//	author <name>
//	author-mail <email>
//	author-time <timestamp>
//	...
//	filename <path>
//	\t<line content>
//
// For subsequent lines from the same commit, the header is shorter:
//
//	<sha> <orig-line> <final-line>
//	\t<line content>
//
// We track author names from the full headers and count the content lines.
func parseBlameOutput(raw string) (map[string]int, int, error) {
	ownership := make(map[string]int)
	totalLines := 0

	// Map SHA → author name (since subsequent blocks omit author info)
	shaToAuthor := make(map[string]string)
	currentSHA := ""
	currentAuthor := ""

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		// Content lines start with a tab
		if strings.HasPrefix(line, "\t") {
			totalLines++
			author := currentAuthor
			if author == "" {
				author = shaToAuthor[currentSHA]
			}
			if author != "" {
				ownership[author]++
			}
			continue
		}

		// Header line: SHA + line numbers
		fields := strings.Fields(line)
		if len(fields) >= 3 && len(fields[0]) == 40 {
			currentSHA = fields[0]
			currentAuthor = shaToAuthor[currentSHA] // may be empty if first occurrence
			continue
		}

		// Author metadata line
		if strings.HasPrefix(line, "author ") {
			currentAuthor = strings.TrimPrefix(line, "author ")
			if currentSHA != "" {
				shaToAuthor[currentSHA] = currentAuthor
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("error scanning blame output: %w", err)
	}

	return ownership, totalLines, nil
}
