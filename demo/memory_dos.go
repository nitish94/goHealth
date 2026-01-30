package main

import (
	"io"
	"net/http"
)

func badReadAll(w http.ResponseWriter, r *http.Request) {
	// Bad: Unlimited read into memory
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(data)
}

func goodReadAll(w http.ResponseWriter, r *http.Request) {
	// Good: Limited read
	data, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024)) // 1MB limit
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(data)
}