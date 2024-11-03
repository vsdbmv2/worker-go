package main

import (
	"fmt"
	"runtime"
	"sync"

	. "github.com/vsdbmv2/worker-go/db"
	. "github.com/vsdbmv2/worker-go/smithWaterman"
)

const (
	BATCH_SIZE = 100000
)

type job struct {
	sequence Sequence
	subtype  SequenceSubtype
}

type result struct {
	mapped LocalMapped
	err    error
}

func main() {
	viruses := GetViruses()

	var targetVirus Virus
	for _, virus := range viruses {
		if virus.Id != 41 {
			continue
		}
		targetVirus = virus
	}

	fmt.Printf("Virus: \"%s\" on db \"%s\"\n", targetVirus.Name, targetVirus.Database_name)
	sequences := GetSequences(targetVirus.Database_name)
	subtypes := GetSequenceSubtypes(targetVirus.Database_name)
	alreadyMapped := GetLocalMapSet(targetVirus.Database_name)

	// Calculate total expected alignments
	totalJobs := 0
	for i := 0; i < len(sequences); i++ {
		for j := 0; j < len(subtypes); j++ {
			if alreadyMapped[fmt.Sprintf("%d-%d", sequences[i].Id, subtypes[j].IdSequence)] || 
			   sequences[i].Id == subtypes[j].IdSequence {
				continue
			}
			totalJobs++
		}
	}

	fmt.Printf("Processing %d alignments in batches of %d\n", totalJobs, BATCH_SIZE)

	// Create channels for jobs and results
	numWorkers := runtime.NumCPU()
	jobs := make(chan job, numWorkers*2) // Buffer size = 2x number of workers
	results := make(chan result, numWorkers*2)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	// Start job producer in a goroutine
	go func() {
		for i := 0; i < len(sequences); i++ {
			for j := 0; j < len(subtypes); j++ {
				if alreadyMapped[fmt.Sprintf("%d-%d", sequences[i].Id, subtypes[j].IdSequence)] || sequences[i].Id == subtypes[j].IdSequence {
					continue
				}
				jobs <- job{
					sequence: sequences[i],
					subtype:  subtypes[j],
				}
			}
		}
		close(jobs)
	}()

	// Start a goroutine to close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and process results in batches
	alignments := make([]LocalMapped, 0, BATCH_SIZE)
	processedCount := 0
	batchCount := 0

	for r := range results {
		if r.err != nil {
			fmt.Printf("Error mapping sequences %d and %d: %v\n",
				r.mapped.IdSequence, r.mapped.IdSubtypeSequence, r.err)
			continue
		}

		alignments = append(alignments, r.mapped)
		processedCount++

		// When we reach batch size or process all jobs, bulk insert
		if len(alignments) >= BATCH_SIZE || processedCount == totalJobs {
			if len(alignments) > 0 {
				batchCount++
				fmt.Printf("Inserting batch %d (%d alignments, total processed: %d/%d)\n", 
					batchCount, len(alignments), processedCount, totalJobs)
				
				BulkInsertLocalAlignments(targetVirus.Database_name, alignments)
				
				// Clear the alignments slice while preserving capacity
				alignments = alignments[:0]
			}
		}
	}

	fmt.Printf("Completed processing %d alignments in %d batches\n", processedCount, batchCount)
}

func worker(jobs <-chan job, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		score, err := SmithWaterman(j.subtype.Sequence, j.sequence.Sequence)
		results <- result{
			mapped: LocalMapped{
				IdSequence:        j.sequence.Id,
				IdSubtypeSequence: j.subtype.IdSequence,
				IdSubtype:        j.subtype.IdSubtype,
				Score:            score,
			},
			err: err,
		}
	}
}