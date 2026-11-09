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
                        x1="610" y1="544" x2="904" y2="278">
                        <stop stopColor="#f7a9c7" stopOpacity="0" />
                        <stop offset=".11" stopColor="#f7aeae" stopOpacity=".6" />
                        <stop offset=".23" stopColor="#ffe5b9" />
                        <stop offset=".31" stopColor="#fff3d5" />
                        <stop offset=".43" stopColor="#f7bfa5" stopOpacity=".66" />
                        <stop offset=".62" stopColor="#da77b9" stopOpacity="0" />
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
                        <stop stopColor="#fff2cc" stopOpacity=".8" />
                        <stop offset=".18" stopColor="#ffda99" stopOpacity=".6" />
                        <stop offset=".56" stopColor="#dc8fac" stopOpacity=".2" />
                        <stop offset="1" stopColor="#a45aaa" stopOpacity="0" />
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

                {/* A sculpted ribbon: atmospheric depth and delicately offset purple
                    traces; only the upper sweep catches gold, the lower return fades. */}
                <g className="auth-orbital-art">
                    <ellipse cx="714" cy="425" rx="116" ry="100"
                        fill="url(#nox-orbit-ambient)" opacity=".65" />
                    <path d="M 945 255 C 815 297 686 386 625 478 C 600 520 603 547 614 564"
                        fill="none" stroke="#7d3c95" strokeWidth="1.25"
                        opacity=".25" />
                    <path d="M 951 283 C 831 314 723 396 661 488 C 629 532 623 550 630 567"
                        fill="none" stroke="#a364a7" strokeWidth="1.15"
                        opacity=".23" />
                    <path d={ORBIT_UPPER} fill="none" stroke="#a94d9f"
                        strokeWidth="26" opacity=".22"
                        filter="url(#nox-atmosphere-glow)" />
                    <path d={ORBIT_UPPER} fill="none" stroke="#d68ab5"
                        strokeWidth="11" opacity=".25"
                        filter="url(#nox-soft-glow)" />
                    <path d={ORBIT_UPPER} fill="none" stroke="url(#nox-orbit-color)"
                        strokeWidth="2.55" strokeLinecap="round" />

                    {/* This is stationary illumination along the curve, not a moving dash. */}
                    <path d={ORBIT_UPPER} fill="none" stroke="url(#nox-orbit-flare)"
                        strokeWidth="18" opacity=".45"
                        filter="url(#nox-soft-glow)" />
                    <path d={ORBIT_UPPER} fill="none" stroke="url(#nox-orbit-flare)"
                        strokeWidth="3.2" opacity=".9" strokeLinecap="round" />

                    <path d={ORBIT_RETURN} fill="none"
                        stroke="url(#nox-orbit-return)" strokeWidth="1.05" />
                    <path d="M 619 567 C 647 599 710 600 794 579 C 855 561 909 539 951 520"
                        fill="none" stroke="#7b3f87" strokeWidth=".65"
                        opacity=".085" />

                    {/* A diffuse lens-like glow travels part of the upper arc, then
                        fades away and rests before starting again. No spinning line. */}
                    <g className="auth-orbit-glimmer" opacity="0">
                        <ellipse rx="31" ry="10" fill="url(#nox-moving-gold)"
                            filter="url(#nox-soft-glow)" />
                        <ellipse rx="16" ry="4" fill="url(#nox-moving-gold)"
                            opacity=".75" />
                        <animateMotion path={ORBIT_UPPER} rotate="auto"
                            dur="17s" repeatCount="indefinite" calcMode="linear"
                            keyPoints="0.19;0.19;0.73;0.73;0.19"
                            keyTimes="0;0.18;0.62;0.78;1" />
                        <animate attributeName="opacity" dur="17s"
                            repeatCount="indefinite"
                            values="0;0;.7;.65;0;0"
                            keyTimes="0;.18;.38;.56;.7;1" />
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
