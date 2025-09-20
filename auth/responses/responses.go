// ВРЕМЕННО: Пакет responses предоставляет вспомогательные функции для ответов
// HTML и JSON. Комментарии временные на период рефакторинга.
package responses

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/andro-kes/Chat/auth/logger"
	"go.uber.org/zap"
)

var templates *template.Template

func Init() {
	templates = template.Must(template.ParseGlob("/app/web/templates/*"))
}

// SendHTMLResponse ВРЕМЕННО: рендерит HTML-шаблон и пишет ответ клиенту
func SendHTMLResponse(w http.ResponseWriter, statusCode int, name string, data map[string]any) {
	var buf bytes.Buffer

	err := templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		if h, ok := w.(http.Hijacker); ok {
            if conn, _, err := h.Hijack(); err == nil {
                conn.Close()
            }
        }
		logger.Log.Error(
			"Не удалось загрузить html страницу",
			zap.String("name", name),
			zap.Error(err),
		)
		SendJSONResponse(w, 500, map[string]any{
			"Error": "Внутренняя ошибка",
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	buf.WriteTo(w)
}

// SendJSONResponse ВРЕМЕННО: отправляет JSON-ответ с указанным статусом
func SendJSONResponse(w http.ResponseWriter, statusCode int, data map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}