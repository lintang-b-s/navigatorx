package alg

import "math"

type Viterbi struct {
	Obs            []ViterbiNode
	States         []ViterbiNode
	TransitionProb map[int]map[int]float64
	EmmissionProb  map[int]map[int]float64
	InitialProb    map[int]float64
}

// ViterbiNode: observations(gps points) atau hidden states(road segments)
type ViterbiNode struct {
	ID     int
	NodeID int32
	Lat    float64
	Lon    float64
}

func NewViterbi(ob []ViterbiNode, sts []ViterbiNode,
	trProb, emissionProb map[int]map[int]float64, initialProb map[int]float64) *Viterbi {
	return &Viterbi{
		Obs:            ob,
		States:         sts,
		TransitionProb: trProb,
		EmmissionProb:  emissionProb,
		InitialProb:    initialProb,
	}
}

// https://web.stanford.edu/~jurafsky/slp3/A.pdf
// https://www.cis.upenn.edu/~cis2620/notes/Example-Viterbi-DNA.pdf
func (v *Viterbi) RunViterbi() (float64, []ViterbiNode) {
	viterbi := []map[int]float64{}
	viterbi = append(viterbi, map[int]float64{})
	path := []ViterbiNode{}
	parent := []map[ViterbiNode]ViterbiNode{}
	for _, s := range v.States {
		var initProb float64 = 0
		if val, ok := v.InitialProb[s.ID]; ok {
			initProb = val
		}
		viterbi[0][s.ID] = initProb + v.EmmissionProb[s.ID][v.Obs[0].ID]
		parent = append(parent, make(map[ViterbiNode]ViterbiNode))
	}

	for t := 1; t < len(v.Obs); t++ {
		viterbi = append(viterbi, make(map[int]float64))
		parent = append(parent, make(map[ViterbiNode]ViterbiNode))
		for _, s := range v.States {
			if _, ok := v.EmmissionProb[s.ID][v.Obs[t].ID]; !ok {
				continue
			}
			state := ViterbiNode{}
			maxTransitionProb := math.Inf(-1)
			for _, prevS := range v.States {
				_, okv := viterbi[t-1][prevS.ID]
				_, okt := v.TransitionProb[prevS.ID][s.ID]
				if !okt || !okv {
					continue
				}

				transitionProb := viterbi[t-1][prevS.ID] + v.TransitionProb[prevS.ID][s.ID]
				if transitionProb > maxTransitionProb {
					maxTransitionProb = transitionProb
					state = prevS
				}
			}

			viterbi[t][s.ID] = maxTransitionProb + v.EmmissionProb[s.ID][v.Obs[t].ID]
			parent[t][s] = state

		}
	}

	prob := math.Inf(-1)
	state := ViterbiNode{}
	for _, s := range v.States {
		_, ok := viterbi[len(v.Obs)-1][s.ID]
		if !ok {
			continue
		}
		if viterbi[len(v.Obs)-1][s.ID] > prob {
			prob = viterbi[len(v.Obs)-1][s.ID]
			state = s
		}
	}

	for t := len(v.Obs) - 1; t > 0; t-- {
		path = append(path, state)
		state = parent[t][state]
	}
	path = append(path, state)

	return prob, reverseG(path)
}

func reverseG[T any](arr []T) (result []T) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}
