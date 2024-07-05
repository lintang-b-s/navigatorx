package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"lintang/coba_osm/alg"
	"lintang/coba_osm/router"
	"lintang/coba_osm/service"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	_ "net/http/pprof"

	"github.com/dhconnelly/rtreego"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

type nodeMapContainer struct {
	mu      sync.Mutex
	nodeMap map[osm.NodeID]*osm.Node
}

func main() {
	// f, err := os.Open("./central_java-latest.osm.pbf")
	surakartaWays := bikinGraphFromOpenstreetmap()
	bikinRtreeStreetNetwork(surakartaWays)
	surakartaWays = surakartaWays[len(surakartaWays)-1:]
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/debug", middleware.Profiler())
	fmt.Println("server started at :3000")

	navigatorSvc := service.NewNavigationService()
	router.NavigatorRouter(r, navigatorSvc)

	http.ListenAndServe(":3000", r)
}

func bikinRtreeStreetNetwork(ways []alg.SurakartaWay) {
	for idx, way := range ways {
		if idx%50000 == 0 {
			fmt.Println("membuat rtree entry untuk way ke: " + fmt.Sprint(idx))
		}
		alg.StRTree.Insert(&alg.StreetRect{Location: rtreego.Point{way.CenterLoc[0], way.CenterLoc[1]},
			Wormhole: nil,
			Street:   &way})
	}

}

// gak bisa simpen rtreenya ke file binary (udah coba)

func bikinGraphFromOpenstreetmap() []alg.SurakartaWay {
	f, err := os.Open("./solo_semarang_jogja_hg_oneway.osm.pbf")

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
	waysNodeID := []osm.NodeID{}
	ways := []*osm.Way{}

	someNodes := [][]float64{}

	for scanner.Scan() {
		o := scanner.Object()
		// do something
		tipe := o.ObjectID().Type()
		typeSet[tipe] = typeSet[tipe] + 1
		if count%50000 == 0 {
			fmt.Println("memproses openstreetmap way ke : " + fmt.Sprint(count))
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

	for key, val := range typeSet {
		fmt.Println(string(key) + " val : " + fmt.Sprint(val))
	}

	scanErr := scanner.Err()
	if scanErr != nil {
		panic(scanErr)
	}
	fmt.Println(count)

	fmt.Println("some nodes...")
	for n := range someNodes {
		fmt.Println(someNodes[n])
	}

	fmt.Println("some way...")
	for w := range waysNodeID {
		nID := waysNodeID[w]
		n := ctr.nodeMap[nID]
		fmt.Println(n.Lat, n.Lon)
	}

	fmt.Println("edges di solo: " + fmt.Sprint(someWayCount))
	for idx, way := range ways {
		for i := 0; i < len(way.Nodes); i++ {
			fromNodeID := way.Nodes[i].ID
			ways[idx].Nodes[i].Lat = ctr.nodeMap[fromNodeID].Lat
			ways[idx].Nodes[i].Lon = ctr.nodeMap[fromNodeID].Lon
		}
	}

	surakartaWays := alg.InitGraph(ways)
	NoteWayTypes(ways)
	return surakartaWays
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

	wayTypesArr := make([][]string, len(wayTypesMap)+1+ len(maspeeds))

	idx := 0
	for key, _ := range wayTypesMap {
		tipe := strings.Split(key, "=")
		wayTypesArr[idx] = []string{tipe[0], tipe[1]}
		idx++
	}
	wayTypesArr[idx] = []string{"total", fmt.Sprint(len(ways))}
	idx++;
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

/*
// filterWay := func(w *osm.Way) bool {

// 	wayNodeID := w.Nodes[0].ID
// 	ctr.mu.Lock()
// 	node, ok := ctr.nodeMap[wayNodeID]
// 	ctr.mu.Unlock()
// 	if !ok {
// 		return true
// 	}
// 	// -7.59534660992167, 110.71172965801168
// 	// -7.513164411569153, 110.84449817651627

// 	if node.Lat < -7.59534660992167 || node.Lon < 110.71172965801168 ||
// 		node.Lat > -7.513164411569153 || node.Lon > 110.84449817651627 {
// 		return true
// 	}

// 	if someWayCount < 10 {
// 		waysNodeID = append(waysNodeID, wayNodeID)
// 	}

// 	ways = append(ways, *w)
// 	someWayCount++
// 	return false
// }
// scanner.FilterWay = filterWay
// filterway ku komen gak ada
// filter

// filterNode := func(n *osm.Node) bool {
// 	if n.Lat < -7.59534660992167 || n.Lon < 110.71172965801168 ||
// 		n.Lat > -7.513164411569153 || n.Lon > 110.84449817651627 {
// 		return true
// 	}
// 	return false
// }
// scanner.FilterNode = filterNode
// nodeMap := make(map[osm.NodeID]*osm.Node)

*/
