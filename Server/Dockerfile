FROM golang:1.7-alpine

COPY . /go/src/github.com/Labutin/KVServer

RUN \
    cd /go/src/github.com/Labutin/KVServer && \
    go build -v -o /go/bin/KVServer && \
    rm -rf /go/src

EXPOSE 8081
ENTRYPOINT ["/go/bin/KVServer"]
