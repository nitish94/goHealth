package main

import (
	"encoding/json"
	"net/http"
)

type User struct {
	Name string
}

func badMarshal(w http.ResponseWriter, user User) {
	// Bad: Ignoring marshal error
	data, _ := json.Marshal(user)
	w.Write(data) // Sends empty/corrupt data if marshal failed
}

func badDBExec() {
	// Bad: Ignoring marshal error again
	_, _ = json.Marshal(User{Name: "bad"})
}

func goodHandling(w http.ResponseWriter, user User) {
	// Good: Handle the error
	data, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(data)
}