package alg

import (
	"math"
)

// pasangan observation ke hidden states nya
// observation = gps point, hidden state = road segment
type StateObservationPair struct {
	Observation CHNode2
	State       []State
}

type State struct {
	ID        int
	NodeID    int32
	Lat       float64
	Lon       float64
	Dist      float64
	EdgeBound Bound
}

// snapping gps ke street nodenya salah, projectionnya gak di dalam jalan
// https://www.microsoft.com/en-us/research/publication/hidden-markov-map-matching-noise-sparseness/
// https://www.ismll.uni-hildesheim.de/lehre/semSpatial-10s/script/6.pdf
// https://github.com/bmwcarit/offline-map-matching/blob/master/src/test/java/com/bmw/mapmatchingutils/OfflineMapMatcherTest.java
// masih salah pas pake data gps trip paper microsoft
func (ch ContractedGraph) HiddenMarkovModelMapMatching(gps []StateObservationPair) []CHNode2 {

	// obsStateDistDiff := []float64{}
	transitionProb := make(map[int]map[int]float64)
	emissionProb := make(map[int]map[int]float64)
	initialProb := make(map[int]float64)

	obs := []ViterbiNode{}
	states := []ViterbiNode{}

	for i := 0; i < len(gps)-1; i++ {
		nextRoadNodes := gps[i+1].State // curr+1 observation hidden states
		currRoadNodes := gps[i].State   //  curr observation hidden states
		currObsLoc := NewLocation(gps[i].Observation.Lat, gps[i].Observation.Lon)
		nextObsLoc := NewLocation(gps[i+1].Observation.Lat, gps[i+1].Observation.Lon)
		for _, currState := range currRoadNodes {
			for _, nextState := range nextRoadNodes {
				var stateRouteLength float64
				var currTransitionProb float64

				if currState.EdgeBound == nextState.EdgeBound {
					stateRouteLength = HaversineDistance(NewLocation(currState.Lat, currState.Lon), NewLocation(nextState.Lat, nextState.Lon))
				} else {
					_, _, dijkstraSp := ch.ShortestPathBiDijkstra(currState.NodeID, nextState.NodeID)
					if dijkstraSp == -1 || dijkstraSp*1000 > 1000 {
						// beda jalan & shortest path antara hidden state & next hidden state nya lebih dari 1 km
						currTransitionProb = -999999999
					} else {
						stateRouteLength = dijkstraSp
					}
				}

				stateRouteLength *= 1000
				obsDistance := HaversineDistance(currObsLoc, nextObsLoc) * 1000
				dt := math.Abs(math.Abs(obsDistance) - math.Abs(stateRouteLength))
				_, ok := transitionProb[currState.ID]
				if !ok {
					transitionProb[currState.ID] = make(map[int]float64)
				}
				if currTransitionProb != -999999999 {
					currTransitionProb = computeLogExpoTransitionProb(dt)
				}

				transitionProb[currState.ID][nextState.ID] = currTransitionProb
			}
			currStateLoc := NewLocation(currState.Lat, currState.Lon)
			obsStateDist := HaversineDistance(currObsLoc, currStateLoc) * 1000
			currEmissionProb := computelogNormalEmissionProb(obsStateDist)
			_, ok := emissionProb[currState.ID]
			if !ok {
				emissionProb[currState.ID] = make(map[int]float64)
			}
			emissionProb[currState.ID][i] = currEmissionProb

			states = append(states, ViterbiNode{ID: currState.ID, NodeID: currState.NodeID, Lat: currState.Lat, Lon: currState.Lon})

			// 
		}

		// transitionProb = updateTransitionProb(transitionProb, dts, currRoadNodes, nextRoadNodes)

		obs = append(obs, ViterbiNode{ID: i, NodeID: -1})

		if i == len(gps)-2 {
			obs = append(obs, ViterbiNode{ID: len(gps) - 1, NodeID: -1})
			for _, nextState := range nextRoadNodes {
				states = append(states, ViterbiNode{ID: nextState.ID, NodeID: nextState.NodeID, Lat: nextState.Lat, Lon: nextState.Lon})

				// emmission probabity buat n-1 observation & hidden statesnya
				nextStateLoc := NewLocation(nextState.Lat, nextState.Lon)
				obsStateDist := HaversineDistance(nextObsLoc, nextStateLoc) * 1000
				nextEmissionProb := computelogNormalEmissionProb(obsStateDist)
				_, ok := emissionProb[nextState.ID]
				if !ok {
					emissionProb[nextState.ID] = make(map[int]float64)
				}
				emissionProb[nextState.ID][i+1] = nextEmissionProb
			}
		}

	}

	viterbi := NewViterbi(obs, states, transitionProb, emissionProb, initialProb)
	_, path := viterbi.RunViterbi()

	nodePath := []CHNode2{}
	for _, p := range path {
		nodePath = append(nodePath, CHNode2{
			Lat: p.Lat,
			Lon: p.Lon,
		})
	}
	return nodePath
}

const (
	sigmaZ = 4.07
	beta   = 0.00959442
)

// https://github.com/bmwcarit/offline-map-matching/blob/master/src/main/java/com/bmw/mapmatchingutils/Distributions.java

func computeTransitionProb(obsStateDiff float64, betaTp float64) float64 {

	return (1 * math.Exp(-obsStateDiff/betaTp)) / betaTp
	// return (1 * math.Exp(-obsStateDiff*betaTp)) * betaTp
}

func computeLogExpoTransitionProb(obsStateDiff float64) float64 {

	return math.Log(1.0/beta) - (obsStateDiff / beta)
}

func computeEmissionProb(obsStateDist float64) float64 {
	return (1 * math.Exp(-0.5*math.Pow(obsStateDist/sigmaZ, 2))) / (math.Sqrt(2*math.Pi) * sigmaZ)
}

func computelogNormalEmissionProb(obsStateDist float64) float64 {
	return math.Log(1.0/(math.Sqrt(2*math.Pi)*sigmaZ)) + (-0.5 * math.Pow(obsStateDist/sigmaZ, 2))
}
