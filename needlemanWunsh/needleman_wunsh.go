package needleman_wunsh

import (
	"errors"
	"math"
)

// GlobalAlignment performs sequence alignment using Needleman-Gotoh algorithm
// Returns map_init, map_end, and coverage percentage
func NeedlemanWunsch(referenceSequence, sequenceToAlign string) (int, int, float64, error) {
	if len(referenceSequence) == 0 {
		return 0, 0, 0, errors.New("empty reference sequence")
	}
	if len(sequenceToAlign) == 0 {
		return 0, 0, 0, errors.New("empty query sequence")
	}

	// Constants for scoring
	const (
		Match     = 2  // match score
		MissMatch = -3 // mismatch penalty
		Gap       = 5  // gap opening penalty
		Ge        = 2  // gap extension penalty
	)

	// Swap sequences if reference is shorter
	if len(referenceSequence) < len(sequenceToAlign) {
		referenceSequence, sequenceToAlign = sequenceToAlign, referenceSequence
	}

	seqRefLength := len(referenceSequence) + 1
	seqAlignLength := len(sequenceToAlign) + 1
	sizeArray := seqRefLength * seqAlignLength

	pointers := make([]int8, sizeArray)
	lengths := make([]int8, sizeArray)

	result := process(referenceSequence, sequenceToAlign, pointers, lengths, Match, MissMatch, Gap, Ge)
	traceResult := traceBack(referenceSequence, sequenceToAlign, result.maxi, result.maxj, pointers, lengths)

	coverage := float64(traceResult.to-traceResult.from) * 100 / float64(len(referenceSequence))
	
	return traceResult.from, traceResult.to, math.Round(coverage*100)/100, nil
}

type processResult struct {
	maxi, maxj int
	score      float64
}

func process(rowString, columnString string, pointers, lengths []int8, M, Ms, G, Ge int) processResult {
	m := len(rowString) + 1
	n := len(columnString) + 1

	// Initialize boundaries
	for i, k := 1, n; i < m; i, k = i+1, k+n {
		pointers[k] = 3
		lengths[k] = int8(i)
	}
	for j := 1; j < n; j++ {
		pointers[j] = 1
		lengths[j] = int8(j)
	}

	v := make([]float64, n)
	vDiagonal := 0.0
	f := math.Inf(-1)
	h := math.Inf(-1)
	g := make([]float64, n)
	for i := range g {
		g[i] = math.Inf(-1)
	}

	lengthOfHorizontalGap := 0
	lengthOfVerticalGap := make([]int, n)

	maximumScore := math.Inf(-1)
	var maxi, maxj int

	// Fill matrices
	for i := 1; i < m; i++ {
		v[0] = float64(-G - (i-1)*Ge)
		k := i * n

		for j := 1; j < n; j++ {
			l := k + j
			similarityScore := Ms
			if rowString[i-1] == columnString[j-1] {
				similarityScore = M
			}

			f = vDiagonal + float64(similarityScore)

			// From left
			if h-float64(Ge) >= v[j-1]-float64(G) {
				h -= float64(Ge)
				lengthOfHorizontalGap++
			} else {
				h = v[j-1] - float64(G)
				lengthOfHorizontalGap = 1
			}

			// From above
			if g[j]-float64(Ge) >= v[j]-float64(G) {
				g[j] = g[j] - float64(Ge)
				lengthOfVerticalGap[j]++
			} else {
				g[j] = v[j] - float64(G)
				lengthOfVerticalGap[j] = 1
			}

			vDiagonal = v[j]
			v[j] = math.Max(f, math.Max(g[j], h))

			if v[j] > maximumScore {
				maximumScore = v[j]
				maxi = i
				maxj = j
			}

			// Set pointer and length
			if v[j] == f {
				pointers[l] = 2
			} else if v[j] == g[j] && v[j] != 0 {
				pointers[l] = 3
				lengths[l] = int8(lengthOfVerticalGap[j])
			} else if v[j] == h && v[j] != 0 {
				pointers[l] = 1
				lengths[l] = int8(lengthOfHorizontalGap)
			}
		}

		h = math.Inf(-1)
		vDiagonal = 0
		lengthOfHorizontalGap = 0
	}

	return processResult{maxi: maxi, maxj: maxj, score: v[n-1]}
}

type traceBackResult struct {
	from, to int
}

func traceBack(als1, als2 string, rowa, cola int, pointers, lengths []int8) traceBackResult {
	maxLength := len(als1) + len(als2)
	reversed1 := make([]rune, maxLength)
	reversed2 := make([]rune, maxLength)

	var len1, len2 int
	i := rowa
	j := cola
	n := len(als2) + 1
	row := i * n

	a := len(als1) - 1
	b := len(als2) - 1

	if a-i > b-j {
		for ; a-i > b-j; a-- {
			reversed1[len1] = rune(als1[a])
			reversed2[len2] = '-'
			len1++
			len2++
		}
		for ; b > j-1; a, b = a-1, b-1 {
			reversed1[len1] = rune(als1[a])
			reversed2[len2] = rune(als2[b])
			len1++
			len2++
		}
	} else {
		for ; b-j > a-i; b-- {
			reversed1[len1] = '-'
			reversed2[len2] = rune(als2[b])
			len1++
			len2++
		}
		for ; a > i-1; a, b = a-1, b-1 {
			reversed1[len1] = rune(als1[a])
			reversed2[len2] = rune(als2[b])
			len1++
			len2++
		}
	}

	stillGoing := true
	for stillGoing {
		l := row + j
		switch pointers[l] {
		case 3:
			for k := 0; k < int(lengths[l]); k++ {
				i--
				reversed1[len1] = rune(als1[i])
				reversed2[len2] = '-'
				len1++
				len2++
				row -= n
			}
			if lengths[l] <= 0 {
				row -= n
			}
		case 2:
			i--
			j--
			reversed1[len1] = rune(als1[i])
			reversed2[len2] = rune(als2[j])
			len1++
			len2++
			row -= n
		case 1:
			if lengths[l] < 0 && row <= 0 {
				stillGoing = false
			} else if lengths[l] < 0 {
				lengths[l] = 1
			}
			for k := 0; k < int(lengths[l]); k++ {
				j--
				reversed1[len1] = '-'
				reversed2[len2] = rune(als2[j])
				len1++
				len2++
			}
		case 0:
			stillGoing = false
		}
	}

	alignedSeq2 := reverseString(reversed2[:len2])
	return traceBackResult{
		from: getFrom(alignedSeq2),
		to:   getTo(string(reversed1[:len1]), alignedSeq2) +1,
	}
}

func getFrom(als2 string) int {
	position := 0
	for position < len(als2) && als2[position] == '-' {
		position++
	}
	return position
}

func getTo(als1, als2 string) int {
	position := len(als1) - 1
	for position > 0 && als2[position] == '-' {
		position--
	}
	return position
}

func reverseString(runes []rune) string {
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}