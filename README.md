# navigatorx


## Quick Start
```
1. download file openstreetmap pbf di: https://drive.google.com/file/d/1pEHN8wwUbB5XpuYMZm141fXQ_ZsIf4CO/view?usp=sharing
2. taruh hasil download ke root project ini
3. go run main.go
(Minimal free ram 5 GB buat data diatas)
4. request ke shortest path
curl --location 'http://localhost:3000/api/navigations/shortestPathCH' \
--header 'Content-Type: application/json' \
--data '{
    "src_lat": -7.550261232598317,
    "src_lon":    110.78210790296636, 
    "dst_lat": -8.024431446370416,
    "dst_lon":   110.32971396395838
}'

Note: Source  & Destination Coordinate harus tempat di sekitaran kota yogyakarta/surakarta/semarang
5. copy polyline string hasil response endpoint tadi ke https://valhalla.github.io/demos/polyline . Centang Unsescape '\'. Rute terdekat akan tampil di peta :) 


```


```
https://jlazarsfeld.github.io/ch.150.project/sections/7-ch-overview/
https://dl.acm.org/doi/pdf/10.1145/971697.602266
https://www.uber.com/en-ID/blog/engineering-routing-engine/
http://theory.stanford.edu/~amitp/GameProgramming/ImplementationNotes.html
```


