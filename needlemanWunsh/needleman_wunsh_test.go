// src/needleman_wunsh_test.go
package needleman_wunsh

import (
	"errors"
	"testing"
)

func TestNeedlemanWunsch(t *testing.T) {
    tests := []struct {
        seq1     string
        seq2     string
        wantInit int
        wantEnd  int
        wantCoverage float64
        expectedError error
    }{
        {
            seq1:     "AACGATATTGCU",
            seq2:     "AACGTGCU",
            wantInit: 0,
            wantEnd:  8,
            wantCoverage: 66.67,
            expectedError: nil,
        },
        {
            seq1:     "GATTACAA",
            seq2:     "GTCGACG",
            wantInit: 0,
            wantEnd:  7,
            wantCoverage: 87.5,
            expectedError: nil,
        },
        {
            seq1:     "GTCGACG",
            seq2:     "GTCGACG",
            wantInit: 0,
            wantEnd:  7,
            wantCoverage: 100,
            expectedError: nil,
        },
        {
            seq1:     "",
            seq2:     "GTCGACG",
            wantInit: 0,
            wantEnd:  7,
            wantCoverage: 100,
            expectedError: errors.New("empty reference sequence"),
        },
        {
            seq1:     "GTCGACG",
            seq2:     "",
            wantInit: 0,
            wantEnd:  7,
            wantCoverage: 100,
            expectedError: errors.New("empty query sequence"),
        },
    }

    for _, tt := range tests {
        init, end, coverage, error := NeedlemanWunsch(tt.seq1, tt.seq2)
        if error != nil && tt.expectedError == nil {
            t.Errorf("NeedlemanWunsch unexpected error (%v, %v) = (%v, %v, %v), want (%v, %v, %v)",
            tt.seq1, tt.seq2, init, end, coverage, tt.wantInit, tt.wantEnd, tt.wantCoverage)
        }else if tt.expectedError == error {
            return
        }
        if init != tt.wantInit || end != tt.wantEnd || coverage != tt.wantCoverage {
            t.Errorf("NeedlemanWunsch(%v, %v) = (%v, %v, %v), want (%v, %v, %v)",
                tt.seq1, tt.seq2, init, end, coverage, tt.wantInit, tt.wantEnd, tt.wantCoverage)
        }
        if coverage <= 0 || coverage > 100 {
            t.Errorf("Coverage percentage out of range: %v", coverage)
        }
    }
}