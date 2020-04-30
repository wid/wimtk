FROM golang

WORKDIR /build

RUN go get -u github.com/c9s/gomon

ADD pucon .

RUN go build

CMD [ "gomon", "-t" ]