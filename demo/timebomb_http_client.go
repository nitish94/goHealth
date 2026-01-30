package main

import (
	"net/http"
	"time"
)

func badGet() {
	// This should be flagged: http.Get uses default client with no timeout
	resp, err := http.Get("http://example.com")
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func badPost() {
	// This should be flagged: http.Post uses default client with no timeout
	resp, err := http.Post("http://example.com", "application/json", nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func badClientLiteral() {
	// This should be flagged: http.Client{} without Timeout
	client := http.Client{}
	client.Get("http://example.com")
}

func badClientPointer() {
	// This should be flagged: &http.Client{} without Timeout
	client := &http.Client{}
	client.Get("http://example.com")
}

func goodClient() {
	// This should NOT be flagged: has Timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	client.Get("http://example.com")
}