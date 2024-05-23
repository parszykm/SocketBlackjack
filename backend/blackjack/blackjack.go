package blackjack

import (
	deck "blackjack/deck"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type GameStage uint8

const (
	PreGame GameStage = iota
	GameActive
	PostGame
)
const (
	MessageTypeSendHand          = "SendHand"
	MessageTypeDealerInitHand    = "DealerInitHand"
	MessageTypeDealerFinalHand   = "DealerFinalHand"
	MessageTypeStartGameResponse = "StartGameResponse"
	MessageTypeGameResult        = "GameResult"
	MessageTypeInitialHandshake  = "InitialHandshake"
	MessageTypeStartGame         = "StartGame"
	MessageTypeOtherHands        = "OtherHands"
	MessageTypeYourTurn          = "YourTurn"
	MessageTypeReconnectState    = "ReconnectState"
	MessageTypeGameReady         = "GameReady"
	MessageTypeGameNotReady      = "GameNotReady"
)

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type InitialHandshakeMessage struct {
	Id  int    `json:"id"`
	Msg string `json:"msg"`
}
type HandMessage struct {
	Hand  []deck.Card `json:"hand"`
	Stage bool        `json:"stage"`
	Count int         `json:"count"`
}

type ResultMessage struct {
	Count  int     `json:"count"`
	Refund float64 `json:"refund"`
	Budget float64 `json:"budget"`
}

type OtherHandStruct struct {
	Id        int         `json:"id"`
	Hand      []deck.Card `json:"hand"`
	Timestamp time.Time   `json:"timestamp"`
	Turn      bool        `json:"turn"`
}

type ReconnectStateMessage struct {
	DealerHand []deck.Card       `json:"dealerHand"`
	Hand       []deck.Card       `json:"hand"`
	OtherHands []OtherHandStruct `json:"otherHands"`
	Turn       bool              `json:"turn"`
	GameStage  GameStage         `json:"gameStage"`
	Count      int               `json:"count"`
}

type Game struct {
	mutex            sync.Mutex
	PlayerIdCounters int
	Deck             []deck.Card
	Players          []*Player
	PlayersWaiting   []*Player
	Dealer           *Dealer
	CurrentTurn      int
	GameStage        GameStage
}

type Player struct {
	Id               int
	SessionId        string
	Username         string
	Budget           float64
	DefaultStake     float64
	Stake            float64
	Game             *Game
	Hand             []deck.Card
	Count            int
	conn             *websocket.Conn
	ActiveConnection bool
}

type Dealer struct {
	Player
}

func sendMessage(ws *websocket.Conn, msg WebSocketMessage) {
	if err := websocket.JSON.Send(ws, msg); err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func NewGame() *Game {
	game := new(Game)
	game.PlayerIdCounters = 0
	game.Deck = deck.New(deck.Deck(1), deck.Shuffle)
	game.Dealer = new(Dealer)
	game.GameStage = PreGame
	game.Dealer.Game = game
	game.CurrentTurn = 0
	return game
}

func (g *Game) FilterPlayers() {
	var ActivePlayers []*Player
	for _, player := range g.Players {
		if player.ActiveConnection {
			ActivePlayers = append(ActivePlayers, player)
		}
	}
	g.Players = ActivePlayers
}
func (g *Game) NewRound() {
	g.FilterPlayers()
	g.CurrentTurn = 0
	g.Deck = deck.New(deck.Deck(1), deck.Shuffle)
	g.Dealer = new(Dealer)
	g.Dealer.Game = g
	g.GameStage = PreGame

	for _, player := range g.Players {
		player.Hand = nil
	}
}
func (g *Game) StartGame() {
	g.FilterPlayers()
	for len(g.Players) == 0 {
		fmt.Println("No players. Waiting 15 seconds....")
		go func() {
			time.Sleep(15 * time.Second)
			g.FilterPlayers()
		}()
	}
	var PlayersWaitingTmp []*Player
	g.GameStage = GameActive
	for _, player := range g.PlayersWaiting {
		if player.ActiveConnection {
			PlayersWaitingTmp = append(PlayersWaitingTmp, player)
			message := WebSocketMessage{MessageTypeGameNotReady, 1}
			sendMessage(player.conn, message)
		}
	}
	g.PlayersWaiting = PlayersWaitingTmp

	for _, player := range g.Players {
		player.Stake = player.DefaultStake
		player.Budget -= player.Stake
		msg := WebSocketMessage{MessageTypeStartGame, player.Budget}
		sendMessage(player.conn, msg)
	}
	g.Dealer.Hit()
	g.Dealer.Hit()

	g.Dealer.ShowHand()
	for i := 0; i < 2; i++ {
		for _, player := range g.Players {
			player.Hit()
		}
	}
	for _, player := range g.Players {
		player.ShowHand()
	}

	g.NextPlayer()
}

func (g *Game) DeletePlayer(id int) {
	for _, player := range g.Players {
		if player.Id == id {
			player.ActiveConnection = false
			break
		}
	}
}
func (p *Player) Reconnect(ws *websocket.Conn) {
	p.conn = ws
	p.ActiveConnection = true
}
func (g *Game) NextPlayer() {
	if g.CurrentTurn == len(g.Players) {
		g.EndGame()
		g.NewRound()
		go func() {
			time.Sleep(15 * time.Second)
			g.StartGame()
		}()
		return
	}
	player := g.Players[g.CurrentTurn]
	if !player.ActiveConnection || player.EvalHand() == 21 {
		g.NextPlayer()
		return
	}
	g.CurrentTurn++
	msg := WebSocketMessage{MessageTypeYourTurn, 1}
	sendMessage(player.conn, msg)
}
func (g *Game) EndGame() error {
	g.GameStage = PostGame
	g.Dealer.ShowHand()

	for g.Dealer.EvalHand() < 17 {
		g.Dealer.Hit()
		fmt.Println("Dealer's hands is less then 17. Drawing...")
		g.Dealer.ShowHand()
		if g.Dealer.EvalHand() > 21 {
			fmt.Println("Everybody wins!")

			for _, player := range g.Players {
				if !player.ActiveConnection {
					continue
				}
				if player.EvalHand() < 21 {
					player.Budget += player.Stake * 2
					message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: player.Stake * 2, Budget: player.Budget}}
					sendMessage(player.conn, message)
				}
				if player.EvalHand() == 21 {
					player.Budget += player.Stake * 2.5
					message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: player.Stake * 2, Budget: player.Budget}}
					sendMessage(player.conn, message)
				}
			}
			for _, player := range g.PlayersWaiting {
				if player.ActiveConnection {
					message := WebSocketMessage{MessageTypeGameReady, 1}
					sendMessage(player.conn, message)
				}
			}
			return nil
		}
	}

	dCount := g.Dealer.EvalHand()
	var currentWin float64 = 0.0
	for _, player := range g.Players {
		if !player.ActiveConnection {
			continue
		}
		currentWin = 0.0
		if player.EvalHand() == 21 {
			currentWin = player.Stake * 2.5
			player.Budget += currentWin
			message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: currentWin, Budget: player.Budget}}
			sendMessage(player.conn, message)
			fmt.Printf("%s has BLACKJACK and wins %f!\nCurrent budget: %f\n", player.Username, currentWin, player.Budget)
		} else if player.EvalHand() < 21 && player.EvalHand() > dCount {
			currentWin = player.Stake * 2
			player.Budget += currentWin
			player.Stake = 0
			message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: currentWin, Budget: player.Budget}}
			sendMessage(player.conn, message)
			fmt.Printf("%s Wins %f!\nCurrent budget: %f\n", player.Username, currentWin, player.Budget)
		} else if player.EvalHand() == dCount {
			currentWin = player.Stake
			player.Budget += currentWin
			player.Stake = 0
			message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: currentWin, Budget: player.Budget}}
			sendMessage(player.conn, message)
			fmt.Printf("Push. %s ties with a dealer. Refunding... %f!\nCurrent budget: %f\n", player.Username, currentWin, player.Budget)
		} else {
			player.Stake = 0
			message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: 0, Budget: player.Budget}}
			sendMessage(player.conn, message)
			fmt.Printf("%s loses!\nCurrent budget: %f\n", player.Username, player.Budget)
		}

	}
	for _, player := range g.PlayersWaiting {
		if player.ActiveConnection {
			message := WebSocketMessage{MessageTypeGameReady, 1}
			sendMessage(player.conn, message)
		}
	}
	return nil
}

