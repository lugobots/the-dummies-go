package main

import (
	"log"
	"net/http"
	"time"

	"github.com/lugobots/lugo4go/v3"

	"my-bot/bot"
)

func main() {
	connectionStarter, defaultFieldMapper, err := lugo4go.NewDefaultStarter()
	if err != nil {
		log.Fatalf("failed to load the bot configuration: %s", err)
	}

	// OPTIONAL
	// define your own field mapper! The default number of col/rows are defined by lugo4go.DefaultFieldMapCols and lugo4go.DefaultFieldMapRows
	//defaultFieldMapper, err = field.NewMapper(NUM_COLS, NUM_ROWS, connectionStarter.Config.TeamSide)
	//if err != nil {
	//	log.Fatalf("failed to create a field mapper: %s", err)
	//}

	// create your bot as you wish
	// in this example, the bot requires the field mapper, the connection config, and a logger.
	myBot := bot.NewBot(
		defaultFieldMapper,
		connectionStarter.Config,
		connectionStarter.Logger,
	)

	// Here you define the initial position of your bot. It's important to use the field mapper instead of points because
	// the field mapper won't be affected when your bot is playing on the away side
	initialPosition := bot.DefaultInitialPositions[connectionStarter.Config.Number]
	region, err := defaultFieldMapper.GetRegion(initialPosition.Col, initialPosition.Row)
	if err != nil {
		log.Fatalf("failed to define initial position using field mapper: %s", err)
	}
	connectionStarter.Config.InitialPosition = region.Center()

	// then lets play
	if err := connectionStarter.Run(myBot); err != nil {
		log.Fatalf("bot stopped: %s", err)
	}
}

func InternetAvailable() bool {
	client := http.Client{
		Timeout: 5 * time.Second, // avoid hanging too long
	}

	resp, err := client.Get("https://www.google.com")
	if err != nil {
		log.Printf("err: %v\n", err)
		return false
	}
	defer resp.Body.Close()
	log.Printf("STATUS: %v\n", resp.StatusCode)
	return resp.StatusCode == http.StatusOK
}
