package game

import (
	"math/rand"
	"strconv"
	"time"
)

const (
	ColorBlack Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
)

type Color uint8

func (c Color) String() string {
	switch c {
	case ColorBlack:
		return "黑色"
	case ColorRed:
		return "红色"
	case ColorGreen:
		return "绿色"
	case ColorYellow:
		return "黄色"
	case ColorBlue:
		return "蓝色"
	}
	panic("unreachable code")
}

type ICard interface {
	Id() uint32
	CanPlay(game *Game, player IPlayer, args ...uint32) bool
	Execute(game *Game, player IPlayer, args ...uint32)
	String() string
	Color() Color
	Number() uint32
}

type baseCard struct {
	id uint32
}

func (c *baseCard) Id() uint32 {
	return c.id
}

type colorfulCard struct {
	baseCard
	color Color
}

func (c *colorfulCard) Color() Color {
	return c.color
}

type numberCard struct {
	colorfulCard
	num uint32
}

func newNumberCard(id, color, num uint32) ICard {
	return &numberCard{colorfulCard{baseCard{id}, Color(color)}, num}
}

func (c *numberCard) CanPlay(game *Game, _ IPlayer, _ ...uint32) bool {
	if game.WantColor == ColorBlack {
		return true
	}
	return game.WantColor == c.Color() || game.LastCard.Number() == c.Number()
}

func (c *numberCard) Execute(game *Game, _ IPlayer, _ ...uint32) {
	game.LastCard = c
	game.WantColor = c.Color()
	game.NextPlayer(1)
}

func (c *numberCard) Number() uint32 {
	return c.num
}

func (c *numberCard) String() string {
	return c.Color().String() + strconv.Itoa(int(c.Number()))
}

type cardSkip struct {
	colorfulCard
}

func newSkipCard(id, color uint32) ICard {
	return &cardSkip{colorfulCard{baseCard{id}, Color(color)}}
}

func (c *cardSkip) CanPlay(game *Game, _ IPlayer, _ ...uint32) bool {
	if game.WantColor == ColorBlack {
		return true
	}
	return game.WantColor == c.Color() || game.LastCard.Number() == c.Number()
}

func (c *cardSkip) Execute(game *Game, _ IPlayer, _ ...uint32) {
	game.LastCard = c
	game.WantColor = c.Color()
	game.NextPlayer(2)
}

func (c *cardSkip) String() string {
	return c.Color().String() + "跳过"
}

func (c *cardSkip) Number() uint32 {
	return 10
}

type cardReverse struct {
	colorfulCard
}

func newReverseCard(id, color uint32) ICard {
	return &cardReverse{colorfulCard{baseCard{id}, Color(color)}}
}

func (c *cardReverse) CanPlay(game *Game, _ IPlayer, _ ...uint32) bool {
	if game.WantColor == ColorBlack {
		return true
	}
	return game.WantColor == c.Color() || game.LastCard.Number() == c.Number()
}

func (c *cardReverse) Execute(game *Game, _ IPlayer, _ ...uint32) {
	game.LastCard = c
	game.WantColor = c.Color()
	game.Dir = !game.Dir
	game.NextPlayer(1)
}

func (c *cardReverse) String() string {
	return c.Color().String() + "转向"
}

func (c *cardReverse) Number() uint32 {
	return 11
}

type cardPlus2 struct {
	colorfulCard
}

func newPlus2Card(id, color uint32) ICard {
	return &cardPlus2{colorfulCard{baseCard{id}, Color(color)}}
}

func (c *cardPlus2) CanPlay(game *Game, _ IPlayer, _ ...uint32) bool {
	if game.WantColor == ColorBlack {
		return true
	}
	return game.WantColor == c.Color() || game.LastCard.Number() == c.Number()
}

func (c *cardPlus2) Execute(game *Game, player IPlayer, _ ...uint32) {
	player.GetNextPlayer(1).Draw(2)
	game.LastCard = c
	game.WantColor = c.Color()
	game.NextPlayer(2)
}

func (c *cardPlus2) String() string {
	return c.Color().String() + "+2"
}

func (c *cardPlus2) Number() uint32 {
	return 12
}

type cardWild struct {
	baseCard
}

func newWildCard(id uint32) ICard {
	return &cardWild{baseCard{id}}
}

func (c *cardWild) CanPlay(_ *Game, _ IPlayer, args ...uint32) bool {
	if len(args) == 1 && Color(args[0]) >= ColorRed && Color(args[0]) <= ColorBlue {
		return true
	}
	logger.Error("参数错误")
	return false
}

func (c *cardWild) Execute(game *Game, player IPlayer, args ...uint32) {
	if player != nil && len(args) > 0 {
		game.WantColor = Color(args[0])
	}
	game.LastCard = c
	game.NextPlayer(1)
}

func (c *cardWild) String() string {
	return "黑色变色"
}

func (c *cardWild) Color() Color {
	return ColorBlack
}

func (c *cardWild) Number() uint32 {
	return 12
}

type cardPlus4 struct {
	baseCard
}

func newPlus4Card(id uint32) ICard {
	return &cardPlus4{baseCard{id}}
}

func (c *cardPlus4) CanPlay(game *Game, player IPlayer, args ...uint32) bool {
	if len(args) != 1 || Color(args[0]) < ColorRed || Color(args[0]) > ColorBlue {
		logger.Error("参数错误")
		return false
	}
	canPlay := true
	player.ForeachCards(func(card ICard) bool {
		if _, ok := card.(*cardPlus4); !ok && card.CanPlay(game, player, args...) {
			canPlay = false
			return false
		}
		return true
	})
	return canPlay
}

func (c *cardPlus4) Execute(game *Game, player IPlayer, _ ...uint32) {
	player.GetNextPlayer(1).Draw(4)
	game.LastCard = c
	game.WantColor = c.Color()
	game.NextPlayer(2)
}

func (c *cardPlus4) String() string {
	return "黑色+4"
}

func (c *cardPlus4) Color() Color {
	return ColorBlack
}

func (c *cardPlus4) Number() uint32 {
	return 12
}

type Deck struct {
	cards       []ICard
	discardPile []ICard
	random      *rand.Rand
}

func NewDeck() *Deck {
	d := new(Deck)
	d.random = rand.New(rand.NewSource(time.Now().Unix()))
	id := uint32(1)
	for i := uint32(1); i < 4; i++ {
		d.cards = append(d.cards, newNumberCard(id, i, 0))
		id++
		for j := uint32(1); j < 10; j++ {
			d.cards = append(d.cards, newNumberCard(id, i, j))
			id++
			d.cards = append(d.cards, newNumberCard(id, i, j))
			id++
		}
		d.cards = append(d.cards, newSkipCard(id, i))
		id++
		d.cards = append(d.cards, newReverseCard(id, i))
		id++
		d.cards = append(d.cards, newPlus2Card(id, i))
	}
	for i := 1; i < 4; i++ {
		d.cards = append(d.cards, newWildCard(id))
		id++
		d.cards = append(d.cards, newPlus4Card(id))
		id++
	}
	d.Shuffle()
	return d
}

func (d *Deck) Shuffle() {
	d.cards = append(d.cards, d.discardPile...)
	d.discardPile = nil
	d.random.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

func (d *Deck) Draw(n int) []ICard {
	if n > len(d.cards) {
		d.Shuffle()
	}
	if n > len(d.cards) {
		n = len(d.cards)
	}
	result := d.cards[:n]
	d.cards = d.cards[n:]
	return result
}

func (d *Deck) Discard(cards ...ICard) {
	d.discardPile = append(d.discardPile, cards...)
}
