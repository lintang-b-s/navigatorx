# navigatorx


## Quick Start
```
1. download file openstreetmap pbf di: https://drive.google.com/file/d/1u8As3D5pbdO0HBk7GT33bU4CEgYmlvVh/view?usp=sharing
2. taruh hasil download ke root project ini
3. go run main.go
4. request ke shortest path
curl --location 'http://localhost:3000/api/navigations/shortestPath' \
--header 'Content-Type: application/json' \
--data '{
    "src_lat":  -7.54868306014711, 
    "src_lon":  110.78270355038615, 
    "dst_lat": -7.771792490144098, 
    "dst_lon": 110.37740457912028
}'

Note: Source  & Destination Coordinate harus tempat di sekitaran kota yogyakarta/surakarta/semarang
5. copy polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. Rute terdekat akan tampil di peta :) 


```

