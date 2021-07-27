import React, { useState } from 'react';
import { BrowserRouter, Route, Switch } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet'
import { CssBaseline, Snackbar, ThemeProvider } from '@material-ui/core';
import { createTheme } from '@material-ui/core/styles';

import './App.css';

/* Style info to use for theming */
const darkTheme = createTheme({
  palette: {
    type: 'dark',
  }
});

const lightTheme = createTheme({
  palette: {
    type: 'light',
  }
});

/**
 * Main Application Component
 * 
 * @param {Object} props 
 */
export default function App(props) {

  // local state
  const [uiTheme, setUITheme] = useState('dark');

  const toggleUITheme = () => {
    if (uiTheme === 'dark') {
      setUITheme('light');
    } else {
      setUITheme('dark');
    }
  };

  return (

    <ThemeProvider theme={uiTheme === 'dark' ? darkTheme : lightTheme}>
      <CssBaseline />
      <MapContainer center={[51.505, -0.09]} zoom={13} scrollWheelZoom={true}>
        <TileLayer
          attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <Marker position={[51.505, -0.09]}>
          <Popup>
            A pretty CSS3 popup. <br /> Easily customizable.
          </Popup>
        </Marker>
      </MapContainer>
    </ThemeProvider>

  );
}
