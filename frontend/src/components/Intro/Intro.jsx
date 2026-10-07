import { useState, useEffect } from "react";
import "./Intro.css";

function Intro() {
    const [showName, setShowName] = useState(false);
    const [showTagline, setShowTagline] = useState(false);
    const [settled, setSettled] = useState(false);

    useEffect(() => {
        const nameTimer = setTimeout(() => setShowName(true), 800);
        const taglineTimer = setTimeout(() => setShowTagline(true), 1400);
        const settleTimer = setTimeout(() => setSettled(true), 3200);
        return () => {
            clearTimeout(nameTimer);
            clearTimeout(taglineTimer);
            clearTimeout(settleTimer);
        };
    }, []);

    return (
        <div className="cinematic-intro">
            <div className={`brand-assembly ${settled ? "is-settled" : ""}`}>
                <img src="/logo.png" className="main-logo" alt="" />
                <img src="/brand-name.png" className={`brand-name ${showName ? "is-visible" : ""}`} alt="Noxorbit" />
                <img src="/brand-tagline.png" className={`brand-tagline ${showTagline ? "is-visible" : ""}`} alt="Connect, Collaborate, Achieve" />
            </div>
        </div>
    );
}

export default Intro;