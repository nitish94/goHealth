package main

import (
	"net/http"
)

func fetch(url string) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	// Missing defer resp.Body.Close()

	// Do something
	_ = resp.Status
}
