import React from 'react'
import Card from './Card.js'
import style from './Hand.css'
const Hand = ({cards}) => {
  return (
    <div className="hand">
        {
            cards.map((card, index) => (
            <div className="card-holder" key={`${card.suit}-${card.rank}`} style={{zIndex: index}}>
            <Card rank={card.rank} suit={card.suit} reversed={card.reversed}/>
            </div>))
        }
    </div>
  )
}

export default Hand