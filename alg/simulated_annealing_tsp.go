package alg

import (
	"math"

	"math/rand/v2"
)

type SimulatedAnnealing struct {
	DistanceMatrix [][]float64
}

func NewSimulatedAnnealing(distanceMatrix [][]float64) *SimulatedAnnealing {
	return &SimulatedAnnealing{
		DistanceMatrix: distanceMatrix,
	}
}

func acceptanceProbability(energy float64, newEnergy float64, temperature float64) float64 {
	if newEnergy < energy {
		return 1.0
	}

	return math.Exp((energy - newEnergy) / temperature)
}

func SimpleNNHeuristics(distanceMatrix [][]float64) []int {
	numCities := len(distanceMatrix)
	visited := make([]bool, numCities)
	tour := make([]int, numCities)
	tour[0] = 0
	visited[0] = true
	for i := 1; i < numCities; i++ {
		minDist := math.MaxFloat64
		minIdx := -1
		for j := 0; j < numCities; j++ {
			if !visited[j] && distanceMatrix[tour[i-1]][j] < minDist {
				minDist = distanceMatrix[tour[i-1]][j]
				minIdx = j
			}
		}
		tour[i] = minIdx
		visited[minIdx] = true
	}
	return tour
}

func (sa *SimulatedAnnealing) Solve() ([]int, float64) {
	temp := 100000.0
	coolingRate := 0.00003

	// currentSolTour, _ := SolveTSPTwoOpt(sa.DistanceMatrix, 100)
	currentSolTour := SimpleNNHeuristics(sa.DistanceMatrix)
	var firstTour = make([]int, len(currentSolTour))
	copy(firstTour, currentSolTour)

	var best = make([]int, len(currentSolTour))
	copy(best, currentSolTour)
	bestDistance := calculateDistanceSA(sa.DistanceMatrix, best)

	numCities := len(sa.DistanceMatrix)

	for temp > 1 {
		var newSolution = make([]int, len(currentSolTour))
		copy(newSolution, currentSolTour)

		tourPosTwo := rand.IntN((numCities+1-1)-2) + 2
		tourPosOne := rand.IntN(numCities + 1 - tourPosTwo)
		for tourPosOne == tourPosTwo {
			tourPosOne = rand.IntN(numCities + 1 - tourPosTwo)
		}

		swapReverseSA(newSolution, tourPosOne, tourPosTwo)

		currentEnergy := calculateDistanceSA(sa.DistanceMatrix, currentSolTour)
		neighbourEnergy := calculateDistanceSA(sa.DistanceMatrix, newSolution)

		if acceptanceProbability(currentEnergy, neighbourEnergy, temp) > rand.Float64() {
			currentSolTour = newSolution
		}

		if calculateDistanceSA(sa.DistanceMatrix, currentSolTour) < calculateDistanceSA(sa.DistanceMatrix, best) {
			best = currentSolTour
			bestDistance = calculateDistanceSA(sa.DistanceMatrix, best)
		}

		temp *= 1 - coolingRate
	}

	// best = append(best, best[0])
	return best, bestDistance
}

func (ch *ContractedGraph) TravelingSalesmanProblemSimulatedAnnealing(cities []int32) ([]CHNode2, float64, float64, [][]float64) {

	spPair := [][]int32{}
	for i := 0; i < len(cities); i++ {
		for j := 0; j < len(cities); j++ {
			if i == j {
				continue
			}
			spPair = append(spPair, []int32{cities[i], cities[j]})
		}
	}

	workers := NewWorkerPool[[]int32, SPSingleResultResult](3, len(spPair))

	for i := 0; i < len(spPair); i++ {
		workers.AddJob(spPair[i])
	}
	close(workers.jobQueue)

	spMap := make(map[int32]map[int32]SPSingleResultResult)

	workers.Start(ch.callBidirectionalDijkstra)
	workers.Wait()

	for i := 0; i < len(spPair); i++ {
		spMap[spPair[i][0]] = make(map[int32]SPSingleResultResult)
	}
	spMap[cities[len(cities)-1]] = make(map[int32]SPSingleResultResult)

	for curr := range workers.CollectResults() {

		spMap[curr.Source][curr.Dest] = curr
	}

	
	distancesMat := make([][]float64, len(cities))
	for i := 0; i < len(cities); i++ {
		distancesMat[i] = make([]float64, len(cities))
		for j := 0; j < len(cities); j++ {
			if i == j {
				distancesMat[i][j] = 0
			}
			distancesMat[i][j] = spMap[cities[i]][cities[j]].Eta // pake eta karena road network ada hierarchy nya
		}
	}

	bestTourCitiesLatLon := [][]float64{}
	saTSP := NewSimulatedAnnealing(distancesMat)
	bestTour, bestETA := saTSP.Solve() // solve tsp pake simulated annealing
	tspTourNodes := []CHNode2{}
	bestDistance := 0.0
	for i := 0; i < len(bestTour)-1; i++ {

		currPathNodes := spMap[cities[bestTour[i]]][cities[bestTour[i+1]]].Paths
		bestDistance += spMap[cities[bestTour[i]]][cities[bestTour[i+1]]].Dist
		bestTourCitiesLatLon = append(bestTourCitiesLatLon, []float64{ch.ContractedNodes[cities[bestTour[i]]].Lat, ch.ContractedNodes[cities[bestTour[i]]].Lon})
		tspTourNodes = append(tspTourNodes, currPathNodes...)
	}
	bestTourCitiesLatLon = append(bestTourCitiesLatLon, []float64{ch.ContractedNodes[cities[bestTour[len(bestTour)-1]]].Lat,
		ch.ContractedNodes[cities[bestTour[len(bestTour)-1]]].Lon})
	tspTourNodes = append(tspTourNodes, spMap[cities[bestTour[len(bestTour)-1]]][cities[bestTour[0]]].Paths...) // dari end tour ke start tour

	return tspTourNodes, bestETA, bestDistance, bestTourCitiesLatLon
}

