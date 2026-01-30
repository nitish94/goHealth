package main

import (
	"time"
)

func badTimeCompare(t1, t2 time.Time) bool {
	// Bad: Direct comparison
	return t1 == t2
}

func goodTimeCompare(t1, t2 time.Time) bool {
	// Good: Use Equal
	return t1.Equal(t2)
}