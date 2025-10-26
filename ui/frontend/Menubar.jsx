// ui/frontend/Menubar.jsx

// React hooks
import { useState } from 'react';
import { useSelector, useDispatch } from 'react-redux';

// MUI components
import { Button, Grid, FormControl, IconButton, InputLabel, MenuItem, Select, TextField, Typography } from '@mui/material';
import { Brightness7, Brightness4, LocationOnSharp } from '@mui/icons-material';

// redux actions and selectors
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

// service for API calls
import service from './poi_service';

/**
 * Menubar component provides navigation and control bar for real estate data visualization.
 * 
 * @component
 * @description
 * Displays average property prices, provides address search with geocoding,
 * year filtering (2016-2023), department filtering, query limit control,
 * UI theme toggle and geolocation finder.
 * 
 * Uses Redux for state management with the following global states:
 * - limit: Maximum number of query results
 * - department: Selected department code
 * - avgprice: Average property price
 * - avgpriceSQM: Average price per square meter
 * - year: Selected year filter
 * - UITheme: Current UI theme ('light' or 'dark')
 * 
 * @example
 * ```jsx
 * <Menubar />
 * ```
 * @returns {JSX.Element} A Material-UI Grid container with navigation controls
 */
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

    /**
     * Toggles between light and dark UI themes
     */
    const toggleUITheme = () => {
        if (UITheme === 'dark') {
            dispatch(changeUITheme('light'));
        } else {
            dispatch(changeUITheme('dark'));
        }
    };

    /**
     * Uses browser geolocation API to set user's current position
     */
    const findMe = () => {
        if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(pos => {
                dispatch(changePosition([pos.coords.latitude, pos.coords.longitude]));
            });
        }
        
    }

    // Local state declarations
    const [ currentLimit, setCurrentLimit ] = useState(limit);
    const [ currentDep, setCurrentDep ] = useState(department);
    const [ currentYear, setYear ] = useState(year);
    const [ currentAddress, setCurrentAddress ] = useState("");
    
    /**
     * Geocodes an address string using the French government's API
     * and updates the position accordingly
     * 
     * {
        "type": "FeatureCollection",
        "features": [
            {
                "type": "Feature",
                "geometry": {
                    "type": "Point",
                    "coordinates": [
                        -2.316402,
                        48.21157
                    ]
                },
                "properties": {
                    "label": "1 La Ville Allain 22230 Trémorel",
                    "score": 0.9434745454545453,
                    "housenumber": "1",
                    "id": "22371_d415_00001",
                    "banId": "bdaa9e02-7eec-496e-8375-cf6aece94acf",
                    "name": "1 La Ville Allain",
                    "postcode": "22230",
                    "citycode": "22371",
                    "x": 305386.22,
                    "y": 6803424.77,
                    "city": "Trémorel",
                    "context": "22, Côtes-d'Armor, Bretagne",
                    "type": "housenumber",
                    "importance": 0.37822,
                    "street": "La Ville Allain",
                    "_type": "address"
                }
            }
        ],
        "query": "1 la ville allain tremorel"
    }
     * 
     * @param {string} addr - The address to geocode
     */
    const loadAddress = (addr) => {
        console.log(addr)
        const baseURL = "https://data.geopf.fr/geocodage/search/?q=" + addr

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
            <Grid style={{ flexGrow: 0 }}>
                <Typography variant='h5'>
                    {avgprice > 0 && `Prix Moyen: ${avgprice.toFixed(0)} € `}
                    {avgpriceSQM > 0 && ` ${avgpriceSQM.toFixed(0)} €/m² `}
                </Typography>
            </Grid>

            <Grid style={{ flexGrow: 1 }}>
                <TextField
                    fullWidth
                    type='text'
                    label={"Address"}
                    value={currentAddress}
                    onChange={(event) => {
                        setCurrentAddress(event.target.value);
                    }}
                    onBlur={() => {
                        loadAddress(currentAddress);
                    }}
                    onKeyUp={(event) => {
                        if (event.key == 'Enter') {
                            loadAddress(currentAddress);
                        }
                    }}
                    disabled={false}
                    variant='outlined'
                    slotProps={{
                        htmlInput: { style: { textAlign: 'left' } }
                    }}
                />
            </Grid>

            <Grid style={{ flexGrow: 0 }}>
                <Button variant="outlined" onClick={() => {loadAddress(currentAddress);}}>
                   Go
                </Button>
            </Grid>

            <Grid style={{ flexGrow: 0 }}>
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
                        <MenuItem value={2024}>2024</MenuItem>
                        <MenuItem value={2025}>2025</MenuItem>
                    </Select>
               </FormControl>
            </Grid>

            <Grid style={{ flexGrow: 0 }}>
                <TextField
                    type='text'
                    label={"Departement"}
                    style={{ width: 110 }}
                    value={currentDep}
                    onChange={(event) => {
                        setCurrentDep(event.target.value);
                    }}
                    onBlur={() => {
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
            
            <Grid style={{ flexGrow: 0 }}>
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
           
            <Grid>
                <IconButton id="change-theme-button" size="small" onClick={() => toggleUITheme()} >
                  {UITheme === 'dark' ? <Brightness4 /> : <Brightness7 />}
                </IconButton>
                <IconButton size="small" onClick={() => findMe()} >
                  <LocationOnSharp />
                </IconButton>
            </Grid>

        </Grid>
    );
}

