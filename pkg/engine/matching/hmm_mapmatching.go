package matching

import (
	"lintang/navigatorx/pkg/datastructure"
	"lintang/navigatorx/pkg/geo"
	"math"
)

// pasangan observation ke hidden states nya
// observation = gps point, hidden state = road segment
type StateObservationPair struct {
	Observation datastructure.CHNode2
	State       []State
}

type State struct {
	ID     int
	NodeID int32
	Lat    float64
	Lon    float64
	Dist   float64
	EdgeID int32
}

type ContractedGraph interface {
	IsChReady() bool
	SnapLocationToRoadNetworkNodeH3(ways []datastructure.SmallWay, wantToSnap []float64) int32
}

type RouteAlgorithm interface {
	ShortestPathBiDijkstra(int32, int32) ([]datastructure.CHNode2, []datastructure.EdgeCH, float64, float64)
	// AStarCH(from, to int32) (pathN []datastructure.CHNode, path string, eta float64, found bool, dist float64)
}
type KVDB interface {
	GetNearestStreetsFromPointCoord(lat, lon float64) ([]datastructure.SmallWay, error)
}

type HMMMapMatching struct {
	ch    ContractedGraph
	KVDB  KVDB
	route RouteAlgorithm
}

func NewHMMMapMatching(ch ContractedGraph, KVDB KVDB, route RouteAlgorithm) *HMMMapMatching {
	return &HMMMapMatching{
		ch:    ch,
		KVDB:  KVDB,
		route: route,
	}
}

func (hmm *HMMMapMatching) HiddenMarkovModelMapMatching(gps []datastructure.StateObservationPair) []datastructure.CHNode2 {

	transitionProb := make(map[int]map[int]float64)
	emissionProb := make(map[int]map[int]float64)
	initialProb := make(map[int]float64)

	obs := []ViterbiNode{}
	states := []ViterbiNode{}

	for i := 0; i < len(gps)-1; i++ {
		nextRoadNodes := gps[i+1].State // curr+1 observation hidden states
		currRoadNodes := gps[i].State   //  curr observation hidden states
		currObsLoc := geo.NewLocation(gps[i].Observation.Lat, gps[i].Observation.Lon)
		nextObsLoc := geo.NewLocation(gps[i+1].Observation.Lat, gps[i+1].Observation.Lon)
		for _, currState := range currRoadNodes {
			for _, nextState := range nextRoadNodes {
				var stateRouteLength float64
				var currTransitionProb float64

				if currState.EdgeID == nextState.EdgeID {
					stateRouteLength = geo.HaversineDistance(geo.NewLocation(currState.Lat, currState.Lon), geo.NewLocation(nextState.Lat, nextState.Lon))
				} else {
					var err error
					if nextState.NodeID == -1 {
						nextState.NodeID, err = hmm.snapLocToStreetNode(nextState.Lat, nextState.Lon)
						if err != nil {
							continue
						}

					}

					if currState.NodeID == -1 {
						currState.NodeID, err = hmm.snapLocToStreetNode(currState.Lat, currState.Lon)
						if err != nil {
							continue
						}

					}
					var dijkstraSp float64
					if hmm.ch.IsChReady() {
						_, _, _, dijkstraSp = hmm.route.ShortestPathBiDijkstra(currState.NodeID, nextState.NodeID)
					} 

					if dijkstraSp == -1 || dijkstraSp*1000 > 150 {
						// beda jalan & shortest path antara hidden state & next hidden state nya lebih dari 150m
						currTransitionProb = -999999999
					} else {
						stateRouteLength = dijkstraSp
					}
				}

				stateRouteLength *= 1000
				obsDistance := geo.HaversineDistance(currObsLoc, nextObsLoc) * 1000
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
			currStateLoc := geo.NewLocation(currState.Lat, currState.Lon)
			obsStateDist := geo.HaversineDistance(currObsLoc, currStateLoc) * 1000
			currEmissionProb := computelogNormalEmissionProb(obsStateDist)
			_, ok := emissionProb[currState.ID]
			if !ok {
				emissionProb[currState.ID] = make(map[int]float64)
			}
			emissionProb[currState.ID][i] = currEmissionProb

			states = append(states, ViterbiNode{ID: currState.ID, NodeID: currState.NodeID, Lat: currState.Lat, Lon: currState.Lon})
		}
		obs = append(obs, ViterbiNode{ID: i, NodeID: -1})

		if i == len(gps)-2 {
			obs = append(obs, ViterbiNode{ID: len(gps) - 1, NodeID: -1})
			for _, nextState := range nextRoadNodes {
				states = append(states, ViterbiNode{ID: nextState.ID, NodeID: nextState.NodeID, Lat: nextState.Lat, Lon: nextState.Lon})

				// emmission probabity buat n-1 observation & hidden statesnya
				nextStateLoc := geo.NewLocation(nextState.Lat, nextState.Lon)
				obsStateDist := geo.HaversineDistance(nextObsLoc, nextStateLoc) * 1000
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

	nodePath := []datastructure.CHNode2{}
	for _, p := range path {
		nodePath = append(nodePath, datastructure.CHNode2{
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

func (hmm *HMMMapMatching) snapLocToStreetNode(lat, lon float64) (int32, error) {
	ways, err := hmm.KVDB.GetNearestStreetsFromPointCoord(lat, lon)
	if err != nil {
		return 0, err
	}
	streetNodeIDx := hmm.ch.SnapLocationToRoadNetworkNodeH3(ways, []float64{lat, lon})

	return streetNodeIDx, nil
}
