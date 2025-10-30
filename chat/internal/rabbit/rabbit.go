package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
	// можно добавить флаг завершения или context для остановки потребителя
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

	// QoS / prefetch — сделаем 1 (по одному сообщению на консьюмер)
	// Можно взять из ENV: RABBITMQ_PREFETCH
	prefetch := 1
	if v := os.Getenv("RABBITMQ_PREFETCH"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			prefetch = p
		}
	}
	if err := rm.ch.Qos(prefetch, 0, false); err != nil {
		logger.Log.Warn("Не удалось установить QoS для канала RabbitMQ", zap.Error(err))
		// не фаталим — но логируем
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
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		logger.Log.Error("Не удалось опубликовать сообщение", zap.Error(err))
		return err
	}
	return nil
}

// ConsumeMessages читает сообщения из очереди и рассылает их через WebSocket.
// Используется manual ack: autoAck=false в Consume, после успешной обработки вызывается delivery.Ack(false).
func (rm *rabbitManager) ConsumeMessages() {
	msgs, err := rm.ch.Consume(
		rm.q.Name, // queue
		"",        // consumer
		false,     // autoAck == false -> подтверждать вручную
		false,     // exclusive
		false,     // noLocal (not supported by rabbitmq server if true)
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		logger.Log.Error("Не удалось подписаться на очередь", zap.Error(err))
		return
	}

	logger.Log.Info("RabbitMQ consumer started", zap.String("queue", rm.q.Name))

	for d := range msgs {
		var msg models.Message
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			logger.Log.Error("Не удалось десериализовать сообщение", zap.Error(err))
			if ackErr := d.Nack(false, false); ackErr != nil {
				logger.Log.Warn("Не удалось Nack сообщение (bad json)", zap.Error(ackErr))
			}
			continue
		}

		room, err := rm.ChatService.GetRoom(msg.RoomID)
		if err != nil {
			logger.Log.Error("Неверное id комнаты", zap.Error(err))
			// если комната не найдена — отдаем Nack с requeue=false (чтобы не зацикливать), либо requeue=true если хотим повторить позже
			if nackErr := d.Nack(false, false); nackErr != nil {
				logger.Log.Warn("Не удалось Nack сообщение (room not found)", zap.Error(nackErr))
			}
			continue
		}

		// Попытка доставки — SendMessage может возвращать ошибку (например, DB down или проблемa с websocket)
		if err := room.SendMessage(&msg); err != nil {
			logger.Log.Warn("Ошибка при отправке сообщения в комнату", zap.Any("message", msg), zap.Error(err))
			if nackErr := d.Nack(false, true); nackErr != nil {
				logger.Log.Warn("Не удалось Nack (requeue) сообщение", zap.Error(nackErr))
			}
			continue
		}

		// Успешно обработано — подтверждаем delivery
		if ackErr := d.Ack(false); ackErr != nil {
			logger.Log.Warn("Не удалось Ack сообщение", zap.Error(ackErr))
		} else {
			logger.Log.Info("Получено и подтверждено сообщение через RabbitMQ", zap.Any("message", msg))
		}
	}
	logger.Log.Info("RabbitMQ consumer loop exited")
}

// Stop корректно закрывает канал и соединение
func (rm *rabbitManager) Stop() {
	if rm.ch != nil {
		_ = rm.ch.Close()
	}
	if rm.conn != nil {
		_ = rm.conn.Close()
	}
}