package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	// Open the database
	db, err := sql.Open("sqlite", "./ecommerce.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Read the migration file
	content, err := os.ReadFile("apply_migration.sql")
	if err != nil {
		log.Fatalf("Error reading migration file: %v", err)
	}

	// Split the file into individual statements
	statements := strings.Split(string(content), ";")

	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Skip SELECT statements in the verification section
		if strings.HasPrefix(strings.ToUpper(stmt), "SELECT") {
			fmt.Println("Executing:", strings.SplitN(stmt, "\n", 2)[0], "...")
			rows, err := db.Query(stmt)
			if err != nil {
				log.Printf("Warning: %v", err)
				continue
			}
			defer rows.Close()

			// Print results
			cols, _ := rows.Columns()
			if len(cols) > 0 {
				for rows.Next() {
					values := make([]interface{}, len(cols))
					for i := range values {
						var s sql.NullString
						values[i] = &s
					}
					if err := rows.Scan(values...); err != nil {
						log.Printf("Error scanning row: %v", err)
						continue
					}
					for i, col := range cols {
						val := values[i].(*sql.NullString)
						if val.Valid {
							fmt.Printf("%s: %s\n", col, val.String)
						} else {
							fmt.Printf("%s: NULL\n", col)
						}
					}
				}
			}
		} else {
			// Execute other statements
			fmt.Println("Executing:", strings.SplitN(stmt, "\n", 2)[0], "...")
			_, err := db.Exec(stmt)
			if err != nil {
				log.Printf("Error executing statement: %v", err)
			}
		}
	}

	fmt.Println("Migration completed successfully!")
}
