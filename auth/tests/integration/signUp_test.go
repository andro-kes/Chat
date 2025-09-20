package auth_tests

// import (
// 	"bytes"

	
// 	"github.com/andro-kes/Chat/auth/internal/utils"
	

// 	"github.com/stretchr/testify/assert"

// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestSignUpHandler(t *testing.T) {
// 	router := SetUpTestRouter()
// 	router.GET("/signup_page", auth.SignUPPageHandler)
// 	router.POST("/api/signup", auth.SignUpHandler)

// 	db := utils.GetTestDB()
// 	tx := db.Begin()
// 	defer tx.Rollback()

// 	router.Use(func(c *gin.Context) {
//         c.Set("DB", tx)
//         c.Next()
//     })

// 	user := models.User{
// 		Email: "test",
// 		Password: "test",
// 	}
// 	jsonUser, err := json.Marshal(user)
// 	assert.NoError(t, err)
// 	w := httptest.NewRecorder()
// 	req, err := http.NewRequest("POST", "/api/signup", bytes.NewBuffer(jsonUser))
// 	assert.NoError(t, err)
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, 200, w.Code)

// 	c := gin.CreateTestContextOnly(w, router)
// 	c.Request = req

// 	var createdUser models.User
// 	obj := tx.Where("email = ?", user.Email).First(&createdUser)
// 	assert.NoError(t, obj.Error)
// 	assert.Equal(t, user.Email, createdUser.Email)

// 	err = utils.CompareHashPasswords(user.Password, createdUser.Password)
// 	assert.NoError(t, err)
// }