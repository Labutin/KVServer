# Simple Key/Value REST Server

## How to start?

`docker-compose up -d`

> ...<br/>
> Creating network "kvserver_default" with the default driver<br/>
> Creating kvserver_mongo_1<br/>
> Creating kvserver_kvserver_1<br/>

`curl -X POST -d '{"key":"t1", "value":"v1"}' http://127.0.0.1:8081/v1/kvstorage`
> {"response":"","ok":true,"error":""}

`curl http://127.0.0.1:8081/v1/kvstorage/get/t1`
> {"response":"v1","ok":true,"error":""}

`curl http://127.0.0.1:8081/v1/kvstorage/get/unknown`
> {"response":null,"ok":false,"error":"Key not found"}

**Working with dict** 

`curl -X POST -d '{"key":"dict", "value":{"k1":1, "k2":2}}' http://127.0.0.1:8081/v1/kvstorage`
> {"response":"","ok":true,"error":""}

`curl http://127.0.0.1:8081/v1/kvstorage/getdict/dict/k1`
> {"response":1,"ok":true,"error":""}

`curl http://127.0.0.1:8081/v1/kvstorage/getdict/t1/k1`
> {"response":null,"ok":false,"error":"Value not Dictionary"}

**Working with list**

`curl -X POST -d '{"key":"list", "value":[0,10,20,30]}' http://127.0.0.1:8081/v1/kvstorage`
> {"response":"","ok":true,"error":""}

`curl http://127.0.0.1:8081/v1/kvstorage/getlist/list/2`
> {"response":20,"ok":true,"error":""}

`curl http://127.0.0.1:8081/v1/kvstorage/getlist/list/10`
> {"response":null,"ok":false,"error":"Out of bound"}

**Working with TTL (in seconds). Exipired keys removes every 60 seconds.**

`curl -X POST -d '{"key":"t1", "value":"I am here", "TTL":30}' http://127.0.0.1:8081/v1/kvstorage`
> {"response":"","ok":true,"error":""}

`curl http://127.0.0.1:8081/v1/kvstorage/get/t1`
> {"response":"I am here","ok":true,"error":""}

_... wait 60 seconds ..._

`curl http://127.0.0.1:8081/v1/kvstorage/get/t1`
> {"response":null,"ok":false,"error":"Key not found"}

**Save all data to Database (MongoDB)**

`curl http://127.0.0.1:8081/v1/kvstorage/saveToDb`
> {"response":"","ok":true,"error":""}

`docker-compose restart`
>Restarting kvserver_kvserver_1 ... done<br/>
>Restarting kvserver_mongo_1 ... done

_Memory cleaned_

`curl http://127.0.0.1:8081/v1/kvstorage/getlist/list/2`
> {"response":null,"ok":false,"error":"Key not found"}

`curl http://127.0.0.1:8081/v1/kvstorage/loadFromDb`
> {"response":"","ok":true,"error":""}

_Try again after restore data_

`curl http://127.0.0.1:8081/v1/kvstorage/getlist/list/2`
> {"response":20,"ok":true,"error":""}
