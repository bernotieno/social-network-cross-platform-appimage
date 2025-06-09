package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database connection
	db, err := sql.Open("sqlite3", "./social_network.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Check if image column already exists
	rows, err := db.Query("PRAGMA table_info(comments)")
	if err != nil {
		log.Fatal("Failed to get table info:", err)
	}
	defer rows.Close()

	hasImageColumn := false
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}

		if name == "image" {
			hasImageColumn = true
			break
		}
	}

	if hasImageColumn {
		fmt.Println("Image column already exists in comments table")
		return
	}

	// Add image column to comments table
	_, err = db.Exec("ALTER TABLE comments ADD COLUMN image TEXT")
	if err != nil {
		log.Fatal("Failed to add image column:", err)
	}

	fmt.Println("Successfully added image column to comments table")

	// Verify the column was added
	rows, err = db.Query("PRAGMA table_info(comments)")
	if err != nil {
		log.Fatal("Failed to verify table info:", err)
	}
	defer rows.Close()

	fmt.Println("Current comments table structure:")
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}

		fmt.Printf("Column: %s, Type: %s, NotNull: %d, PK: %d\n", name, dataType, notNull, pk)
	}
}
