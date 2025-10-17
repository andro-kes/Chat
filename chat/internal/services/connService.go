package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ConnService interface {
	GetConn(userId uuid.UUID) (*websocket.Conn, error)
	SetConn(userId uuid.UUID, conn *websocket.Conn)
	DeleteConn(userId uuid.UUID)
}

type connService struct {
	Conns map[uuid.UUID]*websocket.Conn
}

// NewConnService создает и возвращает новый экземпляр сервиса управления WebSocket-соединениями.
// Сервис использует карту для хранения активных соединений, где ключом является уникальный идентификатор (UUID),
// а значением — объект соединения *websocket.Conn.
// Пример использования:
//   service := NewConnService()
func NewConnService() *connService {
	return &connService{
		Conns: make(map[uuid.UUID]*websocket.Conn),
	}
}

func (cs *connService) GetConn(userId uuid.UUID) (*websocket.Conn, error) {
	if conn, ok := cs.Conns[userId]; ok {
		return conn, nil
	}
	return nil, errors.New("активное соединение не найдено")
}

func (cs *connService) SetConn(userId uuid.UUID, conn *websocket.Conn) {
	if _, ok := cs.Conns[userId]; !ok {
		cs.Conns[userId] = conn
	}
}

func (cs *connService) DeleteConn(userId uuid.UUID) {
	delete(cs.Conns, userId)
}