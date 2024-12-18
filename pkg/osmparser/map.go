package osmparser

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"lintang/navigatorx/pkg/datastructure"
	"lintang/navigatorx/pkg/geo"
	"lintang/navigatorx/pkg/guidance"

	"github.com/k0kubun/go-ansi"
	"github.com/paulmach/osm"
	"github.com/schollz/progressbar/v3"
)

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
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
	"path":           true,
	"road":           true,
	"service":        true,
	"track":          true,
}

// gak ada 1 way dengan multiple road type
func InitGraph(ways []*osm.Way, trafficLightNodeIdMap map[osm.NodeID]bool) ([]datastructure.SurakartaWay, []datastructure.Node,
	[]datastructure.SurakartaWay, map[string][2]bool, map[string]datastructure.StreetExtraInfo) {
	var SurakartaNodeMap = make(map[int64]*datastructure.Node)

	oneWayTypesMap := make(map[string]int64)
	twoWayTypesMap := make(map[string]int64)

	bar := progressbar.NewOptions(len(ways),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()), //you should install "github.com/k0kubun/go-ansi"
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][2/7][reset] saving Openstreetmap way & node  ..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	surakartaWays := []datastructure.SurakartaWay{}
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

		sWay := datastructure.SurakartaWay{
			Nodes: make([]datastructure.CHNode2, 0),
		}

		streetNodeLats := []float64{}
		streetNodeLon := []float64{}

		// creategraph node
		for i := 0; i < len(way.Nodes); i++ {
			currNode := way.Nodes[i]

			node := &datastructure.Node{
				Lat:          currNode.Lat,
				Lon:          currNode.Lon,
				ID:           int64(currNode.ID),
				StreetName:   namaJalan,
				TrafficLight: trafficLightNodeIdMap[currNode.ID],
			}

			if fromRealNode, ok := SurakartaNodeMap[node.ID]; ok {
				node = fromRealNode
				node.UsedInRoad += 1
			} else {
				node.UsedInRoad = 1
				SurakartaNodeMap[node.ID] = node
			}
			

			// add node ke surakartaway
			sWay.Nodes = append(sWay.Nodes, datastructure.CHNode2{
				Lat: node.Lat,
				Lon: node.Lon,
			})
			// append node lat & node lon
			streetNodeLats = append(streetNodeLats, node.Lat)
			streetNodeLon = append(streetNodeLon, node.Lon)

		}
		sort.Float64s(streetNodeLats)
		sort.Float64s(streetNodeLon)

		// https://www.movable-type.co.uk/scripts/latlong.html
		centerLat, centerLon := guidance.MidPoint(streetNodeLats[0], streetNodeLon[0], streetNodeLats[len(streetNodeLats)-1], streetNodeLon[len(streetNodeLon)-1])
		sWay.CenterLoc = []float64{centerLat, centerLon}

		sWay.ID = int32(wayIDx)
		surakartaWays = append(surakartaWays, sWay)
		bar.Add(1)
	}

	WriteWayTypeToCsv(oneWayTypesMap, "onewayTypes.csv")
	WriteWayTypeToCsv(twoWayTypesMap, "twoWayTypes.csv")

	surakartaNodes, surakartaWays, graphEdges, streetDirections, streetExtraInfo := processOnlyIntersectionRoadNodes(SurakartaNodeMap, ways, surakartaWays)

	fmt.Println("")
	return surakartaWays, surakartaNodes, graphEdges, streetDirections, streetExtraInfo
}

