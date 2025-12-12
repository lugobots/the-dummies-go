package bot

import (
	"context"
	"math"

	"github.com/lugobots/lugo4go/v3"
	"github.com/lugobots/lugo4go/v3/field"
	"github.com/lugobots/lugo4go/v3/proto"
	"github.com/lugobots/lugo4go/v3/specs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func NewBot(FieldMapper field.Mapper, Config lugo4go.Config, Logger *zap.SugaredLogger) *Bot {
	return &Bot{
		FieldMapper: FieldMapper,
		Config:      Config,
		Logger:      Logger,
	}
}

type Bot struct {
	FieldMapper field.Mapper
	Config      lugo4go.Config
	Logger      *zap.SugaredLogger
}

func (b *Bot) OnDisputing(_ context.Context, inspector lugo4go.SnapshotInspector) ([]proto.PlayerOrder, string, error) {
	b.Logger.With("method", "OnDisputing").Info("---->>>>> PRINTED <<<<----")
	me := inspector.GetMe()
	ballPosition := inspector.GetBall().GetPosition()
	b.Logger.With("method", "OnDisputing").Info("SSSS")
	ballRegion, err := b.FieldMapper.GetPointRegion(ballPosition)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find the ball region")
	}

	myRegion, err := b.FieldMapper.GetPointRegion(me.GetPosition())
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find my region")
	}

	// if the ball is max 2 blocks away from me, I will move toward the ball
	if !isNear(myRegion, ballRegion) {
		return b.holdPosition(inspector)
	}

	moveOrder, err := inspector.MakeOrderMoveMaxSpeed(*ballPosition)
	if err != nil {
		return nil, "", errors.Wrap(err, "error creating move order")
	}

	// we can ALWAYS try to catch the ball if we are not holding it
	catchOrder := inspector.MakeOrderCatch()

	return []proto.PlayerOrder{moveOrder, catchOrder}, "trying to catch the ball", nil

}

func (b *Bot) OnDefending(_ context.Context, inspector lugo4go.SnapshotInspector) ([]proto.PlayerOrder, string, error) {
	b.Logger.With("method", "OnDefending").Info("---->>>>> PRINTED <<<<----")
	me := inspector.GetMe()
	ballPosition := inspector.GetBall().GetPosition()

	ballRegion, err := b.FieldMapper.GetPointRegion(ballPosition)
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to find the ball region")
	}

	b.Logger.With("asas")

	myRegion, err := b.FieldMapper.GetPointRegion(me.GetPosition())
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to find my region")
	}

	// if the ball is max 2 blocks away from me, I will move toward the ball
	if !isNear(myRegion, ballRegion) {
		return b.holdPosition(inspector)
	}

	moveOrder, err := inspector.MakeOrderMoveMaxSpeed(*ballPosition)
	if err != nil {
		return nil, "", errors.Wrapf(err, "error creating move order")
	}

	// we can ALWAYS try to catch the ball if we are not holding it
	catchOrder := inspector.MakeOrderCatch()

	return []proto.PlayerOrder{moveOrder, catchOrder}, "trying to catch the ball", nil
}

func (b *Bot) OnHolding(_ context.Context, inspector lugo4go.SnapshotInspector) ([]proto.PlayerOrder, string, error) {
	b.Logger.With("method", "OnHolding").Info("---->>>>> PRINTED <<<<----")
	me := inspector.GetMe()

	goal := b.FieldMapper.GetAttackGoal()

	goalRegion, err := b.FieldMapper.GetPointRegion(&goal.Center)
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to find my the goal redion")
	}

	myRegion, err := b.FieldMapper.GetPointRegion(me.GetPosition())
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find my region")
	}

	_ = goalRegion
	_ = myRegion
	// if we are near to the goal, let's kick it!
	if !isNear(myRegion, goalRegion) {
		return []proto.PlayerOrder{}, "trying to catch the ball", nil
		// return []proto.PlayerOrder{inspector.MakeOrderMoveByDirection(field.Forward, specs.PlayerMaxSpeed)}, "trying to catch the ball", nil
	}

	kickOrder, err := inspector.MakeOrderKickMaxSpeed(goal.Center)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create kick order")
	}

	_ = kickOrder
	return []proto.PlayerOrder{}, "trying to catch the ball", nil
}

func (b *Bot) OnSupporting(_ context.Context, inspector lugo4go.SnapshotInspector) ([]proto.PlayerOrder, string, error) {
	b.Logger.With("method", "OnSupporting").Info("---->>>>> PRINTED <<<<----")
	me := inspector.GetMe()

	teammatePosition := inspector.GetBall().GetHolder().GetPosition()

	teammateRegion, err := b.FieldMapper.GetPointRegion(teammatePosition)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find the teammate region")
	}

	myRegion, err := b.FieldMapper.GetPointRegion(me.GetPosition())
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find my region")
	}

	// if the player mate is max 2 blocks away from me, I will help
	if !isNear(myRegion, teammateRegion) {
		return b.holdPosition(inspector)
	}

	if teammatePosition.DistanceTo(*me.Position) < specs.PlayerSize {
		// make order to stop
		return []proto.PlayerOrder{inspector.MakeOrderMoveToStop()}, "supporting teammate", nil
	}

	moveOrder, err := inspector.MakeOrderMoveMaxSpeed(*teammatePosition)
	if err != nil {
		return nil, "", errors.Wrap(err, "error creating move order to teammate position")
	}
	return []proto.PlayerOrder{moveOrder}, "supporting teammate", nil
}

func (b *Bot) AsGoalkeeper(_ context.Context, inspector lugo4go.SnapshotInspector, myState lugo4go.PlayerState) ([]proto.PlayerOrder, string, error) {
	b.Logger.With("method", "AsGoalkeeper").Info("---->>>>> PRINTED <<<<----")
	if myState == lugo4go.HoldingTheBall {
		kickOrder, err := inspector.MakeOrderKickMaxSpeed(field.FieldCenter())
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to create kick order")
		}

		return []proto.PlayerOrder{kickOrder}, "kick the ball", nil
	}

	ballPosition := inspector.GetBall().GetPosition()

	moveOrder, err := inspector.MakeOrderMoveMaxSpeed(*ballPosition)
	if err != nil {
		return nil, "", errors.Wrap(err, "error creating move order")
	}

	return []proto.PlayerOrder{moveOrder, inspector.MakeOrderCatch()}, "trying to catch the ball", nil
}

func (b *Bot) OnGetReady(_ context.Context, _ lugo4go.SnapshotInspector) {
	b.Logger.Debug("game ready to start OR score has been changed")
}

func (b *Bot) holdPosition(inspector lugo4go.SnapshotInspector) ([]proto.PlayerOrder, string, error) {
	expectedRegion := GetPlayerTacticRegion(inspector, b.FieldMapper, b.Config.Number)

	if inspector.GetMe().Position.DistanceTo(*expectedRegion.Center()) < specs.PlayerSize {
		return []proto.PlayerOrder{inspector.MakeOrderMoveToStop()}, "Holding position", nil
	}

	moveOrder, err := inspector.MakeOrderMoveMaxSpeed(*expectedRegion.Center())
	if err != nil {
		return nil, "", errors.Wrap(err, "error creating move order")
	}
	return []proto.PlayerOrder{moveOrder}, "moving to my region", nil
}

func isNear(a, b field.Region) bool {
	const minDist = 2
	colDist := a.Col() - b.Col()
	rowDist := a.Row() - b.Row()
	return math.Hypot(float64(colDist), float64(rowDist)) <= minDist
}
