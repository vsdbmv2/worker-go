package epitope_map

type EpitopeMap struct {
    LinearSequence string `json:"linearSequence"`
    InitPos       int    `json:"init_pos"`
}


// SlideWindow implements the sliding window algorithm for epitope mapping
func SlideWindow(sequence string, epitopes []string) []EpitopeMap {
    var maps []EpitopeMap

    for _, epitope := range epitopes {
        epitopeLen := len(epitope)
        seqLen := len(sequence)

        for i := 0; i <= seqLen-epitopeLen; i++ {
            window := sequence[i:i+epitopeLen]
            if window == epitope {
                maps = append(maps, EpitopeMap{
                    LinearSequence: epitope,
                    InitPos:       i,
                })
            }
        }
    }

    return maps
}

// Helper function
func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}