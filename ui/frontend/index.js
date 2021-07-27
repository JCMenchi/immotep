import React, { Suspense } from "react";
import ReactDOM from "react-dom";

import 'fontsource-roboto';

// import i18n (needs to be bundled ;)) 
import './i18n';

import App from "./App";

ReactDOM.render(
    <Suspense fallback={<h2>loading...</h2>}>
      <App />
    </Suspense>
  ,
  document.getElementById("root"));
