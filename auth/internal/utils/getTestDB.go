package utils

import (
	"gorm.io/gorm"
)

var testDB *gorm.DB

func SetTestDB(db *gorm.DB) {
	testDB = db
}

func GetTestDB() *gorm.DB {
	return testDB
}