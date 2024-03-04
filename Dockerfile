##### BUILDER #####
FROM golang:1.22 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN apt-get update && apt-get install -y build-essential wget

RUN make download-checksum

RUN go build -v -o indexer ./cmd/indexer/
RUN go build -v -o server ./cmd/server/

##### RUNNER #####
FROM debian:12-slim

WORKDIR /app

COPY --from=builder /app/indexer /app/indexer
COPY --from=builder /app/server /app/server
COPY --from=builder /app/checksums.json /app/checksums.json

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

CMD indexer