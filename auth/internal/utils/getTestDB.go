package utils

import (

)

var testDB *gorm.DB

func SetTestDB(db *gorm.DB) {
	testDB = db
}

func GetTestDB() *gorm.DB {
	return testDB
}