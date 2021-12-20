FROM scratch 
 FROM golang:1.11-alpine3.8 AS build-env

RUN apk update && apk upgrade && \
   apk add --no-cache bash git gcc musl-dev

ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org

WORKDIR /go/src/server

ADD . /go/src/server

#disable crosscompiling
ENV CGO_ENABLED=1

#compile linux only
ENV GOOS=linux

#run all tests
RUN go test github.com/PDOK/wmts-kvp-to-restful/operations

#build the binary with debug information removed
RUN go build -ldflags '-w -s -linkmode external -extldflags -static' -a -installsuffix cgo -o /wmts-kvp-to-restful .

FROM scratch as service
WORKDIR /
ENV PATH=/

COPY --from=build-env /wmts-kvp-to-restful /
