package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"lintang/navigatorx/alg"
	"lintang/navigatorx/router"
	"lintang/navigatorx/service"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	_ "net/http/pprof"

	"github.com/dhconnelly/rtreego"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/k0kubun/go-ansi"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/schollz/progressbar/v3"
)

type nodeMapContainer struct {
	nodeMap map[osm.NodeID]*osm.Node
	mu      sync.Mutex
}

func main() {
	surakartaWays, ch, nodeIdxMap := bikinGraphFromOpenstreetmap()
	bikinRtreeStreetNetwork(surakartaWays, ch, nodeIdxMap)
	surakartaWays = nil

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/debug", middleware.Profiler())

	navigatorSvc := service.NewNavigationService(ch)
	router.NavigatorRouter(r, navigatorSvc)

	go func() {
		ch.Contraction()
		ch.AStarGraph = nil
		ch.Ready = true
		runtime.GC()
		runtime.GC() // run garbage collection biar heap size nya ngurang wkwkwk

		fmt.Println("Contraction Hieararchies + Bidirectional Dijkstra Ready!!")
	}()

	fmt.Println("A* Ready!!")
	fmt.Println("server started at :5000")
	err := http.ListenAndServe(":5000", r)
	fmt.Println(err)
}

// gak bisa simpen rtreenya ke file binary (udah coba)
func bikinRtreeStreetNetwork(ways []alg.SurakartaWay, ch *alg.ContractedGraph, nodeIdxMap map[int64]int32) {
	bar := progressbar.NewOptions(len(ways),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][4/6][reset] Membuat rtree entry dari osm way/edge ..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	for i := range len(ways) {
		way := ways[i]
		for j := range len(way.NodesID) {
			way.NodesID[j] = int64(nodeIdxMap[way.NodesID[j]]) // harus int64 (osm.NodeId)
		}
	}

	rtg := rtreego.NewTree(2, 25, 50) // 2 dimension, 25 min entries dan 50 max entries
	rt := alg.NewRtree(rtg)
	for _, way := range ways {
		rt.StRtree.Insert(&alg.StreetRect{Location: rtreego.Point{float64(way.CenterLoc[0]), float64(way.CenterLoc[1])},
			Wormhole: nil,
			Street:   &way})
		bar.Add(1)
	}
	fmt.Println("")
	ch.Rtree = rt
}

func bikinGraphFromOpenstreetmap() ([]alg.SurakartaWay, *alg.ContractedGraph, map[int64]int32) {
	f, err := os.Open("./solo_jogja.osm.pbf") // sololama.osm.pbf

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
		progressbar.OptionSetDescription("[cyan][1/6][reset] memproses openstreetmap way..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	for scanner.Scan() {
		o := scanner.Object()
		// do something
		tipe := o.ObjectID().Type()
		typeSet[tipe] = typeSet[tipe] + 1
		if count%50000 == 0 {
			bar.Add(50000)
		}
		if tipe == osm.TypeNode {
			ctr.nodeMap[o.(*osm.Node).ID] = o.(*osm.Node)
		}

		if tipe == osm.TypeWay {
			ways = append(ways, o.(*osm.Way))
			someWayCount++
		}
		count++
	}
	fmt.Println("")

	scanErr := scanner.Err()
	if scanErr != nil {
		panic(scanErr)
	}
	fmt.Println("jumlah osm object di area sekitar solo,semarang,jogja: " + fmt.Sprint(count))

	trafficLightNodeMap := make(map[string]int64)
	var trafficLightNodeIDMap = make(map[osm.NodeID]bool)

	fmt.Println("jumlah edges/way di area sekitar solo,semarang,jogja: " + fmt.Sprint(someWayCount))
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

	surakartaWays, surakartaNodes := alg.InitGraph(ways, trafficLightNodeIDMap)
	ch := alg.NewContractedGraph()
	nodeIdxMap := ch.InitCHGraph(surakartaNodes, len(ways))

	surakartaNodes = nil

	fmt.Println("")
	NoteWayTypes(ways)

	alg.WriteWayTypeToCsv(trafficLightNodeMap, "traffic_light_node.csv")

	return surakartaWays, ch, nodeIdxMap
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
