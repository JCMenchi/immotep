import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';

import { Grid, TextField, Typography } from '@material-ui/core';

import {
    changeQueryLimit,
    changeQueryDepartment,
    selectQueryLimit,
    selectQueryDepartment,
    selectUIShowMark
} from './store/uiparamSlice';


export default function Menubar() {
    // state from redux global store
    const limit = useSelector(selectQueryLimit);
    const department = useSelector(selectQueryDepartment);
   
    // get reducer dispatcher
    const dispatch = useDispatch();

    const [ currentLimit, setCurrentLimit ] = useState(limit);
    const [ currentDep, setCurrentDep ] = useState(department);

    return (

        <Grid container spacing={1} direction='row' alignItems='center' style={{ paddingTop: 8, paddingBottom: 8 }}>
            <Grid item>
                <Typography variant='h5'>{"Map"}</Typography>
            </Grid>

            <Grid item style={{ flexGrow: 1 }}>
                <TextField
                    type='text'
                    label={"Departement"}
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
            
            <Grid item style={{ flexGrow: 1 }}>
                <TextField
                    type='number'
                    label={"Limit"}
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
           

        </Grid>
    );
}

