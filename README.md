# Lugo - The Dummies Go

[![GoDoc](https://godoc.org/github.com/lugobots/the-dummies-go?status.svg)](https://godoc.org/github.com/lugobots/the-dummies-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lugobots/the-dummies-go)](https://goreportcard.com/report/github.com/lugobots/the-dummies-go)

The Dummies Go is a [Go](http://golang.org/) implementation of a player (bot) for [Lugo](https://lugobots.dev) game.
This bot was made using the [Go Client Player](https://github.com/lugobots/lugo4go).

As this name suggest, _The Dummies_ are not that smart, but they may play well enough to help you to test your bot.

## Requirements

1. Docker >= 18.03 (https://docs.docker.com/install/)
2. Docker Compose >= 1.21 (https://docs.docker.com/compose/install/)

## Before starting

Are you familiar with Lugo?
If not, before continuing, please visit [the project website](https://lugobots.dev) and read about the game.

## How to use this source code

1. **Checkout the code** or download the most recent tag release
2. **Test it out**: Before any change, make The Dummies Go to play to ensure you are not working on a broken code.

   _Note_: this step can take a little longer at the first time.
   ```sh 
   docker-compose up
   ```
   and open [http://localhost:8080/](http://localhost:8080/) to watch the game.
4. **Now, make your changes**: (see :question:[How to change the bot](#how-to-edit-the-bot))
5. Play again to see your changes results:

   ```sh 
   docker-compose up
   ```
6. **Are you ready to compete? Build your Docker image:**

    ```sh 
   docker build -t my-super-bot .
   ```
7. :checkered_flag: Before pushing your changes

   ```sh 
   MY_BOT=my-super-bot docker-compose --file docker-compose-test.yml up
   ```

## How to edit the bot

You will likely never need to change the `main.go` file. You should focus on the files inside the directory `bot/`.

### 1. Define bots' position ([bot/settings.go](bot/settings.go))

Settings file stores configurations and functions that will affect the player behaviour, e.g. positions, tactic, etc.

### 2. Define bots' roles ([bot/helpers.go](bot/helpers.go))

The helper functions will define what are the player roles, and how to determine the team state.

### 3. Define bots' logic ([bot/bot.go](bot/bot.go))

The [bot/bot.go](bot/bot.go) file is the most important one. There you implement the interface that define the bot
logic.

There will be 5 important methods that you must edit to change the bot behaviour.

```go
type Bot interface {
   // OnDisputing is called when no one has the ball possession
   OnDisputing(ctx context.Context, inspector lugo4go.SnapshotInspector) ([]proto.PlayerOrder, string, error)
   // OnDefending is called when an opponent player has the ball possession
   OnDefending(ctx context.Context, inspector lugo4go.SnapshotInspector) ([]proto.PlayerOrder, string, error)
   // OnHolding is called when this bot has the ball possession
   OnHolding(ctx context.Context, inspector lugo4go.SnapshotInspector) ([]proto.PlayerOrder, string, error)
   // OnSupporting is called when a teammate player has the ball possession
   OnSupporting(ctx context.Context, inspector lugo4go.SnapshotInspector) ([]proto.PlayerOrder, string, error)
   // AsGoalkeeper is only called when this bot is the goalkeeper (number 1). This method is called on every turn,
   // and the player state is passed at the last parameter.
   AsGoalkeeper(ctx context.Context, inspector lugo4go.SnapshotInspector, state PlayerState) ([]proto.PlayerOrder, string, error)
}
```

After editing the files, you do not need to build the binary. It will be automatically done by a container ran by the
Docker Compose file.


## Running locally

You may need to run your bot locally for some reasons (e.g. training a bot using machine learning).

The script `play.sh` builds the binary and starts up 11 bots.

1. Start the game server (e.g. `docker run -p 8080:8080 -p 5000:5000 lugobots/server:latest play --dev-mode`)
2. Then, execute `./play.sh home` or `./play.sh away` to start a full team.

