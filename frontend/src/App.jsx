import { useState } from "react";
import Intro from "./components/Intro/Intro";
import Register from "./components/Register";
import "./App.css";

function App() {
  const [introDone, setIntroDone] = useState(false);

  return (
    <div className="app-shell">
      <div className="app-background" />
      <Intro onComplete={() => setIntroDone(true)} />
      {introDone && <Register />}
      <div className="app-content">
        {introDone && <Register />}
      </div>
    </div>
  );
}

export default App;