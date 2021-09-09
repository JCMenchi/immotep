import React, { Suspense } from "react";
import ReactDOM from "react-dom";
import { Provider } from 'react-redux';

import '@fontsource/roboto';

// import i18n (needs to be bundled ;)) 
import './i18n';

import App from "./App";

import { store } from './store';

ReactDOM.render(
  <Provider store={store}>
    <Suspense fallback={<h2>loading...</h2>}>
      <App />
    </Suspense>
  </Provider>
  ,
  document.getElementById("root"));
