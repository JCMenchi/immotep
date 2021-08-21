import React from 'react';
import { useSelector } from 'react-redux';
import { CssBaseline, Grid, ThemeProvider } from '@material-ui/core';
import { createTheme } from '@material-ui/core/styles';

import { selectCenterPosition, selectZoom, selectUITheme} from './store/uiparamSlice';

import { MapViewer } from './MapViewer';
import './App.css';
import Menubar from './Menubar';

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
  const position = useSelector(selectCenterPosition);
  const zoom = useSelector(selectZoom);
  const UITheme = useSelector(selectUITheme);

  return (

    <ThemeProvider theme={UITheme === 'dark' ? darkTheme : lightTheme}>
      <CssBaseline />
      <Grid container direction='column' spacing={1} style={{ height: '100%' }}>
        <Grid item style={{ width: '100%' }}>
            <Menubar/>
        </Grid>
        <Grid item style={{ flexGrow: 1, width: '100%' }}>
          <MapViewer initposition={position} initZoom={zoom} />
        </Grid>
      </Grid>
    </ThemeProvider>

  );
}
