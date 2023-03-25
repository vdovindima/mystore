FROM golang:alpine3.17 AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY main.go ./

RUN go build -o /mystore

FROM alpine:3.17.2

WORKDIR /

COPY --from=build /mystore /mystore

RUN mkdir /data

ENTRYPOINT ["/mystore"]