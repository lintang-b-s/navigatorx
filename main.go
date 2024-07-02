package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"lintang/coba_osm/alg"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/gojek/heimdall/httpclient"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

type nodeMapContainer struct {
	mu      sync.Mutex
	nodeMap map[osm.NodeID]*osm.Node
}

type SurakartaGraph struct {
	Nodes   []*alg.Node
	NodeIdx map[int64]int64
	Counter int64
}

var surakartaGraph = SurakartaGraph{
	Nodes:   make([]*alg.Node, 0),
	NodeIdx: make(map[int64]int64),
	Counter: 0,
}

var surakartaNodeMap = make(map[int64]*alg.Node)

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func main() {
	// f, err := os.Open("./central_java-latest.osm.pbf")
	f, err := os.Open("./solo.osm.pbf")

	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := osmpbf.New(context.Background(), f, 3)
	defer scanner.Close()

	count := 0
	var typeSet = make(map[osm.Type]int)

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
	ctr := nodeMapContainer{
		nodeMap: make(map[osm.NodeID]*osm.Node),
	}

	someWayCount := 0
	waysNodeID := []osm.NodeID{}
	ways := []*osm.Way{}

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

	someNodes := [][]float64{}

	someNodeCount := 0

	for scanner.Scan() {
		o := scanner.Object()
		// do something
		tipe := o.ObjectID().Type()
		typeSet[tipe] = typeSet[tipe] + 1
		fmt.Println(count)
		if tipe == osm.TypeNode {
			ctr.mu.Lock()
			ctr.nodeMap[o.(*osm.Node).ID] = o.(*osm.Node)
			ctr.mu.Unlock()
		}
		if tipe == osm.TypeNode && someNodeCount < 5 {
			someNodes = append(someNodes, []float64{o.(*osm.Node).Lat, o.(*osm.Node).Lon})
			someNodeCount++
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

	initGraph(ways)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	fmt.Println("server started at :3000")
	r.Route("/api", func(r chi.Router) {
		r.Post("/shortestPath", shortestPath)
	})

	http.ListenAndServe(":3000", r)
}

func initGraph(ways []*osm.Way) {

	for _, way := range ways {
		for i := 0; i < len(way.Nodes)-1; i++ {
			fromN := way.Nodes[i]

			from := &alg.Node{
				Lat: roundFloat(fromN.Lat, 6),
				Lon: roundFloat(fromN.Lon, 6),
				ID:  int64(fromN.ID),
			}

			toN := way.Nodes[i+1]
			to := &alg.Node{
				Lat: roundFloat(toN.Lat, 6),
				Lon: roundFloat(toN.Lon, 6),
				ID:  int64(toN.ID),
			}

			if fromRealNode, ok := surakartaNodeMap[from.ID]; ok {
				from = fromRealNode
			} else {
				surakartaNodeMap[from.ID] = from
			}
			if toRealNode, ok := surakartaNodeMap[to.ID]; ok {
				to = toRealNode
			} else {
				surakartaNodeMap[to.ID] = to
			}

			edge := alg.Edge{
				From: from,
				To:   to,
				Cost: euclideanDistance(from, to),
			}
			from.Out_to = append(from.Out_to, edge)

			reverseEdge := alg.Edge{
				From: to,
				To:   from,
				Cost: euclideanDistance(from, to),
			}

			to.Out_to = append(to.Out_to, reverseEdge)

			if _, ok := surakartaGraph.NodeIdx[from.ID]; ok {
				fromIdx := surakartaGraph.NodeIdx[from.ID]
				surakartaGraph.Nodes[fromIdx] = from
			} else {
				surakartaGraph.NodeIdx[from.ID] = surakartaGraph.Counter // save index node saat ini
				surakartaGraph.Nodes = append(surakartaGraph.Nodes, from)
				surakartaGraph.Counter++
			}
			if _, ok := surakartaGraph.NodeIdx[to.ID]; ok {
				toIdx := surakartaGraph.NodeIdx[to.ID]
				surakartaGraph.Nodes[toIdx] = to
			} else {
				surakartaGraph.NodeIdx[to.ID] = surakartaGraph.Counter
				surakartaGraph.Nodes = append(surakartaGraph.Nodes, to)
				surakartaGraph.Counter++
			}

		}
	}

}

func euclideanDistance(from *alg.Node, to *alg.Node) float64 {
	var total float64 = 0
	latDif := math.Abs(from.Lat - to.Lat)
	latDifSq := latDif * latDif

	lonDif := math.Abs(from.Lon - to.Lon)
	lonDifSq := lonDif * lonDif

	total += latDifSq + lonDifSq

	return math.Sqrt(total)
}

type SortestPathRequest struct {
	SrcLat float64 `json:"src_lat"`
	SrcLon float64 `json:"src_lon"`
	DstLat float64 `json:"dst_lat"`
	DstLon float64 `json:"dst_lon"`
}

func (s *SortestPathRequest) Bind(r *http.Request) error {
	if s.SrcLat == 0 || s.SrcLon == 0 || s.DstLat == 0 || s.DstLon == 0 {
		return errors.New("invalid request")
	}
	return nil
}

type ShapeReq struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	Type string  `json:"type"`
}

// buat map matching valhalla
type MapMatchingRequest struct {
	Shape      []ShapeReq `json:"shape"`
	Costing    string     `json:"costing"`
	ShapeMatch string     `json:"shape_match"`
}

type MapMatchingResponse struct {
	MatchedPoints []ShapeReq `json:"matched_points"`
}

type ValhallaErrorResp struct {
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

// router handler
func shortestPath(w http.ResponseWriter, r *http.Request) {
	// get from and to node
	// call shortest path
	// return shortest path
	data := &SortestPathRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	from := &alg.Node{
		Lat: roundFloat(data.SrcLat, 6),
		Lon: roundFloat(data.SrcLon, 6),
	}
	to := &alg.Node{
		Lat: roundFloat(data.DstLat, 6),
		Lon: roundFloat(data.DstLon, 6),
	}

	var err error
	from.Lat, from.Lon, err = SnapLocationToRoadNetworkNode(from.Lat, from.Lon)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("internal server error")))
		return
	}
	to.Lat, to.Lon, err = SnapLocationToRoadNetworkNode(to.Lat, to.Lon)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("internal server error")))
		return
	}

	var fromSurakartaNode *alg.Node
	var toSurakartaNode *alg.Node
	for _, n := range surakartaGraph.Nodes {

		for i := 6; i >= 4; i-- {
			if roundFloat(n.Lat, uint(i)) == roundFloat(from.Lat, uint(i)) && roundFloat(n.Lon, uint(i)) == roundFloat(from.Lon, uint(i)) {
				if fromSurakartaNode != nil &&
					math.Abs(n.Lat-from.Lat) > math.Abs(fromSurakartaNode.Lat-from.Lat) &&
					math.Abs(n.Lon-from.Lon) > math.Abs(fromSurakartaNode.Lon-from.Lon) {
					// node graph (n) saat ini gak lebih dekat dg lokasi from
					continue
				}
				fromSurakartaNode = n
				break
			}
		}

		for i := 6; i >= 3; i-- {
			if roundFloat(n.Lat, uint(i)) == roundFloat(to.Lat, uint(i)) && roundFloat(n.Lon, uint(i)) == roundFloat(to.Lon, uint(i)) {
				if toSurakartaNode != nil &&
					math.Abs(n.Lat-to.Lat) > math.Abs(toSurakartaNode.Lat-to.Lat) &&
					math.Abs(n.Lon-to.Lon) > math.Abs(toSurakartaNode.Lon-to.Lon) {
					// node graph (n) saat ini gak lebih dekat dg lokasi to
					continue
				}
				toSurakartaNode = n
				break
			}
		}

	}

	if fromSurakartaNode == nil || toSurakartaNode == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("node not found")))
		return
	}

	p, dist, found := alg.SorthestPath(fromSurakartaNode, toSurakartaNode)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, NewShortestPathResponse(p, dist, found))
}

