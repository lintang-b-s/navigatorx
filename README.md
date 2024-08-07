# navigatorx


## Quick Start
### Shortest Path Between 2 Place in Openstreetmap
```
1. download file openstreetmap pbf di: https://drive.google.com/file/d/1pEHN8wwUbB5XpuYMZm141fXQ_ZsIf4CO/view?usp=sharing
2. taruh hasil download ke root project ini
3. go run main.go
(Minimal free ram 3 GB buat data diatas)
4. request ke shortest path
curl --location 'http://localhost:5000/api/navigations/shortestPath' \
--header 'Content-Type: application/json' \
--data '{
    "src_lat": -7.550261232598317,
    "src_lon":    110.78210790296636, 
    "dst_lat": -8.024431446370416,
    "dst_lon":   110.32971396395838
}'

Note: Source  & Destination Coordinate harus tempat di sekitaran provinsi yogyakarta/kota surakarta/klaten
5. copy polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. Rute terpendek akan tampil di peta :) 
```

### Hidden Markov Map Matching 
based on https://www.microsoft.com/en-us/research/publication/hidden-markov-map-matching-noise-sparseness/
```
1. request ke server dg data rute list of coordinate (noisy, anggap data gps)
curl --location 'http://localhost:5000/api/navigations/mapMatching' \
--header 'Content-Type: application/json' -d @gps_hmm_map_matching.json

2. copy polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. hasil map matching berupa list of road network node coordinate akan muncul di peta :)
```


#### Theory / Ref
```
https://jlazarsfeld.github.io/ch.150.project/sections/7-ch-overview/
https://www.uber.com/en-ID/blog/engineering-routing-engine/
https://www.uber.com/en-ID/blog/h3/
https://www.microsoft.com/en-us/research/publication/hidden-markov-map-matching-noise-sparseness/
https://www.uber.com/blog/mapping-accuracy-with-catchme/
http://theory.stanford.edu/~amitp/GameProgramming/ImplementationNotes.html
```


go build -gcflags "-m -l" *.go
