import React, { useEffect, useState } from 'react';
import Card from './components/Card';
import Hand from './components/Hand';
// import 'bootstrap/dist/css/bootstrap.min.css';
// const cards = [
//   { value: 'A', suit: 'hearts' },
//   { value: '10', suit: 'spades' },
//   { value: 'K', suit: 'diamonds' },
// ];

function App() {
  const [hand, setHand] = useState([]);
  const [dealerHand, setDealerHand] = useState([]);
  const [budget, setBudget] = useState(2000);
  const [stake, setStake] = useState(0)
  const [wsConn, setWsConn] = useState()
  const [count, setCount] = useState(0)
  const [playable, setPlayable] = useState(true)
  const [resultText, setResultText] = useState('')

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
  }
  function stay() {
    let msg = {
      type: 'Stay',
      msg: ''
    }
    wsConn.send(JSON.stringify(msg));
  }
  function hit() {
    let msg = {
      type: 'Hit',
      msg: ''
    }
    wsConn.send(JSON.stringify(msg));
  }
  

  useEffect(() => {
    // Create a WebSocket connection to the backend
    const ws = new WebSocket('ws://localhost:8080/ws');
    setWsConn(ws)
    // Event handler for receiving a message from the backend
    ws.onmessage = (event) => {
      let msg = JSON.parse(event.data);
      console.log(msg);
      switch (msg.type) {
        case 'InitialHandshake':
          console.log('Received initial handshake:', msg.data);
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
          setDealerHand(initList)
          break
        case 'DealerFinalHand':
          console.log('Received dealer final hand', msg.data);
          const newList = msg.data.hand
          setDealerHand(newList)
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
    <div>
      <h2>Budget: {budget}</h2>
      <h2>Stake: {stake}</h2>
      <h1>Received from WebSocket:</h1>
      <div className="test">
        <Hand cards={hand}/>
      </div>
      <div className="count">
        <p>{count}</p>
      </div>
      <button className='btn' onClick={hit} disabled={!playable}>Hit</button>
      <button className='btn' onClick={stay}>Stay</button>
      <button className='btn btn-primary' onClick={startGame}>Start game</button> 
      <button className='btn btn-primary' onClick={endGame}>End game</button> 
      <div className='results'>
        <p>{resultText}</p>
      </div>

      <div className="test">
        <Hand cards={dealerHand}/>
      </div>
    </div>
  );
}

export default App;
