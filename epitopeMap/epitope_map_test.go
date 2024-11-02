// src/epitope_map_test.go
package epitope_map

import (
	"reflect"
	"testing"
)

func TestSlideWindow(t *testing.T) {
    tests := []struct {
        sequence string
        epitopes []string
        want     []EpitopeMap
    }{
        {
            sequence: "ATGGCTAGCT",
            epitopes: []string{"GCT", "AGC"},
            want: []EpitopeMap{
                {LinearSequence: "GCT", InitPos: 3},
                {LinearSequence: "GCT", InitPos: 7},
                {LinearSequence: "AGC", InitPos: 6},
            },
        },
    }

    for _, tt := range tests {
        got := SlideWindow(tt.sequence, tt.epitopes)
        if !reflect.DeepEqual(got, tt.want) {
            t.Errorf("SlideWindow(%v, %v) = %v, want %v",
                tt.sequence, tt.epitopes, got, tt.want)
        }
    }
}