import React, { useEffect, useState, useRef } from 'react';
import Hand from './Hand';
import './Game.css'
import logo from '../assets/logo-dark.svg';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import CountdownTimer from './CountdownTimer';
import Logo from './Logo';
function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
  });
}

function Game() {
  const stakeInputRef = useRef(null);
  const [hand, setHand] = useState([]);
  const [dealer, setDealer] = useState({count: 0, hand: []});
  const [turn, setTurn] = useState(false)
  const [budget, setBudget] = useState(2000);
  const [stake, setStake] = useState(20)
  const [wsConn, setWsConn] = useState()
  const [count, setCount] = useState(0)
  const [playable, setPlayable] = useState(true)
  const [resultText, setResultText] = useState('')
  const [playerId, setPlayerId] = useState(null)
  const [otherPlayers, setOtherPlayers] = useState([])
  const [timeRem, setTimeRem] = useState(0)
  const [gameReadyState, setGameReadyState] = useState(false)
  const [gameRunning, setGameRunning] = useState(false)

  function bindToGame() {
    let sessionId = sessionStorage.getItem('sessionId');
    if (!sessionId) {
        sessionId = generateUUID();
        sessionStorage.setItem('sessionId', sessionId);
    }
    let oldId = -1
    let LSplayerId = localStorage.getItem(`${sessionId}-PlayerId`)
    if(LSplayerId){
      oldId = parseInt(LSplayerId)
    }
    let msg ={
      type: 'Bind',
      data: {
        username: "Ziom",
        budget: budget,
        stake: stake,
        oldId: oldId,
        sessionId: sessionId
      }
    }
    wsConn.send(JSON.stringify(msg));
  }
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
  function changeStake(){
    let newStake = stakeInputRef.current.value
    if(newStake > budget){
      newStake = budget
      stakeInputRef.current.value = budget
    }
    // setStake(newStake)
    let msg = {
    type: 'ChangeStake',
    data: parseInt(newStake)
    }
    wsConn.send(JSON.stringify(msg));
    console.log('WYSLANE')
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
    if(wsConn){
      setTimeout(() => {
        bindToGame()
      }, 300)
    }
  }, [wsConn])
  useEffect(() => {

    let sessionId = sessionStorage.getItem('sessionId');
    if (!sessionId) {
        // Jeśli nie, wygeneruj nowy identyfikator i zapisz go w sessionStorage
        sessionId = generateUUID();
        sessionStorage.setItem('sessionId', sessionId);
    }

    // Create a WebSocket connection to the backend
    // localStorage.removeItem('PlayerId');
    const ws = new WebSocket('ws://localhost:8080/ws');

    setWsConn(ws)

    // Event handler for receiving a message from the backend
    ws.onmessage = (event) => {
      let msg = JSON.parse(event.data);
      switch (msg.type) {
        case 'InitialHandshake':
          console.log('Received initial handshake:', msg.data);
          if (msg.data.id == -1){
            console.log('Cannot bind right now. Wait for the current game to end...')
            setGameReadyState(false)
            setResultText('Cannot bind right now. Wait for the current game to end...')
            break
          }
          if(localStorage.getItem(`${sessionId}-PlayerId`)){
              let reconnect = {
                type: 'Reconnect',
                data: {
                  storedId: parseInt(localStorage.getItem(`${sessionId}-PlayerId`)),
                  givenId: msg.data.id,
                  sessionId: sessionId
                }
              }
              ws.send(JSON.stringify(reconnect));
          }
          else{
            setPlayerId(msg.data.id)
            localStorage.setItem(`${sessionId}-PlayerId`, msg.data.id)
          }
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
          setTimeRem(15)
          break
        case 'GameReady':
          setTimeRem(15)
          setResultText('You can join to game now...')
          break
        case 'GameNotReady':
          setResultText('Cannot bind right now. Wait for the current game to end...')
          setTimeRem(0)
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
          setDealer(() => ({count: msg.data.count, hand: newList}))
          // setDealer.count(msg.data.count)
          break
        case 'StartGame':
          console.log('Received start game', msg.data);
          setDealer(({...dealer, count: 0, hand: []}))
          setBudget(msg.data)
          setResultText("")
          setTimeRem(0)
          setGameReadyState(true)
          setGameRunning(true)
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
        case 'ReconnectResponse':
          console.log('Received reconnect response', msg.data);
          setPlayerId(msg.data)
          localStorage.setItem(`${sessionId}-PlayerId`, msg.data)
          break
        case 'ReconnectState':
          console.log('Received reconnect state', msg.data);
          setGameReadyState(true)
          if(msg.data.gameStage < 3 ){
            let dHand = msg.data.dealerHand.map((item, index) => {
              if(index === 0){
                return ({suit:item.suit, rank: item.rank, reversed: true})
              }
              else return item})
            setDealer({count: 0, hand: dHand})
          }
          else{
            setDealer({count: 0, hand: msg.data.dealerHand})
          }
          setHand(msg.data.hand)
          setOtherPlayers(msg.data.otherHands)
          setTurn(msg.data.turn)
          setCount(msg.data.count)

          break
        case 'ChangeStakeResponse':
          console.log('Received changeStake response', msg.data)
          if (msg.data == stake){
            setResultText('Cannot change bet right now. Wait for game to end...')
            stakeInputRef.current.value = msg.data
            break
          }
          setStake(msg.data)
          setResultText(`Changed bet to ${msg.data}`)
          stakeInputRef.current.value = msg.data
          
          break
          
        default:
          console.log('Unknown message type:', msg.type);
      }
    };

    return () => {
      ws.close();
    };
  }, []);

  return (
    <div className='game'>
        <div className='game_header'>
          <div className='game_header_top'>
            <div className='game_player_info'>
              {/* <img src={logo} className='logo'/> */}
              <Logo className='logo'/>
              <div className='game_text-info'>
                <h4>Player ID</h4>
                <p>{playerId}</p>
              </div>
              <div className='game_text-info'>
                <h4>Budget</h4>
                <p>{budget}</p>
              </div>

            <div>
          
              <div className='stake_input'>
                <TextField label="Bet"
                  className='stake_input_box'
                  defaultValue={stake}
                  placeholder={stake}
                  InputLabelProps={{ style: { color: 'white'}}}
                  InputProps={{ style: { color: 'white' } }}
                  sx={{ '& .MuiOutlinedInput-root': { color: 'white' } }}
                  inputRef={stakeInputRef}
                  />

                <Button onClick={changeStake} variant="contained" className='button-stake' >
                  Change
                </Button>
              </div>
              </div>
          </div>
            <div className="game_dealer_dashboard" style={{display: gameReadyState ? 'flex' : 'none'}}>
                <h1>Dealer's hand</h1>
                <Hand cards={dealer.hand}/>
                <p>Dealer count: {dealer.count}</p>
            </div>
          </div>
            <div className="game_hand" style={{display: gameReadyState ? 'flex' : 'none'}}>
                <div className="game_hand_cards">
                <h1>Your hand: </h1>
                <Hand cards={hand}/>
                <div className="count">
                    <p>Your count: {count}</p>
                </div>
                </div>
                <div className="game_hand_panel" >
                    <Button variant='contained' color="secondary" onClick={hit} disabled={!playable || !turn}>Hit</Button>
                    <Button variant='contained' color="secondary" onClick={stand} disabled={!playable || !turn}>Stand</Button>
                </div>
            </div>
            <div className='results'>
              <p>{resultText}</p>
            </div>
        </div>
        <div className="buttons">
            <Button variant='contained' onClick={startGame} disabled={gameRunning}>Start game</Button> 
            <Button variant='contained' onClick={bindToGame} disabled={playerId !== null ? true : false}>Join</Button> 
        </div>
      <CountdownTimer className='timer' initialSeconds={timeRem} />
      <div className='game_other'>
        {otherPlayers.map((player) => 
          {
            return(
              <div className='game_other_player' key={player.id}>
                <h2>{player.id == playerId ? "You" : `Player ID: ${player.id}`}<b>{player.turn ? " Turn" : " "}</b></h2>
                <Hand cards={player.id == playerId ? hand : player.hand}/>
              </div>
            )
          
        })

        }
      </div>
    </div>
  );
}

export default Game;
