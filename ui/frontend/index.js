import React, { Suspense } from "react";
import { Provider } from 'react-redux';
import { createRoot } from 'react-dom/client';


import '@fontsource/roboto';

// import i18n (needs to be bundled ;)) 
import './i18n';

import App from "./App";

import { store } from './store';

const container = document.getElementById('root');
const root = createRoot(container);
root.render(
  <Provider store={store}>
    <Suspense fallback={<h2>loading...</h2>}>
      <App />
    </Suspense>
  </Provider>
);
