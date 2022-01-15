package bot

import (
	"context"
	"fmt"
	"github.com/lugobots/lugo4go/v2"
	"github.com/lugobots/lugo4go/v2/lugo"
	"github.com/lugobots/lugo4go/v2/pkg/field"
	"github.com/pkg/errors"
	"math"
)

func (b *Bot) OnHolding(ctx context.Context, sender lugo4go.TurnOrdersSender, snapshot *lugo.GameSnapshot) error {
	opponentSide := field.GetOpponentSide(b.side)
	opponentGoalKeeper := field.GetPlayer(snapshot, opponentSide, field.GoalkeeperNumber)
	opponentTeam := field.GetTeam(snapshot, opponentSide).GetPlayers()
	shouldIShoot, target := ShouldShoot(
		snapshot.Ball.Position,
		opponentGoalKeeper.Position,
		field.GetOpponentGoal(b.side),
		field.GetTeam(snapshot, opponentSide).GetPlayers(),
	)
	if shouldIShoot >= Should {
		kickOrder, err := field.MakeOrderKickMaxSpeed(*snapshot.Ball, *target)
		if err != nil {
			return errors.Wrap(err, "error creating kicking order to shoot on goal")
		}
		b.lastKickTurn = snapshot.Turn
		return processServerResp(sender.Send(ctx, []lugo.PlayerOrder{kickOrder}, "shoot"))
	}

	me := field.GetPlayer(snapshot, b.side, b.number)
	myTeam := field.GetTeam(snapshot, b.side).GetPlayers()
	opponentGoal := field.GetOpponentGoal(b.side)

	moveForwardOrder := field.GoForward(b.side)
	if math.Abs(float64(me.Position.Y-opponentGoal.Center.Y)) < float64(DistanceFar) {
		moveForwardOrder, _ = field.MakeOrderMoveMaxSpeed(*me.Position, opponentGoal.Center)
	}

	shouldIHold := shouldIHoldTheBall(me, opponentGoal, opponentTeam)
	shouldIPass, bestCandidate := shouldIPass(me, snapshot.Ball.Position, opponentGoalKeeper.Position, opponentGoal, myTeam, opponentTeam)

	// we really need to pass the ball, but looks like there are no good options, let's just stop
	// todo needs enhancement We should look for a better path to avoid obstacles
	if shouldIHold < May && shouldIPass < May {
		stopOrder, err := field.MakeOrderMove(*me.Position, opponentGoal.Center, 0)
		if err != nil {
			return errors.Wrap(err, "error creating kicking order to pass")
		}
		return processServerResp(sender.Send(ctx, []lugo.PlayerOrder{stopOrder}, "oops, I am blocked"))
	}

	if shouldIPass < Should && shouldIShoot == May {
		kickOrder, err := field.MakeOrderKickMaxSpeed(*snapshot.Ball, *target)
		if err != nil {
			return errors.Wrap(err, "error creating kicking order to shoot on goal")
		}
		b.lastKickTurn = snapshot.Turn
		return processServerResp(sender.Send(ctx, []lugo.PlayerOrder{kickOrder}, "shoot"))
	}

	if shouldIPass > ShouldNot {
		if shouldIPass == May {
			obstaclesToPlayer, err := findOpponentsOnMyRoute(me.Position, &opponentGoal.Center, field.BallSize, opponentTeam)
			// before passing, if we can advance a little more, let keep moving forward
			if err == nil && (len(obstaclesToPlayer) == 0 || obstaclesToPlayer[0].position.DistanceTo(*me.Position) > 4*field.PlayerSize) {
				return processServerResp(sender.Send(ctx, []lugo.PlayerOrder{moveForwardOrder}, "continue, champ!"))
			}
		}

		kickOrder, err := field.MakeOrderKickMaxSpeed(*snapshot.Ball, *bestCandidate.Position)
		if err != nil {
			return errors.Wrap(err, "error creating kicking order to pass")
		}
		b.lastKickTurn = snapshot.Turn
		return processServerResp(sender.Send(ctx, []lugo.PlayerOrder{kickOrder}, "passing"))
	}
	return processServerResp(sender.Send(ctx, []lugo.PlayerOrder{moveForwardOrder}, fmt.Sprintf("just keep swimming: %v", b.side.String())))
}

