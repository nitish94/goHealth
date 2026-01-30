package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func uglyHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Ugly: Multiple issues in one function

	// Unclosed body
	resp, _ := http.Get("http://example.com")

	// Bad SQL
	rows, _ := db.Query("SELECT * FROM users WHERE id = " + "123")

	// Silenced error
	data, _ := json.Marshal(struct{ Name string }{"test"})

	// Weak randomness
	token := rand.Int()

	// Time comparison
	t1 := time.Now()
	t2 := time.Now()
	if t1 == t2 {
		log.Println("Equal")
	}

	w.Write(data)
	_ = resp
	_ = rows
	_ = token
}