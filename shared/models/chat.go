package models

import (
	"sync"
	"log"
	"slices"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Room struct {
	gorm.Model
	AdminID uint
	Admin User `gorm:"foreignKey:AdminID"`
	Users []*User `gorm:"many2many:room_users"`
	Messages []*Message `gorm:"foreignKey:RoomID"`
}

type RoomData struct {
	Room Room
	Broadcast chan *Message
	Registered chan *UserData
	Unregistered chan *UserData
	Mu sync.Mutex
}

func (room *RoomData) Run() {
	for {
		select {
		case msg := <-room.Broadcast:
			room.Mu.Lock()
			msg.Room = &room.Room
			DB.Create(msg)
			msg.Save()
			room.Mu.Unlock()
		case user := <-room.Registered:
			ok := slices.Contains(room.Room.Users, &user.User)
			if !ok {
				room.Mu.Lock()
				err := DB.Model(&room.Room).Association("Users").Append(user.User)
				if err != nil {
					log.Println("Models chat error: Не удалось добавить пользователя")
				}
				room.Mu.Unlock()
			}
		case user := <-room.Unregistered:
			user.Conn.Close()
		}
	}
}

type Message struct {
	gorm.Model
	SenderID uint
	Sender User `gorm:"foreignKey:SenderID"`
	RoomID uint
	Room *Room `gorm:"foreignKey:RoomID"`
	Text string
}

func (msg *Message) Save() {
	err := DB.Model(&msg.Room).Association("Messages").Append(msg)
	if err != nil {
		log.Println("Не удалось сохранить сообщение")
	}
}