import { useState, useEffect } from "react";
import Intro from "./components/Intro/Intro";
import Register from "./components/Register";
import Login from "./components/Login";
import Home from "./components/Home";
import ForgotPassword from "./components/ForgotPassword";
import { fetchMe } from "./api";
import "./App.css";

function App() {
  const [introDone, setIntroDone] = useState(false);
  const [view, setView] = useState(() => window.location.hash.startsWith("#invite/") ? "login" : "register");
  const [user, setUser] = useState(null);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    let active = true;
    async function restoreSession() {
      try {
        const me = await fetchMe();
        if (active) setUser(me);
      } catch (error) {
        console.error("Could not restore session:", error);
      } finally {
        if (active) setChecking(false);
      }
    }
    void restoreSession();
    return () => { active = false; };
  }, []);

  async function handleLoginSuccess() {
    const me = await fetchMe();
    if (me) setUser(me);
  }

  let screen = null;
  if (!checking && user) {
    // Signed-in refresh goes straight to the existing dark dashboard.
    screen = <Home user={user}
      onUserUpdated={setUser}
      onLogout={() => { setUser(null); setIntroDone(true); setView("login"); }} />;
  } else if (!checking && introDone) {
    if (view === "forgot") {
      screen = <ForgotPassword onLogin={() => setView("login")} />;
    } else if (view === "login") {
      screen = <Login onSuccess={handleLoginSuccess} onRegister={() => setView("register")} onForgot={() => setView("forgot")} />;
    } else {
      screen = <Register onLogin={() => setView("login")} />;
    }
  }

  return (
    <div className={`app-shell${user ? " app-shell--dashboard" : ""}${checking ? " app-shell--checking" : ""}`}>
      {!checking && !user && <>
        <div className="app-background" />
        {!introDone && <Intro onComplete={() => setIntroDone(true)} />}
      </>}
      <div className="app-content">{screen}</div>
    </div>
  );
}

export default App;
