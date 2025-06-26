package models

import (
	"log"

	"gorm.io/gorm"
)

var DB *gorm.DB

type Message struct {
	gorm.Model
	SenderID uint
	Sender User `gorm:"foreignKey:SenderID"`
	RoomID uint
	Room *Room `gorm:"foreignKey:RoomID"`
	Text string
}

func (msg *Message) Save() {
	obj := DB.Create(msg)
	if obj.Error != nil {
		log.Println("Не удалось сохранить сообщение")
		log.Println(obj.Error.Error())
	}
}