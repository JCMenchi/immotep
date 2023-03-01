import React, { useState } from 'react';
import { useSelector, useDispatch } from 'react-redux';

import { Grid, IconButton, TextField, Typography } from '@mui/material';
import { Brightness7, Brightness4, LocationOnSharp } from '@mui/icons-material';

import {
    selectUITheme,
    changeUITheme,
    changeQueryLimit,
    changeQueryDepartment,
    changePosition,
    selectQueryLimit,
    selectQueryDepartment,
    selectAvgPrice,
    selectAvgPriceSQM
} from './store/uiparamSlice';
import service from './poi_service';

export default function Menubar() {
    // state from redux global store
    const limit = useSelector(selectQueryLimit);
    const department = useSelector(selectQueryDepartment);
    const avgprice = useSelector(selectAvgPrice);
    const avgpriceSQM = useSelector(selectAvgPriceSQM);

    // get reducer dispatcher
    const dispatch = useDispatch();

    const UITheme = useSelector(selectUITheme);

    const toggleUITheme = () => {
        if (UITheme === 'dark') {
            dispatch(changeUITheme('light'));
        } else {
            dispatch(changeUITheme('dark'));
        }
    };

    const findMe = () => {
        if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(pos => {
                dispatch(changePosition([pos.coords.latitude, pos.coords.longitude]));
            });
        }
    }

    const [ currentLimit, setCurrentLimit ] = useState(limit);
    const [ currentDep, setCurrentDep ] = useState(department);
    const [ currentAddress, setCurrentAddress ] = useState("");
    
    return (

        <Grid container spacing={1} direction='row' alignItems='center' style={{ paddingTop: 8, paddingBottom: 8 }}>
            <Grid item>
                <Typography variant='h5'>
                    {avgprice > 0 && `Prix Moyen: ${avgprice.toFixed(0)} € `}
                    {avgpriceSQM > 0 && ` ${avgpriceSQM.toFixed(0)} €/m² `}
                </Typography>
            </Grid>

            <Grid item style={{ flexGrow: 1 }}>
                <TextField
                    fullWidth
                    type='text'
                    label={"Address"}
                    value={currentAddress}
                    onChange={(event) => {
                        setCurrentAddress(event.target.value);
                    }}
                    onKeyUp={(event) => {
                        if (event.key == 'Enter') {
                            console.log(currentAddress)
                            const addr = currentAddress;
                            const baseURL = "https://api-adresse.data.gouv.fr/search/?q=" + addr

                            service.get(baseURL)
                            .then((response) => {
                                const features = response.data.features;
                                if (features.length > 0) {
                                    const pos = features[0].geometry.coordinates;
                                    console.log(features[0].properties, pos);
                                    dispatch(changePosition([pos[1], pos[0]]))
                                }
                            }).catch((error) => {
                                console.error('Failed to found addres pos:', error);
                            });
                        }
                    }}
                    disabled={false}
                    variant='outlined'
                    inputProps={{ style: { textAlign: 'right' } }}
                />
            </Grid>

            <Grid item style={{ flexGrow: 0 }}>
                <TextField
                    type='text'
                    label={"Departement"}
                    style={{ width: 110 }}
                    value={currentDep}
                    onChange={(event) => {
                        setCurrentDep(event.target.value);
                    }}
                    onBlur={(event) => {
                        dispatch(changeQueryDepartment(currentDep));
                    }}
                    onKeyUp={(event) => {
                        if (event.key == 'Enter') {
                            dispatch(changeQueryDepartment(currentDep));
                        }
                    }}
                    disabled={false}
                    variant='outlined'
                    inputProps={{ style: { textAlign: 'right' } }}
                />
            </Grid>
            
            <Grid item style={{ flexGrow: 0 }}>
                <TextField
                    type='number'
                    label={"Limit"}
                    style={{ width: 80 }}
                    value={currentLimit}
                    onChange={(event) => {
                        setCurrentLimit(event.target.value)
                    }}
                    onKeyUp={(event) => {
                        if (event.key == 'Enter') {
                            dispatch(changeQueryLimit(Number(currentLimit)));
                        }
                    }}
                    disabled={false}
                    variant='outlined'
                    inputProps={{ style: { textAlign: 'right' } }}
                />
            </Grid>
           
            <Grid item >
                <IconButton size="small" onClick={() => toggleUITheme()} >
                  {UITheme === 'dark' ? <Brightness4 /> : <Brightness7 />}
                </IconButton>
                <IconButton size="small" onClick={() => findMe()} >
                  <LocationOnSharp />
                </IconButton>
            </Grid>

        </Grid>
    );
}

