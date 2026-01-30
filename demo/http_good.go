package main

import (
	"net/http"
	"time"
)

func goodHTTPClient() {
	// Good: Client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	_ = client
}