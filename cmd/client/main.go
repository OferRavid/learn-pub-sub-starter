package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/OferRavid/learn-pub-sub-starter/internal/gamelogic"
	"github.com/OferRavid/learn-pub-sub-starter/internal/pubsub"
	"github.com/OferRavid/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}
	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gs.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gs, publishCh),
	)
	if err != nil {
		log.Fatalf("could not subscribe to queue %s: %v", routing.ArmyMovesPrefix+"."+gs.GetUsername(), err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gs, publishCh),
	)
	if err != nil {
		log.Fatalf("could not subscribe to queue %s: %v", routing.WarRecognitionsPrefix, err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "move":
			armyMove, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+gs.GetUsername(),
				armyMove,
			)
			if err != nil {
				log.Printf("failed to publish move: %v\n", err)
				continue
			}
			log.Println("successfully published move")

		case "spawn":
			err = gs.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(words) != 2 {
				fmt.Println("The spam command consists of the word spam with a number, for example: 'spam 20'")
				continue
			}

			iterations, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Println("To use the spam command, make sure the second word is a number")
				continue
			}

			for range iterations {
				pubsub.PublishGameLog(
					publishCh,
					gs.GetUsername(),
					gamelogic.GetMaliciousLog(),
				)
			}
			// fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
