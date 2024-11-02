package smith_waterman

import (
	"errors"
	"math"
	"strings"
)

// computeSmithWaterman calculates the alignment score using the Smith-Waterman algorithm
func computeSmithWaterman(s1, s2 []string, ge, go_, mt, mst int, currentLine, lastLine []int) int {
	bestScore := 0
	var (
		similarity            int   // similarity between the chars (match or mismatch)
		lastDiag             int   // last score on diagonal
		firstPartialScoreUp  int   // partial up score 1
		secondPartialScoreUp int   // partial up score 2
		firstPartialScoreLeft int  // partial left score 1
		secondPartialScoreLeft int // partial left score 2
		leftScore            int   // left score
	)

	for i := 1; i <= len(s1); i++ {
		leftScore = math.MinInt32
		lastDiag = 0

		for j := 1; j <= len(s2); j++ {
			// Calculate similarity score
			if s1[i-1] == s2[j-1] {
				similarity = lastDiag + mt
			} else {
				similarity = lastDiag + mst
			}

			// Calculate partial up scores
			firstPartialScoreUp = lastLine[j] + ge
			secondPartialScoreUp = currentLine[j] + go_
			if firstPartialScoreUp > secondPartialScoreUp {
				lastLine[j] = firstPartialScoreUp
			} else {
				lastLine[j] = secondPartialScoreUp
			}

			// Calculate partial left scores
			firstPartialScoreLeft = leftScore + ge
			secondPartialScoreLeft = currentLine[j-1] + go_
			if firstPartialScoreLeft > secondPartialScoreLeft {
				leftScore = firstPartialScoreLeft
			} else {
				leftScore = secondPartialScoreLeft
			}

			// Save diagonal for next iteration
			lastDiag = currentLine[j]

			// Find maximum score
			currentLine[j] = max([]int{similarity, lastLine[j], leftScore, 0})

			// Update best score if necessary
			if bestScore <= currentLine[j] {
				bestScore = currentLine[j]
			}
		}
	}

	return bestScore
}

// ComputeLocalAlignment performs local sequence alignment between two sequences
func SmithWaterman(referenceSequence, querySequence string) (int, error) {
	// Validate input sequences
	if len(referenceSequence) == 0 {
		return 0, errors.New("empty reference sequence")
	}
	if len(querySequence) == 0 {
		return 0, errors.New("empty query sequence")
	}

	// Convert sequences to uppercase and split into slices
	reference := strings.Split(strings.ToUpper(referenceSequence), "")
	query := strings.Split(strings.ToUpper(querySequence), "")

	// Initialize parameters
	const (
		ge  = -1  // gap extension penalty
		go_ = -10 // gap opening penalty
		mt  = 5   // match score
		mst = -4  // mismatch score
	)

	// Initialize matrices
	lastLine := make([]int, len(query)+1)
	currentLine := make([]int, len(query)+1)

	// Compute optimal alignment
	score := computeSmithWaterman(reference, query, ge, go_, mt, mst, currentLine, lastLine)

	return score, nil
}

// max returns the maximum value from a slice of integers
func max(values []int) int {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}