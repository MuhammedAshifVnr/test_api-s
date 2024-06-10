package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"test/db"
	"test/handlers"
	"test/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestADDB() {
	dsn := "host=localhost user=postgres dbname=testdb sslmode=disable password=0000"
	db.Init(dsn)
	db.DB.Exec("DELETE FROM admins")
	testAdmin := models.Admin{Name: "Test Admin", Email: "admin@test.com", Password: "12345"}
	db.DB.Create(&testAdmin)
	
}

func TestADLogin(t *testing.T) {

	setupTestADDB()

	gin.SetMode(gin.TestMode)

	t.Run("successful login", func(t *testing.T) {
		router := gin.Default()
		router.POST("/adminlogin", handlers.AdminLogin)

		loginInput := handlers.LoginInput{
			Email:    "admin@test.com",
			Password: "12345",
		}
		jsonValue, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/adminlogin", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Login successful")
	})

	t.Run("invalid email", func(t *testing.T) {
		router := gin.Default()
		router.POST("/adminlogin", handlers.AdminLogin)

		loginInput := handlers.LoginInput{
			Email:    "wrong@example.com",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/adminlogin", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid email or password")
	})

	t.Run("invalid password", func(t *testing.T) {
		router := gin.Default()
		router.POST("/adminlogin", handlers.AdminLogin)

		loginInput := handlers.LoginInput{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		jsonValue, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/adminlogin", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid email or password")
	})
}
