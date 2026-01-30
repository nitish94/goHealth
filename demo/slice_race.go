package main

import "sync"

func badRace() {
	var results []int
	for i := 0; i < 10; i++ {
		go func(i int) {
			results = append(results, i) // Race condition!
		}(i)
	}
}

func goodWithMutex() {
	var results []int
	var mu sync.Mutex
	for i := 0; i < 10; i++ {
		go func(i int) {
			mu.Lock()
			results = append(results, i)
			mu.Unlock()
		}(i)
	}
}