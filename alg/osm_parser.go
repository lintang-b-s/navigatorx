package alg

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/k0kubun/go-ansi"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/schollz/progressbar/v3"
)

type nodeMapContainer struct {
	nodeMap map[osm.NodeID]*osm.Node
	mu      sync.Mutex
}

func BikinGraphFromOpenstreetmap() ([]SurakartaWay, *ContractedGraph, map[int64]int32, []SurakartaWay) {
	f, err := os.Open("./solo_jogja.osm.pbf")

	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := osmpbf.New(context.Background(), f, 3)
	defer scanner.Close()

	count := 0
	var typeSet = make(map[osm.Type]int)

	ctr := nodeMapContainer{
		nodeMap: make(map[osm.NodeID]*osm.Node),
	}

	someWayCount := 0
	ways := []*osm.Way{}

	bar := progressbar.NewOptions(450000,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()), //you should install "github.com/k0kubun/go-ansi"
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][1/7][reset] memproses openstreetmap way..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	nodeCount := 0
	wayNodesMap := make(map[osm.NodeID]bool)
	for scanner.Scan() {
		o := scanner.Object()
		// do something
		tipe := o.ObjectID().Type()
		typeSet[tipe] = typeSet[tipe] + 1
		if count%50000 == 0 {
			bar.Add(50000)
		}

		if tipe != "way" {
			continue
		}
		tag := o.(*osm.Way).TagMap()
		if !isOsmWayUsedByCars(tag) {
			continue
		}

		if tipe == osm.TypeWay {
			ways = append(ways, o.(*osm.Way))
			someWayCount++
			for _, node := range o.(*osm.Way).Nodes {
				wayNodesMap[node.ID] = true
			}
		}
		count++
	}

	f.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	scanner = osmpbf.New(context.Background(), f, 3)
	defer scanner.Close()

	for scanner.Scan() {
		o := scanner.Object()
		if o.ObjectID().Type() == osm.TypeNode {
			node := o.(*osm.Node)
			if _, ok := wayNodesMap[node.ID]; ok {
				ctr.nodeMap[o.(*osm.Node).ID] = o.(*osm.Node)
				nodeCount++
			}
		}
	}

	fmt.Println("")

	scanErr := scanner.Err()
	if scanErr != nil {
		panic(scanErr)
	}
	fmt.Println("jumlah osm nodes: " + fmt.Sprint(nodeCount))

	trafficLightNodeMap := make(map[string]int64)
	var trafficLightNodeIDMap = make(map[osm.NodeID]bool)

	fmt.Println("jumlah osm way: " + fmt.Sprint(someWayCount))
	for idx, way := range ways {
		for i := 0; i < len(way.Nodes); i++ {
			fromNodeID := way.Nodes[i].ID
			ways[idx].Nodes[i].Lat = ctr.nodeMap[fromNodeID].Lat
			ways[idx].Nodes[i].Lon = ctr.nodeMap[fromNodeID].Lon
			for _, tag := range ctr.nodeMap[fromNodeID].Tags {
				if strings.Contains(tag.Value, "traffic_signals") {
					trafficLightNodeMap[tag.Key+"="+tag.Value]++
					trafficLightNodeIDMap[way.Nodes[i].ID] = true
				}
			}
		}
	}

	ctr.nodeMap = nil
	runtime.GC()
	runtime.GC()
	surakartaWays, surakartaNodes, graphEdges := InitGraph(ways, trafficLightNodeIDMap)
	ch := NewContractedGraph()
	nodeIdxMap := ch.InitCHGraph(surakartaNodes, len(ways))
	convertOSMNodeIDToGraphID(surakartaWays, nodeIdxMap)
	surakartaNodes = nil

	fmt.Println("")
	NoteWayTypes(ways)

	WriteWayTypeToCsv(trafficLightNodeMap, "traffic_light_node.csv")

	return surakartaWays, ch, nodeIdxMap, graphEdges
}

func convertOSMNodeIDToGraphID(surakartaWays []SurakartaWay, nodeIDxMap map[int64]int32) {
	for i := range surakartaWays {
		way := &surakartaWays[i]
		for i, nodeID := range way.IntersectionNodesID {

			way.IntersectionNodesID[i] = int64(nodeIDxMap[nodeID])
		}
	}
}

// https://github.com/RoutingKit/RoutingKit/blob/master/src/osm_profile.cpp  [is_osm_way_used_by_cars()]
func isOsmWayUsedByCars(tagMap map[string]string) bool {
	_, ok := tagMap["junction"]
	if ok {
		return true
	}

	route, ok := tagMap["route"]
	if ok && route == "ferry" {
		return true
	}

	ferry, ok := tagMap["ferry"]
	if ok && ferry == "yes" {
		return true
	}

	highway, okHW := tagMap["highway"]
	if !okHW {
		return false
	}

	motorcar, ok := tagMap["motorcar"]
	if ok && motorcar == "no" {
		return false
	}

	motorVehicle, ok := tagMap["motor_vehicle"]
	if ok && motorVehicle == "no" {
		return false
	}

	access, ok := tagMap["access"]
	if ok {
		if !(access == "yes" || access == "permissive" || access == "designated" || access == "delivery" || access == "destination") {
			return false
		}
	}

	if okHW && (highway == "motorway" ||
		highway == "trunk" ||
		highway == "primary" ||
		highway == "secondary" ||
		highway == "tertiary" ||
		highway == "unclassified" ||
		highway == "residential" ||
		highway == "living_street" ||
		highway == "service" ||
		highway == "motorway_link" ||
		highway == "trunk_link" ||
		highway == "primary_link" ||
		highway == "secondary_link" ||
		highway == "tertiary_link") {
		return true
	}

	if highway == "bicycle_road" {
		motorcar, ok := tagMap["motorcar"]
		if ok {
			if motorcar == "yes" {
				return true
			}
		}
		return false
	}

	if highway == "construction" ||
		highway == "path" ||
		highway == "footway" ||
		highway == "cycleway" ||
		highway == "bridleway" ||
		highway == "pedestrian" ||
		highway == "bus_guideway" ||
		highway == "raceway" ||
		highway == "escape" ||
		highway == "steps" ||
		highway == "proposed" ||
		highway == "conveying" {
		return false
	}

	oneway, ok := tagMap["oneway"]
	if ok {
		if oneway == "reversible" || oneway == "alternating" {
			return false
		}
	}

	_, ok = tagMap["maxspeed"]
	if ok {
		return true
	}

	return false
}

func NoteWayTypes(ways []*osm.Way) {

	wayTypesMap := make(map[string]bool)

	maspeeds := make(map[string]int)

	for _, way := range ways {
		for _, wayTag := range way.Tags {
			if !wayTypesMap[wayTag.Key+"="+wayTag.Value] {
				wayTypesMap[wayTag.Key+"="+wayTag.Value] = true
				if strings.Contains(wayTag.Key, "maxspeed") {
					maspeeds[wayTag.Value]++
				}
			}
		}
	}

	wayTypesArr := make([][]string, len(wayTypesMap)+1+len(maspeeds))

	idx := 0
	for key, _ := range wayTypesMap {
		tipe := strings.Split(key, "=")
		wayTypesArr[idx] = []string{tipe[0], tipe[1]}
		idx++
	}
	wayTypesArr[idx] = []string{"total", fmt.Sprint(len(ways))}
	idx++
	for key, val := range maspeeds {
		wayTypesArr[idx] = []string{"maxspeed=" + key, fmt.Sprint(val)}
		idx++
	}

	file, err := os.Create("wayTypes.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.WriteAll(wayTypesArr)

}
