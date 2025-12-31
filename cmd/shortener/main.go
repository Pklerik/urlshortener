// Package main grants cmd entree point for whole application
package main

import (
	"log"

	"github.com/Pklerik/urlshortener/internal/app"
	"github.com/Pklerik/urlshortener/internal/logger"
)

const na = "N/A"

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	parsedArgs := parseArgs()

	err := logger.Initialize(parsedArgs.GetLogLevel())
	if err != nil {
		log.Fatalf("Unable to setup logger: main: %s", err.Error())
	}

	if buildVersion == "" {
		buildVersion = na
	}

	if buildDate == "" {
		buildDate = na
	}

	if buildCommit == "" {
		buildCommit = na
	}

	logger.Sugar.Infof("Build version: <%s>", buildVersion)
	logger.Sugar.Infof("Build date: <%s>", buildDate)
	logger.Sugar.Infof("Build commit: <%s>", buildCommit)

	app.StartApp(parsedArgs)
}
