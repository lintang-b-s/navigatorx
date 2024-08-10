# navigatorx

## Quick Start

### Shortest Path Between 2 Place in Openstreetmap

```
1. download file openstreetmap pbf di: https://drive.google.com/file/d/1pEHN8wwUbB5XpuYMZm141fXQ_ZsIf4CO/view?usp=sharing
2. taruh hasil download ke root project ini
3. go run main.go
(Minimal free ram 1 GB buat data diatas)
4. tunggu sampai ada log "server started at :5000". Jika anda ingin query nya >10x lipat lebih cepat tunggu preprocessing Contraction Hierarchies nya selesai.
5. request ke shortest path (source=surakarta , destination=pantai parangtritis)
curl --location 'http://localhost:5000/api/navigations/shortestPath' \
--header 'Content-Type: application/json' \
--data '{
    "src_lat": -7.550261232598317,
    "src_lon":    110.78210790296636,
    "dst_lat": -8.024431446370416,
    "dst_lon":   110.32971396395838
}'

Note: Source  & Destination Coordinate harus tempat di sekitaran provinsi yogyakarta/kota surakarta/klaten
5. copy path polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. Rute terpendek akan tampil di peta :)
```

### Hidden Markov Map Matching

based on https://www.microsoft.com/en-us/research/publication/hidden-markov-map-matching-noise-sparseness/

```
1.tunggu sampai ada log "server started at :5000". Jika anda ingin query nya >10x lipat lebih cepat tunggu preprocessing Contraction Hierarchies nya selesai.
2. request ke server dg data rute list of coordinate (noisy, anggap data gps)
curl --location 'http://localhost:5000/api/navigations/mapMatching' \
--header 'Content-Type: application/json' -d @gps_hmm_map_matching.json

3. copy path polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. hasil map matching berupa list of road network node coordinate akan muncul di peta :)
```

### Traveling Salesman Problem Using Simulated Annealing & Bidirectional Dijkstra CH
Apa Rute terpendek untuk mengunjungi kampus UGM,UNY,UPNV Jogja, UII Jogja, IAIN Surakarta, UNS, UMS, dan ISI Surakarta tepat sekali?
```
1. Tunggu sampai preprocessing Contraction Hierarchies Selesai 
2. request query traveling salesman problem
curl --location 'http://localhost:5000/api/navigations/travelingSalesmanProblem' \
--header 'Content-Type: application/json' \
--data '{
    "cities_coord": [
        {
            "lat": -7.773700556142326, 
            "lon": 110.37927594982729
        },
        {
            "lat": -7.687798280189743,
            "lon": 110.41397147030537
        },
        {
            "lat": -7.773714842796234, 
            "lon": 110.38625612460329
        },
        {
            "lat": -7.7620859704046135, 
            "lon": 110.40928883503045
        },
        {
            "lat": -7.559256385020671,
            "lon":  110.85624887436603
        },
        {
            "lat": -7.558529640984029,
            "lon": 110.73442218529993
        },
        {
            "lat": -7.5579561088085665,
            "lon":  110.85233572375333
        },
        {
            "lat":  -7.557649260722883, 
            "lon": 110.77068956586514
        }
    ]
}'

3.  copy path polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. Rute terpendek tsp akan tampil di peta :)
```

### Many to Many Shortest Path Query

```
1. tunggu sampai preprocessing contraction hierarchies selesai
2. request  query many to many
curl --location 'http://localhost:5000/api/navigations/manyToManyQuery' \
--header 'Content-Type: application/json' \
--data '{
    "sources": [{
        "lat": -7.550248257898637,
        "lon": 110.78217903249168
    },
    {
        "lat": -7.560347382387681,
        "lon": 110.78879587509478
    },
    {
        "lat": -7.5623445763181945,
        "lon": 110.81010426983109
    }
    ],
    "targets": [{
        "lat": -7.553672205152498,
        "lon": 110.79784256968716
    },
    {
        "lat": -7.564559782091322,
        "lon":  110.80455609811008
    },
    {
        "lat": -7.570135257838102,
        "lon": 110.82292649269334
    },
    {
        "lat": -7.598393719179397,
        "lon": 110.81555588473815
    }


    ]
}'

3.  copy path polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. Rute terpendek many to many query akan tampil di peta :)
```

### Shortest Path with alternative street

```
1. tunggu sampai ada log "server started at :5000". Jika anda ingin query nya >10x lipat lebih cepat tunggu preprocessing Contraction Hierarchies nya selesai.
2. request query shortest path w/ alternative street
curl --location 'http://localhost:5000/api/navigations/shortestPathAlternativeStreet' \
--header 'Content-Type: application/json' \
--data '{
    "src_lat": -7.550261232598317,
    "src_lon":    110.78210790296636,
    "street_alternative_lat": -7.8409667827395815,
    "street_alternative_lon":   110.3472473375829,
      "dst_lat": -8.024431446370416,
    "dst_lon":   110.32971396395838
}'
3. copy path polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. Rute terpendek akan tampil di peta :)
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

go build -gcflags "-m -l" \*.go
