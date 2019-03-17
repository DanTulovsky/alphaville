package utils

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
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

// D2R converts degrees to radians
func D2R(d float64) float64 {
	return d * math.Pi / 180
}

// RotateRect returns a new rect rotated angle degrees
func RotateRect(r pixel.Rect, angle float64) pixel.Rect {

	theta := D2R(angle)

	center := r.Center()

	points := []pixel.Vec{
		pixel.V(r.Min.X, r.Min.Y),
		pixel.V(r.Min.X, r.Max.Y),
		pixel.V(r.Max.X, r.Min.Y),
		pixel.V(r.Max.X, r.Max.Y),
	}

	var minX, minY float64 = math.MaxFloat64, math.MaxFloat64
	var maxX, maxY float64

	for _, p := range points {
		x := center.X + (p.X-center.X)*math.Cos(theta) - (p.Y-center.Y)*math.Sin(theta)
		y := center.Y - (p.X-center.X)*math.Sin(theta) + (p.Y-center.Y)*math.Cos(theta)

		if x < minX {
			minX = x
		}

		if x > maxX {
			maxX = x
		}

		if y < minY {
			minY = y
		}

		if y > maxY {
			maxY = y
		}
	}

	return pixel.R(minX, minY, maxX, maxY)
}
