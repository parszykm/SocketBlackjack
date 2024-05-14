import React, { useEffect, useState } from "react";

const WebSocketComponent = () => {
    const [message, setMessage] = useState("");

    useEffect(() => {
        const socket = new WebSocket("ws://localhost:8080/ws");

        socket.onmessage = (event) => {
            console.log("Received message:", event.data);
            setMessage(event.data);
        };

        socket.onclose = () => {
            console.log("WebSocket connection closed");
        };

        return () => {
            socket.close();
        };
    }, []);

    return (
        <div>
            <h1>WebSocket Example</h1>
            <p>Received message: {message}</p>
        </div>
    );
};

export default WebSocketComponent;
