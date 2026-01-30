package main

import (
	"math/rand"
)

func generateToken() string {
	// Bad: Using math/rand for token generation
	return string(rand.Intn(1000000))
}

func createPassword() string {
	// Bad: Predictable password generation
	return "pass" + string(rand.Int31())
}

// Comment mentioning security
// This function handles session tokens
func handleSessionToken() {
	token := rand.Int63() // Bad
	_ = token
}