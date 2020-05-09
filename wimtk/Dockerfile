FROM golang:alpine

WORKDIR /build

RUN apk add upx bash

COPY go.mod go.sum /build/

RUN go mod download

RUN go get -u github.com/c9s/gomon

COPY . .

CMD [ "sh", "-c", "sleep 3600" ]