func (g *Game) Bind(p *Player, ws *websocket.Conn, stake float64) (int, error) {
	// if g.GameStage == GameActive {
	// 	msg := WebSocketMessage{MessageTypeInitialHandshake, InitialHandshakeMessage{-1, "Welcome to the game!"}}
	// 	sendMessage(ws, msg)
	// 	return -1, nil
	// }
	if p.Budget < stake {
		return -1, fmt.Errorf("Not enough credits in player's budget...")
	}
	g.mutex.Lock()
	defer g.mutex.Unlock()

	found := false
	for _, player := range g.Players {
		if player.Id == p.Id && player.SessionId == p.SessionId {
			found = true
			break
		}
	}
	if (!found || p.Id == -1) && g.GameStage == GameActive {
		g.PlayersWaiting = append(g.PlayersWaiting, &Player{conn: ws, ActiveConnection: true})
		msg := WebSocketMessage{MessageTypeInitialHandshake, InitialHandshakeMessage{-1, "Welcome to the game!"}}
		sendMessage(ws, msg)
		return -1, nil
	}

	for _, waiting := range g.PlayersWaiting {
		if ws == waiting.conn {
			waiting.ActiveConnection = false
		}
	}
	p.Id = g.PlayerIdCounters
	g.PlayerIdCounters++
	p.Stake = stake
	p.DefaultStake = stake
	p.Game = g
	p.conn = ws
	p.ActiveConnection = true

	g.Players = append(g.Players, p)
	fmt.Printf("New player hes been gived ID of  %d\n", p.Id)

	msg := WebSocketMessage{MessageTypeInitialHandshake, InitialHandshakeMessage{p.Id, "Welcome to the game!"}}
	sendMessage(ws, msg)
	return p.Id, nil
}

