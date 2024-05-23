import { BrowserRouter, Routes, Route } from "react-router-dom";
import Game from './components/Game'
import Home from "./components/Home";
function App() {
  return (
    <BrowserRouter>
      <Routes>
          <Route path="/" element={<Home/>}/>
          <Route path ="/room/1" element={<Game port={8080} room={1}/>} />
          <Route path ="/room/2" element={<Game port={8081} room={2}/>} />
          <Route path ="/room/3" element={<Game port={8082} room={3}/>} />
          <Route path ="/room/4" element={<Game port={8083} room={4}/>} />
          <Route path ="/room/5" element={<Game port={8084} room={5}/>} />
      </Routes>
    </BrowserRouter>

  );
}

export default App;
