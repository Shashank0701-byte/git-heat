package render

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Shashank0701-byte/git-heat/internal/model"
)

// jsonOutput is the top-level JSON structure matching the PRD spec.
type jsonOutput struct {
	GeneratedAt string     `json:"generated_at"`
	Repo        string     `json:"repo"`
	WindowDays  int        `json:"window_days"`
	Files       []jsonFile `json:"files"`
}

// jsonFile represents a single file entry in JSON output.
type jsonFile struct {
	Path                   string             `json:"path"`
	HeatScore              float64            `json:"heat_score"`
	RecencyScore           float64            `json:"recency_score"`
	ChurnScore             float64            `json:"churn_score"`
	LastAuthor             string             `json:"last_author"`
	DaysSinceLastCommit    int                `json:"days_since_last_commit"`
	CommitCount            int                `json:"commit_count"`
	OwnershipConcentration float64            `json:"ownership_concentration"`
	Ownership              map[string]float64 `json:"ownership"`
}

// RenderJSON outputs the file heatmap data as structured JSON to stdout.
func RenderJSON(files []model.FileHeat, repoName string, windowDays int) error {
	out := jsonOutput{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Repo:        repoName,
		WindowDays:  windowDays,
		Files:       make([]jsonFile, 0, len(files)),
	}

	for _, f := range files {
		ownership := make(map[string]float64)
		if f.TotalLines > 0 {
			for author, lines := range f.LineOwnership {
				ownership[author] = float64(lines) / float64(f.TotalLines)
			}
		}

		daysSince := int(time.Since(f.LastCommitTime).Hours() / 24)
		if daysSince < 0 {
			daysSince = 0
		}

		out.Files = append(out.Files, jsonFile{
			Path:                   f.Path,
			HeatScore:              f.Heat,
			RecencyScore:           f.RecencyScore,
			ChurnScore:             f.ChurnScore,
			LastAuthor:             f.LastAuthor,
			DaysSinceLastCommit:    daysSince,
			CommitCount:            f.CommitCount90d,
			OwnershipConcentration: f.OwnershipConcentration,
			Ownership:              ownership,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(out); err != nil {
		return fmt.Errorf("failed to encode JSON output: %w", err)
	}

	return nil
}

// jsonSummary is the JSON structure for the summary command.
type jsonSummary struct {
	GeneratedAt      string     `json:"generated_at"`
	Repo             string     `json:"repo"`
	WindowDays       int        `json:"window_days"`
	TotalFiles       int        `json:"total_files"`
	TotalAuthors     int        `json:"total_authors"`
	TotalCommits     int        `json:"total_commits"`
	MostActiveAuthor string     `json:"most_active_author"`
	AvgConcentration float64    `json:"avg_ownership_concentration"`
	TopFiles         []jsonFile `json:"top_files"`
}

// RenderSummaryJSON outputs the summary report as structured JSON.
func RenderSummaryJSON(report model.SummaryReport) error {
	out := jsonSummary{
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		Repo:             report.RepoName,
		WindowDays:       report.WindowDays,
		TotalFiles:       report.TotalFiles,
		TotalAuthors:     report.TotalAuthors,
		TotalCommits:     report.TotalCommits,
		MostActiveAuthor: report.MostActiveAuthor,
		AvgConcentration: report.AvgConcentration,
		TopFiles:         make([]jsonFile, 0, len(report.TopFiles)),
	}

	for _, f := range report.TopFiles {
		ownership := make(map[string]float64)
		if f.TotalLines > 0 {
			for author, lines := range f.LineOwnership {
				ownership[author] = float64(lines) / float64(f.TotalLines)
			}
		}

		daysSince := int(time.Since(f.LastCommitTime).Hours() / 24)
		if daysSince < 0 {
			daysSince = 0
		}

		out.TopFiles = append(out.TopFiles, jsonFile{
			Path:                   f.Path,
			HeatScore:              f.Heat,
			RecencyScore:           f.RecencyScore,
			ChurnScore:             f.ChurnScore,
			LastAuthor:             f.LastAuthor,
			DaysSinceLastCommit:    daysSince,
			CommitCount:            f.CommitCount90d,
			OwnershipConcentration: f.OwnershipConcentration,
			Ownership:              ownership,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(out); err != nil {
		return fmt.Errorf("failed to encode summary JSON: %w", err)
	}

	return nil
}
