package utils

import (
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
	"unicode"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
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

// LoadTTF loads a true type font at path
func LoadTTF(path string, size float64) (font.Face, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	font, err := truetype.Parse(bytes)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(font, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	}), nil
}

// Atoi converts string to integer, errors crash.
func Atoi(a string) (i int) {
	var err error
	if i, err = strconv.Atoi(a); err != nil {
		log.Fatalf("%v is not a valid int", a)
	}
	return i
}

var atlas *text.Atlas
var once sync.Once

// Atlas returns a singleton of the atlas; contains the font to use for text
func Atlas() *text.Atlas {

	once.Do(func() {
		face, err := LoadTTF("fonts/intuitive.ttf", 16)
		if err != nil {
			log.Fatal(err)
		}
		atlas = text.NewAtlas(face, text.ASCII, text.RangeTable(unicode.Sm))
	})

	return atlas
}

// HaveCommonY returns true if r1 and r2 have a common Y value
func HaveCommonY(r1, r2 pixel.Rect) bool {
	if r1.Min.Y > r2.Min.Y && r1.Min.Y < r2.Max.Y {
		return true
	}
	if r1.Max.Y < r2.Max.Y && r1.Max.Y > r2.Min.Y {
		return true
	}

	if r2.Min.Y > r1.Min.Y && r2.Min.Y < r1.Max.Y {
		return true
	}
	if r2.Max.Y < r1.Max.Y && r2.Max.Y > r1.Min.Y {
		return true
	}
	return false
}
