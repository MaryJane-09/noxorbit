import "./AuthCosmos.css";

// Deterministic stars complement a separately illustrated planet asset.
const noise = (n) => {
    const value = Math.sin(n * 127.1 + 71.7) * 43758.5453123;
    return value - Math.floor(value);
};

const STARS = Array.from({ length: 92 }, (_, i) => ({
    x: Math.round(noise(i + 11) * 1672),
    y: Math.round(noise(i + 119) * 920),
    radius: 0.35 + noise(i + 237) * 1.1,
    opacity: 0.12 + noise(i + 351) * 0.48,
}));

const ORBIT =
    "M 890 268 C 747 328 652 414 616 510 C 589 579 633 611 721 603 C 794 595 882 555 946 509";

export default function AuthCosmos() {
    return (
        <div className="auth-cosmos" aria-hidden="true">
            <svg className="auth-cosmos-svg" viewBox="0 0 1672 940"
                preserveAspectRatio="xMidYMid slice" focusable="false">
                <defs>
                    <radialGradient id="nox-nebula">
                        <stop stopColor="#692575" stopOpacity=".62" />
                        <stop offset=".48" stopColor="#3c174b" stopOpacity=".28" />
                        <stop offset="1" stopColor="#110819" stopOpacity="0" />
                    </radialGradient>
                    <linearGradient id="nox-orbit-gold" x1="0" y1="0" x2="1" y2="1">
                        <stop stopColor="#824691" stopOpacity="0" />
                        <stop offset=".24" stopColor="#b84fa8" stopOpacity=".5" />
                        <stop offset=".53" stopColor="#ffd3ad" />
                        <stop offset=".68" stopColor="#f6b878" />
                        <stop offset="1" stopColor="#9541a0" stopOpacity=".05" />
                    </linearGradient>
                    <radialGradient id="nox-twilight">
                        <stop stopColor="#f4a1f0" stopOpacity=".3" />
                        <stop offset="1" stopColor="#f4a1f0" stopOpacity="0" />
                    </radialGradient>
                    <filter id="nox-soft-glow" x="-100%" y="-100%" width="300%" height="300%">
                        <feGaussianBlur stdDeviation="5" />
                    </filter>
                </defs>

                <ellipse cx="140" cy="610" rx="780" ry="720" fill="url(#nox-nebula)" />
                <ellipse cx="1420" cy="220" rx="530" ry="450" fill="url(#nox-nebula)" opacity=".22" />
                <ellipse cx="250" cy="811" rx="880" ry="240" fill="url(#nox-twilight)" opacity=".7" />

                <g className="auth-stars">
                    {STARS.map((star, i) => (
                        <circle key={i} cx={star.x} cy={star.y} r={star.radius}
                            fill="#f9d7f7" opacity={star.opacity}
                            className={i % 11 === 0 ? "auth-star-pulse" : undefined} />
                    ))}
                </g>

                <g className="auth-orbital-art">
                    <path d={ORBIT} fill="none" stroke="#8f3e9f"
                        strokeWidth="32" opacity=".16" filter="url(#nox-soft-glow)" />
                    <path d={ORBIT} fill="none" stroke="#762f83"
                        strokeWidth="9" opacity=".27" />
                    <path d={ORBIT} fill="none" stroke="url(#nox-orbit-gold)"
                        strokeWidth="3.3" strokeLinecap="round" />
                    <path d={ORBIT} className="auth-orbit-sheen" fill="none"
                        stroke="#f7ceaa" strokeWidth="3.1"
                        strokeDasharray="12 1250" strokeLinecap="round" />
                    <circle className="auth-orbit-spark" r="3.5" fill="#ffe6b5">
                        <animateMotion path={ORBIT} dur="13s" repeatCount="indefinite" />
                    </circle>
                </g>

                <g className="auth-distant-moon">
                    <circle cx="764" cy="207" r="34" fill="#110819" />
                    <path d="M 763 173 A 34 34 0 0 0 750 236" fill="none"
                        stroke="#a54ab5" strokeWidth="11" opacity=".24" filter="url(#nox-soft-glow)" />
                    <path d="M 763 173 A 34 34 0 0 0 750 236" fill="none"
                        stroke="#f6a5d2" strokeWidth="2.2" />
                </g>

                {/* Independent original illustration, not a crop of the design reference. */}
                <image className="auth-planet-art" href="/auth-planet.avif"
                    x="0" y="600" width="1200" height="560" />
            </svg>
        </div>
    );
}
