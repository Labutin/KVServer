package api

import (
	"github.com/Labutin/MemoryKeyValueStorage/kvstorage"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
	"net/http"
	"strconv"
	"time"

	"fmt"
)

var (
	storage *kvstorage.Storage
	urlPath = "/v1/kvstorage"
)

type Resp struct {
	Result interface{} `json:"response"`
	Ok     bool        `json:"ok"`
	Error  string      `json:"error"`
}

func handlerPing(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "pong")
}

func InitRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Get("/v1/ping", handlerPing)
	r.Route(urlPath, func(r chi.Router) {
		r.Route("/get/:key", func(r chi.Router) {
			r.Get("/", getRecord)
		})
		r.Route("/getdict/:key/:keydict", func(r chi.Router) {
			r.Get("/", GetDictRecord)
		})
		r.Route("/getlist/:key/:index", func(r chi.Router) {
			r.Get("/", GetListRecord)
		})
		r.Post("/", addRecord)
		r.Post("/dict/", AddDict)
		r.Post("/list/", AddList)
	})

	return r
}

func InitStorage(chunks uint32) {
	storage = kvstorage.NewKVStorage(chunks, true)
}

func addRecord(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int64       `json:"ttl"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.JSON(w, r, Resp{Result: err.Error(), Ok: false})
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Set(data.Key, data.Value, ttl)
	render.JSON(w, r, Resp{Result: "", Ok: true})
}

func getRecord(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var res Resp
	if value, ok := storage.Get(key); !ok {
		res.Ok = false
		res.Error = KeyNotFound.String()
	} else {
		res = Resp{Result: value, Ok: true}
	}
	render.JSON(w, r, res)
	return
}

func GetDictRecord(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	dictKey := chi.URLParam(r, "keydict")
	value, _ := storage.GetDictElement(key, dictKey)
	render.JSON(w, r, value)
	return
}

func GetListRecord(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	index := chi.URLParam(r, "index")
	indexInt, _ := strconv.Atoi(index)
	value, _ := storage.GetListElement(key, indexInt)
	render.JSON(w, r, value)
	return
}

func AddDict(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string         `json:"key"`
		Value kvstorage.Dict `json:"value"`
		TTL   int64          `json:"ttl"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.JSON(w, r, err.Error())
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Set(data.Key, data.Value, ttl)
	render.JSON(w, r, data)
}

func AddList(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string         `json:"key"`
		Value kvstorage.List `json:"value"`
		TTL   int64          `json:"ttl"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.JSON(w, r, err.Error())
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Set(data.Key, data.Value, ttl)
	render.JSON(w, r, data)
}
