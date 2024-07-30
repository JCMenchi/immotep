import React from 'react';
import { useSelector } from 'react-redux';
import { CssBaseline, Grid } from '@mui/material';
import { createTheme, ThemeProvider } from '@mui/material/styles';

import { selectCenterPosition, selectZoom, selectUITheme} from './store/uiparamSlice';

import { MapViewer } from './MapViewer';
import './App.css';
import Menubar from './Menubar';

/* Style info to use for theming */
const darkTheme = createTheme({
  palette: {
    mode: 'dark',
  }
});

const lightTheme = createTheme({
  palette: {
    mode: 'light',
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
      <React.StrictMode>
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
      </React.StrictMode>
  );
}
