package main

import (
	"fmt"
	"lintang/navigatorx/alg"
	"lintang/navigatorx/router"
	"lintang/navigatorx/service"
	"log"
	"net/http"
	"runtime"

	_ "net/http/pprof"

	"github.com/cockroachdb/pebble"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	_, ch, nodeIdxMap, graphEdges := alg.BikinGraphFromOpenstreetmap()

	db, err := pebble.Open("navigatorxDB", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	kvDB := alg.NewKVDB(db)
	defer kvDB.Close()

	go func() {
		kvDB.CreateStreetKV(graphEdges, nodeIdxMap)
		runtime.GC()
		runtime.GC()
	}()
	// bikinRtreeStreetNetwork(surakartaWays, ch, nodeIdxMap)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/debug", middleware.Profiler())
	ch.KVdb = kvDB
	navigatorSvc := service.NewNavigationService(ch, kvDB)
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
	err = http.ListenAndServe(":5000", r)
	fmt.Println(err)
}
