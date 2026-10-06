function OrbitalRings() {
    return (
        <div className="orbit-entrance">
            <svg className="orbit-rings orbit-back" viewBox="0 0 300 300">

                <ellipse
                    className="orbit-ring"
                    cx="150"
                    cy="150"
                    rx="120"
                    ry="55"
                    fill="none"
                    stroke="#E5CC99"
                    strokeWidth="3"
                />

                <ellipse
                    className="orbit-ring"
                    cx="150"
                    cy="150"
                    rx="105"
                    ry="75"
                    fill="none"
                    stroke="#E5CC99"
                    strokeWidth="2"
                    opacity="0.7"
                    transform="rotate(60 150 150)"
                />

                <ellipse
                    className="orbit-ring"
                    cx="150"
                    cy="150"
                    rx="65"
                    ry="115"
                    fill="none"
                    stroke="#E5CC99"
                    strokeWidth="2"
                    opacity="0.35"
                    transform="rotate(120 150 150)"
                />

            </svg>

            <svg className="orbit-rings orbit-front" viewBox="0 0 300 300">

                <ellipse
                    className="orbit-ring"
                    cx="150"
                    cy="150"
                    rx="120"
                    ry="55"
                    fill="none"
                    stroke="#E5CC99"
                    strokeWidth="3"
                />

                <ellipse
                    className="orbit-ring"
                    cx="150"
                    cy="150"
                    rx="105"
                    ry="75"
                    fill="none"
                    stroke="#E5CC99"
                    strokeWidth="2"
                    opacity="0.7"
                    transform="rotate(60 150 150)"
                />

                <ellipse
                    className="orbit-ring"
                    cx="150"
                    cy="150"
                    rx="80"
                    ry="120"
                    fill="none"
                    stroke="#E5CC99"
                    strokeWidth="2"
                    opacity="0.45"
                    transform="rotate(120 150 150)"
                />

            </svg>

        </div>
    );
}

export default OrbitalRings;