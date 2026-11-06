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

    return (
        <div className="cinematic-intro">
            <div className={`brand-assembly ${settled ? "is-settled" : ""}`}>
                <img src="/noxorbit-mark.svg" className="main-logo" alt="" />
                <div className="brand-text">
                    <img src="/noxorbit-wordmark-study-02.svg" className={`brand-name ${showName ? "is-visible" : ""}`} alt="Noxorbit" />
                    <span className={`brand-tagline ${showTagline ? "is-visible" : ""}`}>CONNECT · COLLABORATE · ACHIEVE</span>
                </div>
            </div>
        </div>
    );
}

export default Intro;
