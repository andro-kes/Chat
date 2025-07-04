package models

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

const WORKERS_COUNT = 5

type Room struct {
	gorm.Model
	Name string `json:"name"`
	AdminID uint
	Admin User `gorm:"foreignKey:AdminID"`
	Users []*User `gorm:"many2many:room_users"`
	Messages []*Message `gorm:"foreignKey:RoomID"`
}

type RoomData struct {
	Room Room
	ActiveUsers map[uint]*UserData
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
			room.Mu.Lock()
			if existingUser, ok := room.ActiveUsers[user.User.ID]; ok {
				existingUser.Conn.Close()
			}
			room.ActiveUsers[user.User.ID] = user
			room.Mu.Unlock()
		case user := <-room.Unregistered:
			if user.Conn != nil {
				user.Conn.Close()
			}
			room.Mu.RLock()
			delete(room.ActiveUsers, user.User.ID)
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
	for _, user := range room.ActiveUsers {
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
	for {
		select {
		case task, ok := <- room.TaskQueue: 
			if !ok {
				return
			}
			
			task.User.Mu.Lock()
			err := task.User.Conn.WriteJSON(task.Msg)
			task.User.Mu.Unlock()
			if err != nil {
				select {
				case task.Room.Unregistered <- task.User:
					log.Println("SendMessage: Не удалось отправить сообщение пользователю", task.User.User.Username)
					log.Println("=> Отключение из-за", err.Error())
				default:
					log.Println("Очередь переполнена")
				}	
			}
			
		case <-room.StopWorkers:
			return
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