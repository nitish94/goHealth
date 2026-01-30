package main

import (
	"database/sql"
)

func getUsers(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return
	}
	// Missing defer rows.Close()

	for rows.Next() {
		// ...
	}
}
