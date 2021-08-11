import React, { useState, useEffect } from "react";
import { useSelector } from 'react-redux';

import { GeoJSON, Tooltip } from "react-leaflet";
import service from './poi_service'

import {
    selectQueryDepartment
} from './store/uiparamSlice';

function getColor(v) {
    return v > 3000 ? "#7f0000" :
           v > 2500 ? "#b30000" :
           v > 2000 ? "#d7301f" :
           v > 1600 ? "#ef6548" :
           v > 1400 ? "#fc8d59" :
           v > 1200 ? "#fdbb84" :
           v > 1000 ? "#fdd49e" :
           v > 500  ? "#fee8c8" :
                      "#fff7ec";
}

const citycontour = {
    fillColor: "red",
    weight: 1,
    strokeWidth: 2,
    opacity: 1,
    color: "red",
    fillOpacity: 0.2
};

function computeCityContourStyle(feature) {
    console.log(feature.properties);
    return {
        fillColor: getColor(feature.properties.avgprice),
        weight: 1,
        strokeWidth: 2,
        opacity: 1,
        color: getColor(feature.properties.avgprice),
        fillOpacity: 0.5
    };
}

export const CityStat = () => {

    // state from redux global store
    const department = useSelector(selectQueryDepartment);

    const [cityInfos, setCityInfos] = useState(null)

    useEffect(() => {
        service.get("api/city?limit=600&dep=" + department)
            .then((response) => {
                setCityInfos(response.data);
            }).catch((error) => {
                console.error('Failed to load city info:', error);
            });
    }, [department]);

    return (
        <div>
            {cityInfos && cityInfos.map(item => (
                <GeoJSON key={item.name} data={item.contour} style={computeCityContourStyle}>
                    <Tooltip>
                        {`${item.name}: ${item.avgprice.toFixed(0)}€`}
                    </Tooltip>
                </GeoJSON>
            ))}
        </div>
    )
}