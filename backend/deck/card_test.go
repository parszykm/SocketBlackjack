package deck

import (
	"reflect"
	"testing"
)

func TestCard(t *testing.T) {

	testCases := []struct {
		card     Card
		expected string
	}{
		{Card{Ace, Spade}, "Ace of Spades"},
		{Card{Two, Heart}, "Two of Hearts"},
		{Card{Ten, Diamond}, "Ten of Diamonds"},
		{Card{King, Club}, "King of Clubs"},
	}
	for _, tc := range testCases {
		if tc.card.String() != tc.expected {
			t.Errorf("%s doesnt match %s", tc.card.String(), tc.expected)
		}
	}

	jokerCard := Card{Suit: Joker}
	jokerExpected := "Joker"
	if jokerCard.String() != jokerExpected {
		t.Errorf("Expected %s, but got %s", jokerExpected, jokerCard.String())
	}

}

func TestDeck(t *testing.T) {
	deck := New(Deck(3))
	if len(deck) != 3*52 {
		t.Errorf("Deck is not length of %d", 3*52)
	}
}

func TestDeckShuffle(t *testing.T) {
	deck := New(Deck(3), Shuffle)

	first := deck[:52]
	second := deck[52:104]
	third := deck[104:156]

	if reflect.DeepEqual(first, second) || reflect.DeepEqual(first, third) || reflect.DeepEqual(second, third) {
		t.Error("Deck is shuffled periodicaly")
	}
}

func TestValued(t *testing.T) {
	card := Card{Ace, Diamond}
	if card.GetValue() != 11 {
		t.Error("Wrong value")
	}
	card = Card{Two, Diamond}
	if card.GetValue() != 2 {
		t.Error("Wrong value")
	}
	card = Card{Queen, Diamond}
	if card.GetValue() != 10 {
		t.Error("Wrong value")
	}
}
