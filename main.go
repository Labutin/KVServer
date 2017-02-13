package main

import (
	"github.com/Labutin/KVServer/api"
	"github.com/Labutin/KVServer/logs"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
	"log"
	"net/http"
	"os"
)

var opts struct {
	Chunks       uint32 `long:"chunks" env:"CHUNKS" description:"Number chunks in cocurrent map" required:"true"`
	LoggingLevel string `long:"logginglevel" env:"LOGGING_LEVEL" description:"Logging level" default:"INFO" required:"true"`
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
	http.ListenAndServe(":8081", api.InitRouter())
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
}
