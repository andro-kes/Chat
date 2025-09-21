package auth_tests

import (
	"bytes"

	"github.com/andro-kes/Chat/auth/internal/models"

	"github.com/stretchr/testify/assert"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPILogin(t *testing.T) {
	authHandlers := SetUp(t)
	
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
	_, ok := cookiesSet["access_token"]
	assert.True(t, ok)
	_, ok = cookiesSet["refresh_token"]
	assert.True(t, ok)
}