// // harus pake rute a->b dan b->a, udah coba misalin undirected pathnya salah (ada yang gak ke jalan) karena pake road network beneran
// func (ch *ContractedGraph) TravelingSalesmanProblemSimulatedAnnealing(cities []int32) ([]CHNode2, float64, float64, [][]float64) {
// 	srcNode := make(map[int32]bool)

// 	spPair := [][]int32{}
// 	for i := 0; i < len(cities)-1; i++ {
// 		for j := i + 1; j < len(cities); j++ {
// 			srcNode[cities[i]] = true
// 			spPair = append(spPair, []int32{cities[i], cities[j]})
// 		}
// 	}

// 	workers := NewWorkerPool[[]int32, SPSingleResultResult](3, len(spPair))

// 	for i := 0; i < len(spPair); i++ {
// 		workers.AddJob(spPair[i])
// 	}
// 	close(workers.jobQueue)

// 	spMap := make(map[int32]map[int32]SPSingleResultResult)

// 	workers.Start(ch.callBidirectionalDijkstra)
// 	workers.Wait()

// 	for i := 0; i < len(spPair); i++ {
// 		spMap[spPair[i][0]] = make(map[int32]SPSingleResultResult)
// 	}
// 	spMap[cities[len(cities)-1]] = make(map[int32]SPSingleResultResult)

// 	for curr := range workers.CollectResults() {

// 		spMap[curr.Source][curr.Dest] = curr
// 		spMap[curr.Dest][curr.Source] = curr
// 	}

// 	// for i := 0; i < len(cities); i++ {
// 	// 	for j := 0; j <= i; j++ {

// 	// 		if i == j {
// 	// 			continue
// 	// 		}
// 	// 		spMap[cities[i]][cities[j]] = spMap[cities[j]][cities[i]]
// 	// 	}
// 	// }

// 	distancesMat := make([][]float64, len(cities))
// 	for i := 0; i < len(cities); i++ {
// 		distancesMat[i] = make([]float64, len(cities))
// 		for j := 0; j < len(cities); j++ {
// 			if i == j {
// 				distancesMat[i][j] = 0
// 			}
// 			distancesMat[i][j] = spMap[cities[i]][cities[j]].Eta // pake eta karena road network ada hierarchy nya
// 		}
// 	}

// 	bestTourCitiesLatLon := [][]float64{}
// 	saTSP := NewSimulatedAnnealing(distancesMat)
// 	bestTour, bestETA := saTSP.Solve() // solve tsp pake simulated annealing
// 	tspTourNodes := []CHNode2{}
// 	bestDistance := 0.0
// 	for i := 0; i < len(bestTour)-1; i++ {
// 		var currPathNodes []CHNode2
// 		if srcNode[cities[bestTour[i]]] {
// 			currPathNodes = spMap[cities[bestTour[i]]][cities[bestTour[i+1]]].Paths
// 		} else {
// 			currPathNodes = reverseG(spMap[cities[bestTour[i]]][cities[bestTour[i+1]]].Paths)
// 		}
// 		bestDistance += spMap[cities[bestTour[i]]][cities[bestTour[i+1]]].Dist
// 		bestTourCitiesLatLon = append(bestTourCitiesLatLon, []float64{ch.ContractedNodes[cities[bestTour[i]]].Lat, ch.ContractedNodes[cities[bestTour[i]]].Lon})
// 		tspTourNodes = append(tspTourNodes, currPathNodes...)
// 	}
// 	bestTourCitiesLatLon = append(bestTourCitiesLatLon, []float64{ch.ContractedNodes[cities[bestTour[len(bestTour)-1]]].Lat,
// 		ch.ContractedNodes[cities[bestTour[len(bestTour)-1]]].Lon})
// 	tspTourNodes = append(tspTourNodes, spMap[cities[bestTour[len(bestTour)-1]]][cities[bestTour[0]]].Paths...) // dari end tour ke start tour

// 	return tspTourNodes, bestETA, bestDistance, bestTourCitiesLatLon
// }

func calculateDistanceSA(disanceMat [][]float64, route []int) float64 {
	distance := 0.0
	for i := 0; i < len(route); i++ {

		distance += disanceMat[route[i%len(route)]][route[(i+1)%len(route)]]
	}
	return distance
}

func swapReverseSA(route []int, tourPosOne int, tourPosTwo int) {
	var reversedRoute = make([]int, len(route))
	copy(reversedRoute, route)
	reversedRoute = reversedRoute[tourPosOne : tourPosOne+tourPosTwo]
	reverseG(reversedRoute)
	idx := 0
	for i := tourPosOne; i < tourPosOne+tourPosTwo; i++ {
		route[i] = reversedRoute[idx]
		idx++
	}
}
