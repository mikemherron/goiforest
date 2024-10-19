package main

import (
	"fmt"
	"math"
)

// Harmonic number function
func harmonicNumber(n int) float64 {
	var sum float64 = 0
	for i := 1; i <= n; i++ {
		sum += 1.0 / float64(i)
	}
	return sum
}

// Calculate expected average path length
func expectedAveragePathLength(sampleSize int) float64 {
	return 2*harmonicNumber(sampleSize-1) - (2*float64(sampleSize-1))/float64(sampleSize)
}

func main() {
	sampleSize := 512
	expectedAverage := expectedAveragePathLength(sampleSize)
	maxDepth := math.Ceil(math.Log2(float64(sampleSize)))

	fmt.Printf("Expected average path length: %f\n", expectedAverage)
	fmt.Printf("Maximum possible depth: %f\n", maxDepth)

	if expectedAverage > maxDepth {
		fmt.Println("Error: Expected average path length exceeds maximum depth!")
	} else {
		fmt.Println("Expected average path length is within the valid range.")
	}
}
