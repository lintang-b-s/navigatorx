package alg

import (
	"time"

	"golang.org/x/exp/rand"
)

func GenerateInitialRoute(numCities int) []int {
	res := rand.Perm(numCities)
	for i := 0; i < len(res); i++ {
		if res[i] == 0 {
			res = append(res[:i], res[i+1:]...)
			break
		}
	}
	return res
}

func SolveTSPTwoOpt(disanceMat [][]float64, maxIteration int) ([]int, float64) {
	rand.Seed(uint64(time.Now().Unix()))
	bestDistance := 0.0
	bestRoute := []int{}
	iteration := 0
	for iteration < maxIteration {
		num_cities := len(disanceMat)
		initialRoute := []int{}
		initialRoute = append(initialRoute, 0)
		pickRoute := GenerateInitialRoute(num_cities)
		initialRoute = append(initialRoute, pickRoute...)

		newRoute, newDistance := TwoOpt(disanceMat, initialRoute, 0.0001)
		if iteration == 0 {
			bestDistance = newDistance
			bestRoute = newRoute
		}

		if newDistance < bestDistance {
			bestDistance = newDistance
			bestRoute = newRoute
		}
		iteration += 1
	}

	return bestRoute, bestDistance
}

func SolveTSPTwoOptToStart(disanceMat [][]float64, maxIteration int) ([]int, float64) {
	rand.Seed(uint64(time.Now().Unix()))
	bestDistance := 0.0
	bestRoute := []int{}
	iteration := 0
	// route = [0,1,2,3]
	for iteration < maxIteration {
		num_cities := len(disanceMat)
		initialRoute := []int{}
		initialRoute = append(initialRoute, 0)
		pickRoute := GenerateInitialRoute(num_cities)
		initialRoute = append(initialRoute, pickRoute...)

		initialRoute = append(initialRoute, 0)
		newRoute, newDistance := TwoOpt(disanceMat, initialRoute, 0.0001)
		if iteration == 0 {
			bestDistance = newDistance
			bestRoute = newRoute
		}

		if newDistance < bestDistance {
			bestDistance = newDistance
			bestRoute = newRoute
		}
		iteration += 1
	}

	return bestRoute, bestDistance
}

// https://en.wikipedia.org/wiki/2-opt
func TwoOpt(disanceMat [][]float64, route []int, improvmentThreshold float64) ([]int, float64) {
	var bestRoute = make([]int, len(route))
	copy(bestRoute, route)
	bestDistance := calculateDistance(disanceMat, route)
	improvmentFactor := 1.0
	numCities := len(disanceMat)

	for improvmentFactor > improvmentThreshold {
		previousBestDist := bestDistance
		for swapFirst := 1; swapFirst < numCities-2; swapFirst++ {
			for swapLast := swapFirst + 1; swapLast < numCities-1; swapLast++ {
				beforeStart := bestRoute[swapFirst-1]
				start := bestRoute[swapFirst]
				end := bestRoute[swapLast]
				afterEnd := bestRoute[swapLast+1]
				before := disanceMat[beforeStart][start] + disanceMat[end][afterEnd]
				after := disanceMat[beforeStart][end] + disanceMat[start][afterEnd]
				if after < before {
					newRoute := swapReverseOpt(bestRoute, swapFirst, swapLast)
					newDistance := calculateDistance(disanceMat, newRoute)
					bestDistance = newDistance
					bestRoute = newRoute
				}
			}
		}
		improvmentFactor = 1 - bestDistance/previousBestDist
	}
	return bestRoute, bestDistance
}

func swapReverseOpt(route []int, swapFirst int, swapLast int) []int {
	var reversedRoute = []int{}

	for i := swapLast; i >= swapFirst; i-- {
		reversedRoute = append(reversedRoute, route[i])
	}

	var lastRoute = []int{}

	for i := swapLast + 1; i < len(route); i++ {
		lastRoute = append(lastRoute, route[i])
	}

	var newRoute = []int{}
	newRoute = append(newRoute, route[:swapFirst]...)
	newRoute = append(newRoute, reversedRoute...)
	newRoute = append(newRoute, lastRoute...)

	return newRoute
}

func calculateDistance(disanceMat [][]float64, route []int) float64 {
	distance := 0.0
	for i := 0; i < len(route)-1; i++ {
		distance += disanceMat[route[i]][route[i+1]]
	}
	return distance
}
