package main

import (
	"database/sql"
	"fmt"
)

func queryUser(db *sql.DB, name string) {
	// Bad: Sprintf
	query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)
	db.Exec(query)

	// Bad: Concatenation
	db.Query("SELECT * FROM users WHERE name = '" + name + "'")
}
