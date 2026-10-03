// src/components/CinematicIntro.jsx
import React, { useState, useEffect } from 'react';
import './Intro.css';
import logoAsset from '../assets/logo.png'; 

export default function Intro() {
    const [startAnimation, setStartAnimation] = useState(false);

    useEffect(() => {
        // Triggers the animation sequence a split-second after mounting
        const timer = setTimeout(() => {
            setStartAnimation(true);
        }, 100);
        return () => clearTimeout(timer);
    }, []);

    return (
        <div className="intro-container">
            {/* The main assembly block that coordinates scaling and moving */}
            <div className={`brand-assembly ${startAnimation ? 'is-moving' : ''}`}>

                {/* Logo Aspect wrapping your actual graphic asset */}
                <div className="logo-graphic-box">
                    <svg className="orbit-rings" viewBox="0 0 200 200">
                        <ellipse className="ring ring-1" cx="100" cy="100" rx="90" ry="40" />
                        <ellipse className="ring ring-2" cx="100" cy="100" rx="90" ry="40" />
                        <ellipse className="ring ring-3" cx="100" cy="100" rx="90" ry="40" />
                    </svg>
                    <img src={logoAsset} className="main-logo-asset" alt="Noxorbit Logo" />
                </div>
                
                {/* The Text Layout reveal track */}
                <div className="brand-text-block">
                    <span className="brand-name">NOXORBIT</span>
                    <span className="brand-tagline">CONNECT • COLLABORATE • ACHIEVE</span>
                </div>

            </div>
        </div>
    );
}
