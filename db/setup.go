package db

import (
	"log"
	"test/models"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

var DB *gorm.DB

func Init(dsn string) {
	var err error
	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	DB.AutoMigrate(&models.User{}, &models.Admin{})
}

func SetDB(database *gorm.DB) {
	DB = database
}
