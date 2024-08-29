# navigatorx

Simple Openstreetmap routing engine in go. This project uses Contraction Hierarchies to speed up shortest path queries by preprocessing the road network graph (adding many shortcut edges) and Bidirectional Dijsktra for shortest path queries. H3 is used as a nearest neighbor query.

## Quick Start

#### Docker

```
1. docker compose up  --build
2. wait for preprocessing contraction hierarchies to complete (about 3 minutes) [you can check it with 'docker logs navigatorx -f', wait until the text 'Contraction Hierarchies + Bidirectional Dijkstra Ready!!' appears  ]
```

#### Local

```
1. download the jogja & solo openstreetmap pbf file at: https://drive.google.com/file/d/1pEHN8wwUbB5XpuYMZm141fXQ_ZsIf4CO/view?usp=sharing
Note: or you can also use another openstreetmap file with the osm.pbf format (https://download.geofabrik.de/)
2.  put the download results into the ./bin directory of this project
3.  go mod tidy &&  mkdir bin
4. CGO_ENABLED=1  go build -o ./bin/navigatorx .
5. ./bin/navigatorx
(Minimum free RAM 1 GB for the above data)
note: or you can also do it with "make run"
5.  wait for preprocessing contraction hierarchies to complete (about 3 minutes)
```

#### Change Map Data

```
1. provide openstreetmap filename flag when running the program
 ./bin/navigatorx -f jakarta.osm.pbf
2. for docker setup, change the Filename args "MAP_FILE" and google drive file id "DRIVE_FILE_ID" in docker-compose
 args:
        MAP_FILE: solo_jogja
        DRIVE_FILE_ID: 1pEHN8wwUbB5XpuYMZm141fXQ_ZsIf4CO
```

### Shortest Path Between 2 Place in Openstreetmap

```
1. wait until there is a log "server started at :5000". If you want the query to be >10x faster, wait for the Contraction Hierarchies preprocessing to complete.
2. request ke shortest path (source=surakarta , destination=pantai parangtritis) [untuk data openstreetmap pada step setup]
curl --location 'http://localhost:5000/api/navigations/shortest-path' \
--header 'Content-Type: application/json' \
--data '{
    "src_lat": -7.550261232598317,
    "src_lon":    110.78210790296636,
    "dst_lat": -8.024431446370416,
    "dst_lon":   110.32971396395838
}'

Note: Source & Destination Coordinates must be around Yogyakarta Province/Surakarta City/Klaten if using OpenStreetMap data in the setup step
5. Copy the polyline string path of the response endpoint result to https://valhalla.github.io/demos/polyline . Check Unsescape '\'. The shortest route will appear on the map. :)
```

### Hidden Markov Map Matching

based on https://www.microsoft.com/en-us/research/publication/hidden-markov-map-matching-noise-sparseness/

```
1. wait until there is a log "server started at :5000". If you want the query to be >10x faster, wait for the Contraction Hierarchies preprocessing to complete.
2. request to the server with route list of gps coordinate data (or fake route coordinate data)
curl --location 'http://localhost:5000/api/navigations/map-matching' \
--header 'Content-Type: application/json' -d @gps_hmm_map_matching.json

Note:  GPS Coordinates must be around the Yogyakarta province/Surakarta city/Klaten if using OpenStreetMap data in the setup step
3. Copy the polyline string path of the response endpoint result to https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. The map matching results in the form of a list of road network node coordinates will appear on the map. :)
```

### Traveling Salesman Problem Using Simulated Annealing

What is the shortest (suboptimal) route to visit UGM, UNY, UPNV Jogja, UII Jogja, IAIN Surakarta, UNS, UMS, and ISI Surakarta campuses exactly once and return to the original campus location?

```
1. Wait until Contraction Hierarchies preprocessing is complete
2. request query traveling salesman problem
curl --location 'http://localhost:5000/api/navigations/tsp' \
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
Note:  "cities_coord" must be a place around the province of Yogyakarta/Surakarta/Klaten if using OpenStreetMap data in the setup step
3.  Copy the polyline string path of the response endpoint result to https://valhalla.github.io/demos/polyline . Check Unsescape '\'. The shortest (suboptimal) TSP route will be displayed on the map. :)
```

### Many to Many Shortest Path Query

```
1. wait until preprocessing contraction hierarchies is complete
2. request  query many to many
curl --location 'http://localhost:5000/api/navigations/many-to-many' \
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

Note:  "sources" and "targets" must be around the province of Yogyakarta/Surakarta/Klaten if using OpenStreetMap data in the setup step
3.  Copy the polyline string path of the response endpoint result to https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. Check Unsescape '\'. The shortest route of many to many query will be displayed on the map. :)
```

### Shortest Path with alternative street

```
1. wait until there is a log "server started at :5000". If you want the query to be >10x faster, wait for the Contraction Hierarchies preprocessing to complete.
2. request query shortest path w/ alternative street
curl --location 'http://localhost:5000/api/navigations/shortest-path-alternative-street' \
--header 'Content-Type: application/json' \
--data '{
    "src_lat": -7.550261232598317,
    "src_lon":    110.78210790296636,
    "street_alternative_lat": -7.8409667827395815,
    "street_alternative_lon":   110.3472473375829,
      "dst_lat": -8.024431446370416,
    "dst_lon":   110.32971396395838
}'

Note:  "sources" and "targets" must be around the province of Yogyakarta/Surakarta/Klaten if using OpenStreetMap data in the setup step
3. Copy the polyline string path of the response endpoint result to https://valhalla.github.io/demos/polyline . Check Unsescape '\'. The shortest route will appear on the map. :)
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
