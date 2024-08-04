import {ADD_MESSAGE_DESC, ADD_MESSAGE_URGENT_DESC, RESET_DESC} from '@/actions/DescriptionActions';

const initialState = {
    messages: [{
        date: new Date().toISOString(), text: 'Descriptions are now being generated. Point your camera at something!', urgent: false
    }]
};

export default function descriptionReducer(state = initialState, action) {
    switch (action.type) {
        case ADD_MESSAGE_DESC:
            let message = {date: new Date().toISOString(), text: action.payload, urgent: false}
            return {
                ...state, messages: [...state.messages, message]
            };
        case ADD_MESSAGE_URGENT_DESC:
            let urgent_message = {date: new Date().toISOString(), text: action.payload, urgent: true}
            return {
                ...state, messages: [urgent_message]
            };
        case RESET_DESC:
            return {
                ...initialState
            }
        default:
            return state;
    }
}
