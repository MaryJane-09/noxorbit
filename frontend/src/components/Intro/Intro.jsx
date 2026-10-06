import "./Intro.css";
import GoldNLogo from "./GoldNlogo";
import OrbitalRings from "./OrbitalRings";

function Intro() {
    return (
        <div className="cinematic-intro">
            <div className="logo-composition">
                <OrbitalRings />
                <GoldNLogo />
            </div>
        </div>
    );
}

export default Intro;