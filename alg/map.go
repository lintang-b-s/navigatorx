package alg

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/k0kubun/go-ansi"
	"github.com/paulmach/osm"
	"github.com/schollz/progressbar/v3"
)

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type SurakartaWay struct {
	ID                  int32
	CenterLoc           []float64 // [lat, lon]
	NodesID             []int64   // ini harus int64 karena id dari osm int64  (osm.NodeId)
	IntersectionNodesID []int64
}

var ValidRoadType = map[string]bool{
	"motorway":       true,
	"trunk":          true,
	"primary":        true,
	"secondary":      true,
	"tertiary":       true,
	"unclassified":   true,
	"residential":    true,
	"motorway_link":  true,
	"trunk_link":     true,
	"primary_link":   true,
	"secondary_link": true,
	"tertiary_link":  true,
	"living_street":  true,
}

// gak ada 1 way dengan multiple road type
func InitGraph(ways []*osm.Way, trafficLightNodeIdMap map[osm.NodeID]bool) ([]SurakartaWay, []Node) {
	var SurakartaNodeMap = make(map[int64]*Node)

	oneWayTypesMap := make(map[string]int64)
	twoWayTypesMap := make(map[string]int64)

	bar := progressbar.NewOptions(len(ways),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()), //you should install "github.com/k0kubun/go-ansi"
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][2/7][reset] Menyimpan way & node Openstreetmap ..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	surakartaWays := []SurakartaWay{}
	for wayIDx, way := range ways {

		namaJalan := ""
		roadTypes := make(map[string]int)

		for _, tag := range way.Tags {
			if tag.Key == "highway" {
				twoWayTypesMap[tag.Key+"="+tag.Value] += 1
				roadTypes[tag.Value] += 1
			}
			if strings.Contains(tag.Key, "oneway") && !strings.Contains(tag.Value, "no") {
				oneWayTypesMap[tag.Key+"="+tag.Value] += 1

			}
			if tag.Key == "name" {
				namaJalan = tag.Value
			}
		}

		if !isOsmWayUsedByCars(way.TagMap()) {
			continue
		}

		sWay := SurakartaWay{
			NodesID: make([]int64, 0),
		}

		streetNodeLats := []float64{}
		streetNodeLon := []float64{}

		// creategraph node
		for i := 0; i < len(way.Nodes); i++ {
			currNode := way.Nodes[i]

			node := &Node{
				Lat:          currNode.Lat,
				Lon:          currNode.Lon,
				ID:           int64(currNode.ID),
				StreetName:   namaJalan,
				TrafficLight: trafficLightNodeIdMap[currNode.ID],
			}

			if fromRealNode, ok := SurakartaNodeMap[node.ID]; ok {
				node = fromRealNode
			} else {
				SurakartaNodeMap[node.ID] = node
			}
			node.UsedInRoad += 1

			// add node ke surakartaway
			sWay.NodesID = append(sWay.NodesID, node.ID)
			// append node lat & node lon
			streetNodeLats = append(streetNodeLats, node.Lat)
			streetNodeLon = append(streetNodeLon, node.Lon)

		}
		sort.Float64s(streetNodeLats)
		sort.Float64s(streetNodeLon)

		// https://www.movable-type.co.uk/scripts/latlong.html
		centerLat, centerLon := MidPoint(streetNodeLats[0], streetNodeLon[0], streetNodeLats[len(streetNodeLats)-1], streetNodeLon[len(streetNodeLon)-1])
		sWay.CenterLoc = []float64{centerLat, centerLon}

		sWay.ID = int32(wayIDx)
		surakartaWays = append(surakartaWays, sWay)
		bar.Add(1)
	}

	WriteWayTypeToCsv(oneWayTypesMap, "onewayTypes.csv")
	WriteWayTypeToCsv(twoWayTypesMap, "twoWayTypes.csv")

	surakartaNodes, surakartaWays := processOnlyIntersectionRoadNodes(SurakartaNodeMap, ways, surakartaWays)

	fmt.Println("")
	return surakartaWays, surakartaNodes
}

func processOnlyIntersectionRoadNodes(nodeMap map[int64]*Node, ways []*osm.Way, surakartaWays []SurakartaWay) ([]Node, []SurakartaWay) {
	surakartaNodes := []Node{}
	alreadyAdded := make(map[int64]struct{})
	intersectionNodes := []int64{}
	for wayIDx, way := range ways {
		maxSpeed, isOneWay, reversedOneWay, roadType := getMaxspeedOneWayRoadType(*way)
		if !isOsmWayUsedByCars(way.TagMap()) {
			continue
		}

		if !ValidRoadType[roadType] {
			continue
		}
		currSurakartaWay := &surakartaWays[wayIDx]

		from := nodeMap[int64(way.Nodes[0].ID)]

		if _, ok := alreadyAdded[from.ID]; !ok {
			intersectionNodes = append(intersectionNodes, from.ID)
			alreadyAdded[from.ID] = struct{}{}
		}
		for i := 1; i < len(way.Nodes); i++ {
			currNode := way.Nodes[i]
			// idnya masih pake id osm
			to := nodeMap[int64(currNode.ID)]

			if to.UsedInRoad >= 2 {
				// nodenya ada di intersection of 2  or more roads

				// add edge antara dua node intersection
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
					currSurakartaWay.IntersectionNodesID = append(currSurakartaWay.IntersectionNodesID, from.ID)
				} else if isOneWay && reversedOneWay {
					reverseEdge := Edge{
						From:     to,
						To:       from,
						Cost:     fromToDistance,
						MaxSpeed: maxSpeed,
					}
					to.Out_to = append(to.Out_to, reverseEdge)
					currSurakartaWay.IntersectionNodesID = append(currSurakartaWay.IntersectionNodesID, to.ID)
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
					currSurakartaWay.IntersectionNodesID = append(currSurakartaWay.IntersectionNodesID, from.ID)
					currSurakartaWay.IntersectionNodesID = append(currSurakartaWay.IntersectionNodesID, to.ID)
				}
				if _, ok := alreadyAdded[to.ID]; !ok {
					intersectionNodes = append(intersectionNodes, to.ID)
					alreadyAdded[to.ID] = struct{}{}
				}
				from = to
			}
		}

	}

	for _, node := range intersectionNodes {

		// hanya append node intersection
		surakartaNodes = append(surakartaNodes, *nodeMap[node])

	}
	return surakartaNodes, surakartaWays
}

func getMaxspeedOneWayRoadType(way osm.Way) (float64, bool, bool, string) {

	maxSpeed := 50.0

	isOneWay := false // 0, 1
	reversedOneWay := false

	roadTypes := make(map[string]int)

	roadType := ""

	for _, tag := range way.Tags {
		if tag.Key == "highway" {
			roadTypes[tag.Value] += 1
			roadType = tag.Value
		}
		if strings.Contains(tag.Key, "oneway") && !strings.Contains(tag.Value, "no") {
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

	}
	if maxSpeed == 50.0 || maxSpeed == 0 {
		maxSpeed = RoadTypeMaxSpeed(roadType)
	}
	return maxSpeed, isOneWay, reversedOneWay, roadType
}

func WriteWayTypeToCsv(wayTypesMap map[string]int64, filename string) {
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
