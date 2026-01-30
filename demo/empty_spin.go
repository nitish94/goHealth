package main

import "time"

func badSpin() {
	// Bad: Empty spin loop
	for {
		select {
		default:
			// Do nothing, just spin
		}
	}
}

func goodSpin() {
	// Good: With timeout
	for {
		select {
		case <-time.After(time.Second):
			// Do something
		default:
			// Optional work
		}
	}
}