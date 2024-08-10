import React, { useState, useEffect } from "react";

import { GeoJSON, Tooltip } from "react-leaflet";
import service from './poi_service'
import { getColorFromPrice } from "./maputils";

function computeRegionContourStyle(feature) {
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

export const RegionStat = () => {

    const [regionInfos, setRegionInfos] = useState(null)

    useEffect(() => {
        service.get("api/regions")
            .then((response) => {
                setRegionInfos(response.data);
            }).catch((error) => {
                console.error('Failed to load Region info:', error);
            });
    }, []);

    return (
        <div>
            {regionInfos && regionInfos.map(item => (
                <GeoJSON key={item.name} data={item.contour} style={computeRegionContourStyle}>
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