import {combineReducers, configureStore} from '@reduxjs/toolkit';
import timelineReducer from '@/reducers/TimelineReducer';
import descriptionReducer from '@/reducers/DescriptionReducer';

const initialState = {
  timeline: {messages: []}, description: {messages: []}
};

const rootReducer = combineReducers({
  timeline: timelineReducer, description: descriptionReducer
});

const store = configureStore({
  reducer: rootReducer
});

export default store;
