# navigatorx
Sementara pake data openstreetmap kota solo/surakarta

## Quick Start
```
1. download file openstreetmap pbf di: https://drive.google.com/file/d/1CF1aydZf6j4ula_CYhYxezZo805w44Rn/view?usp=sharing
2. taruh hasil download ke root project ini
3. go run main.go
4. request ke shortest path
curl --location 'http://localhost:3000/api/navigations/shortestPath' \
--header 'Content-Type: application/json' \
--data '{
    "src_lat": -7.556474471231671,
    "src_lon":   110.80444085178604, 
    "dst_lat": -7.576648776305183, 
    "dst_lon":    110.81751879114158
}'

Note: Source  & Destination Coordinate harus tempat di sekitaran kota surakarta
5. copy polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. Rute terdekat akan tampil di peta :) 
(sementara anggap semua jalan 2 arah)
```

