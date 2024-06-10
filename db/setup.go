package db

import (
	"log"
	"test/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	DB.AutoMigrate(&models.User{}, &models.Admin{})
}

func SetDB(database *gorm.DB) {
	DB = database
}
