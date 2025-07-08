package models

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

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
	TaskQueue chan MessageTask
	StopWorkers chan struct{}
	Mu sync.RWMutex
	Wg sync.WaitGroup
}

type MessageTask struct {
	User *UserData
	Msg *Message
}

func (roomData *RoomData) SendMessage (msg *Message) {
	roomData.Mu.RLock()
	msg.Save()
	roomData.Mu.RUnlock()

	for _, user := range roomData.ActiveUsers {
		select{
			case roomData.TaskQueue <- MessageTask{
				User: user,
				Msg: msg,
			}:
			default:
				log.Println("Очередь сообщений переполнена")
		}
	}
}

func (roomData *RoomData) Registered (user *UserData) {
	roomData.Mu.Lock()
	defer roomData.Mu.Unlock()

	if activeUser, ok := roomData.ActiveUsers[user.User.ID]; ok {
		activeUser.CloseConn()
	}

	roomData.ActiveUsers[user.User.ID] = user
}

func (roomData *RoomData) Unregistered (user *UserData) {
	roomData.Mu.Lock()
	defer roomData.Mu.Unlock()

	user.CloseConn()
	delete(roomData.ActiveUsers, user.User.ID)
}

func (room *RoomData) CheckActive() bool {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	return len(room.ActiveUsers) > 0
}

func (roomData *RoomData) StartWork(n int) {
	roomData.StopWorkers = make(chan struct{})
	roomData.TaskQueue = make(chan MessageTask, 1000)

	for range n {
		roomData.Wg.Add(1)
		go roomData.worker()
	}
}

func (roomData *RoomData) worker() {
	defer func() {
		roomData.Wg.Done()
        if r := recover(); r != nil {
            log.Printf("Worker восстановлен: %v", r)
        }
    }()

	for {
		select {
		case task, ok := <-roomData.TaskQueue:
			if !ok {
				return
			}
			roomData.CompleteTask(task)
			
		case <-roomData.StopWorkers:
			return
		}
	}
}

func (roomData *RoomData) CompleteTask(task MessageTask) {
	task.User.Mu.Lock()
	defer task.User.Mu.Unlock()

	task.User.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := task.User.Conn.WriteJSON(task.Msg)
	if err != nil {
		log.Println("Ошибка при получении сообщения:", err.Error())
		roomData.Unregistered(task.User)
	}
}

func (roomData *RoomData) Stop() {
	roomData.Mu.Lock()
	defer roomData.Mu.Unlock()

	close(roomData.StopWorkers)

	for _, user := range roomData.ActiveUsers {
		user.CloseConn()
	}
	roomData.ActiveUsers = make(map[uint]*UserData)

	go func() {
		roomData.Wg.Wait()
		safeChanClose(roomData.TaskQueue)
	}()
}

func safeChanClose[T any](ch chan T) {
	defer func() { recover() }()
    close(ch)
}