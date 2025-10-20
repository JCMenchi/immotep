import React, { useState, useEffect } from "react";
import { useSelector, useDispatch } from 'react-redux';

import { GeoJSON, Tooltip } from "react-leaflet";
import service from './poi_service'
import { getColorFromPrice } from "./maputils";

import { useMapEvents } from "react-leaflet";

import {
    changeAvgPrice,
    changeAvgPriceSQM,
    selectQueryDepartment,
    selectQueryLimit,
    selectYear
} from './store/uiparamSlice';

function computeCityContourStyle(feature) {
    const color = getColorFromPrice(feature.properties.avgprice);
    return {
        fillColor: color,
        weight: 1,
        strokeWidth: 2,
        opacity: 1,
        color: "red",
        fillOpacity: 0.5
    };
}

export const CityStat = () => {

    // state from redux global store
    const department = useSelector(selectQueryDepartment);
    const limit = useSelector(selectQueryLimit);
    const year = useSelector(selectYear);

    const [cityInfos, setCityInfos] = useState(null)
    const [transactions, setTransactions] = useState(null)

    // get reducer dispatcher
    const dispatch = useDispatch();

    const map = useMapEvents({
        moveend(_event) {
            const bounds = {
                northEast: map.getBounds()._northEast,
                southWest: map.getBounds()._southWest,
                code: department
            }
            
            // need to reload data with new bounds
            service.get("api/cities?limit=600&dep=" + department)
            .then((response) => {
                setCityInfos(response.data);
            }).catch((error) => {
                console.error('Failed to load city info:', error);
            });
            
            service.post("api/pois/filter?limit=" + limit + "&year=" + year, bounds)
                .then((response) => {
                    const tr = response.data;
                    setTransactions(tr.transactions);

                    if (tr.avgprice > 0) {
                        dispatch(changeAvgPrice(tr.avgprice));
                    } else {
                        dispatch(changeAvgPrice(-1));
                    }
                    if (tr.avgprice_sqm > 0) {
                        dispatch(changeAvgPriceSQM(tr.avgprice_sqm));
                    } else {
                        dispatch(changeAvgPriceSQM(-1));
                    }

                }).catch((error) => {
                    console.error('Failed to load pois:', error);
                });

        }

    })

    useEffect(() => {
        service.get("api/cities?limit=600&dep=" + department)
            .then((response) => {
                if (Array.isArray(response.data)) {
                    setCityInfos(response.data);
                }
            }).catch((error) => {
                console.error('Failed to load city info:', error);
            });
    }, [department]);

    return (
        <div>
            {cityInfos && cityInfos.map(item => (
                <GeoJSON key={item.name} data={item.contour} style={computeCityContourStyle}>
                    <Tooltip>
                        {`(${item.zip}) ${item.name}: ${item.avgprice.toFixed(0)}€`}<br /> {`Population: ${item.population}`}<br /> 
                        { item.stat && Object.keys(item.stat).map((k) => {
                                return (
                                    <span>&nbsp; {k + ": " + item.stat[k]}<br /></span>
                                )
                            })
                        }
                    </Tooltip>
                </GeoJSON>
            ))}
        </div>
    )
}