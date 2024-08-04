import {useSelector} from 'react-redux';
import {useEffect, useRef} from 'react';
import moment from 'moment-timezone';

export const Timeline = () => {
    const state = useSelector(state => state.timeline);
    const timelineRef = useRef(null);

    useEffect(() => {
        const timeline = timelineRef.current;
        if (timeline) {
            timeline.scrollTop = timeline.scrollHeight;
        }
    }, [state.messages])
    ;
    const formatDate = (date) => {
        return moment(date).local().format('MMMM Do YYYY, hh:mm:ss');
    }

    return (<div className="timeline-container" aria-label={'Timeline with Object Description list'} ref={timelineRef}>
        {state.messages.map((msg, index) => (<div key={index} className="message" aria-label={'Timeline Message'}>
            <div className="message-time">{formatDate(msg.date)}</div>
            <div>{msg.text}</div>
        </div>))}
    </div>);
};