package main

import (
	"github.com/geoffjay/crm/util"

	log "github.com/sirupsen/logrus"
)

func initLogging() {
	formatter := util.Getenv("LOG_FORMATTER", "text")
	level := util.Getenv("LOG_LEVEL", "info")

	if logLevel, err := log.ParseLevel(level); err == nil {
		log.SetLevel(logLevel)
	}

	if formatter == "json" {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
}
