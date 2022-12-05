package main

import (
	"github.com/streadway/amqp"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Text struct {
	gorm.Model
	Text string `json:"text" gorm:"unique;not null"`
}

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database")
	}

	err = database.AutoMigrate(&Text{})

	if err != nil {
		return
	}

	DB = database
}

var CONNECTION *amqp.Connection
var CHANNEL *amqp.Channel

func ConnectQueue() {
	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	CONNECTION = connection

	channel, err := connection.Channel()
	if err != nil {
		panic(err)
	}
	CHANNEL = channel
}

func main() {
	ConnectDatabase()

	ConnectQueue()
	defer CONNECTION.Close()
	defer CHANNEL.Close()

	messages, err := CHANNEL.Consume(
		"text-sub", // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		panic(err)
	}

	forever := make(chan bool)
	go func() {
		for message := range messages {
			DB.Create(&Text{Text: string(message.Body)})
		}
	}()
	<-forever
}
