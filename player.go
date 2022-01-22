package main

import "time"

type IPlayer interface {
	Init(game *Game, beginHandCard []Card)
	NotifyAddHandCard(card ...Card)
	NotifyOtherAddHandCard(playerId uint32, num uint32)
	NotifyDeckNum()
	NotifyDiscardCard(playerId uint32, card Card, wantColor uint32)
	NotifyTurn(playerId uint32, dir bool)
	AskForDiscardCard() (*Card, uint32)
	IsWin() bool
	NotifyWin(playerId uint32)
}

type RobotPlayer struct {
	game  *Game
	cards map[uint32]Card
}

func (r *RobotPlayer) Init(game *Game, beginHandCard []Card) {
	r.game = game
	r.cards = make(map[uint32]Card)
	for _, card := range beginHandCard {
		r.cards[card.Id] = card
	}
}

func (r *RobotPlayer) NotifyAddHandCard(card ...Card) {
	for _, card := range card {
		r.cards[card.Id] = card
	}
}

func (r *RobotPlayer) NotifyOtherAddHandCard(uint32, uint32) {
}

func (r *RobotPlayer) NotifyDeckNum() {
}

func (r *RobotPlayer) NotifyDiscardCard(playerId uint32, card Card, _ uint32) {
	if playerId == 0 {
		delete(r.cards, card.Id)
	}
}

func (r *RobotPlayer) NotifyTurn(uint32, bool) {
}

func (r *RobotPlayer) AskForDiscardCard() (*Card, uint32) {
	time.Sleep(time.Second / 2)
	for _, card := range r.cards {
		if card.Color != 0 && card.Num >= 10 {
			if card.CanPlay(r.game.LastCard, r.game.WantColor) {
				return &card, 0
			}
		}
	}
	for _, card := range r.cards {
		if card.Color != 0 && card.Num < 10 {
			if card.CanPlay(r.game.LastCard, r.game.WantColor) {
				return &card, 0
			}
		}
	}
	for _, card := range r.cards {
		if card.Color == 0 && card.Num == 13 {
			return &card, r.getMaxNumColor()
		}
	}
	for _, card := range r.cards {
		if card.Color == 0 && card.Num == 14 {
			return &card, r.getMaxNumColor()
		}
	}
	return nil, 0
}

func (r *RobotPlayer) IsWin() bool {
	return len(r.cards) == 0
}

func (r *RobotPlayer) NotifyWin(uint32) {
}

func (r *RobotPlayer) getMaxNumColor() uint32 {
	nums := make([]uint32, 5)
	for _, card := range r.cards {
		nums[card.Color]++
	}
	maxI := 1
	for i := 2; i <= 4; i++ {
		if nums[i] > nums[maxI] {
			maxI = i
		}
	}
	return uint32(maxI)
}
