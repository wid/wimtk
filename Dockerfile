FROM golang

WORKDIR /build

COPY wimtk/go.mod wimtk/go.sum /build/

RUN go mod download

RUN go get -u github.com/c9s/gomon

ADD wimtk .

RUN go build

CMD [ "gomon", "-t" ]