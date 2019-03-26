package main

import (
	"math/rand"
	"time"

	"github.com/makeitplay/arena"
	"github.com/makeitplay/client-player-go"
	"github.com/makeitplay/the-dummies-go/dummy"
	"github.com/makeitplay/the-dummies-go/strategy"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
)

func main() {
	rand.Seed(time.Now().Unix())
	serverConfig := new(client.Configuration)
	serverConfig.ParseFromFlags()

	gamer := &client.Gamer{}

	dummy.PlayerNumber = serverConfig.PlayerNumber
	dummy.TeamPlace = serverConfig.TeamPlace
	dummy.MyRule = strategy.DefinePlayerRule(serverConfig.PlayerNumber)
	dummy.TeamBallPossession = dummy.TeamPlace
	dummy.ClientResponder = gamer

	gamer.OnAnnouncement = reactToNewState

	if err := gamer.Play(dummy.GetInitialRegion().Center(serverConfig.TeamPlace), serverConfig); err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	select {
	case <-signalChan:
		logrus.Print("*********** INTERRUPTION SIGNAL ****************")
		gamer.StopToPlay(true)
	}

}

func reactToNewState(ctx client.TurnContext) {

	switch ctx.GameMsg().GameInfo.State {
	case arena.Listening:
		if ctx.GameMsg().Ball().Holder != nil {
			dummy.TeamBallPossession = ctx.GameMsg().Ball().Holder.TeamPlace
		}

		player := &dummy.Dummy{
			GameMsg:     ctx.GameMsg(),
			Player:      ctx.Player(),
			PlayerState: strategy.DetermineMyState(ctx),
			TeamState:   strategy.DetermineMyTeamState(ctx, dummy.TeamBallPossession),
			Logger:      ctx.Logger(),
		}

		ctx.Logger().Infof("my state: %s", player.PlayerState)
		player.React()
		dummy.LastHolderFrom = ctx.GameMsg().Ball().Holder
	}
}
