package database

import (
	"os"

	"github.com/TOomaAh/qbrdt/pkg/logger"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewDatabase(logger logger.Interface) *gorm.DB {

	if os.Getenv("QBRDT_DB") == "" {
		os.Setenv("QBRDT_DB", "qbrdt.db")
	}

	db, err := gorm.Open(sqlite.Open(os.Getenv("QBRDT_DB")), &gorm.Config{})

	if err != nil {
		logger.Fatal("failed to connect database")
	}

	return db

}
