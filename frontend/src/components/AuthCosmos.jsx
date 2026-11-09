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
                    {/* Atmospheric light field: no rings, outlines or traced arcs. */}
                    <radialGradient id="nox-celestial-violet">
                        <stop stopColor="#a765be" stopOpacity=".34" />
                        <stop offset=".28" stopColor="#7f3b97" stopOpacity=".26" />
                        <stop offset=".65" stopColor="#4b1c62" stopOpacity=".11" />
                        <stop offset="1" stopColor="#2a1038" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-celestial-amber">
                        <stop stopColor="#ffe6ac" stopOpacity=".58" />
                        <stop offset=".16" stopColor="#e7ba83" stopOpacity=".4" />
                        <stop offset=".38" stopColor="#ba789d" stopOpacity=".18" />
                        <stop offset=".75" stopColor="#8b4b9c" stopOpacity=".065" />
                        <stop offset="1" stopColor="#693979" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-celestial-spark">
                        <stop stopColor="#fff8e5" stopOpacity=".85" />
                        <stop offset=".17" stopColor="#ffe3b8" stopOpacity=".64" />
                        <stop offset=".4" stopColor="#e7a3b4" stopOpacity=".25" />
                        <stop offset="1" stopColor="#a65ca7" stopOpacity="0" />
                    </radialGradient>
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

                {/* Quiet celestial atmosphere in the negative space between copy
                    and card. No orbit, no continuous line or artificial loop. */}
                <g className="auth-celestial-field">
                    <g className="auth-celestial-mist">
                        <ellipse cx="759" cy="470" rx="219" ry="112"
                            transform="rotate(-27 759 470)"
                            fill="url(#nox-celestial-violet)" />
                        <ellipse cx="816" cy="428" rx="149" ry="70"
                            transform="rotate(-32 816 428)"
                            fill="url(#nox-celestial-violet)" opacity=".58" />
                        <ellipse cx="766" cy="475" rx="111" ry="84"
                            fill="url(#nox-celestial-amber)" opacity=".7" />
                    </g>

                    {/* A single warm focal star with diffuse depth, not a bright
                        streak crossing the description or registration card. */}
                    <g className="auth-celestial-focal">
                        <circle cx="774" cy="461" r="73"
                            fill="url(#nox-celestial-spark)" opacity=".52" />
                        <circle cx="774" cy="461" r="29"
                            fill="url(#nox-celestial-spark)" opacity=".67" />
                        <path d="M 774 451 Q 775 459 784 461 Q 775 463 774 471 Q 773 463 764 461 Q 773 459 774 451 Z"
                            fill="#ffe9c4" opacity=".8" />
                        <circle cx="774" cy="461" r="1.25" fill="#fff8e2" />
                    </g>

                    {/* Small related points give the space texture without
                        competing with the form or turning into a constellation line. */}
                    <g className="auth-celestial-dust">
                        <circle cx="695" cy="396" r="1" fill="#f8d9d0" opacity=".55" />
                        <circle cx="732" cy="372" r=".8" fill="#dfbadb" opacity=".45" />
                        <circle cx="838" cy="409" r="1.1" fill="#f7d4b8" opacity=".63" />
                        <circle cx="860" cy="482" r=".8" fill="#ffe7c7" opacity=".58" />
                        <circle cx="687" cy="533" r=".85" fill="#eec1d7" opacity=".47" />
                        <circle cx="814" cy="546" r="1" fill="#f1d2c4" opacity=".52" />
                        <circle cx="767" cy="572" r=".7" fill="#ddb6d9" opacity=".4" />
                        <circle cx="643" cy="459" r=".7" fill="#cc9ccc" opacity=".38" />
                    </g>
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
