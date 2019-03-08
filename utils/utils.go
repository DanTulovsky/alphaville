package utils

import (
	"log"
	"math/rand"
	"time"
)

// Elapsed is used to time function execution
func Elapsed(what string) func() {
	start := time.Now()
	return func() {
		log.Printf("%s took %v\n", what, time.Since(start))
	}
}

// RandomInt returns a random number in [min, max)
func RandomInt(min, max int) int {
	return rand.Intn(max-min) + min
}

// RandomFloat64 returns a random number in [min, max)
func RandomFloat64(min, max float64) float64 {
	return float64(rand.Int63n(int64(max-min)) + int64(min))
}
