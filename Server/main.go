package main

import (
	"github.com/Labutin/KVServer/Server/api"
	"github.com/Labutin/KVServer/Server/api/persist"
	"github.com/Labutin/KVServer/Server/logs"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
	"log"
	"net/http"
	"os"
)

var opts struct {
	Chunks              uint32 `long:"chunks" env:"CHUNKS" description:"Number chunks in cocurrent map" required:"true"`
	LoggingLevel        string `long:"loggingLevel" env:"LOGGING_LEVEL" description:"Logging level" default:"INFO" required:"true"`
	MDBConnectionString string `long:"mdbConnectionString" env:"MDB_CONNECTION_STRING" description:"MongoDB connection string" required:"true"`
	MDBDbName           string `long:"mdbDbName" env:"MDB_DATABASE" description:"MongoDB database name" required:"true"`
	MDBCollection       string `long:"mdbCollection" env:"MDB_COLLECTION" description:"MongoDB collection name" required:"true"`
}

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
	api.InitStorage(opts.Chunks)
	api.InitPersistentStorage(persist.NewMongoStorage(opts.MDBConnectionString, opts.MDBDbName, opts.MDBCollection))
	log.Println(logs.MakeLogString(logs.INFO, "main", "Ready to recieve requests", nil))
	http.ListenAndServe(":8081", api.InitRouter())
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
}
