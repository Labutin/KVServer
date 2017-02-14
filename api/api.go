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
	Response interface{} `json:"response"`
	Ok       bool        `json:"ok"`
	Error    string      `json:"error"`
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
			r.Get("/", getDictRecord)
		})
		r.Route("/getlist/:key/:index", func(r chi.Router) {
			r.Get("/", GetListRecord)
		})
		r.Post("/", addRecord)
		r.Post("/dict/", addDict)
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
		render.JSON(w, r, Resp{Error: err.Error(), Ok: false})
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Set(data.Key, data.Value, ttl)
	render.JSON(w, r, Resp{Response: "", Ok: true})
}

func getRecord(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var res Resp
	if value, ok := storage.Get(key); !ok {
		render.Status(r, 404)
		res.Ok = false
		res.Error = KeyNotFound.String()
	} else {
		res = Resp{Response: value, Ok: true}
	}
	render.JSON(w, r, res)
	return
}

func addDict(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string         `json:"key"`
		Value kvstorage.Dict `json:"value"`
		TTL   int64          `json:"ttl"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.JSON(w, r, Resp{Error: err.Error(), Ok: false})
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Set(data.Key, data.Value, ttl)
	render.JSON(w, r, Resp{Response: "", Ok: true})
}

func getDictRecord(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	dictKey := chi.URLParam(r, "keydict")
	var res Resp
	if value, err := storage.GetDictElement(key, dictKey); err != nil {
		render.Status(r, 404)
		res.Ok = false
		res.Error = err.Error()
	} else {
		res = Resp{Response: value, Ok: true}
	}
	render.JSON(w, r, res)
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
