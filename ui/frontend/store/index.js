// REDUX import
import { configureStore, combineReducers } from '@reduxjs/toolkit';

// REDUX persistence
import { persistStore, persistReducer } from 'redux-persist';
import autoMergeLevel2 from 'redux-persist/lib/stateReconciler/autoMergeLevel2'
import storage from 'redux-persist/lib/storage';

// REDUX middleware
import logger from "redux-logger";
import thunk from "redux-thunk";

// my REDUX reducer
import { uiParam } from './uiparamSlice';

const reducers = combineReducers({
  uiParam: uiParam.reducer
})

export const persistConfig = {
  key: 'imt_cache',
  version: 1,
  debug: true,
  storage,
  timeout: 5000,
  whitelist: ['uiParam'],
  stateReconciler: autoMergeLevel2
}

const persistedReducer = persistReducer(persistConfig, reducers)

const middlewareList = [thunk];
/*if (process.env.NODE_ENV !== 'production') {
  middlewareList.push(logger);
}*/

export const store = configureStore({
  reducer: persistedReducer,
  middleware: middlewareList,
})

export const persistor = persistStore(store, undefined, () => { console.log('rehydration complete')});

