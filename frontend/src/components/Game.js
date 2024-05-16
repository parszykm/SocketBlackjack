import React, { useEffect, useState } from 'react';
import Card from './Card';
import Hand from './Hand';
import './Game.css'

function Game() {
  const [hand, setHand] = useState([]);
  const [dealer, setDealer] = useState({count: 0, hand: []});
  const [turn, setTurn] = useState(false)
  const [budget, setBudget] = useState(2000);
  const [stake, setStake] = useState(0)
  const [wsConn, setWsConn] = useState()
  const [count, setCount] = useState(0)
  const [playable, setPlayable] = useState(true)
  const [resultText, setResultText] = useState('')
  const [playerId, setPlayerId] = useState(null)
  const [otherPlayers, setOtherPlayers] = useState([])

  useEffect(() => {
    console.log('DEALER',dealer)
    if(dealer.hand.length === 0){
      console.log("WYCZYSZCZONE ESSA",dealer.hand)
    }
  }, [dealer])
  function startGame() {
    let msg ={
      type: 'StartGame',
      msg: 'Start Game'
    }
    wsConn.send(JSON.stringify(msg));
  }
  function endGame() {
    let msg ={
      type: 'EndGame',
      msg: 'End Game'
    }
    wsConn.send(JSON.stringify(msg));
    // setTimeout(()=>{
    //   startGame();
    // },5000)
  }
  function stand() {
    let msg = {
      type: 'Stand',
      msg: ''
    }
    wsConn.send(JSON.stringify(msg));
    setTurn(false)
  }
  function hit() {
    let msg = {
      type: 'Hit',
      msg: ''
    }
    wsConn.send(JSON.stringify(msg));
  }
  useEffect(() => {
    if(!playable){
      stand()
    }
  }, [playable])
  useEffect(() => {
    // Create a WebSocket connection to the backend
    const ws = new WebSocket('ws://localhost:8080/ws');
    setWsConn(ws)
    // Event handler for receiving a message from the backend
    ws.onmessage = (event) => {
      let msg = JSON.parse(event.data);
      switch (msg.type) {
        case 'InitialHandshake':
          console.log('Received initial handshake:', msg.data);
          setPlayerId(msg.data.id)
          break
        case 'SendHand':
          console.log('Received hand:', msg.data);
          setHand(msg.data.hand)
          setCount(msg.data.count)
          setPlayable(msg.data.stage)
          break
        case 'GameResult':
          console.log('Received game result:', msg.data);
          setBudget(msg.data.budget)
          setResultText(`You have won ${msg.data.refund}. Congrats!`)
          break
        case 'DealerInitHand':
          console.log('Received dealer init hand', msg.data);
          let initList = msg.data.map((item, index) => {
            if(index === 0){
              return ({suit:item.suit, rank: item.rank, reversed: true})
            }
            else return item})
          setDealer(old => ({...old, hand: initList}))
          break
        case 'DealerFinalHand':
          console.log('Received dealer final hand', msg.data);
          const newList = msg.data.hand
          setDealer(old => ({count: msg.data.count, hand: newList}))
          // setDealer.count(msg.data.count)
          break
        case 'StartGame':
          console.log('Received start game', msg.data);
          setDealer(({...dealer, count: 0, hand: []}))
          // setDealer.count(0)
          setHand([])
          break
        case 'OtherHands':
          console.log('Received other hands', msg.data);
          setOtherPlayers(msg.data)
          break
        case 'YourTurn':
          console.log('Received your turn', msg.data);
          setTurn(true)
          break
        default:
          console.log('Unknown message type:', msg.type);
      }
    };
    // Clean up the WebSocket connection when the component is unmounted
    return () => {
      ws.close();
    };
  }, []);

  return (
    <div className='game'>
      <h1>Player ID: {playerId}</h1>
      <h2>Budget: {budget}</h2>
      <h2>Stake: {stake}</h2>

        <div className='game_header'>
            <div className="game_dealer_dashboard">
                <h1>Dealer's hand</h1>
                <Hand cards={dealer.hand}/>
                <p>Dealer count: {dealer.count}</p>
            </div>
            <div className="game_hand">
                <div className="game_hand_cards">
                <h1>Your hand: </h1>
                <Hand cards={hand}/>
                <div className="count">
                    <p>Your count: {count}</p>
                </div>
                </div>
                <div className="game_hand_panel">
                    <button className='button-17' onClick={hit} disabled={!playable || !turn}>Hit</button>
                    <button className='button-17' onClick={stand} disabled={!playable || !turn}>Stand</button>
                </div>
            </div>
        </div>
        <div className="buttons">
            <button className='button-17' onClick={startGame}>Start game</button> 
            <button className='button-17' onClick={endGame}>End game</button> 
            <button className='button-17' onClick={() => {startGame(); endGame();}}>Refresh game</button> 
        </div>
      <div className='results'>
        <p>{resultText}</p>
      </div>

      <div className='game_other'>
        {otherPlayers.map((player) => 
          {if(player.id != playerId){
            return(
              <div className='game_other_player' key={player.id}>
                <h2>Player ID: {player.id}</h2>
                <Hand cards={player.hand}/>
              </div>
            )
          }
        })

        }
      </div>
    </div>
  );
}

export default Game;
