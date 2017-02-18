package api

import (
	"fmt"
	"github.com/Labutin/KVServer/Server/api/persist"
	"github.com/Labutin/KVServer/Server/logs"
	"github.com/Labutin/MemoryKeyValueStorage/kvstorage"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	storage        *kvstorage.Storage
	chuncks        uint32
	urlPath        = "/v1/kvstorage"
	persistStorage persist.PersistStorage
)

const GOROUTINE_NAME = "main"

type Resp struct {
	Response interface{} `json:"response"`
	Ok       bool        `json:"ok"`
	Error    string      `json:"error"`
}

// handlerPing just ping pong
func handlerPing(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "pong")
}

// InitRouter creates router
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
			r.Get("/", getListRecord)
		})
		r.Route("/", func(r chi.Router) {
			r.Post("/", addRecord)
			r.Put("/", updateRecord)
			r.Delete("/", removeRecord)
		})
		r.Post("/dict/", addDict)
		r.Post("/list/", addList)
		r.Get("/keys", getAllKeys)
		r.Get("/saveToDb", saveToDb)
		r.Get("/loadFromDb", loadFromDb)
	})

	return r
}

// InitStorage creates Key/Value storage
func InitStorage(totalChunks uint32) {
	storage = kvstorage.NewKVStorage(totalChunks, true)
	chuncks = totalChunks
}

// InitPersistentStorage sets MongoDb params
func InitPersistentStorage(pStorage persist.PersistStorage) {
	persistStorage = pStorage
}

// addRecord puts record to storage
func addRecord(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int64       `json:"ttl"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.Status(r, http.StatusNotAcceptable)
		render.JSON(w, r, Resp{Error: err.Error(), Ok: false})
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Set(data.Key, data.Value, ttl)
	log.Println(logs.MakeLogString(logs.INFO, GOROUTINE_NAME, "Added record with key: "+data.Key, nil))
	render.JSON(w, r, Resp{Response: "", Ok: true})
}

// updateRecord updates record with given key
func updateRecord(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int64       `json:"ttl"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.Status(r, http.StatusNotAcceptable)
		render.JSON(w, r, Resp{Error: err.Error(), Ok: false})
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Update(data.Key, data.Value, ttl)
	log.Println(logs.MakeLogString(logs.INFO, GOROUTINE_NAME, "Updated record with key: "+data.Key, nil))
	render.JSON(w, r, Resp{Response: "", Ok: true})
}

// getAllKeys returns all keys in storage
func getAllKeys(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Resp{Response: storage.Keys(), Ok: true})
}

// getRecord returns record from storage
func getRecord(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var res Resp
	if value, ok := storage.Get(key); !ok {
		render.Status(r, http.StatusNotFound)
		res.Ok = false
		res.Error = KeyNotFound.String()
	} else {
		res = Resp{Response: value, Ok: true}
	}
	log.Println(logs.MakeLogString(logs.INFO, GOROUTINE_NAME, "Fetched record with key: "+key, nil))
	render.JSON(w, r, res)
}

// removeRecord removes record with given key from storage
func removeRecord(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key string `json:"key"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.Status(r, http.StatusNotAcceptable)
		render.JSON(w, r, Resp{Error: err.Error(), Ok: false})
		return
	}
	storage.Remove(data.Key)
	render.JSON(w, r, Resp{Response: "", Ok: true})
}

// addDict puts dictionary to storage
func addDict(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string                 `json:"key"`
		Value map[string]interface{} `json:"value"`
		TTL   int64                  `json:"ttl"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.Status(r, http.StatusNotAcceptable)
		render.JSON(w, r, Resp{Error: err.Error(), Ok: false})
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Set(data.Key, data.Value, ttl)
	log.Println(logs.MakeLogString(logs.INFO, GOROUTINE_NAME, "Added dictionary with key: "+data.Key, nil))
	render.JSON(w, r, Resp{Response: "", Ok: true})
}

// getDictRecord returns value from dictionary with given key and nested key
func getDictRecord(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	dictKey := chi.URLParam(r, "keydict")
	var res Resp
	if value, err := storage.GetDictElement(key, dictKey); err != nil {
		render.Status(r, http.StatusNotFound)
		res.Ok = false
		res.Error = err.Error()
	} else {
		res = Resp{Response: value, Ok: true}
	}
	log.Println(logs.MakeLogString(logs.INFO, GOROUTINE_NAME, "Fetched dictionary with key: "+key+" and subkey: "+dictKey, nil))
	render.JSON(w, r, res)
}

// addList puts list to storage
func addList(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string        `json:"key"`
		Value []interface{} `json:"value"`
		TTL   int64         `json:"ttl"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.Status(r, http.StatusNotAcceptable)
		render.JSON(w, r, Resp{Error: err.Error(), Ok: false})
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Set(data.Key, data.Value, ttl)
	log.Println(logs.MakeLogString(logs.INFO, GOROUTINE_NAME, "Added list with key: "+data.Key, nil))
	render.JSON(w, r, Resp{Response: "", Ok: true})
}

// getListRecord returns element from list with given key and index
func getListRecord(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	index := chi.URLParam(r, "index")
	indexInt, _ := strconv.Atoi(index)
	var res Resp
	if value, err := storage.GetListElement(key, indexInt); err != nil {
		render.Status(r, http.StatusNotFound)
		res.Ok = false
		res.Error = err.Error()
	} else {
		res = Resp{Response: value, Ok: true}
	}
	log.Println(logs.MakeLogString(logs.INFO, GOROUTINE_NAME, "Fetched list with key: "+key+" and index: "+strconv.Itoa(indexInt), nil))
	render.JSON(w, r, res)
}

// saveToDb store all data to MongoDB
func saveToDb(w http.ResponseWriter, r *http.Request) {
	err := persistStorage.SaveToDb(storage)
	var res Resp
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		res.Ok = false
		res.Error = err.Error()
		render.JSON(w, r, res)
		return
	}
	res.Ok = true
	res.Response = ""
	render.JSON(w, r, res)
}

// saveToDb restore all data from MongoDB
func loadFromDb(w http.ResponseWriter, r *http.Request) {
	storage.StopTTLProcessing()
	InitStorage(chuncks)
	err := persistStorage.LoadFromDb(storage)
	var res Resp
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		res.Ok = false
		res.Error = err.Error()
		render.JSON(w, r, res)
		return
	}
	res.Ok = true
	res.Response = ""
	render.JSON(w, r, res)
}
