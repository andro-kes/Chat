package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/andro-kes/Chat/chat/internal/models"
	"github.com/andro-kes/Chat/chat/internal/services"
	"github.com/andro-kes/Chat/chat/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type RabbitManager interface {
	PublishMessage(msg models.Message) error
	ConsumeMessages()
	Stop()
}

type rabbitManager struct {
	conn        *amqp.Connection
	ch          *amqp.Channel
	q           amqp.Queue
	ChatService services.ChatService
}

func Init(chatSvc services.ChatService) (*rabbitManager, error) {
	var rm rabbitManager
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	name := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	addr := os.Getenv("RABBITMQ_ADDR")
	if addr == "" {
		addr = "localhost:5672"
	}
	connURL := fmt.Sprintf("amqp://%s:%s@%s/", name, password, addr)

	// simple retry/backoff
	backoff := time.Second
	for attempts := 0; attempts < 5; attempts++ {
		rm.conn, err = amqp.Dial(connURL)
		if err == nil {
			break
		}
		logger.Log.Warn("RabbitMQ dial failed, retrying", zap.Int("attempt", attempts+1), zap.Error(err))
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}
	if err != nil {
		logger.Log.Error("Не удалось установить соединение с очередью", zap.Error(err))
		return nil, err
	}

	rm.ch, err = rm.conn.Channel()
	if err != nil {
		logger.Log.Error("Не удалось создать канал для очереди", zap.Error(err))
		return nil, err
	}

	// durable queue
	rm.q, err = rm.ch.QueueDeclare(
		"chat",
		true,  // durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Log.Error("Не удалось создать очередь", zap.Error(err))
		return nil, err
	}

	rm.ChatService = chatSvc

	// старт consumer в отдельной горутине
	go rm.ConsumeMessages()

	return &rm, nil
}

// PublishMessage публикует сообщение в очередь RabbitMQ.
func (rm *rabbitManager) PublishMessage(msg models.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Error("Не удалось сериализовать сообщение", zap.Error(err))
		return err
	}

	err = rm.ch.PublishWithContext(
		context.Background(),
		"",        // default exchange
		rm.q.Name, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		logger.Log.Error("Не удалось опубликовать сообщение", zap.Error(err))
		return err
	}
	return nil
}

// consumeMessages считывает сообщения из очереди и рассылает их через WebSocket.
func (rm *rabbitManager) ConsumeMessages() {
	msgs, err := rm.ch.Consume(
		rm.q.Name,  // Имя очереди
		"",      // Имя потребителя
		true,    // Auto-ack
		false,   // Exclusive
		false,   // No-local
		false,   // No-wait
		nil,     // Args
	)
	if err != nil {
		logger.Log.Error("Не удалось подписаться на очередь", zap.Error(err))
		return
	}

	for d := range msgs {
		var msg models.Message
		err := json.Unmarshal(d.Body, &msg)
		if err != nil {
			logger.Log.Error("Не удалось десериализовать сообщение", zap.Error(err))
			continue
		}

		room, err := rm.ChatService.GetRoom(msg.RoomID)
		if err != nil {
			logger.Log.Error("Неверное id комнаты", zap.Error(err))
			continue
		}
		room.SendMessage(&msg)

		logger.Log.Info("Получено сообщение через RabbitMQ", zap.Any("message", msg))
	}
}

func (rm *rabbitManager) Stop() {
	rm.ch.Close()
	rm.conn.Close()
}