package main

import (
	"os"
	"time"

	"github.com/TOomaAh/qbrdt/internal/config"
	"github.com/TOomaAh/qbrdt/internal/qbrdt"
	"github.com/TOomaAh/qbrdt/pkg/logger"
)

func init() {

	tz := os.Getenv("TZ")
	if tz == "" {
		time.Local = time.UTC
		return
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		time.Local = time.UTC
		return
	}
	time.Local = loc

}

func main() {

	config := config.NewQBRDTConfig()

	log := logger.New(
		config.Logger.Level,
	)

	log.Info("Using timezone: %s", time.Local.String())

	app := qbrdt.New(log, config)

	app.Run()
}
