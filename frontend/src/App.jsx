import { useState, useEffect } from "react";
import Intro from "./components/Intro/Intro";
import Register from "./components/Register";
import Login from "./components/Login";
import Home from "./components/Home";
import { fetchMe } from "./api";
import "./App.css";

function App() {
  const [introDone, setIntroDone] = useState(false);
  const [view, setView] = useState("register"); 
  const [user, setUser] = useState(null);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    fetchMe().then((me) => {
      setUser(me);
      setChecking(false);
    });
  }, []);

  async function handleLoginSuccess() {
    setUser(await fetchMe());
  }

  let screen = null;
  if (introDone && !checking) {
    if (user) {
      screen = <Home user={user} onLogout={() => { setUser(null); setView("login"); }} />;
    } else if (view === "login") {
      screen = <Login onSuccess={handleLoginSuccess} onRegister={() => setView("register")} />;
    } else {
      screen = <Register onLogin={() => setView("login")} />;
    }
  }

  return (
    <div className="app-shell">
      <div className="app-background" />
      <Intro onComplete={() => setIntroDone(true)} />
      <div className="app-content">{screen}</div>
    </div>
  );
}

export default App;