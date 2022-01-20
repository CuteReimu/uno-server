package main

import (
	"net"
	"time"
)

type Game struct {
	Dir              bool
	Players          []IPlayer
	TotalPlayerCount uint32
	Deck             *Deck
	LastCard         Card
	WantColor        uint32
	WhoseTurn        uint32
}

func main() {
	listen, err := net.Listen("tcp", config.GetString("listen_address"))
	if err != nil {
		logger.Fatalln(err)
	}

	game := &Game{
		Dir:              true,
		TotalPlayerCount: config.GetUint32("player.total_count"),
		Deck:             NewDeck(),
	}

	for i := 0; i < config.GetInt("player.robot_count"); i++ {
		game.Players = append(game.Players, new(RobotPlayer))
	}

	logger.Infof("已加入%d个机器人，等待%d人加入。。。", game.TotalPlayerCount, int(game.TotalPlayerCount)-len(game.Players))

	for i := 0; i < int(game.TotalPlayerCount)-len(game.Players); i++ {
		conn, err := listen.Accept()
		if err != nil {
			panic(err)
		}
		player := &HumanPlayer{Connection: conn}
		go player.Recv()
		game.Players = append(game.Players, player)
	}

	for _, player := range game.Players {
		cards := game.Deck.Draw(7)
		player.Init(game, cards)
		for i := uint32(1); i < game.TotalPlayerCount; i++ {
			player.NotifyOtherAddHandCard(i, 7)
		}
	}
	card := game.Deck.Draw(1)
	game.NotifyDeckNum()
	game.DiscardCard(99999, card[0], 0)
	switch card[0].Num {
	case 10:
		game.NextPlayer()
	case 11:
		game.Dir = !game.Dir
	case 12:
		game.AddHandCard(2)
		game.NextPlayer()
	case 14:
		game.AddHandCard(4)
		game.NextPlayer()
	}
	for {
		game.NotifyTurn()
		playerCard, wantColor := game.Players[game.WhoseTurn].AskForDiscardCard()
		if playerCard == nil {
			logger.Infof("%d号玩家不出牌", game.WhoseTurn)
			game.AddHandCard(1)
		} else {
			game.DiscardCard(game.WhoseTurn, *playerCard, wantColor)
			if game.Players[game.WhoseTurn].IsWin() {
				logger.Infof("%d号玩家获胜", game.WhoseTurn)
				break
			}
			switch playerCard.Num {
			case 10:
				game.NextPlayer()
			case 11:
				game.Dir = !game.Dir
			case 12:
				game.NextPlayer()
				game.AddHandCard(2)
			case 14:
				game.NextPlayer()
				game.AddHandCard(4)
			}
		}
		game.NextPlayer()
	}
	time.Sleep(10 * time.Second)
}

func (game *Game) AddHandCard(count int) {
	card := game.Deck.Draw(count)
	game.NotifyDeckNum()
	logger.Infof("%d号玩家摸了%d张牌", game.WhoseTurn, count)
	game.Players[game.WhoseTurn].NotifyAddHandCard(card...)
	for id, player := range game.Players {
		if uint32(id) != game.WhoseTurn {
			player.NotifyOtherAddHandCard((game.WhoseTurn+game.TotalPlayerCount-uint32(id))%game.TotalPlayerCount, uint32(len(card)))
		}
	}
}

func (game *Game) NextPlayer() {
	if game.Dir {
		game.WhoseTurn++
	} else {
		game.WhoseTurn += game.TotalPlayerCount - 1
	}
	game.WhoseTurn %= game.TotalPlayerCount
}

func (game *Game) DiscardCard(playerId uint32, card Card, wantColor uint32) {
	if card.Color == 0 && wantColor != 0 {
		logger.Printf("%d号玩家打出%s，并选择%s", playerId, card.String(), GetColorString(wantColor))
	} else {
		logger.Printf("%d号玩家打出%s", playerId, card.String())
	}
	game.LastCard = card
	game.WantColor = wantColor
	game.Deck.Discard(card)
	if playerId >= game.TotalPlayerCount {
		for _, player := range game.Players {
			player.NotifyDiscardCard(playerId, card, wantColor)
		}
	} else {
		for id, player := range game.Players {
			player.NotifyDiscardCard((playerId+game.TotalPlayerCount-uint32(id))%game.TotalPlayerCount, card, wantColor)
		}
	}
}

func (game *Game) NotifyDeckNum() {
	for _, player := range game.Players {
		player.NotifyDeckNum()
	}
}

func (game *Game) NotifyTurn() {
	for playerId, player := range game.Players {
		player.NotifyTurn((game.WhoseTurn+game.TotalPlayerCount-uint32(playerId))%game.TotalPlayerCount, game.Dir)
	}
}
