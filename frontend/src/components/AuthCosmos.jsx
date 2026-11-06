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

// Reference geometry: a lit upper sweep and an almost invisible return.
// The orbit should read as a crescent of light, never as a uniformly glowing oval.
const ORBIT_UPPER =
    "M 940 270 C 829 302 706 387 642 477 C 616 514 604 543 613 558";
const ORBIT_RETURN =
    "M 613 558 C 632 599 695 612 775 588 C 833 571 886 552 945 529";
const ORBIT_FLARE =
    "M 633 519 C 655 477 681 434 713 404";

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
                        x1="610" y1="544" x2="904" y2="278">
                        <stop stopColor="#a850ae" stopOpacity=".44" />
                        <stop offset=".16" stopColor="#f1a2bd" stopOpacity=".84" />
                        <stop offset=".26" stopColor="#ffe9c2" />
                        <stop offset=".34" stopColor="#ffd5ae" />
                        <stop offset=".47" stopColor="#d778bb" stopOpacity=".62" />
                        <stop offset=".72" stopColor="#854090" stopOpacity=".34" />
                        <stop offset="1" stopColor="#764081" stopOpacity="0" />
                    </linearGradient>
                    <linearGradient id="nox-orbit-return" gradientUnits="userSpaceOnUse"
                        x1="614" y1="557" x2="940" y2="529">
                        <stop stopColor="#b34dba" stopOpacity=".36" />
                        <stop offset=".16" stopColor="#74327d" stopOpacity=".19" />
                        <stop offset=".5" stopColor="#572767" stopOpacity=".11" />
                        <stop offset="1" stopColor="#542663" stopOpacity="0" />
                    </linearGradient>
                    <linearGradient id="nox-orbit-flare" gradientUnits="userSpaceOnUse"
                        x1="630" y1="521" x2="718" y2="399">
                        <stop stopColor="#f7a9c7" stopOpacity="0" />
                        <stop offset=".22" stopColor="#ffd9b0" />
                        <stop offset=".47" stopColor="#fff2c9" />
                        <stop offset=".75" stopColor="#f5b69e" stopOpacity=".66" />
                        <stop offset="1" stopColor="#f2a4be" stopOpacity="0" />
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

                {/* Layered orbital sweep: purple ghost trails, a golden focal flare,
                    and a very dark return into the card — matching the reference. */}
                <g className="auth-orbital-art">
                    <path d="M 940 264 C 820 301 698 384 636 473 C 604 520 602 544 614 564"
                        fill="none" stroke="#854094" opacity=".23" strokeWidth="1.1" />
                    <path d="M 943 280 C 826 316 717 396 655 484 C 622 532 613 548 623 574"
                        fill="none" stroke="#9c448f" opacity=".24" strokeWidth=".95" />
                    <path d={ORBIT_UPPER} fill="none" stroke="#ab599e"
                        strokeWidth="9" opacity=".12" filter="url(#nox-soft-glow)" />
                    <path d={ORBIT_UPPER} fill="none" stroke="url(#nox-orbit-color)"
                        strokeWidth="2.2" strokeLinecap="round" />

                    <path d={ORBIT_FLARE} fill="none" stroke="#f5b8ae"
                        strokeWidth="11" opacity=".45" filter="url(#nox-soft-glow)" />
                    <path d={ORBIT_FLARE} fill="none" stroke="url(#nox-orbit-flare)"
                        strokeWidth="3.5" strokeLinecap="round" />
                    <circle cx="667" cy="458" r="3.1" fill="#ffefcb"
                        opacity=".9" filter="url(#nox-soft-glow)" />

                    <path d={ORBIT_RETURN} fill="none" stroke="url(#nox-orbit-return)"
                        strokeWidth="1.15" strokeLinecap="round" />
                    <path d="M 618 563 C 646 606 710 607 800 579 C 854 562 909 540 952 521"
                        fill="none" stroke="#783784" opacity=".1" strokeWidth=".8" />

                    <path d={ORBIT_UPPER} className="auth-orbit-sheen auth-orbit-aura"
                        fill="none" stroke="#ffe4bb" strokeWidth="7"
                        strokeLinecap="round" filter="url(#nox-soft-glow)"
                        pathLength="1000" strokeDasharray="34 966" />
                    <path d={ORBIT_UPPER} className="auth-orbit-sheen"
                        fill="none" stroke="#ffefd5" strokeWidth="1.8"
                        strokeLinecap="round" pathLength="1000"
                        strokeDasharray="34 966" />
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
