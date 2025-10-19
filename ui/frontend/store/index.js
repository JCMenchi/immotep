// REDUX import
import { configureStore, combineReducers, Tuple } from '@reduxjs/toolkit';

// REDUX middleware
import {thunk} from "redux-thunk";

// my REDUX reducer
import { uiParam } from './uiparamSlice';

const reducers = combineReducers({
  uiParam: uiParam.reducer
})

/*
import logger from "redux-logger";

const middlewareList = [thunk];
if (process.env.NODE_ENV !== 'production') {
  middlewareList.push(logger);
}
*/

export const store = configureStore({
  reducer: reducers,
  middleware: () => new Tuple(thunk),
})

