package game

import (
	"fmt"
	"github.com/CuteReimu/uno-server/config"
	_ "github.com/CuteReimu/uno-server/core"
	"github.com/CuteReimu/uno-server/protos"
	"github.com/CuteReimu/uno-server/utils"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"os"
	"time"
)

var logger = utils.GetLogger("game")

type Game struct {
	Dir              bool
	Players          []IPlayer
	TotalPlayerCount int
	Deck             *Deck
	LastCard         ICard
	WantColor        Color
	WhoseTurn        int
	cellnet.EventQueue
}

func (game *Game) NextPlayer(location int) {
	if game.Dir {
		location = -location
	}
	game.WhoseTurn += location
	if game.WhoseTurn < 0 {
		game.WhoseTurn += game.TotalPlayerCount
	}
	game.WhoseTurn %= game.TotalPlayerCount
	for _, player := range game.Players {
		player.NotifyTurn(game.WhoseTurn, game.Dir)
	}
}

func (game *Game) Start(totalCount, robotCount int) {
	humanCount := totalCount - robotCount
	game.TotalPlayerCount = totalCount
	index := 0
	for ; index < robotCount; index++ {
		game.Players = append(game.Players, new(RobotPlayer))
	}
	logger.Info(fmt.Sprintf("已加入%d个机器人，等待%d人加入。。。", robotCount, humanCount))

	if !config.GlobalConfig.GetBool("log.tcp_debug_log") {
		msglog.SetCurrMsgLogMode(msglog.MsgLogMode_Mute)
	}
	// 创建一个事件处理队列，整个服务器只有这一个队列处理事件，服务器属于单线程服务器
	game.EventQueue = cellnet.NewEventQueue()

	// 创建一个tcp的侦听器，名称为server，所有连接将事件投递到queue队列,单线程的处理
	p := peer.NewGenericPeer("tcp.Acceptor", "server", config.GlobalConfig.GetString("listen_address"), game.EventQueue)

	humanMap := make(map[int64]*HumanPlayer)
	proc.BindProcessorHandler(p, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			if index < totalCount {
				player := &HumanPlayer{Session: ev.Session()}
				game.Players = append(game.Players, player)
				humanMap[player.Session.ID()] = player
				index++
				logger.Info("server accepted", player.Session)
				if index == totalCount {
					game.Post(game.start)
				}
			} else {
				ev.Session().Close()
				logger.Info("房间人数已满")
			}
		case *cellnet.SessionClosed:
			logger.Info("session closed: ", ev.Session().ID())
			if _, ok := humanMap[ev.Session().ID()]; ok {
				logger.Info("目前不支持断线重连，程序将在3秒后关闭")
				time.Sleep(time.Second * 3)
				os.Exit(1)
			}
		case *protos.DiscardCardTos:
			r := humanMap[ev.Session().ID()]
			r.PlayCard(msg.CardId, msg.WantColor)
		case *protos.RestartGameTos:
			if index == totalCount {
				game.Post(game.start)
			}
		}
	})
	p.Start()
	game.StartLoop()
	if humanCount == 0 {
		game.Post(game.start)
	}
	game.Wait()
}

func (game *Game) start() {
	game.Deck = NewDeck()
	game.Dir = true
	for location, player := range game.Players {
		player.Init(game, location)
	}
	for _, player := range game.Players {
		player.Draw(7)
	}
	cards := game.Deck.Draw(1)
	logger.Info("翻出了", cards[0])
	for _, player := range game.Players {
		player.NotifyDeckNum(len(game.Deck.cards))
		player.NotifyDiscardCard(99999, cards[0], 0)
	}
	game.WhoseTurn = len(game.Players) - 1
	game.Deck.Discard(cards...)
	cards[0].Execute(game, game.Players[game.WhoseTurn])
}
