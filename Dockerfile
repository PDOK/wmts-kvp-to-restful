FROM scratch 
 FROM golang:1.11-alpine3.8 AS build-env

ENV http_proxy http://www-proxy.cs.kadaster.nl:8082
ENV https_proxy http://www-proxy.cs.kadaster.nl:8082
ENV no_proxy localhost,127.0.0.1,.so.kadaster.nl,.in.kadaster.nl,.fto.kadaster.nl,.cs.kadaster.nl

RUN apk update && apk upgrade && \
   apk add --no-cache bash git gcc musl-dev

ENV GO111MODULE=on

WORKDIR /go/src/server

ADD . /go/src/server

#disable crosscompiling
ENV CGO_ENABLED=1

#compile linux only
ENV GOOS=linux

#build the binary with debug information removed
RUN go build -ldflags '-w -s -linkmode external -extldflags -static' -a -installsuffix cgo -o /wmts-kvp-to-restful wmts-kvp-to-restful.go

FROM scratch as service
WORKDIR /
ENV PATH=/

COPY --from=build-env /wmts-kvp-to-restful /