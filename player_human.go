package main

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"os"
	"time"
	"uno/protos"
	_ "uno/tcp"
)

type HumanPlayer struct {
	game  *Game
	cards map[uint32]Card
	cellnet.Session
	isTurn   bool
	cardChan chan *protos.DiscardCardTos
}

func (r *HumanPlayer) Init(game *Game, beginHandCard []Card) {
	r.game = game
	r.cards = make(map[uint32]Card)
	r.cardChan = make(chan *protos.DiscardCardTos)
	msg := &protos.InitToc{
		PlayerNum: game.TotalPlayerCount,
	}
	for _, card := range beginHandCard {
		r.cards[card.Id] = card
		id, color, num := card.Id, card.Color, card.Num
		msg.Cards = append(msg.Cards, &protos.UnoCard{
			CardId: id,
			Color:  color,
			Num:    num,
		})
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyAddHandCard(cards ...Card) {
	msg := &protos.DrawCardToc{}
	for _, card := range cards {
		r.cards[card.Id] = card
		id, color, num := card.Id, card.Color, card.Num
		msg.Card = append(msg.Card, &protos.UnoCard{
			CardId: id,
			Color:  color,
			Num:    num,
		})
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyOtherAddHandCard(playerId uint32, num uint32) {
	msg := &protos.OtherAddHandCardToc{
		PlayerId: playerId,
		Num:      num,
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyDeckNum() {
	length := uint32(len(r.game.Deck.cards))
	msg := &protos.SetDeckNumToc{
		Num: length,
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyDiscardCard(playerId uint32, card Card, wantColor uint32) {
	if playerId == 0 {
		delete(r.cards, card.Id)
	}
	msg := &protos.DiscardCardToc{
		PlayerId: playerId,
		Card: &protos.UnoCard{
			CardId: card.Id,
			Color:  card.Color,
			Num:    card.Num,
		},
		WantColor: wantColor,
	}
	r.Send(msg)
}

func (r *HumanPlayer) NotifyTurn(playerId uint32, dir bool) {
	msg := &protos.NotifyTurnToc{
		PlayerId: playerId,
		Dir:      dir,
	}
	r.Send(msg)
}

func (r *HumanPlayer) AskForDiscardCard() (*Card, uint32) {
	for {
		r.isTurn = true
		msg := <-r.cardChan
		if msg.GetCardId() == 0 {
			return nil, 0
		}
		card, ok := r.cards[msg.GetCardId()]
		if !ok {
			logger.Info("你没有这张牌")
			continue
		}
		if !card.CanPlay(r.game.LastCard, r.game.WantColor) {
			logger.Info("你不能出这张牌")
			continue
		}
		if card.Color == 0 && card.Num == 14 {
			canPlay := true
			for _, c := range r.cards {
				if c.Id != card.Id && c.Num != 14 && c.CanPlay(r.game.LastCard, r.game.WantColor) {
					canPlay = false
					break
				}
			}
			if !canPlay {
				logger.Info("你还有其他牌可以出，不能出+4")
				continue
			}
		}
		if card.Color == 0 && (msg.GetWantColor() > 4 || msg.GetWantColor() < 1) {
			logger.Info("你选的颜色不对")
			continue
		}
		delete(r.cards, msg.GetCardId())
		return &card, msg.GetWantColor()
	}
}

func (r *HumanPlayer) IsWin() bool {
	return len(r.cards) == 0
}

func (r *HumanPlayer) NotifyWin(playerId uint32) {
	r.Send(&protos.NotifyWinToc{PlayerId: playerId})
}

func StartListen(humanCount uint32) (players []IPlayer) {
	// 创建一个事件处理队列，整个服务器只有这一个队列处理事件，服务器属于单线程服务器
	queue := cellnet.NewEventQueue()

	// 创建一个tcp的侦听器，名称为server，所有连接将事件投递到queue队列,单线程的处理
	p := peer.NewGenericPeer("tcp.Acceptor", "server", config.GetString("listen_address"), queue)

	humanMap := make(map[int64]*HumanPlayer)
	var index uint32
	ch := make(chan struct{})
	proc.BindProcessorHandler(p, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			if index < humanCount {
				player := &HumanPlayer{Session: ev.Session()}
				players = append(players, player)
				humanMap[player.Session.ID()] = player
				index++
				logger.Info("server accepted", player.Session)
				if index == humanCount {
					ch <- struct{}{}
				}
			} else {
				ev.Session().Close()
				logger.Info("房间人数已满")
			}
		case *cellnet.SessionClosed:
			logger.Info("session closed: ", ev.Session().ID())
			logger.Info("目前不支持断线重连，程序将在3秒后关闭")
			time.Sleep(time.Second * 3)
			os.Exit(1)
		case *protos.DiscardCardTos:
			r := humanMap[ev.Session().ID()]
			if r.isTurn {
				r.isTurn = false
				r.cardChan <- msg
			} else {
				logger.Info("还没到你出牌的回合")
			}
		}
	})
	p.Start()
	queue.StartLoop()
	//queue.Wait()
	<-ch
	return
}
