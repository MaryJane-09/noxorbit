import { useState } from 'react'
import Intro from './components/Intro'
import Register from './components/register'
import './App.css'

function App() {
  const [showMainContent, setShowMainContent] = useState(false);

  return (
    <div className="app-container">
      {/* The isolated cinematic sequence */}
      <intro onAnimationComplete={() => setShowMainContent(true)} />

      {/* Main content grid area (Triggers after the intro completes) */}
      <div className={`main-canvas ${showMainContent ? 'content-fade-in' : 'content-hidden'}`}>
        {/* WE WILL BUILD YOUR FIRST ACTUAL SCREEN RIGHT HERE NEXT */}
        <p style={{ color: '#64748b', fontSize: '18px' }}>
          <Register />
        </p>
      </div>
    </div>
  );
}

export default App