package rabbit

import (
	// "context"
	// "time"

	"github.com/andro-kes/Chat/chat/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func Init() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Log.Error(
			"Не удалось установить соединение с очередью",
			zap.Error(err),
		)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Log.Error(
			"Не удалось создать канал для очереди",
			zap.Error(err),
		)
		return err
	}
	defer ch.Close()

	// q, err := ch.QueueDeclare(
	// 	"chat", // name
	// 	false,   // durable
	// 	false,   // delete when unused
	// 	false,   // exclusive
	// 	false,   // no-wait
	// 	nil,     // arguments
	// )	

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	return nil
}