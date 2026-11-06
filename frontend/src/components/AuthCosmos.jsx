import "./AuthCosmos.css";

// Deterministic detail makes the scene feel textured without external image assets.
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

const LIGHT_CLUSTERS = [
    [110, 826, 102, 35], [248, 855, 120, 41], [398, 891, 115, 49],
    [548, 943, 125, 51], [664, 998, 95, 36],
];

const CITY_LIGHTS = Array.from({ length: 460 }, (_, i) => {
    const cluster = LIGHT_CLUSTERS[i % LIGHT_CLUSTERS.length];
    return {
        x: cluster[0] + (noise(i * 3 + 7) - 0.5) * cluster[2] * 2,
        y: cluster[1] + (noise(i * 3 + 19) - 0.5) * cluster[3] * 2,
        r: 0.3 + noise(i * 3 + 53) * 1.1,
        opacity: 0.2 + noise(i * 3 + 79) * 0.65,
    };
});

const PLANET =
    "M -260 818 C 32 675 414 705 763 900 C 1070 1072 1220 1420 1280 1740 L -260 1740 Z";
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
                    <radialGradient id="nox-atmosphere" cx=".42" cy=".1" r=".85">
                        <stop stopColor="#8b338e" stopOpacity=".76" />
                        <stop offset=".35" stopColor="#451c57" stopOpacity=".69" />
                        <stop offset="1" stopColor="#0d0815" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-planet-surface" cx=".46" cy=".02" r=".96">
                        <stop stopColor="#723875" />
                        <stop offset=".3" stopColor="#32153e" />
                        <stop offset=".73" stopColor="#150b23" />
                        <stop offset="1" stopColor="#090711" />
                    </radialGradient>
                    <linearGradient id="nox-horizon-light" x1="0" y1="0" x2="1" y2=".15">
                        <stop stopColor="#c270dd" stopOpacity=".5" />
                        <stop offset=".19" stopColor="#f3a2f5" />
                        <stop offset=".5" stopColor="#ffd6ec" />
                        <stop offset=".74" stopColor="#dc88da" />
                        <stop offset="1" stopColor="#6e327b" stopOpacity="0" />
                    </linearGradient>
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
                    <clipPath id="nox-planet-clip"><path d={PLANET} /></clipPath>
                    <filter id="nox-wide-glow" x="-100%" y="-100%" width="300%" height="300%">
                        <feGaussianBlur stdDeviation="19" />
                    </filter>
                    <filter id="nox-soft-glow" x="-100%" y="-100%" width="300%" height="300%">
                        <feGaussianBlur stdDeviation="5" />
                    </filter>
                    <filter id="nox-surface-noise" x="0" y="0" width="100%" height="100%">
                        <feTurbulence type="fractalNoise" baseFrequency=".014 .027"
                            numOctaves="4" seed="17" stitchTiles="stitch" />
                        <feColorMatrix type="saturate" values=".25" />
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

                <g className="auth-planet">
                    <path d={PLANET} fill="url(#nox-planet-surface)" />
                    <g clipPath="url(#nox-planet-clip)">
                        <rect x="-100" y="680" width="1250" height="720"
                            filter="url(#nox-surface-noise)" opacity=".24"
                            style={{ mixBlendMode: "screen" }} />
                        <g fill="#9e598a" opacity=".23">
                            <path d="M-40 818 26 796 82 802 125 784 178 806 242 793 297 823 278 846 221 826 180 845 130 837 86 857 31 843Z" />
                            <path d="M218 858 284 821 345 842 386 831 456 862 474 902 409 881 354 902 287 883Z" />
                            <path d="M420 913 473 888 522 911 586 927 634 972 592 1002 546 976 489 974Z" />
                            <path d="M60 918 109 891 168 907 201 942 183 974 119 954Z" />
                        </g>
                        <g fill="#ed9d69" opacity=".26" filter="url(#nox-soft-glow)">
                            <ellipse cx="175" cy="827" rx="73" ry="14" />
                            <ellipse cx="390" cy="897" rx="111" ry="20" />
                            <ellipse cx="555" cy="942" rx="95" ry="20" />
                        </g>
                        {CITY_LIGHTS.map((point, i) => (
                            <circle key={i} cx={point.x} cy={point.y} r={point.r}
                                fill={i % 4 === 0 ? "#fff3bd" : "#f4a661"}
                                opacity={point.opacity} />
                        ))}
                    </g>
                    <path d={PLANET} fill="none" stroke="#e16de3"
                        strokeWidth="34" opacity=".35" filter="url(#nox-wide-glow)" />
                    <path d={PLANET} fill="none" stroke="url(#nox-horizon-light)"
                        strokeWidth="4.5" />
                    <path d={PLANET} fill="none" stroke="#f9ddf6"
                        strokeWidth="1.1" opacity=".6" />
                </g>
            </svg>
        </div>
    );
}
