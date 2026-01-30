package main

import (
	"context"
	"database/sql"
)

func badTransaction(ctx context.Context, db *sql.DB) error {
	// Bad: Begin transaction without defer rollback
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Do some work
	_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "John")
	if err != nil {
		return err // Transaction not rolled back!
	}

	return tx.Commit()
}

func goodTransaction(ctx context.Context, db *sql.DB) error {
	// Good: Defer rollback immediately
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Safe rollback

	// Do some work
	_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "Jane")
	if err != nil {
		return err // Rollback will happen
	}

	return tx.Commit() // Rollback ignored on success
}

func badBegin(ctx context.Context, db *sql.DB) error {
	// Bad: Using old Begin without defer
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Work...
	_, err = tx.Exec("SELECT 1")
	return err // No rollback
}