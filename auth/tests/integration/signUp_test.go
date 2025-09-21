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

func TestAPISignUp(t *testing.T) {
	authHandlers := SetUp(t)

	user := models.User{
		Username: "newuser",
		Password: "newpassword",
		Email: "newemail",
	}
	jsonUser, err := json.Marshal(user)
	assert.NoError(t, err)
	
	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", "/api/sign_up", bytes.NewBuffer(jsonUser))
	assert.NoError(t, err)
	authHandlers.SignUpHandler(w, r)
	assert.Equal(t, 200, w.Code)

	createdUser, err := authHandlers.UserService.GetUserByEmail(user.Email)
	assert.NoError(t, err)

	assert.Equal(t, createdUser.Username, "newuser")
	assert.Equal(t, createdUser.Email, "newemail")
}