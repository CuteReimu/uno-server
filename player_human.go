package main

import (
	"bytes"
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"uno/protos"
)

type HumanPlayer struct {
	lastCard   Card
	game       *Game
	cards      map[uint32]Card
	Connection net.Conn
	isTurn     bool
	cardChan   chan *protos.DiscardCardTos
}

func (r *HumanPlayer) send(protoName string, message proto.Message) {
	var byteBuffer bytes.Buffer
	buf, err := proto.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}
	buff := make([]byte, 2)
	binary.BigEndian.PutUint16(buff, uint16(len(protoName)+len(buf)+2))
	_, err = byteBuffer.Write(buff)
	if err != nil {
		log.Fatalln(err)
	}
	binary.BigEndian.PutUint16(buff, uint16(len(protoName)))
	_, err = byteBuffer.Write(buff)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = byteBuffer.Write([]byte(protoName))
	if err != nil {
		log.Fatalln(err)
	}
	_, err = byteBuffer.Write(buf)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = r.Connection.Write(byteBuffer.Bytes())
	if err != nil {
		log.Fatalln(err)
	}
}

func (r *HumanPlayer) Recv() {
	r.cardChan = make(chan *protos.DiscardCardTos)
	for {
		buff := make([]byte, 2)
		_, err := r.Connection.Read(buff)
		if err != nil {
			panic(err)
		}
		totalLen := binary.BigEndian.Uint16(buff)
		buf := make([]byte, totalLen)
		n, err := r.Connection.Read(buf)
		if err != nil || n != int(totalLen) {
			panic(err)
		}
		nameLen := binary.BigEndian.Uint16(buf[:2])
		protoName := string(buf[2 : 2+nameLen])
		if protoName != "discard_card_tos" {
			log.Println("错误的协议", protoName)
		} else {
			msg := &protos.DiscardCardTos{}
			err = proto.Unmarshal(buf[2+nameLen:], msg)
			if err != nil {
				panic(err)
			}
			if r.isTurn {
				r.isTurn = false
				r.cardChan <- msg
			} else {
				log.Println("还没到你出牌的回合")
			}
		}
	}
}

func (r *HumanPlayer) Init(game *Game, beginHandCard []Card) {
	r.game = game
	r.cards = make(map[uint32]Card)
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
	r.send("init_toc", msg)
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
	r.send("draw_card_toc", msg)
}

func (r *HumanPlayer) NotifyOtherAddHandCard(playerId uint32, num uint32) {
	msg := &protos.OtherAddHandCardToc{
		PlayerId: playerId,
		Num:      num,
	}
	r.send("other_add_hand_card_toc", msg)
}

func (r *HumanPlayer) NotifyDeckNum() {
	length := uint32(len(r.game.Deck.cards))
	msg := &protos.SetDeckNumToc{
		Num: length,
	}
	r.send("set_deck_num_toc", msg)
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
	r.send("discard_card_toc", msg)
}

func (r *HumanPlayer) NotifyTurn(playerId uint32, dir bool) {
	msg := &protos.NotifyTurnToc{
		PlayerId: playerId,
		Dir:      dir,
	}
	r.send("notify_turn_toc", msg)
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
			log.Println("你没有这张牌")
			continue
		}
		if !card.CanPlay(r.game.LastCard, r.game.WantColor) {
			log.Println("你不能出这张牌")
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
				log.Println("你还有其他牌可以出，不能出+4")
				continue
			}
		}
		if card.Color == 0 && (msg.GetWantColor() > 4 || msg.GetWantColor() < 1) {
			log.Println("你选的颜色不对")
			continue
		}
		delete(r.cards, msg.GetCardId())
		return &card, msg.GetWantColor()
	}
}

func (r *HumanPlayer) IsWin() bool {
	return len(r.cards) == 0
}
