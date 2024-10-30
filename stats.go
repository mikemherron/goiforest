package goiforest

import (
	"math"
)

func min(values []float64) float64 {
	min := math.Inf(1)
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func max(values []float64) float64 {
	max := math.Inf(-1)
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func kurtosis(values []float64) float64 {
	n := float64(len(values))
	if n < 4.0 {
		return 0
	}

	mean := mean(values)
	variance := variance(values)
	stdDev := math.Sqrt(variance)

	kurtosis := 0.0
	for _, v := range values {
		kurtosis += math.Pow((v-mean)/stdDev, 4)
	}

	kurtosis *= ((n * (n + 1)) / ((n - 1) * (n - 2) * (n - 3)))
	kurtosis -= (3 * (math.Pow((n - 1), 2))) / ((n - 2) * (n - 3))

	return kurtosis
}

func mean(values []float64) float64 {
	n := float64(len(values))
	if n == 0 {
		return 0
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= n

	return mean
}

func variance(values []float64) float64 {
	mean := mean(values)

	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values)) - 1

	return variance
}
