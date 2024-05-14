//go:generate stringer -type=Suit,Rank

package deck

import (
	"fmt"
	"math/rand"
	"time"
)

type Suit uint8
type Rank uint8

const (
	Spade Suit = iota
	Diamond
	Club
	Heart
	Joker
)

const (
	_ Rank = iota
	Ace
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

type Card struct {
	Rank `json:"rank"`
	Suit `json:"suit"`
}

func (c Card) String() string {
	if c.Suit == Joker {
		return c.Suit.String()
	}
	return fmt.Sprintf("%s of %ss", c.Rank.String(), c.Suit.String())

}

var suits = [...]Suit{Spade, Diamond, Club, Heart}
var ranks = [...]Rank{Ace, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King}

func New(opts ...func([]Card) []Card) []Card {
	var cards []Card
	for _, rank := range ranks {
		for _, suite := range suits {
			cards = append(cards, Card{rank, suite})
		}
	}
	for _, opt := range opts {
		cards = opt(cards)
	}

	return cards
}

func Shuffle(cards []Card) []Card {
	var seed = rand.New(rand.NewSource(time.Now().Unix()))
	indexes := seed.Perm(len(cards))
	shuffled_cards := make([]Card, len(cards))
	for i, rand := range indexes {
		shuffled_cards[i] = cards[rand]
	}
	return shuffled_cards
}

func Deck(n int) func(cards []Card) []Card {
	return func(cards []Card) []Card {
		var deck []Card
		for i := 0; i < n; i++ {
			deck = append(deck, cards...)
		}
		return deck
	}
}

func PrintDeck(cards []Card) {
	for _, card := range cards {
		fmt.Println(card)
	}
}

func (c Card) GetValue() int {
	if c.Rank >= Two && c.Rank <= Ten {
		return int(c.Rank)
	} else if c.Rank >= Jack && c.Rank <= King {
		return 10

	} else if c.Rank == Ace {
		return 11
	}
	return 0
}
