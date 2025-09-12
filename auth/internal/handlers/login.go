package auth

import (
	"net/http"

	"github.com/andro-kes/Chat/auth/binding"
	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/internal/services"
	"github.com/andro-kes/Chat/auth/internal/utils"
	"github.com/andro-kes/Chat/auth/responses"
)

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	responses.SendHTMLResponse(w, 200, "login.html", map[string]any{
		"title": "login",
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	userService := services.NewUserService()
	tokenService := services.NewTokenService()

	var user models.User
	err := binding.BindUserWithJSON(r, &user)
	if err != nil {
		responses.SendJSONResponse(w, 400, map[string]any{
			"Error": err.Error(),
		})
		return
	}

	loginUser, err := userService.Login(user)
	if err != nil {
		responses.SendJSONResponse(w, 500, map[string]any{
			"Error": err.Error(),
		})
		return
	}

	refreshTokenString, err := tokenService.GenerateRefreshToken(loginUser.ID)
	if err != nil {
		responses.SendJSONResponse(w, 500, map[string]any{
			"Error": err.Error(),
		})
		return
	}

	accessTokenString, err := utils.GenerateAccessToken(loginUser.ID)
	if err != nil {
		responses.SendJSONResponse(w, 500, map[string]any{
			"Error": err.Error(),
		})
		return
	}

	tokenService.SetTokenCookie(w, "refresh_token", refreshTokenString)
	tokenService.SetTokenCookie(w, "access_token", accessTokenString)

	responses.SendJSONResponse(w, 200, map[string]any{
		"Message": "Успешный вход в систему",
		"User": loginUser,
	})
}