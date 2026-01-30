package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting loop...")
	for i := 0; i < 10; i++ {
		fmt.Printf("Processing %d\n", i)
		time.Sleep(1 * time.Second) // Doctor should catch this!
	}
}
