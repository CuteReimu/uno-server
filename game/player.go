package game

import (
	"fmt"
	"time"
)

type IPlayer interface {
	Init(game *Game, location int)
	Location() int
	NotifyAddHandCard(card ...ICard)
	NotifyOtherAddHandCard(location int, count int)
	NotifyDeckNum(count int)
	NotifyDiscardCard(location int, card ICard, args ...uint32)
	NotifyTurn(location int, dir bool)
	PlayCard(cardId uint32, args ...uint32)
	IsWin() bool
	GetNextPlayer(location int) IPlayer
	NotifyWin(location int)
	Draw(count int)
	ForeachCards(func(card ICard) bool)
}

type basePlayer struct {
	game     *Game
	location int
	cards    map[uint32]ICard
}

func (p *basePlayer) Init(game *Game, location int) {
	p.game = game
	p.location = location
	p.cards = make(map[uint32]ICard)
}

func (p *basePlayer) ForeachCards(f func(card ICard) bool) {
	for _, card := range p.cards {
		if !f(card) {
			break
		}
	}
}

func (p *basePlayer) Location() int {
	return p.location
}

func (p *basePlayer) NotifyAddHandCard(_ ...ICard) {
}

func (p *basePlayer) NotifyOtherAddHandCard(int, int) {
}

func (p *basePlayer) NotifyDeckNum(int) {
}

func (p *basePlayer) NotifyDiscardCard(location int, card ICard, _ ...uint32) {
	if location == p.location {
		delete(p.cards, card.Id())
		p.game.Deck.Discard(card)
	}
}

func (p *basePlayer) NotifyTurn(int, bool) {
	panic("implement me")
}

func (p *basePlayer) PlayCard(cardId uint32, args ...uint32) {
	if p.game.WhoseTurn != p.location {
		logger.Error("还没到你的回合，不能出牌")
		return
	}
	if cardId == 0 {
		p.Draw(1)
		p.game.NextPlayer(1)
		return
	}
	card := p.cards[cardId]
	if card == nil {
		logger.Error("你没有这张牌")
		return
	}
	if card.CanPlay(p.game, p, args...) {
		for _, player := range p.game.Players {
			player.NotifyDiscardCard(p.location, card, args...)
		}
		if card.Color() == ColorBlack && p.game.WantColor != ColorBlack {
			logger.Info(fmt.Sprintf("%d号玩家打出%s，并选择%s", p.location, card, p.game.WantColor))
		} else {
			logger.Info(fmt.Sprintf("%d号玩家打出%s", p.location, card))
		}
		if p.IsWin() {
			logger.Info(fmt.Sprintf("%d号玩家获胜", p.location))
			for _, player := range p.game.Players {
				player.NotifyWin(p.location)
			}
			logger.Info("游戏将在10秒后重新开始。。。")
			time.AfterFunc(time.Second*10, func() {
				p.game.Post(func() {
					if p.IsWin() {
						p.game.start()
					}
				})
			})
			return
		}
		card.Execute(p.game, p, args...)
	} else {
		logger.Error("你不能打这张牌", card)
	}
}

func (p *basePlayer) IsWin() bool {
	return len(p.cards) == 0
}

func (p *basePlayer) GetNextPlayer(location int) IPlayer {
	if !p.game.Dir {
		location = -location
	}
	location += p.location
	if location < 0 {
		location += p.game.TotalPlayerCount
	}
	location %= p.game.TotalPlayerCount
	return p.game.Players[location]
}

func (p *basePlayer) NotifyWin(int) {
}

func (p *basePlayer) Draw(count int) {
	cards := p.game.Deck.Draw(count)
	for _, card := range cards {
		p.cards[card.Id()] = card
	}
	logger.Info(fmt.Sprintf("%d号玩家摸了%d张牌, 现在还有%d张牌", p.location, count, len(p.cards)))
	for _, player := range p.game.Players {
		if player.Location() == p.Location() {
			player.NotifyDeckNum(len(p.game.Deck.cards))
			player.NotifyAddHandCard(cards...)
		} else {
			player.NotifyDeckNum(len(p.game.Deck.cards))
			player.NotifyOtherAddHandCard(p.location, len(cards))
		}
	}
}

type RobotPlayer struct {
	basePlayer
}

func (r *RobotPlayer) NotifyTurn(location int, _ bool) {
	time.AfterFunc(time.Second/2, func() {
		r.game.Post(func() {
			if location != r.location {
				return
			}
			cardId, wantColor := func() (uint32, uint32) {
				for _, card := range r.cards {
					if card.Color() != ColorBlack && card.Number() >= 10 {
						if card.CanPlay(r.game, r) {
							return card.Id(), 0
						}
					}
				}
				for _, card := range r.cards {
					if card.Color() != ColorBlack && card.Number() < 10 {
						if card.CanPlay(r.game, r) {
							return card.Id(), 0
						}
					}
				}
				for _, card := range r.cards {
					if card.Color() == 0 && card.Number() == 13 {
						return card.Id(), r.getMaxNumColor()
					}
				}
				for _, card := range r.cards {
					if card.Color() == 0 && card.Number() == 14 {
						return card.Id(), r.getMaxNumColor()
					}
				}
				return 0, 0
			}()
			r.PlayCard(cardId, wantColor)
		})
	})
}

func (r *RobotPlayer) getMaxNumColor() uint32 {
	nums := make([]uint32, 5)
	for _, card := range r.cards {
		nums[card.Color()]++
	}
	maxI := 1
	for i := 2; i <= 4; i++ {
		if nums[i] > nums[maxI] {
			maxI = i
		}
	}
	return uint32(maxI)
}
