import React, { useState, useEffect } from "react";

import { GeoJSON, Tooltip } from "react-leaflet";
import service from './poi_service'

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

function computeDepartmentContourStyle(feature) {
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

export const DepartmentStat = (props) => {

    const [departmentInfos, setDepartmentInfos] = useState(null)

    useEffect(() => {
        service.get("api/departments")
            .then((response) => {
                setDepartmentInfos(response.data);
            }).catch((error) => {
                console.error('Failed to load Department info:', error);
            });
    }, []);

    return (
        <div>
            {departmentInfos && departmentInfos.map(item => (
                <GeoJSON key={item.name} data={item.contour} style={computeDepartmentContourStyle}>
                    <Tooltip>
                        {`(${item.code}) ${item.name}: ${item.avgprice.toFixed(0)}€`}
                    </Tooltip>
                </GeoJSON>
            ))}
        </div>
    )
}