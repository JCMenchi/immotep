import React, { useState } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useTranslation } from 'react-i18next';
import { CssBaseline, IconButton, Grid, Paper, ThemeProvider } from '@material-ui/core';
import { createTheme } from '@material-ui/core/styles';
import { Brightness7, Brightness4 } from '@material-ui/icons';

import { selectUITheme, changeUITheme } from './store/uiparamSlice';

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
  const [uiTheme, setUITheme] = useState('dark');

  // get reducer dispatcher
  const dispatch = useDispatch();
  const UITheme = useSelector(selectUITheme);

  const toggleUITheme = () => {
    if (uiTheme === 'dark') {
      dispatch(changeUITheme('light'));
      setUITheme('light');
    } else {
      dispatch(changeUITheme('dark'));
      setUITheme('dark');
    }
  };

  return (

    <ThemeProvider theme={UITheme === 'dark' ? darkTheme : lightTheme}>
      <CssBaseline />
      <Grid container spacing={1}>
        <Grid item style={{ width: '100%' }}>
          <Paper variant='outlined'>
            <Grid item container direction='row' alignItems='center'>
              <Grid item >
                <IconButton size="small" onClick={() => toggleUITheme()} >
                  {UITheme === 'dark' ? <Brightness4 /> : <Brightness7 />}
                </IconButton>
              </Grid>
              <Grid item style={{ flexGrow: 1 }}>
                <Menubar/>
              </Grid>
            </Grid>
          </Paper>
        </Grid>
        <Grid item style={{ width: '100%' }}>
          <MapViewer />
        </Grid>
      </Grid>
    </ThemeProvider>

  );
}
