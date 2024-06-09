package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"test/db"
	"test/handlers"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetupTestDB() (*gorm.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		panic("failed to open sqlmock database connection")
	}
	gormDB, err := gorm.Open("postgres", mockDB)
	if err != nil {
		panic("failed to open gorm db connection")
	}
	return gormDB, mock
}

func TestSignup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid signup", func(t *testing.T) {
		testDB, mock := SetupTestDB()
		db.SetDB(testDB)

		defer testDB.Close()
		
		mock.ExpectBegin()

		mock.ExpectQuery(`INSERT INTO "users" \("created_at","updated_at","deleted_at","name","email","password"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6\) RETURNING "users"."id"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "Test User", "testuser@example.com", sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectCommit()

		router := gin.Default()
		router.POST("/signup", handlers.Signup)

		payload := `{"name": "Test User", "email": "testuser@example.com", "password": "password123"}`
		req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

	
		if w.Code != http.StatusOK {
			t.Logf("Response Code: %d", w.Code)
			t.Logf("Response Body: %s", w.Body.String())
		}

		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "User created successfully", response["message"])

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("mock expectations were not met: %v", err)
		}
	})
	t.Run("Invalid signup", func(t *testing.T) {
		
		router := gin.Default()
		router.POST("/signup", handlers.Signup)

		payload := `{"name": "", "email": "", "password": ""}`
		req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

	
		if w.Code != 400 {
			t.Logf("Response Code: %d", w.Code)
			t.Logf("Response Body: %s", w.Body.String())
		}

		require.Equal(t, 400, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "All fields are required", response["error"])

		
	})
}
