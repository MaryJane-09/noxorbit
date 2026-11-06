import "./AuthCosmos.css";

// Reproducible scene details; never re-randomize during React renders.
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

// City lights appear in a few settlements instead of covering the planet in glitter.
const CITIES = [
    [130, 176, 65, 32], [274, 203, 62, 35],
    [412, 249, 56, 38], [556, 308, 49, 34],
];
const CITY_LIGHTS = Array.from({ length: 72 }, (_, i) => {
    const area = CITIES[i % CITIES.length];
    return {
        x: area[0] + (noise(i * 3 + 3) - 0.5) * area[2] * 2,
        y: area[1] + (noise(i * 3 + 17) - 0.5) * area[3] * 2,
        r: 0.34 + noise(i * 3 + 39) * 0.75,
        opacity: 0.16 + noise(i * 3 + 63) * 0.43,
    };
});

const ORBIT = "M 918 284 C 782 326 679 407 623 510 C 592 567 608 595 648 597 C 692 599 730 580 766 554";
const HORIZON = "M -240 130 C 63 -31 407 -6 704 157 C 950 292 1117 471 1252 720";
const PLANET = HORIZON + " L 1252 760 L -240 760 Z";

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
                        <stop stopColor="#f0a5e8" stopOpacity=".24" />
                        <stop offset="1" stopColor="#f0a5e8" stopOpacity="0" />
                    </radialGradient>
                    <linearGradient id="nox-orbit-color" gradientUnits="userSpaceOnUse"
                        x1="625" y1="540" x2="880" y2="290">
                        <stop stopColor="#a64aab" stopOpacity=".07" />
                        <stop offset=".28" stopColor="#b65bba" stopOpacity=".35" />
                        <stop offset=".59" stopColor="#efb8ae" stopOpacity=".74" />
                        <stop offset=".76" stopColor="#f4d49e" stopOpacity=".6" />
                        <stop offset="1" stopColor="#94519a" stopOpacity=".07" />
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
                    <radialGradient id="nox-surface" cx=".47" cy=".06" r=".96">
                        <stop stopColor="#61336b" />
                        <stop offset=".28" stopColor="#2d183d" />
                        <stop offset=".68" stopColor="#170e29" />
                        <stop offset="1" stopColor="#070711" />
                    </radialGradient>
                    <linearGradient id="nox-continent" x1="0" y1="0" x2=".82" y2="1">
                        <stop stopColor="#4b2945" />
                        <stop offset=".5" stopColor="#281928" />
                        <stop offset="1" stopColor="#120d20" />
                    </linearGradient>
                    <linearGradient id="nox-horizon-rim" gradientUnits="userSpaceOnUse"
                        x1="0" y1="0" x2="1000" y2="330">
                        <stop stopColor="#d67be3" stopOpacity=".58" />
                        <stop offset=".22" stopColor="#f4a0ea" />
                        <stop offset=".47" stopColor="#ffe0ed" />
                        <stop offset=".73" stopColor="#db89dd" />
                        <stop offset="1" stopColor="#713d98" stopOpacity=".03" />
                    </linearGradient>
                    <radialGradient id="nox-city-haze">
                        <stop stopColor="#e2a06c" stopOpacity=".28" />
                        <stop offset=".5" stopColor="#b26977" stopOpacity=".08" />
                        <stop offset="1" stopColor="#c07783" stopOpacity="0" />
                    </radialGradient>
                    <filter id="nox-soft-glow" x="-100%" y="-100%" width="300%" height="300%">
                        <feGaussianBlur stdDeviation="5" />
                    </filter>
                    <filter id="nox-atmosphere-glow" x="-100%" y="-100%" width="300%" height="300%">
                        <feGaussianBlur stdDeviation="13" />
                    </filter>
                    <filter id="nox-terrain-grain" x="-5%" y="-5%" width="110%" height="110%">
                        <feTurbulence type="fractalNoise" baseFrequency=".009 .021"
                            numOctaves="3" seed="19" />
                        <feColorMatrix type="saturate" values=".18" />
                    </filter>
                    <clipPath id="nox-planet-clip"><path d={PLANET} /></clipPath>
                </defs>

                <ellipse cx="140" cy="610" rx="780" ry="720" fill="url(#nox-nebula)" />
                <ellipse cx="1420" cy="220" rx="530" ry="450" fill="url(#nox-nebula)" opacity=".2" />
                <ellipse cx="250" cy="811" rx="880" ry="240" fill="url(#nox-twilight)" opacity=".64" />

                <g className="auth-stars">
                    {STARS.map((star, i) => (
                        <circle key={i} cx={star.x} cy={star.y} r={star.radius}
                            fill="#f9d7f7" opacity={star.opacity}
                            className={i % 13 === 0 ? "auth-star-pulse" : undefined} />
                    ))}
                </g>

                <g className="auth-orbital-art">
                    <path d={ORBIT} fill="none" stroke="#a44b9f"
                        strokeWidth="12" opacity=".1" filter="url(#nox-soft-glow)" />
                    <path d={ORBIT} fill="none" stroke="url(#nox-orbit-color)"
                        strokeWidth="2.4" strokeLinecap="round" />
                    <path d={ORBIT} className="auth-orbit-sheen" fill="none"
                        stroke="#ffdaa8" strokeWidth="2.5" strokeLinecap="round"
                        pathLength="1000" strokeDasharray="24 976" />
                </g>

                <g className="auth-distant-moon">
                    <circle cx="764" cy="207" r="41" fill="#a84aab" opacity=".16"
                        filter="url(#nox-atmosphere-glow)" />
                    <circle cx="764" cy="207" r="31" fill="url(#nox-moon-disc)" />
                    <path d="M 760 177 C 737 183 726 214 749 236" fill="none"
                        stroke="#ad5baf" strokeWidth="7" opacity=".3"
                        filter="url(#nox-soft-glow)" />
                    <path d="M 760 177 C 737 183 726 214 749 236" fill="none"
                        stroke="url(#nox-moon-crescent)" strokeWidth="2" strokeLinecap="round" />
                </g>

                {/* Original vector planet with shaped terrain, soft cloud detail and sparse settlement lights. */}
                <g className="auth-planet" transform="translate(0 600)">
                    <path d={PLANET} fill="url(#nox-surface)" />
                    <g clipPath="url(#nox-planet-clip)">
                        <g fill="url(#nox-continent)" opacity=".8">
                            <path d="M -85 165 C -24 115 66 125 141 143 C 194 134 227 148 267 180 C 303 205 285 226 254 225 C 205 204 187 238 140 229 C 88 213 61 245 18 220 C -24 218 -77 251 -110 222 Z" />
                            <path d="M 201 233 C 258 206 329 225 376 257 C 431 253 469 271 493 315 C 475 348 423 334 390 352 C 339 348 310 326 279 322 C 254 294 225 289 201 233 Z" />
                            <path d="M 452 302 C 527 284 590 319 650 361 C 703 376 739 417 760 463 C 705 478 650 441 614 426 C 553 416 537 399 502 380 C 464 361 449 336 452 302 Z" />
                        </g>
                        <g className="auth-planet-clouds">
                            <rect x="-10" y="70" width="1220" height="570"
                                filter="url(#nox-terrain-grain)" opacity=".12" />
                            <path d="M -100 167 C 100 80 245 118 402 171 C 520 207 598 241 710 311"
                                fill="none" stroke="#b66aae" strokeWidth="16"
                                opacity=".09" filter="url(#nox-atmosphere-glow)" />
                        </g>
                        {CITIES.map((area, i) => (
                            <ellipse key={i} cx={area[0]} cy={area[1]}
                                rx={area[2] * 2} ry={area[3] * 1.6}
                                fill="url(#nox-city-haze)" />
                        ))}
                        <g className="auth-city-lights">
                            {CITY_LIGHTS.map((point, i) => (
                                <circle key={i} cx={point.x} cy={point.y} r={point.r}
                                    fill={i % 6 === 0 ? "#ffe4b0" : "#db926d"}
                                    opacity={point.opacity} />
                            ))}
                        </g>
                    </g>
                    <path d={HORIZON} fill="none" stroke="#e591dc"
                        strokeWidth="32" opacity=".25" filter="url(#nox-atmosphere-glow)" />
                    <path d={HORIZON} fill="none" stroke="#bc5ca5"
                        strokeWidth="12" opacity=".19" filter="url(#nox-soft-glow)" />
                    <path d={HORIZON} fill="none" stroke="url(#nox-horizon-rim)"
                        strokeWidth="3.7" strokeLinecap="round" />
                    <path d={HORIZON} fill="none" stroke="#f5dbeb"
                        strokeWidth=".9" opacity=".54" />
                </g>
            </svg>
        </div>
    );
}
