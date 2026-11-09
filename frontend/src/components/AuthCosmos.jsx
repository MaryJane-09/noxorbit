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

// A larger sculpted oval, scaled to the annotated visual reference.
// The card hides both upper and lower ends; the upper edge is luminous
// while the lower return is a barely visible shadow.
const ORBIT_UPPER =
    "M 1010 235 C 835 264 697 357 594 496 C 512 607 488 705 570 749";
const ORBIT_RETURN =
    "M 570 749 C 650 810 836 718 1020 586";

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
                        x1="545" y1="665" x2="952" y2="252">
                        <stop stopColor="#a64fbb" stopOpacity=".19" />
                        <stop offset=".13" stopColor="#b86bb5" stopOpacity=".45" />
                        <stop offset=".31" stopColor="#e49ab9" stopOpacity=".77" />
                        <stop offset=".49" stopColor="#ffe1b4" stopOpacity=".95" />
                        <stop offset=".59" stopColor="#d995b7" stopOpacity=".72" />
                        <stop offset=".78" stopColor="#9860a7" stopOpacity=".37" />
                        <stop offset="1" stopColor="#764081" stopOpacity="0" />
                    </linearGradient>
                    <linearGradient id="nox-orbit-return" gradientUnits="userSpaceOnUse"
                        x1="570" y1="749" x2="1020" y2="586">
                        <stop stopColor="#b269b7" stopOpacity=".28" />
                        <stop offset=".24" stopColor="#8a438d" stopOpacity=".15" />
                        <stop offset=".58" stopColor="#5d286c" stopOpacity=".07" />
                        <stop offset="1" stopColor="#542663" stopOpacity="0" />
                    </linearGradient>
                    <linearGradient id="nox-orbit-flare" gradientUnits="userSpaceOnUse"
                        x1="522" y1="711" x2="935" y2="243">
                        <stop stopColor="#f7a9c7" stopOpacity="0" />
                        <stop offset=".22" stopColor="#f6b3c1" stopOpacity=".23" />
                        <stop offset=".36" stopColor="#ffd4b2" stopOpacity=".62" />
                        <stop offset=".49" stopColor="#fff5d7" stopOpacity=".98" />
                        <stop offset=".59" stopColor="#f5c7ad" stopOpacity=".72" />
                        <stop offset=".72" stopColor="#dd85be" stopOpacity=".16" />
                        <stop offset="1" stopColor="#da77b9" stopOpacity="0" />
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
                    <radialGradient id="nox-moving-gold">
                        <stop stopColor="#fff8df" stopOpacity="1" />
                        <stop offset=".18" stopColor="#ffe4a9" stopOpacity=".93" />
                        <stop offset=".46" stopColor="#ecb383" stopOpacity=".57" />
                        <stop offset=".78" stopColor="#d77bb9" stopOpacity=".19" />
                        <stop offset="1" stopColor="#a45aaa" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-flare-halo">
                        <stop stopColor="#fff7d8" stopOpacity=".9" />
                        <stop offset=".17" stopColor="#f9d69b" stopOpacity=".7" />
                        <stop offset=".46" stopColor="#d692ae" stopOpacity=".33" />
                        <stop offset=".78" stopColor="#a858ab" stopOpacity=".12" />
                        <stop offset="1" stopColor="#8f488f" stopOpacity="0" />
                    </radialGradient>
                    <radialGradient id="nox-orbit-ambient">
                        <stop stopColor="#e19ab5" stopOpacity=".27" />
                        <stop offset=".42" stopColor="#a451a4" stopOpacity=".14" />
                        <stop offset="1" stopColor="#6a316f" stopOpacity="0" />
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

                {/* A broad orbital ribbon following the user's annotated silhouette.
                    Layered traces give depth, while a faint lower return disappears
                    behind the card. Only a localized bloom moves, never the ring. */}
                <g className="auth-orbital-art">
                    <ellipse cx="671" cy="520" rx="203" ry="224"
                        fill="url(#nox-orbit-ambient)" opacity=".72" />
                    <path d="M 1022 221 C 825 254 691 340 580 487 C 495 605 468 716 550 758"
                        fill="none" stroke="#7d3c95" strokeWidth="1.5"
                        opacity=".26" />
                    <path d="M 1010 255 C 843 289 710 370 614 508 C 539 616 521 700 592 743"
                        fill="none" stroke="#b16ab3" strokeWidth="1.15"
                        opacity=".28" />
                    <path d={ORBIT_UPPER} fill="none" stroke="#a951a6"
                        strokeWidth="33" opacity=".24"
                        filter="url(#nox-atmosphere-glow)" />
                    <path d={ORBIT_UPPER} fill="none" stroke="#d487b9"
                        strokeWidth="14" opacity=".3"
                        filter="url(#nox-soft-glow)" />
                    <path d={ORBIT_UPPER} fill="none" stroke="url(#nox-orbit-color)"
                        strokeWidth="2.8" strokeLinecap="round" />
                    <path d={ORBIT_UPPER} fill="none" stroke="url(#nox-orbit-flare)"
                        strokeWidth="19" opacity=".52" filter="url(#nox-soft-glow)" />
                    <path d={ORBIT_UPPER} fill="none" stroke="url(#nox-orbit-flare)"
                        strokeWidth="3.3" strokeLinecap="round" />

                    <path d={ORBIT_RETURN} fill="none"
                        stroke="url(#nox-orbit-return)" strokeWidth="1.15" />
                    <path d="M 571 758 C 651 821 836 734 1024 590"
                        fill="none" stroke="#87408d" strokeWidth="1"
                        opacity=".095" />

                    {/* A persistent champagne flare, plus a slow, diffuse shimmer.
                        The glow never turns into a spinning outline. */}
                    <g className="auth-orbit-beacon">
                        <ellipse cx="650" cy="430" rx="97" ry="39"
                            transform="rotate(-50 650 430)"
                            fill="url(#nox-flare-halo)" opacity=".83" />
                        <ellipse cx="650" cy="430" rx="49" ry="16"
                            transform="rotate(-50 650 430)"
                            fill="url(#nox-moving-gold)" opacity=".92" />
                        <circle cx="650" cy="430" r="3.8" fill="#fff0cb"
                            opacity=".68" filter="url(#nox-soft-glow)" />
                        <circle cx="650" cy="430" r="1.3" fill="#fff6df"
                            opacity=".95" />
                    </g>
                    <g className="auth-orbit-glimmer">
                        <ellipse rx="39" ry="13" transform="rotate(-48)"
                            fill="url(#nox-moving-gold)" opacity=".82" />
                        <animateMotion
                            path="M 775 324 C 735 352 696 384 660 418 C 626 456 594 496 571 529"
                            dur="14s" repeatCount="indefinite" calcMode="linear" />
                        <animate attributeName="opacity" dur="14s"
                            repeatCount="indefinite"
                            values="0;.72;.84;.12;0"
                            keyTimes="0;.26;.52;.79;1" />
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
