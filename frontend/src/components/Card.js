import React from 'react'
import { useState, useEffect, useRef} from 'react'
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
    const canvasRef = useRef(null);
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

    useEffect(() => {
        if(!canvasRef.current) return
        const canvas = canvasRef.current;
        const ctx = canvas.getContext('2d');
    
        ctx.fillStyle = '#f8f8f8';
        ctx.strokeStyle = '#ccc';
        ctx.lineWidth = 5;
        const borderRadius = 20;
        const width = canvas.width;
        const height = canvas.height;
    
        ctx.beginPath();
        ctx.moveTo(borderRadius, 0);
        ctx.lineTo(width - borderRadius, 0);
        ctx.quadraticCurveTo(width, 0, width, borderRadius);
        ctx.lineTo(width, height - borderRadius);
        ctx.quadraticCurveTo(width, height, width - borderRadius, height);
        ctx.lineTo(borderRadius, height);
        ctx.quadraticCurveTo(0, height, 0, height - borderRadius);
        ctx.lineTo(0, borderRadius);
        ctx.quadraticCurveTo(0, 0, borderRadius, 0);
        ctx.closePath();
        ctx.fill();
        ctx.stroke();
      }, [canvasRef]);

    if(reverseState){
        return (
            <div className="card">
                <img className="reverse" src={reverse}/>
            </div>
        )
    }
    return (
        <div className="card">
             <canvas ref={canvasRef} width={400} height={600} className="card-canvas" />
            <div className="header">
                <div className='suit-rank'>
                    <p className="card-rank">{rankState}</p>
                    <img className="card-image" src={suitState} />
                </div>
            </div>
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