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
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
	ChatService services.ChatService
}

func Init() (*rabbitManager, error) {
	var rm rabbitManager
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	name := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	conn := fmt.Sprintf("amqp://%s:%s@localhost:5672/", name, password)
	rm.conn, err = amqp.Dial(conn)
	var backoff = 1 * time.Second
	for range 5 {
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			time.Sleep(backoff)
			backoff *= 2
			rm.conn, err = amqp.Dial("amqp://andde:androkes@localhost:5672/")
		}
	}
	if err != nil {
		logger.Log.Error(
			"Не удалось установить соединение с очередью",
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	rm.ch, err = rm.conn.Channel()
	if err != nil {
		logger.Log.Error(
			"Не удалось создать канал для очереди",
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	rm.q, err = rm.ch.QueueDeclare(
		"chat", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		logger.Log.Error(
			"Не удалось создать очередь",
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	for range 5 {
		go rm.ConsumeMessages()
	}

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
		"",     // Обменник
		rm.q.Name, // Имя очереди
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
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