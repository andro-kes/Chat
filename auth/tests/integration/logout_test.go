package auth_tests

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/andro-kes/Chat/auth/internal/middlewares"
	"github.com/andro-kes/Chat/auth/internal/models"

	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPILogout(t *testing.T) {
	authHandlers := SetUp(t)
	authMiddlewares := middlewares.NewAuthMiddlewares()

	user := models.User{
		Email: "testemail",
		Password: "testpassword",
	}
	jsonUser, err := json.Marshal(user)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonUser))
	assert.NoError(t, err)
	authHandlers.LoginHandler(w, r)
	assert.Equal(t, 200, w.Code)

	cookies := w.Result().Cookies()
	cookiesSet := make(map[string]*http.Cookie)
	for _, cookie := range cookies {
		cookiesSet[cookie.Name] = cookie
	}
	accessToken, ok := cookiesSet["access_token"]
	assert.True(t, ok)
	refreshToken, ok := cookiesSet["refresh_token"]
	assert.True(t, ok)

	W := httptest.NewRecorder()
	R, err := http.NewRequest("POST", "/api/logout", nil)
	assert.NoError(t, err)
	
	// Создаем куки с правильным временем истечения
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken.Value,
		Expires:  time.Now().Add(5 * time.Minute),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	refreshCookie := &http.Cookie{
		Name:     "refresh_token", 
		Value:    refreshToken.Value,
		Expires:  time.Now().Add(720 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	
	R.AddCookie(accessCookie)
	R.AddCookie(refreshCookie)
	R.Header.Set("Authentication", accessToken.Value)
	logout := authMiddlewares.AuthMiddleware(http.HandlerFunc(authHandlers.LogoutHandler))
	logout.(http.HandlerFunc)(W, R)

	assert.Equal(t, 200, W.Code)

	cookies = W.Result().Cookies()
	logoutCookies := make(map[string]string)
	for _, cookie := range cookies {
		logoutCookies[cookie.Name] = cookie.Value
	}
	val, ok := logoutCookies["access_token"]
	assert.Equal(t, "", val)
	val, ok = logoutCookies["refresh_token"]
	assert.Equal(t, "", val)
}