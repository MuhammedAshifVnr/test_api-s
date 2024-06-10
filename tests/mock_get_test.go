package tests

import (
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
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTesGetDB(t *testing.T) (sqlmock.Sqlmock, func()) {
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

	cleanup := func() {
		mockDB.Close()
	}

	return mock, cleanup
}

func TestGetAllUsers(t *testing.T) {
	mock, cleanup := setupTesGetDB(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	t.Run("successful fetch users", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "users"`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "User One", "userone@example.com", "password1").
				AddRow(2, "User Two", "usertwo@example.com", "password2"))

		router := gin.Default()
		router.GET("/users", handlers.GetAllUsers)

		req, _ := http.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string][]models.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response["users"], 2)
		assert.Equal(t, "User One", response["users"][0].Name)
		assert.Equal(t, "userone@example.com", response["users"][0].Email)
		assert.Equal(t, "User Two", response["users"][1].Name)
		assert.Equal(t, "usertwo@example.com", response["users"][1].Email)
	})

	t.Run("failure to fetch users", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "users"`).
			WillReturnError(gorm.ErrInvalidTransaction)

		router := gin.Default()
		router.GET("/users", handlers.GetAllUsers)

		req, _ := http.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Failed to fetch users", response["error"])
	})
}
