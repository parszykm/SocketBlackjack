import React from 'react';
import { Outlet, Link } from 'react-router-dom';
import './Home.css';
import logo from '../assets/logo-light.svg'

const Home = () => {
    return (
        <div className="home-container">
            <header className="welcome-header">
                <img src={logo} className="home-logo"/>
                <h1>Welcome to Blackjack Casino!</h1>
            </header>
            <h2>Rooms:</h2>
            <nav className="room-nav">
                <ul>
                    <li>
                        <Link to="/room/1">Room 1</Link>
                    </li>
                    <li>
                        <Link to="/room/2">Room 2</Link>
                    </li>
                    <li>
                        <Link to="/room/3">Room 3</Link>
                    </li>
                    <li>
                        <Link to="/room/4">Room 4</Link>
                    </li>
                    <li>
                        <Link to="/room/5">Room 5</Link>
                    </li>
                </ul>
            </nav>
            <Outlet />
            <footer className="rules-footer">
                <h2>Blackjack Rules:</h2>
                <p>1. Players aim to have a hand value closer to 21 than the dealer's hand without exceeding 21.</p>
                <p>2. Face cards (Jack, Queen, King) are worth 10 points, Aces can are worth 11 points, and other cards are worth their face value.</p>
                <p>3. The dealer must hit until their hand value is 17 or higher.</p>
                <p>4. If a player's hand value exceeds 21, they bust and lose the game.</p>
                <p>5. If the dealer busts, all remaining players win.</p>
            </footer>
        </div>
    );
};

export default Home;
