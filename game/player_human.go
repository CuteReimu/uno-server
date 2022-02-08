package game

import (
	"github.com/CuteReimu/uno-server/protos"
	"github.com/davyxu/cellnet"
)

type HumanPlayer struct {
	basePlayer
	cellnet.Session
}

func (r *HumanPlayer) Init(game *Game, location int) {
	r.basePlayer.Init(game, location)
	msg := &protos.InitToc{
		PlayerNum: uint32(r.game.TotalPlayerCount),
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyAddHandCard(cards ...ICard) {
	msg := &protos.DrawCardToc{}
	for _, card := range cards {
		msg.Card = append(msg.Card, &protos.UnoCard{
			CardId: card.Id(),
			Color:  uint32(card.Color()),
			Num:    card.Number(),
		})
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyOtherAddHandCard(location int, count int) {
	msg := &protos.OtherAddHandCardToc{
		PlayerId: r.getAlternativeLocation(location),
		Num:      uint32(count),
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyDeckNum(count int) {
	msg := &protos.SetDeckNumToc{
		Num: uint32(count),
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyDiscardCard(location int, card ICard, args ...uint32) {
	r.basePlayer.NotifyDiscardCard(location, card, args...)
	msg := &protos.DiscardCardToc{
		PlayerId: r.getAlternativeLocation(location),
		Card: &protos.UnoCard{
			CardId: card.Id(),
			Color:  uint32(card.Color()),
			Num:    card.Number(),
		},
	}
	if len(args) > 0 {
		msg.WantColor = args[0]
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyTurn(location int, dir bool) {
	msg := &protos.NotifyTurnToc{
		PlayerId: r.getAlternativeLocation(location),
		Dir:      dir,
	}
	r.Send(msg)
}

func (r *HumanPlayer) IsWin() bool {
	return len(r.cards) == 0
}

func (r *HumanPlayer) NotifyWin(location int) {
	r.Send(&protos.NotifyWinToc{
		PlayerId: r.getAlternativeLocation(location),
	})
}

func (r *HumanPlayer) getAlternativeLocation(location int) uint32 {
	if location == 99999 {
		return 99999
	}
	location -= r.Location()
	if location < 0 {
		location += r.game.TotalPlayerCount
	}
	return uint32(location % r.game.TotalPlayerCount)
}
