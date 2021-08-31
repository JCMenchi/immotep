import React, { useState, useEffect } from "react";
import { useSelector } from 'react-redux';

import { GeoJSON, Tooltip } from "react-leaflet";
import service from './poi_service'
import { getColorFromPrice } from "./maputils";

import {
    selectQueryDepartment
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

    const [cityInfos, setCityInfos] = useState(null)

    useEffect(() => {
        service.get("api/cities?limit=600&dep=" + department)
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
                        {`(${item.zip}) ${item.name}: ${item.avgprice.toFixed(0)}€`}
                    </Tooltip>
                </GeoJSON>
            ))}
        </div>
    )
}