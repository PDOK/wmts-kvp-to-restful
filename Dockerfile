FROM golang:1.20 AS build-env

WORKDIR /go/src/server
ADD . /go/src/server

# disable crosscompiling
ENV CGO_ENABLED=0
# compile linux only
ENV GOOS=linux

RUN go mod download all

#run all tests
RUN go test github.com/PDOK/wmts-kvp-to-restful/operations

#build the binary with debug information removed
RUN go build  -ldflags '-w -s' -a -installsuffix cgo -o /wmts-kvp-to-restful .

FROM scratch as service
WORKDIR /
ENV PATH=/

COPY --from=build-env /wmts-kvp-to-restful /
