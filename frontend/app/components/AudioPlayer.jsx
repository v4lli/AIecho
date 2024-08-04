import {useEffect, useRef, useState} from 'react';
import IconButton from '@mui/material/IconButton';
import BoltIcon from '@mui/icons-material/Bolt';
import {useSelector} from 'react-redux';

export const AudioPlayer = ({toggleFastMode, onStream}) => {
    const state = useSelector(state => state.description);
    const spokenMessages = useRef(new Set());
    const [isClicked, setClicked] = useState(true);

    useEffect(() => {
        const synth = window.speechSynthesis;
        if (onStream) {
            state.messages.forEach(message => {
                if (!spokenMessages.current.has(message.date)) {
                    const utterance = new SpeechSynthesisUtterance(message.text);
		    utterance.rate = 1.1;
                    synth.speak(utterance);
                    spokenMessages.current.add(message.date);
                }
            });
        }else{
            synth.cancel()
        }
    }, [onStream, state.messages]);

    const handleClick = () => {
        setClicked(!isClicked);
        toggleFastMode();
    }


    return (<div className="mode" aria-label="Fast Mode switches the reading such that the The explanation is directly read">
        <button className={`mode-button ${isClicked ? 'clicked':''}`} onClick={handleClick} aria-label="Toggle fast mode">
           <BoltIcon className="bolt-icon" /> {`${isClicked ? 'Summary Only':'Continous Output'}`}
        </button>
    </div>);
};
