package models

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

const WORKERS_COUNT = 20

type Room struct {
	gorm.Model
	Name string `json:"name"`
	AdminID uint
	Admin User `gorm:"foreignKey:AdminID"`
	Users []*User `gorm:"many2many:room_users"`
	Messages []*Message `gorm:"foreignKey:RoomID"`
}

func (room *Room) CheckAccess(user *User) bool {
	for _, u := range room.Users {
		if u.ID == user.ID {
			return true
		}
	}
	return false
}

type RoomData struct {
	Room Room
	ActiveUsers map[*UserData]bool
	Broadcast chan *Message
	Registered chan *UserData
	Unregistered chan *UserData
	TaskQueue chan MessageTask
	Close chan bool
	StopWorkers chan bool
	Mu sync.RWMutex
}

type MessageTask struct {
	User *UserData
	Msg *Message
	Room *RoomData
}

func (room *RoomData) Run() {
	room.StopWorkers = make(chan bool)
	room.StartWork(WORKERS_COUNT)
	for {
		select {
		case msg := <-room.Broadcast:
			room.Mu.RLock()
			msg.Save()
			room.SendMessage(msg)
			room.Mu.RUnlock()
		case user := <-room.Registered:
			if _, ok := room.ActiveUsers[user]; !ok {
				room.ActiveUsers[user] = true
			}
		case user := <-room.Unregistered:
			user.Conn.Close()
			room.Mu.RLock()
			delete(room.ActiveUsers, user)
			room.Mu.RUnlock()
		case <-room.Close:
			room.StopWorkers <- true
			return
		}
	}
}

func (room *RoomData) CheckActive() bool {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	return len(room.ActiveUsers) > 0
}

func (room *RoomData) SendMessage(msg *Message) {
	for user := range room.ActiveUsers {
		select{
			case room.TaskQueue <- MessageTask{
				User: user,
				Msg: msg,
				Room: room,
			}:
			default:
				log.Println("Очередь сообщений переполнена")
		}
	}
}

func (room *RoomData) StartWork(n int) {
	for range n{
		go room.worker()
	}
}

func (room *RoomData) worker() {
	defer func() {
        if r := recover(); r != nil {
            log.Printf("Worker восстановлен: %v", r)
        }
    }()

	for task := range room.TaskQueue {
		task.User.Mu.Lock()
		err := task.User.Conn.WriteJSON(task.Msg)
		task.User.Mu.Unlock()
		if err != nil {
			select {
			case task.Room.Unregistered <- task.User:
				log.Println("SendMessage: Не удалось отправить сообщение пользователю", task.User.User.Username)
				log.Println("=> Отключение")
			case <-room.StopWorkers:
				return
			default:
				log.Println("Очередь переполнена")
			}
		}	
	}
}

func (room *RoomData) Stop() {
	close(room.TaskQueue)
	time.Sleep(100 * time.Millisecond)

	room.Close <- true
	SafeChanClose(room.Close)
	SafeChanClose(room.Registered)
	close(room.Unregistered)
	close(room.Broadcast)
	close(room.StopWorkers)
}

func SafeChanClose[T any](ch chan T) {
	defer func() { recover() }()
    close(ch)
}