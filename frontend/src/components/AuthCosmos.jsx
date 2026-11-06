import "./AuthCosmos.css";

// Stable star positions avoid visual jumps between React renders.
const noise = (n) => {
    const value = Math.sin(n * 127.1 + 71.7) * 43758.5453123;
    return value - Math.floor(value);
};

const STARS = Array.from({ length: 64 }, (_, i) => ({
    x: Math.round(noise(i + 11) * 1672),
    y: Math.round(noise(i + 119) * 920),
    radius: 0.35 + noise(i + 237) * 0.85,
    opacity: 0.1 + noise(i + 351) * 0.27,
}));

// Both ends disappear behind the form: the orbit never stops in open space.
const ORBIT =
    "M 990 284 C 837 313 708 399 632 513 C 569 612 646 623 765 563 C 869 521 955 494 1054 485";

export default function AuthCosmos() {
    return (
        <div className="auth-cosmos" aria-hidden="true">
            <svg className="auth-cosmos-svg" viewBox="0 0 1672 940"
                preserveAspectRatio="xMidYMid slice" focusable="false">
                <defs>
                    <radialGradient id="nox-nebula">
                        <stop stopColor="#692575" stopOpacity=".53" />
                        <stop offset=".48" stopColor="#3c174b" stopOpacity=".23" />
                        <stop offset="1" stopColor="#110819" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-twilight">
                        <stop stopColor="#f0a5e8" stopOpacity=".22" />
                        <stop offset="1" stopColor="#f0a5e8" stopOpacity="0" />
                    </radialGradient>
                    <linearGradient id="nox-orbit-color" gradientUnits="userSpaceOnUse"
                        x1="628" y1="526" x2="900" y2="285">
                        <stop stopColor="#a64aab" stopOpacity=".1" />
                        <stop offset=".28" stopColor="#b65bba" stopOpacity=".38" />
                        <stop offset=".59" stopColor="#efb8ae" stopOpacity=".8" />
                        <stop offset=".76" stopColor="#f4d49e" stopOpacity=".63" />
                        <stop offset="1" stopColor="#94519a" stopOpacity=".04" />
                    </linearGradient>
                    <radialGradient id="nox-moon-disc" cx="35%" cy="27%" r="82%">
                        <stop stopColor="#27132f" />
                        <stop offset=".58" stopColor="#140b20" />
                        <stop offset="1" stopColor="#0b0715" />
                    </radialGradient>
                    <linearGradient id="nox-moon-crescent" x1="0" y1="1" x2="1" y2="0">
                        <stop stopColor="#ad5baf" stopOpacity=".35" />
                        <stop offset=".5" stopColor="#f4a5d2" />
                        <stop offset="1" stopColor="#e6b3d7" stopOpacity=".65" />
                    </linearGradient>
                    <filter id="nox-soft-glow" x="-100%" y="-100%" width="300%" height="300%">
                        <feGaussianBlur stdDeviation="5" />
                    </filter>
                    <filter id="nox-atmosphere-glow" x="-100%" y="-100%" width="300%" height="300%">
                        <feGaussianBlur stdDeviation="13" />
                    </filter>
                </defs>

                <ellipse cx="140" cy="610" rx="780" ry="720" fill="url(#nox-nebula)" />
                <ellipse cx="1420" cy="220" rx="530" ry="450"
                    fill="url(#nox-nebula)" opacity=".2" />
                <ellipse cx="250" cy="811" rx="880" ry="240"
                    fill="url(#nox-twilight)" opacity=".5" />

                <g className="auth-stars">
                    {STARS.map((star, i) => (
                        <circle key={i} cx={star.x} cy={star.y} r={star.radius}
                            fill="#f9d7f7" opacity={star.opacity}
                            className={i % 13 === 0 ? "auth-star-pulse" : undefined} />
                    ))}
                </g>

                {/* Independent animated orbital light; never part of a static screenshot. */}
                <g className="auth-orbital-art">
                    <path d={ORBIT} fill="none" stroke="#a44b9f"
                        strokeWidth="12" opacity=".09" filter="url(#nox-soft-glow)" />
                    <path d={ORBIT} fill="none" stroke="url(#nox-orbit-color)"
                        strokeWidth="2.4" strokeLinecap="round" />
                    <path d={ORBIT} className="auth-orbit-sheen" fill="none"
                        stroke="#ffdaa8" strokeWidth="2.3" strokeLinecap="round"
                        pathLength="1000" strokeDasharray="24 976" />
                </g>

                <g className="auth-distant-moon">
                    <circle cx="764" cy="207" r="40" fill="#a84aab"
                        opacity=".12" filter="url(#nox-atmosphere-glow)" />
                    <circle cx="764" cy="207" r="31" fill="url(#nox-moon-disc)" />
                    <path d="M 760 177 C 737 183 726 214 749 236" fill="none"
                        stroke="#ad5baf" strokeWidth="7" opacity=".27"
                        filter="url(#nox-soft-glow)" />
                    <path d="M 760 177 C 737 183 726 214 749 236" fill="none"
                        stroke="url(#nox-moon-crescent)" strokeWidth="2" strokeLinecap="round" />
                </g>

                {/* The planet is a standalone illustration, not a crop of the reference UI. */}
                <image className="auth-planet-art" href="/auth-planet-horizon.webp"
                    x="-25" y="690" width="1090" height="320"
                    preserveAspectRatio="xMinYMin meet" />
            </svg>
        </div>
    );
}
