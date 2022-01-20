package main

import (
	"math/rand"
	"strconv"
	"time"
)

func GetColorString(color uint32) string {
	switch color {
	case 1:
		return "红色"
	case 2:
		return "绿色"
	case 3:
		return "黄色"
	case 4:
		return "蓝色"
	case 0:
		return "黑色"
	default:
		panic(color)
	}
}

type Card struct {
	Id    uint32
	Color uint32
	Num   uint32
}

func (card *Card) CanPlay(lastCard Card, wantColor uint32) bool {
	if lastCard.Color != 0 {
		if card.Color == lastCard.Color || card.Num == lastCard.Num || card.Color == 0 {
			return true
		}
	} else {
		if card.Color == wantColor || card.Color == 0 || wantColor == 0 {
			return true
		}
	}
	return false
}

func (card *Card) String() string {
	color := GetColorString(card.Color)
	var num string
	switch card.Num {
	case 10:
		num = "跳过"
	case 11:
		num = "转向"
	case 12:
		num = "+2"
	case 13:
		num = "变色"
	case 14:
		num = "+4"
	default:
		num = strconv.Itoa(int(card.Num))
	}
	return color + num
}

type Deck struct {
	cards       []Card
	pos         int
	discardPile []Card
	random      rand.Source
}

func NewDeck() *Deck {
	d := new(Deck)
	d.random = rand.NewSource(time.Now().Unix())
	id := uint32(1)
	for i := uint32(1); i < 4; i++ {
		d.cards = append(d.cards, Card{id, i, 0})
		id++
		for j := uint32(1); j <= 12; j++ {
			d.cards = append(d.cards, Card{id, i, j})
			id++
			d.cards = append(d.cards, Card{id, i, j})
			id++
		}
	}
	for i := 1; i < 4; i++ {
		d.cards = append(d.cards, Card{id, 0, 13})
		id++
		d.cards = append(d.cards, Card{id, 0, 14})
		id++
	}
	d.pos = len(d.cards)
	d.Shuffle()
	return d
}

func (d *Deck) Shuffle() {
	l := d.pos
	for i := 0; i < l; i++ {
		j := int(d.random.Int63())%(l-i) + i
		if i != j {
			d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
		}
	}
}

func (d *Deck) Draw(n int) []Card {
	if n > d.pos {
		var newDeck []Card
		for i := 0; i < d.pos; i++ {
			newDeck = append(newDeck, d.cards[i])
		}
		newDeck = append(newDeck, d.discardPile...)
		d.cards = newDeck
		d.discardPile = make([]Card, 0)
		d.pos = len(d.cards)
		d.Shuffle()
	}
	result := make([]Card, 0, n)
	for i := 0; i < n; i++ {
		d.pos--
		result = append(result, d.cards[d.pos])
	}
	return result
}

func (d *Deck) Discard(cards ...Card) {
	d.discardPile = append(d.discardPile, cards...)
}
