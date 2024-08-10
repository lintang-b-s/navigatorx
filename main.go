package main

import (
	"flag"
	"fmt"
	"lintang/navigatorx/alg"
	"lintang/navigatorx/api"
	_ "lintang/navigatorx/docs"
	"lintang/navigatorx/service"
	"log"
	"net/http"
	"runtime"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "net/http/pprof"

	"github.com/cockroachdb/pebble"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	listenAddr = flag.String("listenaddr", ":5000", "server listen address")
	mapFile    = flag.String("f", "solo_jogja.osm.pbf", "openstreeetmap file buat road network graphnya")
)

//	@title			navigatorx lintangbs API
//	@version		1.0
//	@description	simple openstreetmap routing engine in go

//	@contact.name	lintang birda saputra
//	@description 	simple openstreetmap routing engine in go. Using Contraction Hierarchies for preprocessing and Bidirectioanl Dijkstra for shortest path query

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @host		localhost:5000
// @BasePath	/api
// @schemes	http
func main() {
	flag.Parse()
	_, ch, nodeIdxMap, graphEdges := alg.BikinGraphFromOpenstreetmap(*mapFile)

	db, err := pebble.Open("navigatorxDB", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	kvDB := alg.NewKVDB(db)
	defer kvDB.Close()

	go func () {
		kvDB.CreateStreetKV(graphEdges, nodeIdxMap) 
	}()
	
	// alg.BikinRtreeStreetNetwork(graphEdges, ch, nodeIdxMap)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/debug", middleware.Profiler())
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:5000/swagger/doc.json"), //The url pointing to API definition
	))
	ch.KVdb = kvDB
	navigatorSvc := service.NewNavigationService(ch, kvDB)
	api.NavigatorRouter(r, navigatorSvc)

	go func() {
		ch.Contraction()
		ch.AStarGraph = nil
		ch.Ready = true
		runtime.GC()
		runtime.GC() // run garbage collection biar heap size nya ngurang wkwkwk
		fmt.Printf("\n Contraction Hieararchies + Bidirectional Dijkstra Ready!!")
	}()

	log.Fatal(http.ListenAndServe(*listenAddr, r))
}
