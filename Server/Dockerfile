FROM golang:1.8-alpine

COPY . /go/src/github.com/Labutin/KVServer/Server

RUN \
    cd /go/src/github.com/Labutin/KVServer/Server && \
    go build -v -o /go/bin/KVServer && \
    rm -rf /go/src

EXPOSE 8081
ENTRYPOINT ["/go/bin/KVServer"]
