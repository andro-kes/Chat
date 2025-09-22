package responses

import (
	"encoding/json"
	"net/http"

	"github.com/andro-kes/Chat/chat/logger"
	"go.uber.org/zap"
)

func SendJSONResponse(w http.ResponseWriter, statusCode int, data map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	logger.Log.Error(
		"Не удалось сериализовать ответ",
		zap.String("service", "auth"),
		zap.Error(err),
	)
}