package xmath

import (
	"math"
	"math/rand"
	"time"
)

// Rand generates a random integer
func Rand(min ...int) int {
	if len(min) == 0 {
		return rand.Intn(math.MaxInt32)
	}
	if len(min) == 1 {
		return rand.Intn(min[0])
	}
	return min[0] + rand.Intn(min[1]-min[0]+1)
}

// MtRand generates a random value via the Mersenne Twister
func MtRand(min ...int) int {
	return Rand(min...)
}

// Srand seeds the random number generator
func Srand(seed ...int64) {
	if len(seed) > 0 {
		rand.Seed(seed[0])
	} else {
		rand.Seed(time.Now().UnixNano())
	}
}

// MtSrand seeds the Mersenne Twister
func MtSrand(seed ...int64) {
	Srand(seed...)
}

// Getrandmax returns the largest possible random value
func Getrandmax() int {
	return math.MaxInt32
}

// MtGetrandmax returns the largest possible random value
func MtGetrandmax() int {
	return math.MaxInt32
}

// LcgValue returns a pseudo random number
func LcgValue() float64 {
	return rand.Float64()
}

// RandFloat generates a random float between 0 and 1
func RandFloat() float64 {
	return rand.Float64()
}

// RandInt generates a random integer between min and max (inclusive)
func RandInt(min, max int) int {
	if min > max {
		min, max = max, min
	}
	return min + rand.Intn(max-min+1)
}

// RandFloatRange generates a random float between min and max
func RandFloatRange(min, max float64) float64 {
	if min > max {
		min, max = max, min
	}
	return min + rand.Float64()*(max-min)
}