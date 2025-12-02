package main

import (
	"fmt"
	"log"

	"github.com/OferRavid/learn-pub-sub-starter/internal/gamelogic"
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
	gamelogic.PrintServerHelp()
	amqpChannel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Could not create channel: %v", err)
	}

	for {
		inputs := gamelogic.GetInput()

		if len(inputs) == 0 {
			continue
		}
		switch inputs[0] {
		case "pause":
			fmt.Println("Sending a pause message to server...")
			err = pubsub.PublishJSON(
				amqpChannel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}

		case "resume":
			fmt.Println("Sending a resume message to server...")
			err = pubsub.PublishJSON(
				amqpChannel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}

		case "quit":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Couldn't identify the command. Try again.")
			gamelogic.PrintServerHelp()
		}
	}
}
