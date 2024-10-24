# coba pake alpine gabisa
FROM debian:bookworm-slim as builder
RUN apt-get update
RUN apt-get install -y wget libzstd-dev  build-essential
RUN wget https://go.dev/dl/go1.22.6.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.22.6.linux-amd64.tar.gz
RUN apt-get install -y pkg-config  zlib1g-dev
ARG MAP_FILE
ARG DRIVE_FILE_ID
COPY . /app
WORKDIR /app
ENV PATH="/usr/local/go/bin:${PATH}"
# uber h3 & zstd butuh cgo 
RUN   CGO_ENABLED=1 GOOS=linux  go build -o /bin/app .   
RUN  wget --no-check-certificate "https://docs.google.com/uc?export=download&id=$DRIVE_FILE_ID" -O /app/$MAP_FILE.osm.pbf
CMD ["sh", "-c", "/bin/app", "-f=/app/$MAP_FILE.osm.pbf"] 


