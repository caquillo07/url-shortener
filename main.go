package main

import (
	"flag"
	"log"

	"go.uber.org/zap"

	"github.com/caquillo07/sample_url_shortener/pkg/server"
	"github.com/caquillo07/sample_url_shortener/pkg/storage"
)

var devLog bool

func main() {
	flag.BoolVar(&devLog, "dev-log", false, "Show logs in development format instead of JSON format")
	flag.Parse()
	initLogging(devLog)

	sv := server.NewSever(storage.NewMemoryStore())
	log.Fatal(sv.Run())
}

func initLogging(devMode bool) {
	var logger *zap.Logger
	if devMode {
		logger, _ = zap.NewDevelopment()
		logger.Info("Development logging enabled")
	} else {
		logger, _ = zap.NewProduction()
	}

	logger.Info("Starting URL Shortener service")
	zap.ReplaceGlobals(logger)
}
