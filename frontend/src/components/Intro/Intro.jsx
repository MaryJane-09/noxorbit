import { useState, useEffect } from "react";
import "./Intro.css";

function Intro({ onComplete }) {
    const [showName, setShowName] = useState(false);
    const [showTagline, setShowTagline] = useState(false);
    const [settled, setSettled] = useState(false);

    useEffect(() => {
        const nameTimer = setTimeout(() => setShowName(true), 800);
        const taglineTimer = setTimeout(() => setShowTagline(true), 1400);
        const settleTimer = setTimeout(() => setSettled(true), 3200);
        const doneTimer = setTimeout(() => onComplete?.(), 4600);
        return () => {
            clearTimeout(nameTimer);
            clearTimeout(taglineTimer);
            clearTimeout(settleTimer);
            clearTimeout(doneTimer);
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);
    // ...the return (JSX) stays exactly as it is

    

    return (
        <div className="cinematic-intro">
            <div className={`brand-assembly ${settled ? "is-settled" : ""}`}>
                <img src="/logo.png" className="main-logo" alt="" />
                <div className="brand-text">
                    <img src="/brand-name.png" className={`brand-name ${showName ? "is-visible" : ""}`} alt="Noxorbit" />
                    <img src="/brand-tagline.png" className={`brand-tagline ${showTagline ? "is-visible" : ""}`} alt="Connect, Collaborate, Achieve" />
                </div>
            </div>
        </div>
    );
}

export default Intro;