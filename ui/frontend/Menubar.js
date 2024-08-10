import React, { useState } from 'react';
import { useSelector, useDispatch } from 'react-redux';

import { Button, Grid, FormControl, IconButton, InputLabel, MenuItem, Select, TextField, Typography } from '@mui/material';
import { Brightness7, Brightness4, LocationOnSharp } from '@mui/icons-material';

import {
    selectUITheme,
    changeUITheme,
    changeQueryLimit,
    changeQueryDepartment,
    changePosition,
    changeYear,
    selectQueryLimit,
    selectQueryDepartment,
    selectAvgPrice,
    selectAvgPriceSQM,
    selectYear
} from './store/uiparamSlice';
import service from './poi_service';

export default function Menubar() {
    // state from redux global store
    const limit = useSelector(selectQueryLimit);
    const department = useSelector(selectQueryDepartment);
    const avgprice = useSelector(selectAvgPrice);
    const avgpriceSQM = useSelector(selectAvgPriceSQM);
    const year = useSelector(selectYear);

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
    const [ currentYear, setYear ] = useState(year);
    const [ currentAddress, setCurrentAddress ] = useState("");
    
    const loadAddress = (addr) => {
        console.log(addr)
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

    return (

        <Grid container spacing={1} direction='row' alignItems='center' style={{ paddingTop: 8, paddingBottom: 8 }}>
            <Grid item style={{ flexGrow: 0 }}>
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
                            loadAddress(currentAddress);
                        }
                    }}
                    disabled={false}
                    variant='outlined'
                    inputProps={{ style: { textAlign: 'right' } }}
                />
            </Grid>

            <Grid item style={{ flexGrow: 0 }}>
                <Button variant="outlined" onClick={() => {loadAddress(currentAddress);}}>
                   Go
                </Button>
            </Grid>

            <Grid item style={{ flexGrow: 0 }}>
                <FormControl>
                    <InputLabel size="small" id="select-year-label">Year</InputLabel>
                    <Select
                        labelId="select-year-label"
                        id="select-year"
                        variant='outlined'
                        value={currentYear}
                        label={"Year"}
                        onChange={(event) => {
                            setYear(event.target.value);
                            dispatch(changeYear(event.target.value));
                        }}>
                        <MenuItem value={-1}>
                        <em>All</em>
                        </MenuItem>
                        <MenuItem value={2016}>2016</MenuItem>
                        <MenuItem value={2017}>2017</MenuItem>
                        <MenuItem value={2018}>2018</MenuItem>
                        <MenuItem value={2019}>2019</MenuItem>
                        <MenuItem value={2020}>2020</MenuItem>
                        <MenuItem value={2021}>2021</MenuItem>
                        <MenuItem value={2022}>2022</MenuItem>
                        <MenuItem value={2023}>2023</MenuItem>
                    </Select>
               </FormControl>
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

