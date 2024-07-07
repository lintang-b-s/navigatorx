package alg

import (
	"encoding/csv"
	"fmt"
	"lintang/coba_osm/util"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/paulmach/osm"
)

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Bound struct {
	MinLat float64 `json:"minlat"`
	MaxLat float64 `json:"maxlat"`
	MinLon float64 `json:"minlon"`
	MaxLon float64 `json:"maxlon"`
}

type SurakartaWay struct {
	ID        int64
	Nodes     []*Node
	Bound     Bound
	CenterLoc []float64 // [lat, lon]
}

// gak ada 1 way dengan multiple road type

func InitGraph(ways []*osm.Way) []SurakartaWay {
	var SurakartaNodeMap = make(map[int64]*Node)

	oneWayTypesMap := make(map[string]int64)
	twoWayTypesMap := make(map[string]int64)

	surakartaWays := []SurakartaWay{}
	for idx, way := range ways {

		namaJalan := ""

		maxSpeed := 50.0

		isOneWay := false // 0, 1
		reversedOneWay := false

		roadTypes := make(map[string]int)

		roadType := ""

		for _, tag := range way.Tags {
			if tag.Key == "highway" {
				twoWayTypesMap[tag.Key+"="+tag.Value] += 1
				roadTypes[tag.Value] += 1
				roadType = tag.Value
			}
			if strings.Contains(tag.Key, "oneway") && !strings.Contains(tag.Value, "no") {
				oneWayTypesMap[tag.Key+"="+tag.Value] += 1
				isOneWay = true
				if strings.Contains(tag.Value, "-1") {
					reversedOneWay = true
				}
			}
			if strings.Contains(tag.Key, "maxspeed") {
				_, err := strconv.ParseFloat(tag.Value, 64)
				if err != nil {
					maxSpeed, _ = strconv.ParseFloat(tag.Value, 64)
				}
			}

			if tag.Key == "name" {
				namaJalan = tag.Value
			}
		}
		// path,cycleway, construction,steps,platform,bridleway,footway are not for cars
		if maxSpeed == 50.0 {
			maxSpeed = RoadTypeMaxSpeed(roadType)
		}

		if roadType == "path" || roadType == "cycleway" || roadType == "construction" || roadType == "steps" || roadType == "platform" ||
			roadType == "bridleway" || roadType == "footway" {
			continue
		}

		if idx%50000 == 0 {
			fmt.Println("membuat graph dari openstreetmap way ke: " + fmt.Sprint(idx))

		}
		sWay := SurakartaWay{
			ID:    int64(way.ID),
			Nodes: make([]*Node, 0),
		}

		streetNodeLats := []float64{}
		streetNodeLon := []float64{}

		// creategraph node
		for i := 0; i < len(way.Nodes)-1; i++ {
			fromN := way.Nodes[i]

			from := &Node{
				Lat:        util.TruncateFloat64(fromN.Lat, 6),
				Lon:        util.TruncateFloat64(fromN.Lon, 6),
				ID:         int64(fromN.ID),
				StreetName: namaJalan,
			}

			toN := way.Nodes[i+1]
			to := &Node{
				Lat:        util.TruncateFloat64(toN.Lat, 6),
				Lon:        util.TruncateFloat64(toN.Lon, 6),
				ID:         int64(toN.ID),
				StreetName: namaJalan,
			}

			if fromRealNode, ok := SurakartaNodeMap[from.ID]; ok {
				from = fromRealNode
			} else {
				SurakartaNodeMap[from.ID] = from
			}
			if toRealNode, ok := SurakartaNodeMap[to.ID]; ok {
				to = toRealNode
			} else {
				SurakartaNodeMap[to.ID] = to
			}

			// fromToDistance := EuclideanDistance(from, to)
			fromLoc := NewLocation(from.Lat, from.Lon)
			toLoc := NewLocation(to.Lat, to.Lon)
			fromToDistance := HaversineDistance(fromLoc, toLoc) * 1000 // meter
			if isOneWay && !reversedOneWay {
				edge := Edge{
					From:     from,
					To:       to,
					Cost:     fromToDistance,
					MaxSpeed: maxSpeed,
				}
				from.Out_to = append(from.Out_to, edge)
			} else if isOneWay && reversedOneWay {
				reverseEdge := Edge{
					From:     to,
					To:       from,
					Cost:     fromToDistance,
					MaxSpeed: maxSpeed,
				}
				to.Out_to = append(to.Out_to, reverseEdge)
			} else {
				edge := Edge{
					From:     from,
					To:       to,
					Cost:     fromToDistance,
					MaxSpeed: maxSpeed,
				}
				from.Out_to = append(from.Out_to, edge)

				reverseEdge := Edge{
					From:     to,
					To:       from,
					Cost:     fromToDistance,
					MaxSpeed: maxSpeed,
				}
				to.Out_to = append(to.Out_to, reverseEdge)
			}

			// add node ke surakartaway
			sWay.Nodes = append(sWay.Nodes, from)
			// append node lat & node lon
			streetNodeLats = append(streetNodeLats, from.Lat)
			streetNodeLon = append(streetNodeLon, from.Lon)
			if i == len(way.Nodes)-2 {
				sWay.Nodes = append(sWay.Nodes, to)
				streetNodeLats = append(streetNodeLats, to.Lat)
				streetNodeLon = append(streetNodeLon, to.Lon)
			}
		}
		sort.Sort(sort.Float64Slice(streetNodeLats))
		sort.Sort(sort.Float64Slice(streetNodeLon))
		sWay.Bound.MinLat = streetNodeLats[0]
		sWay.Bound.MaxLat = streetNodeLats[len(streetNodeLats)-1]
		sWay.Bound.MinLon = streetNodeLon[0]
		sWay.Bound.MaxLon = streetNodeLon[len(streetNodeLon)-1]
		// https://www.movable-type.co.uk/scripts/latlong.html
		centerLat, centerLon := MidPoint(sWay.Bound.MinLat, sWay.Bound.MinLon, sWay.Bound.MaxLat, sWay.Bound.MaxLon)
		sWay.CenterLoc = []float64{centerLat, centerLon}

		surakartaWays = append(surakartaWays, sWay)
	}
	clear(SurakartaNodeMap)

	writeWayTypeToCsv(oneWayTypesMap, "onewayTypes.csv")
	writeWayTypeToCsv(twoWayTypesMap, "twoWayTypes.csv")
	return surakartaWays
}

func writeWayTypeToCsv(wayTypesMap map[string]int64, filename string) {
	wayTypesArr := make([][]string, len(wayTypesMap)+1)

	count := 0
	idx := 0
	for key, val := range wayTypesMap {
		tipe := strings.Split(key, "=")
		wayTypesArr[idx] = []string{tipe[0], tipe[1], strconv.FormatInt(val, 10)}
		idx++
		count += int(val)
	}
	wayTypesArr[idx] = []string{"total", fmt.Sprint(count)}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.WriteAll(wayTypesArr)
}
