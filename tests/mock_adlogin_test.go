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

func setupAdminTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
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

	// Clean the database
	db.SetDB(database)
	mock.ExpectExec("DELETE FROM admins").WillReturnResult(sqlmock.NewResult(0, 0))
	db.DB.Exec("DELETE FROM admins")

	// Create a test admin
	password, _ := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
	testAdmin := models.Admin{Name: "Test Admin", Email: "admin@example.com", Password: string(password)}
	db.DB.Create(&testAdmin)

	cleanup := func() {
		mockDB.Close()
	}

	return mock, cleanup
}

func TestAdminLogin(t *testing.T) {
	mock, cleanup := setupAdminTestDB(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	t.Run("successful admin login", func(t *testing.T) {
		password, _ := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
		mock.ExpectQuery(`SELECT \* FROM "admins" WHERE email = \$1 ORDER BY "admins"\."name" LIMIT \$2`).
			WithArgs("admin@example.com", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "Test Admin", "admin@example.com", string(password)))

		router := gin.Default()
		router.POST("/admin/login", handlers.AdminLogin)
		loginInput := handlers.AdminLoginInput{
			Email:    "admin@example.com",
			Password: "adminpassword",
		}
		jsonValue, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/admin/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Login successful")
	})

	t.Run("invalid email", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "admins" WHERE email = \$1 ORDER BY "admins"\."name" LIMIT \$2`).
			WithArgs("wrongadmin@example.com", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password"}))

		router := gin.Default()
		router.POST("/admin/login", handlers.AdminLogin)
		loginInput := handlers.AdminLoginInput{
			Email:    "wrongadmin@example.com",
			Password: "adminpassword",
		}
		jsonValue, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/admin/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid email or password")
	})

	t.Run("invalid password", func(t *testing.T) {
		password, _ := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
		mock.ExpectQuery(`SELECT \* FROM "admins" WHERE email = \$1 ORDER BY "admins"\."name" LIMIT \$2`).
			WithArgs("admin@example.com", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "Test Admin", "admin@example.com", string(password)))

		router := gin.Default()
		router.POST("/admin/login", handlers.AdminLogin)
		loginInput := handlers.AdminLoginInput{
			Email:    "admin@example.com",
			Password: "wrongpassword",
		}
		jsonValue, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/admin/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid email or password")
	})
}
