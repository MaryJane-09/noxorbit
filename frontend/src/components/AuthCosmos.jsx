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

// One continuous orbital sweep. Both endpoints sit behind the form,
// leaving the illuminated curve visible without an abrupt cut in open space.
const ORBIT =
    "M 1040 274 C 862 301 724 378 638 493 C 566 592 604 626 732 574 C 863 520 962 488 1084 475";

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
                        x1="605" y1="562" x2="857" y2="300">
                        <stop stopColor="#a450b5" stopOpacity=".28" />
                        <stop offset=".23" stopColor="#c37bbb" stopOpacity=".62" />
                        <stop offset=".48" stopColor="#eab49e" stopOpacity=".91" />
                        <stop offset=".62" stopColor="#ffe3b6" stopOpacity=".98" />
                        <stop offset=".79" stopColor="#c58cbb" stopOpacity=".67" />
                        <stop offset="1" stopColor="#a251af" stopOpacity=".16" />
                    </linearGradient>
                    <radialGradient id="nox-moon-disc" cx="26%" cy="38%" r="90%">
                        <stop stopColor="#4b2858" />
                        <stop offset=".26" stopColor="#2b1737" />
                        <stop offset=".68" stopColor="#140b22" />
                        <stop offset="1" stopColor="#090713" />
                    </radialGradient>
                    <linearGradient id="nox-moon-crescent" gradientUnits="userSpaceOnUse"
                        x1="735" y1="211" x2="763" y2="185">
                        <stop stopColor="#ad5baf" stopOpacity=".46" />
                        <stop offset=".48" stopColor="#f4a5d2" />
                        <stop offset="1" stopColor="#f2d3df" stopOpacity=".68" />
                    </linearGradient>
                    <clipPath id="nox-moon-clip">
                        <circle cx="764" cy="207" r="31" />
                    </clipPath>
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

                {/* A continuous, low-contrast orbit with one travelling warm reflection. */}
                <g className="auth-orbital-art">
                    <path d={ORBIT} fill="none" stroke="#ae5bba"
                        strokeWidth="12" opacity=".16" filter="url(#nox-soft-glow)" />
                    <path d={ORBIT} fill="none" stroke="#a85fb2"
                        strokeWidth="1.25" opacity=".38" strokeLinecap="round" />
                    <path d={ORBIT} fill="none" stroke="url(#nox-orbit-color)"
                        strokeWidth="2.65" strokeLinecap="round" />
                    <path d={ORBIT} className="auth-orbit-sheen auth-orbit-aura"
                        fill="none" stroke="#ffd7a3" strokeWidth="8"
                        strokeLinecap="round" filter="url(#nox-soft-glow)"
                        pathLength="1000" strokeDasharray="52 948" />
                    <path d={ORBIT} className="auth-orbit-sheen"
                        fill="none" stroke="#fff0c7" strokeWidth="2.5"
                        strokeLinecap="round" pathLength="1000"
                        strokeDasharray="52 948" />
                </g>

                <g className="auth-distant-moon">
                    <circle className="auth-moon-halo" cx="764" cy="207" r="43"
                        fill="#b763bf" opacity=".19" filter="url(#nox-atmosphere-glow)" />
                    <g clipPath="url(#nox-moon-clip)">
                        <circle cx="764" cy="207" r="31" fill="url(#nox-moon-disc)" />
                        <path d="M 756 179 C 735 192 730 216 751 235
                                 C 739 218 742 193 756 179 Z"
                            fill="url(#nox-moon-crescent)" opacity=".58" />
                        <ellipse cx="757" cy="201" rx="8" ry="3" fill="#865d89"
                            opacity=".12" transform="rotate(-24 757 201)" />
                        <ellipse cx="770" cy="220" rx="11" ry="4" fill="#76547d"
                            opacity=".1" transform="rotate(18 770 220)" />
                        <path d="M 744 220 Q 753 211 762 215" fill="none"
                            stroke="#a277a7" strokeWidth=".7" opacity=".18" />
                    </g>
                    <path d="M 758 177 C 735 185 729 215 750 237" fill="none"
                        stroke="#c96bca" strokeWidth="7" opacity=".25"
                        filter="url(#nox-soft-glow)" strokeLinecap="round" />
                    <path d="M 758 177 C 735 185 729 215 750 237" fill="none"
                        stroke="url(#nox-moon-crescent)" strokeWidth="1.8"
                        strokeLinecap="round" />
                </g>

                {/* The planet is a standalone illustration, not a crop of the reference UI. */}
                <image className="auth-planet-art" href="/auth-planet-horizon.webp"
                    x="-25" y="690" width="1090" height="320"
                    preserveAspectRatio="xMinYMin meet" />
            </svg>
        </div>
    );
}
