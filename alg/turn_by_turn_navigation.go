package alg

import (
	"errors"
	"fmt"
	"lintang/coba_osm/util"
	"math"
)

type TURN string

const (
	KIRI  TURN = "KIRI"
	KANAN TURN = "KANAN"
	LURUS TURN = "LURUS"
)

type Navigation struct {
	Instruction string  `json:"instruction"`
	StreetName  string  `json:"street_name"`
	TurnETA     float64 `json:"turn_eta"`
	TurnDist    float64 `json:"distance_before_turn"`
	Turn        TURN    `json:"turn"`
}

func CreateTurnByTurnNavigation(p []Pather) ([]Navigation, error) {
	n := []Navigation{}
	if len(p) < 4 {
		return n, nil
	}

	startSTNodeBeforeTurn := *p[0].(*Node)
	currStreet := p[0].(*Node).GetStreetName()
	currStDist := 0.0
	currStETA := 0.0

	for i := 0; i < len(p)-3; i++ {
		pathN2 := *p[i+1].(*Node)
		pathN3 := *p[i+2].(*Node)
		pathN4 := *p[i+3].(*Node)
		if currStreet != pathN3.GetStreetName() &&
			(pathN3.GetStreetName() != "") {

			if pathN3.GetStreetName() != pathN4.GetStreetName() {
				continue
			}

			stNode := MakeSixDigitsAfterComa(startSTNodeBeforeTurn, 6)
			pathN3 := MakeSixDigitsAfterComa(pathN3, 6)
			pathN4 := MakeSixDigitsAfterComa(pathN4, 6)

			b1 := Bearing(util.TruncateFloat64(stNode.Lat, 6), util.TruncateFloat64(stNode.Lon, 6), util.TruncateFloat64(pathN3.Lat, 6),
				util.TruncateFloat64(pathN3.Lon, 6))

			b2 := Bearing(util.TruncateFloat64(pathN3.Lat, 6), util.TruncateFloat64(pathN3.Lon, 6),
				util.TruncateFloat64(pathN4.Lat, 6), util.TruncateFloat64(pathN4.Lon, 6))

			if b1 == 0 || b2 == 0 {
				continue
			}

			turn := CalculateTurn(b1, b2)
			turnDirection := GetTurnDirection(PredictTurn(turn))

			for j := i + 4; j <= i+4+2; j++ {
				// biar turn directionnya makin akurat (ada node simpangan pathN4 yang agak gajelas)
				if j < len(p) {
					pathN5 := *p[j].(*Node)
					pathN5 = MakeSixDigitsAfterComa(pathN5, 6)
					if pathN5.GetStreetName() == pathN3.GetStreetName() {
						b3 := Bearing(util.TruncateFloat64(stNode.Lat, 6), util.TruncateFloat64(stNode.Lon, 6), util.TruncateFloat64(pathN3.Lat, 6),
							util.TruncateFloat64(pathN3.Lon, 6))

						b4 := Bearing(util.TruncateFloat64(pathN3.Lat, 6), util.TruncateFloat64(pathN3.Lon, 6),
							util.TruncateFloat64(pathN5.Lat, 6), util.TruncateFloat64(pathN5.Lon, 6))

						if b3 == 0 || b4 == 0 {
							continue
						}

						turnAdjacentLines := CalculateTurn(b3, b4)
						turnDirectionAdjacentLines := GetTurnDirection(PredictTurn(turnAdjacentLines))

						if math.Abs(turnAdjacentLines) > math.Abs(turn) {
							turnDirection = turnDirectionAdjacentLines
						}
					}
				}

			}

			n = append(n, Navigation{
				StreetName: pathN3.GetStreetName(),
				TurnETA:    util.RoundFloat(currStETA, 2),  //CalculateETA(startSTNodeBeforeTurn, pathN3),
				TurnDist:   util.RoundFloat(currStDist, 2), //  HaversineDistance(stLoc, turnLoc),
				Turn:       turnDirection,
			})

			startSTNodeBeforeTurn = pathN3
			currStreet = pathN3.GetStreetName()
			currStDist = 0
			currStETA = 0
		} else {
			stLoc := NewLocation(startSTNodeBeforeTurn.Lat, startSTNodeBeforeTurn.Lon)
			pathN2Loc := NewLocation(pathN2.Lat, pathN2.Lon)
			currStDist = HaversineDistance(stLoc, pathN2Loc) * 1000
			maxSpeed := 30 * 1000 / 60
			currStETA = HaversineDistance(stLoc, pathN2Loc) * 1000 / float64(maxSpeed)
		}
	}

	beforeDestionationLat := p[len(p)-1].(*Node).Lat
	beforeDestionationLon := p[len(p)-1].(*Node).Lon
	stLoc := NewLocation(startSTNodeBeforeTurn.Lat, startSTNodeBeforeTurn.Lon)
	pathN2Loc := NewLocation(beforeDestionationLat, beforeDestionationLon)
	currStDist = HaversineDistance(stLoc, pathN2Loc) * 1000
	maxSpeed := 30 * 1000 / 60
	currStETA = HaversineDistance(stLoc, pathN2Loc) * 1000 / float64(maxSpeed)

	if len(n) == 0 {
		return []Navigation{{StreetName: "maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)"}},
			errors.New("maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)")
	}

	if n[len(n)-1].StreetName == "" {
		n[len(n)-1].StreetName = "Jalan Unknown"
	}
	n = append(n, Navigation{
		StreetName: n[len(n)-1].StreetName,
		TurnETA:    util.RoundFloat(currStETA, 2),  //CalculateETA(startSTNodeBeforeTurn, pathN3),
		TurnDist:   util.RoundFloat(currStDist, 2), //  HaversineDistance(stLoc, turnLoc),
		Turn:       LURUS,
	})

	// buat instruction
	for i := 0; i < len(n); i++ {
		if i == len(n)-1 {
			n[i].Instruction = fmt.Sprintf(`LURUS dari awal %s ke tempat tujuan (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
		} else if n[i].Turn != LURUS {
			n[i].Instruction = fmt.Sprintf(`Belok %s ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].Turn, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
		} else {
			n[i].Instruction = fmt.Sprintf(`LURUS ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
		}
	}

	return n, nil
}

// buat semua coordinate ada 6 angka dibelakang koma
// biar itungan bearingnya ga ngaco
func MakeSixDigitsAfterComa(n Node, precision int) Node {

	if util.CountDecimalPlaces(n.Lat) != precision {
		n.Lat = util.TruncateFloat64(n.Lat+0.000001, precision)
	}
	if util.CountDecimalPlaces(n.Lon) != precision {
		n.Lon = util.TruncateFloat64(n.Lon+0.000001, precision)
	}
	return n
}

func GetTurnDirection(turn string) TURN {
	switch turn {
	case "KIRI":
		return KIRI
	case "KANAN":
		return KANAN
	case "LURUS":
		return LURUS
	}

	return LURUS
}

func reverse(p []Pather) []Pather {
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
	return p
}
