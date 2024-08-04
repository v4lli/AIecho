import {ADD_MESSAGE_TL, RESET_TL} from '@/actions/TimelineActions';

const initialState = {
    messages: [{
        date: new Date().toISOString(), text: 'Descriptions are now being generated. Point your camera at something!'
    }]
};

export default function timelineReducer(state = initialState, action) {
    switch (action.type) {
        case ADD_MESSAGE_TL:
            let message = {date: new Date().toISOString(), text: action.payload}
            return {
                ...state, messages: [...state.messages, message]
            };
        case RESET_TL:
            return {
                ...initialState
            }
        default:
            return state;
    }
}
