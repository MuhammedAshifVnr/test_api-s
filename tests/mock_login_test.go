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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})

	database, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	db.SetDB(database)
	mock.ExpectExec("DELETE FROM users").WillReturnResult(sqlmock.NewResult(0, 0))
	db.DB.Exec("DELETE FROM users")
	password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := models.User{Name: "Test User", Email: "test@example.com", Password: string(password)}
	db.DB.Create(&testUser)

	cleanup := func() {
		mockDB.Close()
	}
	return mock, cleanup
}

func TestLogin(t *testing.T) {
	mock, cleanup := setupTestDB(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	t.Run("successful login", func(t *testing.T) {
		password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 ORDER BY "users"\."id" LIMIT \$2`).
			WithArgs("test@example.com", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "Test User", "test@example.com", string(password)))

		router := gin.Default()
		router.POST("/login", handlers.Login)

		loginInput := handlers.LoginInput{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Login successful")
	})

	t.Run("invalid email", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 ORDER BY "users"\."id" LIMIT \$2`).
			WithArgs("wrong@example.com", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password"}))

		router := gin.Default()
		router.POST("/login", handlers.Login)

		loginInput := handlers.LoginInput{
			Email:    "wrong@example.com",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid email or password")
	})

	t.Run("invalid password", func(t *testing.T) {
		password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1 ORDER BY "users"\."id" LIMIT \$2`).
			WithArgs("test@example.com", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "Test User", "test@example.com", string(password)))

		router := gin.Default()
		router.POST("/login", handlers.Login)

		loginInput := handlers.LoginInput{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		jsonValue, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid email or password")
	})
}
