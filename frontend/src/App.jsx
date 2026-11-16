import { useState, useEffect } from "react";
import Intro from "./components/Intro/Intro";
import Register from "./components/Register";
import Login from "./components/Login";
import Home from "./components/Home";
import ForgotPassword from "./components/ForgotPassword";
import { fetchMe, fetchAppearancePreferences } from "./api";
import "./App.css";

const DEFAULT_APPEARANCE = { theme: "dark", reduced_motion: false };

async function loadSignedInAppearance() {
  try {
    const appearance = await fetchAppearancePreferences();
    try { window.localStorage.setItem("noxorbit.appearance.hint", appearance.theme); } catch { /* storage unavailable */ }
    return appearance;
  } catch (error) {
    // Never display a placeholder dark dashboard while a theme is loading.
    // If loading genuinely fails, keep the safe dark default.
    console.error("Could not load appearance preferences:", error);
    return DEFAULT_APPEARANCE;
  }
}

function App() {
  const [introDone, setIntroDone] = useState(false);
  const [view, setView] = useState(() => window.location.hash.startsWith("#invite/") ? "login" : "register");
  const [session, setSession] = useState(null);
  const [checking, setChecking] = useState(true);
  const [loadingTheme] = useState(() => {
    try {
      const hint = window.localStorage.getItem("noxorbit.appearance.hint");
      if (hint === "light") return "light";
      if (hint === "system") return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
    } catch { /* use dark loading surface */ }
    return "dark";
  });

  useEffect(() => {
    let active = true;
    async function restoreSession() {
      const me = await fetchMe();
      const appearance = me ? await loadSignedInAppearance() : null;
      if (!active) return;
      setSession(me ? { user: me, appearance } : null);
      setChecking(false);
    }
    void restoreSession();
    return () => { active = false; };
  }, []);

  async function handleLoginSuccess() {
    const me = await fetchMe();
    if (!me) return;
    const appearance = await loadSignedInAppearance();
    setSession({ user: me, appearance });
  }

  const user = session?.user ?? null;
  let screen = null;
  if (!checking && user) {
    // Authenticated refreshes skip Intro entirely. Appearance is resolved
    // before mounting Home, so no dark-to-light repaint occurs.
    screen = <Home user={user} initialAppearance={session.appearance}
      onUserUpdated={(updated) => setSession((current) => current ? { ...current, user: updated } : current)}
      onLogout={() => { setSession(null); setIntroDone(true); setView("login"); }} />;
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
    <div className={`app-shell${user ? " app-shell--dashboard" : ""}${checking ? " app-shell--checking" : ""}`}
      data-loading-theme={loadingTheme}>
      {!checking && !user && <>
        <div className="app-background" />
        {!introDone && <Intro onComplete={() => setIntroDone(true)} />}
      </>}
      <div className="app-content">{screen}</div>
    </div>
  );
}

export default App;
