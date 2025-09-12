package auth_tests

import (
	"bytes"
	"encoding/json"

	
	"github.com/andro-kes/Chat/auth/internal/utils"

	
	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogoutHandler(t *testing.T) {
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
	cookies := w.Result().Cookies()

	router.POST("/logout", auth.LogoutHandler)
	w = httptest.NewRecorder()
	req, err = http.NewRequest("POST", "/logout", nil)
	assert.NoError(t, err)

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	c := gin.CreateTestContextOnly(w, router)
	c.Request = req
	
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	cookies = w.Result().Cookies()
	var (
		token string
		refresh_token string
	)
	for _, cookie := range cookies {
		if cookie.Name == "token" {
			token = cookie.Value
		}
		if cookie.Name == "refresh_token" {
			refresh_token = cookie.Value
		}
	}
	assert.Empty(t, token)
	assert.Empty(t, refresh_token)

	var RefreshToken models.RefreshTokens
	tx.Where("user_id = ?", uint(1)).First(&RefreshToken)
	assert.Equal(t, "", RefreshToken.Token)
}