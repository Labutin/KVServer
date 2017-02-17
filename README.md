# Simple Key/Value REST Server

## How to start?

`docker-compose up -d`

> ...
> Creating network "kvserver_default" with the default driver

> Creating kvserver_mongo_1

> Creating kvserver_kvserver_1

`curl -X POST -d '{"key":"t1", "value":"v1"}' http://127.0.0.1:8081/v1/kvstorage`
> {"response":"","ok":true,"error":""}