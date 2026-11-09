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

// Small, scattered light points support the nebula without a central star.
const NEBULA_DUST = Array.from({ length: 23 }, (_, i) => ({
    x: Math.round(628 + noise(i * 7 + 401) * 310),
    y: Math.round(353 + noise(i * 7 + 529) * 238),
    radius: 0.35 + noise(i * 7 + 607) * 0.72,
    opacity: 0.22 + noise(i * 7 + 719) * 0.38,
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
                    {/* Soft overlapping gas clouds, with no drawn ring or focal star. */}
                    <radialGradient id="nox-veil-violet">
                        <stop stopColor="#aa59b6" stopOpacity=".32" />
                        <stop offset=".26" stopColor="#87459a" stopOpacity=".26" />
                        <stop offset=".59" stopColor="#532265" stopOpacity=".12" />
                        <stop offset="1" stopColor="#381744" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-veil-plum">
                        <stop stopColor="#8f408e" stopOpacity=".28" />
                        <stop offset=".38" stopColor="#743176" stopOpacity=".17" />
                        <stop offset=".78" stopColor="#451b52" stopOpacity=".04" />
                        <stop offset="1" stopColor="#2d1435" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-veil-rose">
                        <stop stopColor="#d18cae" stopOpacity=".21" />
                        <stop offset=".24" stopColor="#b573a8" stopOpacity=".16" />
                        <stop offset=".62" stopColor="#803c83" stopOpacity=".05" />
                        <stop offset="1" stopColor="#4e254e" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-veil-champagne">
                        <stop stopColor="#f3cbaf" stopOpacity=".18" />
                        <stop offset=".18" stopColor="#d6a5aa" stopOpacity=".16" />
                        <stop offset=".5" stopColor="#a064a2" stopOpacity=".065" />
                        <stop offset="1" stopColor="#763f81" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-veil-dust">
                        <stop stopColor="#fff0d9" stopOpacity=".53" />
                        <stop offset=".16" stopColor="#e7bda3" stopOpacity=".3" />
                        <stop offset=".65" stopColor="#a66da3" stopOpacity=".065" />
                        <stop offset="1" stopColor="#803c80" stopOpacity="0" />
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

                {/* A broad, layered atmosphere in the negative space.
                    No isolated star, ring, outline or obvious repeating shape. */}
                <g className="auth-celestial-field">
                    <g className="auth-nebula-back">
                        <ellipse cx="774" cy="478" rx="250" ry="132"
                            transform="rotate(-24 774 478)"
                            fill="url(#nox-veil-violet)" opacity=".76" />
                        <ellipse cx="842" cy="419" rx="185" ry="83"
                            transform="rotate(-34 842 419)"
                            fill="url(#nox-veil-plum)" opacity=".72" />
                        <ellipse cx="700" cy="567" rx="170" ry="95"
                            transform="rotate(-24 700 567)"
                            fill="url(#nox-veil-plum)" opacity=".62" />
                    </g>
                    <g className="auth-nebula-front">
                        <ellipse cx="795" cy="474" rx="160" ry="67"
                            transform="rotate(-29 795 474)"
                            fill="url(#nox-veil-rose)" opacity=".73" />
                        <ellipse cx="738" cy="526" rx="152" ry="63"
                            transform="rotate(-28 738 526)"
                            fill="url(#nox-veil-violet)" opacity=".46" />
                        <ellipse cx="803" cy="471" rx="116" ry="47"
                            transform="rotate(-27 803 471)"
                            fill="url(#nox-veil-champagne)" opacity=".7" />
                    </g>

                    {/* Scattered champagne light, never a single oversized star. */}
                    <g className="auth-nebula-sparks">
                        <circle cx="692" cy="432" r="22"
                            fill="url(#nox-veil-dust)" opacity=".55" />
                        <circle cx="842" cy="493" r="19"
                            fill="url(#nox-veil-dust)" opacity=".5" />
                        <circle cx="754" cy="557" r="15"
                            fill="url(#nox-veil-dust)" opacity=".37" />
                        {NEBULA_DUST.map((particle, i) => (
                            <circle key={i} cx={particle.x} cy={particle.y}
                                r={particle.radius}
                                fill={i % 3 === 0 ? "#e6b7be" : "#e4c8d8"}
                                opacity={particle.opacity} />
                        ))}
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
