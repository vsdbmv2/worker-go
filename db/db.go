package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/vsdbmv2/worker-go/smithWaterman"
)

type Virus struct {
	Id int
	Name string
	Database_name string
}

type Sequence struct {
	Id int
	Sequence string
}
type SequenceSubtype struct {
	IdSequence int
	IdSubtype int
	Sequence string
}

func executeQuery(dbName, query string) (*sql.Rows, error) {
	// Connection string format: username:password@protocol(host:port)/dbname
	connectionString := fmt.Sprintf("root:220311@tcp(localhost:3306)/%s", dbName)

	// Open database connection
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(query)
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	db.Close()
	return rows, nil
}

func GetViruses() []Virus{
		// Example usage
		dbName := "vsdbmv2"
		query := "SELECT id, name, database_name FROM virus"
	
		fmt.Printf("Reading database\n")
		rows, err := executeQuery(dbName, query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		fmt.Printf("Read finished\n")
		viruses := make([]Virus, 0)
		// Process the results
		fmt.Printf("Iterating over rows\n")
		for rows.Next() {
			// Example: scanning results into variables
			var virus Virus
			err := rows.Scan(&virus.Id, &virus.Name, &virus.Database_name)
			if err != nil {
				log.Fatal(err)
				continue
			}
			viruses = append(viruses, virus)
			// fmt.Printf("ID: %d, Name: %s, db_name: %s\n", virus.Id, virus.Name, virus.Database_name)
		}
		fmt.Printf("Iteration finished\n")
	
		if err = rows.Err(); err != nil {
			log.Fatal(err)
		}
		return viruses
}

func GetSequences(database string) []Sequence {
	fmt.Printf("Reading sequences\n")
	query := "SELECT id, sequence FROM sequence WHERE country like '%brazil%';"
	sequences := make([]Sequence, 0)

	rows, err := executeQuery(database, query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var sequence Sequence
		err := rows.Scan(&sequence.Id, &sequence.Sequence)
		if err != nil {
			log.Fatal(err)
			continue
		}
		sequences = append(sequences, sequence)
	}
	fmt.Printf("Got %d sequences\n", len(sequences))
	return sequences
}

func GetSequenceSubtypes(database string) []SequenceSubtype {
	fmt.Printf("Reading subtypes\n")
	query := "SELECT s.id as idsequence, s.sequence, sub.id FROM subtype_reference_sequence srs "
	query += "JOIN subtype sub on sub.id = srs.idsubtype "
	query += "JOIN sequence s on s.id = srs.idsequence "
	query += "WHERE srs.is_refseq IS NOT NULL;"

	sequences := make([]SequenceSubtype, 0)

	rows, err := executeQuery(database, query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var sequence SequenceSubtype
		err := rows.Scan(&sequence.IdSequence, &sequence.Sequence, &sequence.IdSubtype)
		if err != nil {
			log.Fatal(err)
			continue
		}
		sequences = append(sequences, sequence)
	}
	fmt.Printf("Got %d sequence subtypes\n", len(sequences))
	return sequences
}

func GetLocalMapSet(database string) map[string]bool {
	fmt.Printf("Reading already mapped\n")
	query := "SELECT idsequence, idsequencesubtype FROM subtype_reference_sequence WHERE is_refseq IS NULL OR is_refseq = false"
	const t = true
	mapped := map[string]bool{}

	rows, err := executeQuery(database, query)
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	defer rows.Close()
	for rows.Next() {
		var idsequence int
		var idsubtypesequence int
		err := rows.Scan(&idsequence, &idsubtypesequence)
		if err != nil {
			log.Fatal(err)
			continue
		}
		mapped[fmt.Sprintf("%d-%d", idsequence, idsubtypesequence)] = t
		count++
	}
	fmt.Printf("Got %d mappings\n", count)
	return mapped
}

func BulkInsertLocalAlignments(database string, mapped []LocalMapped){
	fmt.Printf("Bulk inserting %d maps\n", len(mapped))
	if len(mapped) == 0 {
		return
	}
	query := "INSERT INTO subtype_reference_sequence (idsequence, idsubtype, idsequencesubtype, subtype_score) VALUES "

	for i := 0; i < len(mapped); i++ {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("(%d, %d, %d, %d)", mapped[i].IdSequence, mapped[i].IdSubtype, mapped[i].IdSubtypeSequence, mapped[i].Score)
	}
	query += ";"
	_, err := executeQuery(database, query)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Bulk inserted %d maps\n", len(mapped))
}