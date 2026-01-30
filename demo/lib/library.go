package lib

import (
	"log"
	"os"
)

func badExit() {
	// Bad: Library calling os.Exit
	os.Exit(1)
}

func badFatal() {
	// Bad: Library calling log.Fatal
	log.Fatal("Error in library")
}

func badPanic() {
	// Bad: Library panicking
	panic("Something went wrong")
}

func goodFunction() error {
	// Good: Return error instead
	return nil
}