// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	. "github.com/vsdbmv2/worker-go/epitopeMap"
	. "github.com/vsdbmv2/worker-go/needlemanWunsh"
	. "github.com/vsdbmv2/worker-go/smithWaterman"
)

// WorkType represents the type of mapping job
type WorkType string

const (
    GlobalMapping  WorkType = "global-mapping"
    LocalMapping   WorkType = "local-mapping"
    EpitopeMapping WorkType = "epitope-mapping"
)

// Work represents a single mapping job
type Work struct {
    Type       WorkType `json:"type"`
    ID1        int      `json:"id1"`
    Organism   string   `json:"organism"`
    Sequence1  string   `json:"sequence1"`
    Sequence2  interface{} `json:"sequence2"` // Can be string or []string for epitope mapping
    ID2        int      `json:"id2"`
    Identifier string   `json:"identifier"`
    IDSubtype  int      `json:"idSubtype,omitempty"`
}

// Result represents the mapping result
type Result struct {
    Organism     string     `json:"organism"`
    Type         WorkType   `json:"type"`
    Identifier   string     `json:"identifier"`
    MapInit      int       `json:"map_init,omitempty"`
    MapEnd       int       `json:"map_end,omitempty"`
    CoveragePct  float64   `json:"coverage_pct,omitempty"`
    IDSequence   int       `json:"idSequence,omitempty"`
    AlignmentScore int     `json:"alignment_score,omitempty"`
    IDSequenceSubtype int  `json:"idSequenceSubtype,omitempty"`
    IDSubtype    int       `json:"idSubtype,omitempty"`
    EpitopeMaps  []EpitopeMap `json:"epitope_maps,omitempty"`
}

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using default values")
    }

    // Get max concurrency
    maxConcurrency := runtime.NumCPU()
    if maxConcurrencyEnv := os.Getenv("maxConcurrency"); maxConcurrencyEnv != "" {
        if mc, err := strconv.Atoi(maxConcurrencyEnv); err == nil {
            maxConcurrency = mc
        }
    }

    // Get websocket host
    wsHost := "ws://vsdbm-api.hfabio.dev"
    if wsHostEnv := os.Getenv("websocketHost"); wsHostEnv != "" {
        wsHost = wsHostEnv
    }

    // Connect to WebSocket
    c, _, err := websocket.DefaultDialer.Dial(wsHost, nil)
    if err != nil {
        log.Fatal("WebSocket connection error:", err)
    }
    defer c.Close()

    log.Println("Connected to WebSocket server")

    // Channel for work results
    results := make(chan Result)
    var activeWorks int

    for {
        _, message, err := c.ReadMessage()
        if err != nil {
            log.Println("WebSocket read error:", err)
            break
        }

        var event struct {
            Type    string          `json:"type"`
            Payload json.RawMessage `json:"payload"`
        }

        if err := json.Unmarshal(message, &event); err != nil {
            log.Println("JSON parse error:", err)
            continue
        }

        switch event.Type {
        case "work":
            var works []Work
            if err := json.Unmarshal(event.Payload, &works); err != nil {
                log.Println("Work parse error:", err)
                continue
            }

            // Process works concurrently
            activeWorks = len(works)
            for _, work := range works {
                go processWork(work, results)
            }

        case "ping":
            if activeWorks == 0 {
                response := struct {
                    WorksAmount int `json:"worksAmount"`
                }{
                    WorksAmount: maxConcurrency,
                }
                c.WriteJSON(map[string]interface{}{
                    "type":    "get-work",
                    "payload": response,
                })
            }
        }

        // Collect results
        select {
        case result := <-results:
            activeWorks--
            if activeWorks == 0 {
                c.WriteJSON(map[string]interface{}{
                    "type":    "work-complete",
                    "payload": result,
                })
            }
        default:
        }
    }

    log.Println("Disconnected from WebSocket server")
}

func processGlobalMapping(work Work) Result {
  if sequence, ok := work.Sequence2.(string); ok {
    /* act on str */
    startPos, endPos, coverage, err := NeedlemanWunsch(work.Sequence1, sequence)

    if err != nil {
      fmt.Println(err)
    }

    return Result{
      Type: work.Type,
      Organism: work.Organism,
      MapInit: startPos,
      MapEnd: endPos,
      Identifier: work.Identifier,
      CoveragePct: coverage,
      IDSequence: work.ID2,
    }
  } else {
    panic("String expected in global mapping")
  }
}
func processLocalMapping(work Work) Result {
  if sequence, ok := work.Sequence2.(string); ok {
    /* act on str */
    score, err := SmithWaterman(work.Sequence1, sequence)
    if err != nil {
      fmt.Println(err)
    }

    return Result{
      Type: work.Type,
      Organism: work.Organism,
      Identifier: work.Identifier,
      AlignmentScore: score,
      IDSequence: work.ID1,
      IDSequenceSubtype: work.ID2,
      IDSubtype: work.IDSubtype,
    }
  } else {
    panic("String expected in local mapping")
  }
}
func processEpitopeMapping(work Work) Result {
  epitopes := make([]string, 0)
  switch work.Sequence2.(type) {
  case string:
    epitopes = append(epitopes, work.Sequence2.(string))
  case []string:
    epitopes = work.Sequence2.([]string)
  default:
    panic("Expected string or string array for epitopes")
  }
  mapped := SlideWindow(work.Sequence1, epitopes)

  return Result{
    Type: work.Type,
    Organism: work.Organism,
    Identifier: work.Identifier,
    IDSequence: work.ID1,
    IDSequenceSubtype: work.ID2,
    IDSubtype: work.IDSubtype,
    EpitopeMaps: mapped,
  }
}

func processWork(work Work, results chan<- Result) {
    var result Result
    
    switch work.Type {
    case GlobalMapping:
        result = processGlobalMapping(work)
    case LocalMapping:
        result = processLocalMapping(work)
    case EpitopeMapping:
        result = processEpitopeMapping(work)
    }

    results <- result
}
