package binding

import (
	"encoding/json"
	"net/http"

	"github.com/andro-kes/Chat/chat/logger"
)

func BindWithJSON(r *http.Request, obj any) error  {
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		logger.Log.Warn(
			"Не удалось декодировать данные",
		)
	}
	
	return err
}