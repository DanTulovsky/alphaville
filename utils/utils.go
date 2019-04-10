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
	// return float64(rand.Int63n(int64(max-min)) + int64(min))
	return AffineTransform(rand.Float64(), 0, 1, min, max)
}

// AffineTransform x (in the range [a, b] to a number in [c, d]
func AffineTransform(x, a, b, c, d float64) float64 {
	// log.Printf("in: %v [%v, %v] -> [%v, %v]", x, a, b, c, d)
	if x < a {
		log.Print("invalid input into AffineTransform, returning min.")
		log.Printf("AffineTransform -> in: %v [%v, %v] -> [%v, %v]", x, a, b, c, d)
		return c
	}
	if x > b {
		log.Print("invalid input into AffineTransform, returning max.")
		log.Printf("AffineTransform -> in: %v [%v, %v] -> [%v, %v]", x, a, b, c, d)
		return d
	}
	return (x-a)*((d-c)/(b-a)) + c
}

// D2R converts degrees to radians
func D2R(d float64) float64 {
	return d * math.Pi / 180
}

// RotateRect returns a new rect rotated angle degrees
func RotateRect(r pixel.Rect, angle float64) pixel.Rect {

	theta := D2R(angle)
	center := r.Center()
	points := r.Vertices()

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
	if r1.Min.Y >= r2.Min.Y && r1.Min.Y <= r2.Max.Y {
		return true
	}
	if r1.Max.Y <= r2.Max.Y && r1.Max.Y >= r2.Min.Y {
		return true
	}

	if r2.Min.Y >= r1.Min.Y && r2.Min.Y <= r1.Max.Y {
		return true
	}
	if r2.Max.Y <= r1.Max.Y && r2.Max.Y >= r1.Min.Y {
		return true
	}
	return false
}

// LineSlope returns the slope of the line passing through a and b
func LineSlope(a, b pixel.Vec) float64 {
	return (b.Y - a.Y) / (b.X - a.X)
}

// RotatedAroundOrigin returns r moved to origin and rotated 180 degrees
func RotatedAroundOrigin(r pixel.Rect) pixel.Rect {
	ro := r.Moved(pixel.V(-r.Center().X, -r.Center().Y))
	return pixel.R(-ro.Min.X, -ro.Min.Y, -ro.Max.X, -ro.Max.Y).Norm()
}

// MinkowskiSum returns the minkowski sum of r1 and r2
func MinkowskiSum(r1, r2 pixel.Rect) pixel.Rect {
	return pixel.R(r1.Min.X+r2.Min.X, r1.Min.Y+r2.Min.Y, r1.Max.X+r2.Max.X, r1.Max.Y+r2.Max.Y)
}

// Intersect returns true if the two rectangles intersect
func Intersect(r1, r2 pixel.Rect) bool {
	return r1.Intersect(r2) != pixel.R(0, 0, 0, 0)
}

// IntersectAny returns true if r1 intersects any rect in r2
func IntersectAny(r1 pixel.Rect, r2 []pixel.Rect) bool {

	for _, f := range r2 {
		if r1 == f {
			continue // skip yourself
		}
		if Intersect(r1, f) {
			return true
		}
	}
	return false
}

//VecLen returns the length of the vector that is the distance between a and b
func VecLen(a, b pixel.Vec) float64 {
	return a.Sub(b).Len()
}
