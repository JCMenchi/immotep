import React, { Suspense } from "react";
import ReactDOM from "react-dom";
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';

import 'fontsource-roboto';

// import i18n (needs to be bundled ;)) 
import './i18n';

import App from "./App";

import { store, persistor } from './store';

ReactDOM.render(
  <Provider store={store}>
    <PersistGate
      persistor={persistor}
      onBeforeLift={() => new Promise(resolve => setTimeout(resolve, 500))}>
      <Suspense fallback={<h2>loading...</h2>}>
        <App />
      </Suspense>
    </PersistGate>
  </Provider>
  ,
  document.getElementById("root"));
