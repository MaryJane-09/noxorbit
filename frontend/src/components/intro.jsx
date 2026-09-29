import React, { useState, useEffect } from 'react';
import './intro.css';


export default function CinematicIntro({ onAnimationComplete }) {
    const [isSettled, setIsSettled] = useState(false);

    useEffect(() => {
        const timer = setTimeout(() => {
            setIsSettled(true);
            if (onAnimationComplete) onAnimationComplete();
        }, 3800); 
        return () => clearTimeout(timer);
    }, [onAnimationComplete]);

    return (
        <div className={`intro-overlay ${isSettled ? 'intro-minimized' : ''}`}>
            {}
            <div className="brand-assembly">

                {}
                <div className="logo-container">
                    <svg className="orbital-ring" viewBox="0 0 100 100">
                        <ellipse cx="50" cy="50" rx="45" ry="15" />
                    </svg>
                    <div className="brand-letter">N</div>
                </div>

                {}
                <div className="brand-text-block">
                    <span className="brand-name">NOXORBIT</span>
                    <span className="brand-tagline">CONNECT • COLLABORATE • ACHIEVE</span>
                </div>

            </div>
        </div>
    );
}

