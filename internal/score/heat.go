// Package score implements the heat score calculation from the PRD.
package score

import (
	"math"
	"time"

	"github.com/Shashank0701-byte/git-heat/internal/model"
)

const (
	recencyWeight = 0.6
	churnWeight   = 0.4
	recencyWindow = 365.0 // days
	churnCap      = 100.0
	churnMultiple = 5.0
)

// CalcRecency computes a recency score (0–100) based on how recently
// the file was last modified. A file touched today scores 100;
// a file untouched for 365+ days scores 0.
func CalcRecency(lastCommit time.Time) float64 {
	daysSince := time.Since(lastCommit).Hours() / 24.0
	score := 100.0 * (1.0 - daysSince/recencyWindow)
	return math.Max(0, score)
}

// CalcChurn computes a churn score (0–100) based on the number of
// commits in the last 90 days. Each commit contributes 5 points, capped at 100.
func CalcChurn(commitCount90d int) float64 {
	return math.Min(churnCap, float64(commitCount90d)*churnMultiple)
}

// CalcHeat computes the weighted composite heat score.
//   heat = 0.6 × recency + 0.4 × churn
func CalcHeat(recency, churn float64) float64 {
	return recencyWeight*recency + churnWeight*churn
}

// CalcOwnershipConcentration computes the Herfindahl-Hirschman Index (HHI)
// of line ownership. A value of 1.0 means a single author owns all lines;
// values closer to 0 indicate distributed ownership.
func CalcOwnershipConcentration(ownership map[string]int) float64 {
	if len(ownership) == 0 {
		return 0
	}

	total := 0
	for _, count := range ownership {
		total += count
	}

	if total == 0 {
		return 0
	}

	hhi := 0.0
	for _, count := range ownership {
		share := float64(count) / float64(total)
		hhi += share * share
	}

	return hhi
}

// ScoreFile computes all heat metrics for a single file entry.
func ScoreFile(entry model.FileEntry) model.HeatScore {
	recency := CalcRecency(entry.LastCommitTime)
	churn := CalcChurn(entry.CommitCount90d)
	heat := CalcHeat(recency, churn)
	concentration := CalcOwnershipConcentration(entry.LineOwnership)

	return model.HeatScore{
		RecencyScore:           math.Round(recency*100) / 100,
		ChurnScore:             math.Round(churn*100) / 100,
		Heat:                   math.Round(heat*100) / 100,
		OwnershipConcentration: math.Round(concentration*1000) / 1000,
	}
}
