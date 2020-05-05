FROM golang

WORKDIR /build

COPY go.mod go.sum /build/

RUN go mod download

RUN go get -u github.com/c9s/gomon

COPY . .

RUN go build

CMD [ "gomon", "-t" ]