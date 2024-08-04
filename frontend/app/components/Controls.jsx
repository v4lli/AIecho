import React, { useState } from 'react';
import BoltIcon from "@mui/icons-material/Bolt";
import {Visibility, VisibilityOff} from "@mui/icons-material";

export const Controls = ({ onStart, onStop }) => {
    const [visible, setVisible] = useState(true);

    const toggleVisibility = () => {
        setVisible(!visible);
    };

    return (
        <div className="startStop" aria-label={'Control'}>
            {visible && <button onClick={() => { onStart(); toggleVisibility(); }} aria-label={'Start AIEcho'}><Visibility className="bolt-icon" /> Start Describing</button>}
            {!visible && <button onClick={() => { onStop(); toggleVisibility(); }} aria-label={'Stop AIEcho'}><VisibilityOff className="bolt-icon" /> Stop Describing</button>}
        </div>
    );
};
