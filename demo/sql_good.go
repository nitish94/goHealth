package main

import (
	"database/sql"
	"context"
)

func goodSQL(ctx context.Context, db *sql.DB, userID int) (*sql.Rows, error) {
	// Good: Parameterized query
	return db.QueryContext(ctx, "SELECT name FROM users WHERE id = $1", userID)
}