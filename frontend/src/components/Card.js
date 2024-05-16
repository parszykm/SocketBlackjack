import React from 'react'
import { useState, useEffect } from 'react'
import heart from '../assets/heart.svg'
import club from '../assets/club.svg'
import diamond from '../assets/diamond.svg'
import spade from '../assets/spade.svg'
import style from './Card.css'
import queen  from '../assets/queen.svg'
import king from '../assets/king.svg'
import jack from '../assets/jack.svg'
import queenb  from '../assets/queenb.svg'
import kingb from '../assets/kingb.svg'
import jackb from '../assets/jackb.svg'
import reverse from '../assets/reverse.svg'
function getSuitPath(suit){
    switch(suit){
        case 0:
            return spade
        case 1:
            return diamond
        case 2:
            return club
        case 3:
            return heart
        default:
            return ''
            
    }
}

function getRank(rank){
    if(rank>1 && rank<=10){
        return rank.toString()
    }
    else if(rank == 11) return 'J'
    else if(rank == 12) return 'Q'
    else if(rank == 13) return 'K'
    else if(rank == 1) return 'A'
    else return ''

}

function getCenterIcon(rank, suit){
    let color = getColor(suit)
    if(color === 'red')
    {
        if(rank == 11) return jack
        else if(rank == 12) return queen
        else if(rank == 13) return king
        return getSuitPath(suit)
    }
    else{
        if(rank == 11) return jackb
        else if(rank == 12) return queenb
        else if(rank == 13) return kingb
        return getSuitPath(suit)
    }

}

function getColor(suit){
    if(suit == 1 || suit == 3) return 'red'
    return 'black'
}
function Card({suit,rank, reversed=false}) {
    const [suitState, setSuitState] = useState()
    const [rankState, setRankState] = useState()
    const [centerIconState, setCenterIconState] = useState()
    const [reverseState, setReverseState] = useState(true)
    // console.log("Suit:",suitState, suit)
    // console.log("rank:", rankState, rank)
    useEffect(() => {
        setReverseState(reversed)
    }, [reversed])
    useEffect(() => {
        setSuitState(getSuitPath(suit))
        setRankState(getRank(rank))
        setCenterIconState(getCenterIcon(rank,suit))
        // setReverseState(reversed)
    },[])
    const isOddRank = rank % 2 !== 0;
    const oddIndex = isOddRank ? Math.ceil(rank / 3) : 0
    if(reverseState){
        return (
            <div className="card">
                <img className="reverse" src={reverse}/>
            </div>
        )
    }
    return (
        <div className="card">
            <div className="header">
                <div className='suit-rank'>
                    <p className="card-rank">{rankState}</p>
                    <img className="card-image" src={suitState} />
                </div>
            </div>
            {/* <div className={`content ${isOddRank ? 'odd-rank' : 'even-rank'}`}>
                {(() => {
                    let icons = [];

                    for(let i=0; i<rank; i++){
                        icons.push(<img key={i} className="card-icon" src={suitState} />);
                    }
                    return icons;
                })()}
            </div> */}
            <div className="content">
                <img className="card-image-center" src={centerIconState}/>
            </div>
            <div className="footer">
                <div className='suit-rank'>
                    <p className="card-rank">{rankState}</p>
                    <img className="card-image" src={suitState}/>
                </div>
            </div>
        </div>
      );
}

export default Card