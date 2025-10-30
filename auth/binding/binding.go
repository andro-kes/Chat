package binding

import (
	"encoding/json"
	"net/http"

	"github.com/andro-kes/Chat/auth/internal/models"
	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
)

// BindUserWithJSON декодирует JSON тело запроса в переданную структуру user
func BindUserWithJSON(r *http.Request, user *models.User) error {
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		logger.Log.Warn("Не удалось декодировать данные пользователя", zap.Error(err))
	}
	return err
}