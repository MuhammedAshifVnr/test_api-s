package main

import (
	"test/db"
	"test/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := "host=localhost user=postgres password=0000 dbname=sample sslmode=disable"
	db.Init(dsn)

	r := gin.Default()
	r.POST("/signup", handlers.Signup)
	r.Run()
}
