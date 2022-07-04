# syntax=docker/dockerfile:1

##
## Build
##

FROM golang:latest AS build

WORKDIR /appdir

COPY . ./

RUN go mod tidy

USER root

RUN go test ./load_balancer/...

RUN go build ./load_balancer/app/

CMD [ "./app" ]