# coba pake alpine gabisa
FROM debian:bookworm-slim as builder
RUN apt-get update
RUN apt-get install -y wget libzstd-dev  build-essential
RUN wget https://go.dev/dl/go1.22.6.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.22.6.linux-amd64.tar.gz
RUN apt-get install -y pkg-config  zlib1g-dev
COPY . /app
WORKDIR /app
ENV PATH="/usr/local/go/bin:${PATH}"
RUN go clean -modcache
# uber h3 & zstd butuh cgo 
RUN   CGO_ENABLED=1 GOOS=linux  go build -o /bin/app .   
RUN wget --no-check-certificate 'https://docs.google.com/uc?export=download&id=1pEHN8wwUbB5XpuYMZm141fXQ_ZsIf4CO' -O /bin/solo_jogja.osm.pbf
CMD ["/bin/app"]