// func (p *Player) Bind(game *Game) {
// 	p.Game = game
// 	game.Players = append(game.Players, p)
// }

func (p *Player) Hit() (deck.Card, error) {

	if len(p.Game.Deck) == 0 {
		return deck.Card{}, fmt.Errorf("Deck is empty")
	}
	pop := p.Game.Deck[len(p.Game.Deck)-1]
	p.Game.Deck = p.Game.Deck[:len(p.Game.Deck)-1]
	p.Hand = append(p.Hand, pop)
	return pop, nil
}

func (p *Player) EvalHand() int {
	p.Count = 0
	for _, card := range p.Hand {
		p.Count += card.GetValue()
	}
	return p.Count
}

func (p *Player) ShowHand() bool {
	if !p.ActiveConnection {
		return false
	}

	defer func() {
		var hands []OtherHandStruct
		for index, player := range p.Game.Players {
			if player.ActiveConnection {
				turn := ((p.Game.CurrentTurn) == index)
				hands = append(hands, OtherHandStruct{player.Id, player.Hand, time.Now(), turn})
			}

		}
		msg := WebSocketMessage{MessageTypeOtherHands, hands}
		for _, player := range p.Game.Players {
			if player.Id != p.Id && player.ActiveConnection {
				sendMessage(player.conn, msg)
			}
		}
	}()

	count := p.EvalHand()
	if count == 21 {
		// p.Budget += p.Stake * 2.5
		fmt.Printf("Count: %d\n-----------\n", p.Count)

		fmt.Println("You have BLACKJACK. Congrats...")
		message := WebSocketMessage{MessageTypeSendHand, HandMessage{Hand: p.Hand, Stage: false, Count: count}}
		sendMessage(p.conn, message)
		return false
	} else if count > 21 {
		// p.Stake = 0
		fmt.Printf("Count: %d\n-----------\n", p.Count)
		fmt.Println("Your hand exceeded the 21 score. You lost...")
		message := WebSocketMessage{MessageTypeSendHand, HandMessage{Hand: p.Hand, Stage: false, Count: count}}
		sendMessage(p.conn, message)
		return false
	}
	fmt.Printf("Count: %d\n-----------\n", p.Count)

	message := WebSocketMessage{MessageTypeSendHand, HandMessage{Hand: p.Hand, Stage: true, Count: count}}
	sendMessage(p.conn, message)

	return true
}

func (d *Dealer) ShowHand() {
	fmt.Printf("Dealer's hand:\n")

	if d.Game.GameStage == PostGame {
		message := WebSocketMessage{MessageTypeDealerFinalHand, HandMessage{Count: d.EvalHand(), Hand: d.Hand, Stage: true}}
		for _, player := range d.Game.Players {
			if !player.ActiveConnection {
				continue
			}
			sendMessage(player.conn, message)
		}
		d.EvalHand()
		fmt.Println(d.Count, "\n-------")
	} else {
		message := WebSocketMessage{MessageTypeDealerInitHand, d.Hand}
		for _, player := range d.Game.Players {
			if !player.ActiveConnection {
				continue
			}
			sendMessage(player.conn, message)
		}
		fmt.Println("\n-------")
	}
}

func (p *Player) SendReconnectState() {

	if p.Game.GameStage == PreGame {

		msg := WebSocketMessage{MessageTypeReconnectState, ReconnectStateMessage{
			DealerHand: []deck.Card{},
			Hand:       []deck.Card{},
			OtherHands: []OtherHandStruct{},
			Turn:       false,
			GameStage:  PreGame,
			Count:      p.EvalHand(),
		}}

		sendMessage(p.conn, msg)
		return
	}
	var index int
	var turn bool
	for i, player := range p.Game.Players {
		if player.Id == p.Id {
			index = i
			break
		}
	}
	turn = index != p.Game.CurrentTurn

	var hands []OtherHandStruct
	for index, player := range p.Game.Players {
		if player.ActiveConnection {
			turn := ((p.Game.CurrentTurn) == index)
			hands = append(hands, OtherHandStruct{player.Id, player.Hand, time.Now(), turn})
		}
	}

	msg := WebSocketMessage{MessageTypeReconnectState, ReconnectStateMessage{
		DealerHand: p.Game.Dealer.Hand,
		Hand:       p.Hand,
		OtherHands: hands,
		Turn:       turn,
		GameStage:  p.Game.GameStage,
		Count:      p.EvalHand(),
	}}

	sendMessage(p.conn, msg)

}

// func (d *Dealer) ResetDealer() {

// }