func processOnlyIntersectionRoadNodes(nodeMap map[int64]*datastructure.Node, ways []*osm.Way, surakartaWays []datastructure.SurakartaWay) ([]datastructure.Node,
	[]datastructure.SurakartaWay, []datastructure.SurakartaWay, map[string][2]bool, map[string]datastructure.StreetExtraInfo) {
	surakartaNodes := []datastructure.Node{}
	alreadyAdded := make(map[int64]struct{})
	intersectionNodes := []int64{}
	streetDirection := make(map[string][2]bool)

	streetExtraInfo := make(map[string]datastructure.StreetExtraInfo)
	for wayIDx, way := range ways {
		maxSpeed, isOneWay, reversedOneWay, roadType, namaJalan, jumlahLanes, roadclassLink, roundabout, streetInfo := getMaxspeedOneWayRoadType(*way)
		if !isOsmWayUsedByCars(way.TagMap()) {
			continue
		}

		streetExtraInfo[namaJalan] = datastructure.StreetExtraInfo{
			Destination:      streetInfo.Destination,
			DestinationRef:   streetInfo.Destination,
			MotorwayJunction: streetInfo.MotorwayJunction}

		if !ValidRoadType[roadType] {
			continue
		}
		currSurakartaWay := &surakartaWays[wayIDx]

		startIdx := 0
		var from *datastructure.Node = nil
		for ; startIdx < len(way.Nodes); startIdx++ {
			curr := nodeMap[int64(way.Nodes[startIdx].ID)]
			if curr.UsedInRoad >= 2 {
				from = curr
				break
			}
		}
		
		if from == nil {
			continue
		}

		if _, ok := alreadyAdded[from.ID]; !ok {
			intersectionNodes = append(intersectionNodes, from.ID)
			alreadyAdded[from.ID] = struct{}{}
		}
		for i := startIdx + 1; i < len(way.Nodes); i++ {
			currNode := way.Nodes[i]
			// idnya masih pake id osm
			to := nodeMap[int64(currNode.ID)]

			if to.UsedInRoad >= 2 {

				// nodenya ada di intersection of 2  or more roads

				// add edge antara dua node intersection
				fromLoc := geo.NewLocation(from.Lat, from.Lon)
				toLoc := geo.NewLocation(to.Lat, to.Lon)
				fromToDistance := geo.HaversineDistance(fromLoc, toLoc) * 1000 // meter
				if isOneWay && !reversedOneWay {
					edge := datastructure.Edge{
						From:          from,
						To:            to,
						Cost:          fromToDistance,
						MaxSpeed:      maxSpeed,
						StreetName:    namaJalan,
						RoadClass:     roadType,
						RoadClassLink: roadclassLink,
						Lanes:         jumlahLanes,
						Roundabout:    roundabout,
					}
					from.Out_to = append(from.Out_to, edge)
					currSurakartaWay.IntersectionNodesID = append(currSurakartaWay.IntersectionNodesID, from.ID)
					streetDirection[namaJalan] = [2]bool{true, false}
				} else if isOneWay && reversedOneWay {
					reverseEdge := datastructure.Edge{
						From:          to,
						To:            from,
						Cost:          fromToDistance,
						MaxSpeed:      maxSpeed,
						StreetName:    namaJalan,
						RoadClass:     roadType,
						RoadClassLink: roadclassLink,
						Lanes:         jumlahLanes,
						Roundabout:    roundabout,
					}
					to.Out_to = append(to.Out_to, reverseEdge)
					currSurakartaWay.IntersectionNodesID = append(currSurakartaWay.IntersectionNodesID, to.ID)
					streetDirection[namaJalan] = [2]bool{false, true}
				} else {
					edge := datastructure.Edge{
						From:          from,
						To:            to,
						Cost:          fromToDistance,
						MaxSpeed:      maxSpeed,
						StreetName:    namaJalan,
						RoadClass:     roadType,
						RoadClassLink: roadclassLink,
						Lanes:         jumlahLanes,
						Roundabout:    roundabout,
					}
					from.Out_to = append(from.Out_to, edge)

					reverseEdge := datastructure.Edge{
						From:          to,
						To:            from,
						Cost:          fromToDistance,
						MaxSpeed:      maxSpeed,
						StreetName:    roadType,
						RoadClass:     roadType,
						RoadClassLink: roadclassLink,
						Lanes:         jumlahLanes,
						Roundabout:    roundabout,
					}
					to.Out_to = append(to.Out_to, reverseEdge)
					currSurakartaWay.IntersectionNodesID = append(currSurakartaWay.IntersectionNodesID, from.ID)
					currSurakartaWay.IntersectionNodesID = append(currSurakartaWay.IntersectionNodesID, to.ID)
					streetDirection[namaJalan] = [2]bool{true, true}
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
	graphEdges := []datastructure.SurakartaWay{}
	for _, way := range surakartaWays {

		if len(way.IntersectionNodesID) > 0 {
			graphEdges = append(graphEdges, way)
		}
	}

	return surakartaNodes, surakartaWays, graphEdges, streetDirection, streetExtraInfo
}

func getMaxspeedOneWayRoadType(way osm.Way) (float64, bool, bool, string, string, int, string, bool, datastructure.StreetExtraInfo) {
	jumlahLanes := 0

	maxSpeed := 50.0

	isOneWay := false // 0, 1
	reversedOneWay := false

	roadTypes := make(map[string]int)
	namaJalan := ""

	roadType := ""
	roadClassLink := ""
	roundAbout := false
	streetInfo := datastructure.StreetExtraInfo{}

	for _, tag := range way.Tags {
		if tag.Key == "highway" && !strings.Contains(tag.Value, "link") {
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

		if tag.Key == "highway" && strings.Contains(tag.Value, "link") {
			roadClassLink = tag.Value
		}

		if strings.Contains(tag.Value, "roundabout") {
			roundAbout = true
		}

		if tag.Key == "name" {
			namaJalan = tag.Value
		}

		if tag.Key == "lanes" {
			jumlahLanes, _ = strconv.Atoi(tag.Value)
		}
		if tag.Key == "destination" {
			streetInfo.Destination = tag.Value
		}
		if tag.Key == "destination:ref" {
			streetInfo.DestinationRef = tag.Value
		}
		if (tag.Key == "highway" && tag.Value == "motorway") || (tag.Key == "highway" && tag.Value == "motorway_link") {
			for _, tag2 := range way.Tags {
				if tag2.Key == "highway" && tag2.Value == "motorway_junction" {
					streetInfo.MotorwayJunction = tag.Value
				}
			}
		}

	}
	if maxSpeed == 50.0 || maxSpeed == 0 {
		maxSpeed = datastructure.RoadTypeMaxSpeed(roadType)
	}
	return maxSpeed, isOneWay, reversedOneWay, roadType, namaJalan, jumlahLanes, roadClassLink, roundAbout, streetInfo
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
