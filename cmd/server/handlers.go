package main

import (
	"fmt"

	"github.com/OferRavid/learn-pub-sub-starter/internal/gamelogic"
	"github.com/OferRavid/learn-pub-sub-starter/internal/pubsub"
	"github.com/OferRavid/learn-pub-sub-starter/internal/routing"
)

func handlerLogs() func(routing.GameLog) pubsub.AckType {
	defer fmt.Print("> ")
	return func(gl routing.GameLog) pubsub.AckType {
		err := gamelogic.WriteLog(gl)
		if err != nil {
			return pubsub.NackRequeue
		}

		return pubsub.Ack
	}
}
