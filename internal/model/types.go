// Package model defines the core data types used across git-heat.
package model

import "time"

// Commit represents a parsed git commit with its metadata and changed files.
type Commit struct {
	SHA          string
	Author       string
	Email        string
	Timestamp    time.Time
	Message      string
	FilesChanged []string
}

// FileEntry holds aggregated data about a single file's contribution history.
type FileEntry struct {
	Path           string
	LastCommitSHA  string
	LastAuthor     string
	LastCommitTime time.Time
	CommitCount90d int
	LineOwnership  map[string]int // author → line count
	TotalLines     int
}

// HeatScore contains the calculated heat metrics for a file or directory.
type HeatScore struct {
	RecencyScore           float64 // 0–100: how recently the file was touched
	ChurnScore             float64 // 0–100: how frequently it was modified
	Heat                   float64 // weighted composite score
	OwnershipConcentration float64 // 0–1: Herfindahl index (1 = single owner)
}

// FileHeat combines a file's metadata with its computed heat score.
type FileHeat struct {
	FileEntry
	HeatScore
}

// DirHeat represents aggregated heat data for a directory.
type DirHeat struct {
	Path           string
	FileCount      int
	TotalCommits   int
	LastCommitTime time.Time
	LastAuthor     string
	LineOwnership  map[string]int
	TotalLines     int
	HeatScore
}

// AuthorStat holds summary statistics for a single author.
type AuthorStat struct {
	Name        string
	Email       string
	CommitCount int
}

// AuthorFileOwnership represents an author's ownership of a specific file.
type AuthorFileOwnership struct {
	Path            string
	LinesOwned      int
	TotalLines      int
	OwnershipPct    float64
	CommitCount     int
	LastCommitTime  time.Time
}

// SummaryReport holds the data for the `summary` command output.
type SummaryReport struct {
	RepoName         string
	WindowDays       int
	TotalFiles       int
	TotalAuthors     int
	TotalCommits     int
	MostActiveAuthor string
	TopFiles         []FileHeat
	AvgConcentration float64
}