func ShouldShoot(ballPosition, goalKeeperPosition *lugo.Point, goal field.Goal, opponentTeam []*lugo.Player) (FuzzyScale, *lugo.Point) {
	distanceToShoot := DistanceForShooting(ballPosition, goal)
	if distanceToShoot >= DistanceFar {
		return MustNot, nil
	}

	betterTargetShoot := FindBestPointShootTheBall(goalKeeperPosition, goal)

	// @todo needs enhancement: if an opponent player stays in our way inside the goal zone, the player won't kick neither advance
	obstaclesToTarget, err := findOpponentsOnMyRoute(ballPosition, &betterTargetShoot, field.PlayerSize, opponentTeam)
	if err != nil {
		return MustNot, &goal.Center
	}

	if len(obstaclesToTarget) == 0 {
		if distanceToShoot <= DistanceBeside {
			return Must, &betterTargetShoot
		}
		if distanceToShoot <= DistanceNear {
			return Should, &betterTargetShoot
		}
		if distanceToShoot <= DistanceFar {
			return May, &betterTargetShoot
		}
		return ShouldNot, &betterTargetShoot
	}
	return ShouldNot, &betterTargetShoot
}

// DistanceForShooting gets the distance from the ball to the goal center or one of the goal's poles. Whatever is closer
func DistanceForShooting(ballPosition *lugo.Point, goal field.Goal) float64 {
	ref := goal.Center
	if ballPosition.Y < field.GoalMinY {
		ref = goal.BottomPole
	} else if ballPosition.Y > field.GoalMaxY {
		ref = goal.TopPole
	}
	return ballPosition.DistanceTo(ref)
}

// FindBestPointShootTheBall find a good target in the opponent goal to shoot the ball at.
// @todo needs enhancement: the method is only choosing a side of the goal, but could consider the player position
func FindBestPointShootTheBall(opponentGoalKeeperPosition *lugo.Point, opponentGoal field.Goal) (target lugo.Point) {
	if opponentGoalKeeperPosition.Y > opponentGoal.Center.Y {
		return lugo.Point{
			X: opponentGoal.BottomPole.X,
			Y: opponentGoal.BottomPole.Y + (field.BallSize / 2),
		}
	} else {
		return lugo.Point{
			X: opponentGoal.TopPole.X,
			Y: opponentGoal.TopPole.Y - (field.BallSize / 2),
		}
	}
}

func shouldIPass(me *lugo.Player, ballPosition, goalKeeperPosition *lugo.Point, goal field.Goal, myTeam, opponentTeam []*lugo.Player) (FuzzyScale, *lugo.Player) {
	passingDecision := MustNot
	bestCandidateScore := 0
	var bestCandidate *lugo.Player
	for _, teammate := range myTeam {
		if teammate.Number == me.Number {
			continue
		}
		distanceFromMe := me.Position.DistanceTo(*teammate.Position)
		obstaclesToPlayer, err := findOpponentsOnMyRoute(me.Position, teammate.Position, field.BallSize, opponentTeam)
		if err != nil || len(obstaclesToPlayer) > 0 {
			continue
		}

		teammateScore := 1
		shouldShoot, _ := ShouldShoot(ballPosition, goalKeeperPosition, goal, myTeam)
		switch shouldShoot {
		case May:
			teammateScore = 3
		case Should:
			teammateScore = 5
		case Must:
			teammateScore = 7
		}
		if distanceFromMe <= DistanceBeside {
			teammateScore *= 3
		} else if distanceFromMe <= DistanceNear {
			teammateScore *= 2
		} else if distanceFromMe >= DistanceTooFar {
			teammateScore *= 0
		}

		if teammateScore > bestCandidateScore {
			bestCandidate = teammate
			bestCandidateScore = teammateScore
			if teammateScore <= 7 {
				passingDecision = May
			} else if teammateScore < 14 {
				passingDecision = Should
			} else {
				passingDecision = Must
			}
		}
	}
	return passingDecision, bestCandidate
}

func shouldIHoldTheBall(me *lugo.Player, goal field.Goal, opponentTeam []*lugo.Player) FuzzyScale {
	obstaclesToPlayer, err := findOpponentsOnMyRoute(me.Position, &goal.Center, field.PlayerSize*2, opponentTeam)
	if err != nil {
		return ShouldNot
	}
	if len(obstaclesToPlayer) == 0 {
		return Must
	}
	closerObstacleDist := obstaclesToPlayer[0].position.DistanceTo(*me.Position)
	if closerObstacleDist > DistanceFar {
		return Should
	}

	if closerObstacleDist > DistanceNear {
		return May
	}

	return MustNot
}
