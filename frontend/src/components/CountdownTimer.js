import React, { useState, useEffect } from 'react';

function CountdownTimer({ initialSeconds }) {
  const [seconds, setSeconds] = useState(initialSeconds);
    
  useEffect(() => {
      setSeconds(initialSeconds)
      if (initialSeconds > 0) {
        console.log("Changes", initialSeconds)
      const intervalId = setInterval(() => {
        setSeconds(prevSeconds => Math.max(prevSeconds - 1,0));
      }, 1000);
      
      // Clear the interval on component unmount
      return () => clearInterval(intervalId);
    }
  }, [initialSeconds]);

  return (
    <div className='timer'>
      <h1>Countdown Timer</h1>
      <p>{seconds} seconds remaining to new game</p>
    </div>
  );
}

export default CountdownTimer;
