package alg

import (
	"fmt"
	"log"
	"math"

	"github.com/cockroachdb/pebble"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
	"github.com/uber/h3-go/v4"
)

type KVDB struct {
	db *pebble.DB
}

func NewKVDB(db *pebble.DB) *KVDB {
	return &KVDB{db}
}

func (k *KVDB) CreateStreetKV(way []SurakartaWay, nodeIDxMap map[int64]int32, listenAddr string) {
	bar := progressbar.NewOptions(len(way),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][4/7][reset] Membuat h3 index untuk osm street..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	kv := make(map[string][]SurakartaWay)
	for i, w := range way {
		street := way[i]
		centerWayLat := w.CenterLoc[0]
		centerWayLon := w.CenterLoc[1]

		h3LatLon := h3.NewLatLng(centerWayLat, centerWayLon)
		cell := h3.LatLngToCell(h3LatLon, 9)

		kv[cell.String()] = append(kv[cell.String()], street)

		bar.Add(1)
	}

	fmt.Println("")
	bar = progressbar.NewOptions(len(kv),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][5/7][reset] Menyimpan h3 indexed street ke pebble db..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	workers := NewWorkerPool[SaveWayJobItem, interface{}](4, len(kv))

	for keyStr, valArr := range kv {
		workers.AddJob(SaveWayJobItem{keyStr, valArr})
		bar.Add(1)
	}
	close(workers.jobQueue)

	workers.Start(k.SaveWay)
	workers.Wait()

	fmt.Println("")
	fmt.Printf("A* Ready!!\n")
	fmt.Printf("\nserver started at %v\n", listenAddr)
	fmt.Println("")
}

type SaveWayJobItem struct {
	KeyStr string
	ValArr []SurakartaWay
}

func (k *KVDB) SaveWay(wayItem SaveWayJobItem) interface{} {
	keyStr := wayItem.KeyStr
	valArr := wayItem.ValArr
	key := []byte(keyStr)
	val, err := CompressWay(valArr)

	if err != nil {
		log.Fatal(err)
	}
	if err := k.db.Set(key, val, pebble.Sync); err != nil {
		log.Fatal(err)
	}
	return nil
}

// GetNearestStreetsFromPointCoord buat road snaping. Untuk menentukan jalan-jalan yang paling dekat dengan titik start/end rute yang di tunjuk sama pengguna
func (k *KVDB) GetNearestStreetsFromPointCoord(lat, lon float64) ([]SurakartaWay, error) {
	ways := []SurakartaWay{}

	home := h3.NewLatLng(lat, lon)
	cell := h3.LatLngToCell(home, 9)
	val, closer, err := k.db.Get([]byte(cell.String()))
	if err == nil {
		defer closer.Close()
	}

	streets, _ := LoadWay(val)

	ways = append(ways, streets...)

	// cells := h3.GridDisk(cell, 1)
	cells := kRingIndexesArea(lat, lon, 1) // search cell neighbor dari homeCell yang radius nya 1 km 
	for _, currCell := range cells {
		if currCell == cell {
			continue
		}
		val, closer, err := k.db.Get([]byte(currCell.String()))
		if closer == nil || err != nil {
			continue
		}

		streets, err := LoadWay(val)
		if err != nil {
			return []SurakartaWay{}, err
		}
		ways = append(ways, streets...)
		closer.Close()
	}

	// kalau dari radius 1 km dari titik start/end rute pengguna gak ada jalan (misal di bandara, hutan, dll). maka cari jalan dari neighbor h3 cell yang lebih jauh dari titik start/end rute
	for lev := 2; lev <= 10; lev++ {
		if len(ways) == 0 {
			cells := h3.GridDisk(cell, lev)
			for _, currCell := range cells {
				if currCell == cell {
					continue
				}
				val, closer, err := k.db.Get([]byte(currCell.String()))
				if closer == nil || err != nil {
					continue
				}

				streets, err := LoadWay(val)
				if err != nil {
					return []SurakartaWay{}, err
				}
				ways = append(ways, streets...)
				closer.Close()
			}
		}
	}

	if len(ways) == 0 {

		return []SurakartaWay{}, fmt.Errorf("tidak ada jalan di sekitar lokasi")
	}

	return ways, nil
}



/*
*
  - https://observablehq.com/@nrabinowitz/h3-radius-lookup?collection=@nrabinowitz/h3
    search cell neighbor dari cell dari lat,lon  yang radius nya = searchRadiusKm
*/
func kRingIndexesArea(lat, lon, searchRadiusKm float64) []h3.Cell {
	home := h3.NewLatLng(lat, lon)
	origin := h3.LatLngToCell(home, 9)
	originArea := h3.CellAreaKm2(origin)
	searchArea := math.Pi * searchRadiusKm * searchRadiusKm

	radius := 0
	diskArea := originArea

	for diskArea < searchArea {
		radius++
		cellCount := float64(3*radius*(radius+1) + 1)
		diskArea = cellCount * originArea
	}

	return h3.GridDisk(origin, radius)
}

func (k *KVDB) Close() {
	k.db.Close()
}
