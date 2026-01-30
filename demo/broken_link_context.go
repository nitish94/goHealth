package main

import (
	"context"
	"database/sql"
)

func badDatabaseCall(ctx context.Context, db *sql.DB) error {
	// Bad: Using context.Background() instead of ctx
	return db.QueryRowContext(context.Background(), "SELECT 1").Scan()
}

func badWithTODO(ctx context.Context) {
	// Bad: Using context.TODO()
	context.WithValue(context.TODO(), "key", "value")
}

func goodCall(ctx context.Context, db *sql.DB) error {
	// Good: Using the passed ctx
	return db.QueryRowContext(ctx, "SELECT 1").Scan()
}

func functionWithoutContext(db *sql.DB) error {
	// This should not be flagged: no ctx available
	return db.QueryRowContext(context.Background(), "SELECT 1").Scan()
}