import React from 'react'
import Card from './Card.js'
import style from './Hand.css'
const Hand = ({cards}) => {
    console.log(cards)
  return (
    <div className="hand">
        {
            cards.map((card, index) => (
            <div className="card-holder" key={index} style={{zIndex: index}}>
            <Card rank={card.rank} suit={card.suit} reversed={card.reversed}/>
            </div>))
        }
    </div>
  )
}

export default Hand