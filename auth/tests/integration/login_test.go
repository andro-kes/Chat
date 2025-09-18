package auth_tests

import (
	"bytes"

	"github.com/andro-kes/Chat/auth/internal/handlers"
	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/internal/utils"

	"github.com/stretchr/testify/assert"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPILogin(t *testing.T) {
	authHandlers := NewAuthHandlers()
	http.HandleFunc("/api/login", authHandlers.LoginHandler)
	
	pool := SetUp(t)
	
	user := models.User{
		Email: "testemail",
		Password: "testpassword",
	}
	jsonUser, err := json.Marshal(user)
	assert.NoError(t, err)
	r, err := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonUser))
	assert.NoError(t, err)
	
	client := &http.Client{}
	resp, err := client.Do(r)
	assert.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, 200, resp.StatusCode)
}