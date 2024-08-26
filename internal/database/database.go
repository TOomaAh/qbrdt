package database

import (
	"github.com/TOomaAh/qbrdt/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDatabase(logger logger.Interface) *gorm.DB {

	db, err := gorm.Open(sqlite.Open("qbrdt.db"), &gorm.Config{})

	if err != nil {
		logger.Fatal("failed to connect database")
	}

	return db

}
