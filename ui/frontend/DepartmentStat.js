import React, { useState, useEffect } from "react";

import { GeoJSON, Tooltip } from "react-leaflet";
import service from './poi_service'
import { getColorFromPrice } from "./maputils";

function computeDepartmentContourStyle(feature) {
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

export const DepartmentStat = () => {

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
                        { item.stat && Object.keys(item.stat).map((k,i) => {
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