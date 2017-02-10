package main

import (
	"fmt"
	"github.com/Labutin/KVServer/logs"
	"github.com/Labutin/MemoryKeyValueStorage/kvstorage"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var opts struct {
	Chunks       uint32 `long:"chunks" env:"CHUNKS" description:"Number chunks in cocurrent map" required:"true"`
	LoggingLevel string `long:"logginglevel" env:"LOGGING_LEVEL" description:"Logging level" default:"INFO" required:"true"`
}

func handlerPing(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "pong")
}

var storage *kvstorage.Storage

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		log.Fatalln(logs.MakeLogString(
			logs.ERROR,
			"main",
			"Can't parse options.",
			err))
	}
	filter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel{
			logutils.LogLevel(logs.DEBUG.String()),
			logutils.LogLevel(logs.INFO.String()),
			logutils.LogLevel(logs.WARN.String()),
			logutils.LogLevel(logs.ERROR.String()),
		},
		MinLevel: logutils.LogLevel(opts.LoggingLevel),
		Writer:   os.Stdout,
	}
	log.SetOutput(filter)
	storage = kvstorage.NewKVStorage(opts.Chunks, true)
	r := initRouter()
	http.ListenAndServe(":8081", r)
}

func initRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Get("/v1/ping", handlerPing)
	r.Route("/v1/kvstorage", func(r chi.Router) {
		r.Route("/get/:key", func(r chi.Router) {
			r.Get("/", GetRecord)
		})
		r.Route("/getdict/:key/:keydict", func(r chi.Router) {
			r.Get("/", GetDictRecord)
		})
		r.Route("/getlist/:key/:index", func(r chi.Router) {
			r.Get("/", GetListRecord)
		})
		r.Post("/", AddRecord)
		r.Post("/dict/", AddDict)
		r.Post("/list/", AddList)
	})
	return r
}

func GetRecord(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	value, _ := storage.Get(key)
	render.JSON(w, r, value)
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

func AddRecord(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int64       `json:"ttl"`
	}
	if err := render.Bind(r.Body, &data); err != nil {
		render.JSON(w, r, err.Error())
		return
	}
	ttl := time.Second * time.Duration(data.TTL)
	storage.Set(data.Key, data.Value, ttl)
	render.JSON(w, r, data)
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

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
}
