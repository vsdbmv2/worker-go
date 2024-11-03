package main

import (
	"fmt"

	. "github.com/vsdbmv2/worker-go/db"
	. "github.com/vsdbmv2/worker-go/smithWaterman"
)

func main2() {
	viruses := GetViruses()

	var targetVirus Virus

	for _, virus := range viruses {
		if virus.Id != 41 { continue }
		targetVirus = virus
	}
	fmt.Printf("Virus: \"%s\" on db \"%s\"\n", targetVirus.Name, targetVirus.Database_name)
	sequences := GetSequences(targetVirus.Database_name)
	subtypes := GetSequenceSubtypes(targetVirus.Database_name)
	alreadyMapped := GetLocalMapSet(targetVirus.Database_name)

	alignments := make([]LocalMapped, 0)
	for i := 0; i < len(sequences); i++ {
		for j := 0; j < len(subtypes); j++ {
			score, err := SmithWaterman(subtypes[i].Sequence, sequences[i].Sequence)
			if err != nil {
				fmt.Printf("Error mapping sequences %d and %d", subtypes[i].IdSequence, sequences[i].Id)
			}
			if alreadyMapped[fmt.Sprintf("%d-%d", sequences[i].Id, subtypes[i].IdSequence)] || sequences[i].Id == subtypes[i].IdSequence{
				continue
			}
			alignments = append(alignments, LocalMapped{
				IdSequence: sequences[i].Id,
				IdSubtypeSequence: subtypes[i].IdSequence,
				IdSubtype: subtypes[i].IdSubtype,
				Score: score,
			})
		}
	}

	BulkInsertLocalAlignments(targetVirus.Database_name, alignments)
}