func SnapLocationToRoadNetworkNode(lat, lon float64) (snappedLat, snappedLon float64, err error) {
	timeout := 2000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	mapMatchBody := MapMatchingRequest{
		Shape: []ShapeReq{
			{
				Lat:  lat,
				Lon:  lon,
				Type: "break",
			},
			{
				Lat:  lat + 0.0000000001,
				Lon:  lon + 0.0000000001,
				Type: "via",
			},
		},
		Costing:    "auto",
		ShapeMatch: "map_snap",
	}
	bodyBytes, _ := json.Marshal(&mapMatchBody)
	reader := bytes.NewReader(bodyBytes)

	res, err := client.Post("http://localhost:8002/trace_attributes?json", reader, http.Header{})

	// res, err := http.Post("http://localhost:8002/trace_attributes?json", "application/json", reader)
	if err != nil {
		err = errors.New("internal server error")
		return
	}
	var errorValhalla = &ValhallaErrorResp{}
	if res.StatusCode == 400 {
		err = json.NewDecoder(res.Body).Decode(errorValhalla)
		if err != nil {
			err = errors.New("internal server error")
			return
		}
	}
	fmt.Println(errorValhalla)
	defer res.Body.Close()

	matchedPoints := &MapMatchingResponse{}
	err = json.NewDecoder(res.Body).Decode(matchedPoints)
	if err != nil {
		err = errors.New("internal server error")
		return
	}

	snappedLat = matchedPoints.MatchedPoints[0].Lat
	snappedLon = matchedPoints.MatchedPoints[0].Lon
	return
}

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
type ShortestPathResponse struct {
	Path  string       `json:"path"`
	Dist  float64      `json:"distance"`
	Found bool         `json:"found"`
	Route []Coordinate `json:"route"`
}

func NewShortestPathResponse(p []alg.Pather, distance float64, found bool) *ShortestPathResponse {
	var route []Coordinate = make([]Coordinate, 0)
	for _, path := range p {
		pathN := path.(*alg.Node)
		route = append(route, Coordinate{
			Lat: pathN.Lat,
			Lon: pathN.Lon,
		})
	}

	return &ShortestPathResponse{
		Path:  alg.RenderPath(p),
		Dist:  distance,
		Found: found,
		Route: route,
	}
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}
