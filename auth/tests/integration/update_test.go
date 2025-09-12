package auth_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	
	"github.com/andro-kes/Chat/auth/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestUpdateUserHandler(t *testing.T) {
	router := SetUpTestRouter()
	router.PATCH("/api/update", auth.UpdateUser)

	db := utils.GetTestDB()
	tx := db.Begin()
	defer tx.Rollback()

	router.Use(func(c *gin.Context) {
        c.Set("DB", tx)
        c.Next()
    })

	user := models.User{
		Email: "testemail",
		Password: "newpassword",
	}
	jsonUser, err := json.Marshal(user)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("PATCH", "/api/update", bytes.NewBuffer(jsonUser))
	router.ServeHTTP(w, req)
	assert.NoError(t, err)
	assert.Equal(t, 200, w.Code)

	c := gin.CreateTestContextOnly(w, router)
	c.Request = req

	var updateUser models.User
	obj := tx.First(&updateUser)
	assert.NoError(t, obj.Error)

	err = utils.CompareHashPasswords("testpassword", updateUser.Password)
	assert.NotEqual(t, nil, err)
}