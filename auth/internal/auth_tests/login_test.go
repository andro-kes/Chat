package auth_tests

import (
	"bytes"

	"github.com/andro-kes/Chat/auth/internal/auth"
	"github.com/andro-kes/Chat/auth/internal/utils"
	"github.com/andro-kes/Chat/shared/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPILogin(t *testing.T) {
	router := SetUpTestRouter()
	router.POST("/api/login", auth.LoginHandler)
	
	db := utils.GetTestDB()
	tx := db.Begin()
	defer tx.Rollback()

	router.Use(func(c *gin.Context) {
        c.Set("DB", tx)
        c.Next()
    })

	user := models.User{
		Email: "testemail",
		Password: "testpassword",
	}
	jsonUser, err := json.Marshal(user)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonUser))
	assert.NoError(t, err)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}