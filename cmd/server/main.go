package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/OferRavid/learn-pub-sub-starter/internal/pubsub"
	"github.com/OferRavid/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const connectionString = `amqp://guest:guest@localhost:5672/`
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer connection.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")
	amqpChannel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Could not create channel: %v", err)
	}

	pubsub.PublishJSON(amqpChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: true,
	})

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("\nClosing AMQP connection gracefully...")
	fmt.Println("\nAMQP connection closed.")
}
