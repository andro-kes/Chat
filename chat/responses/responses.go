package responses

import (
	"encoding/json"
	"net/http"
	"html/template"
	"bytes"

	"github.com/andro-kes/Chat/chat/logger"
	"go.uber.org/zap"
)

var templates *template.Template

func Init() {
	templates = template.Must(template.ParseGlob("/app/web/templates/*"))
}

func SendHTMLResponse(w http.ResponseWriter, statusCode int, name string, data map[string]any) {
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		// Попытка закрыть соединение если необходимо
		if h, ok := w.(http.Hijacker); ok {
			if conn, _, err := h.Hijack(); err == nil {
				conn.Close()
			}
		}
		logger.Log.Error("Не удалось загрузить html страницу", zap.String("name", name), zap.Error(err))
		SendJSONResponse(w, 500, map[string]any{"Error": "Внутренняя ошибка"})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	buf.WriteTo(w)
}

func SendJSONResponse(w http.ResponseWriter, statusCode int, data map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Log.Error("Не удалось сериализовать ответ", zap.Error(err))
	}
}