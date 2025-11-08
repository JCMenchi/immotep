import React, { useState, useEffect } from "react";
import { useSelector, useDispatch } from 'react-redux';

import { GeoJSON, Tooltip } from "react-leaflet";
import service from './poi_service'
import { getColorFromPrice } from "./maputils";

import { useMapEvents } from "react-leaflet";

import {
    changeAvgPrice,
    changeAvgPriceSQM,
    selectQueryLimit
} from './store/uiparamSlice';

export function computeCityContourStyle(feature) {
    let color = "grey";

    if (feature && feature.properties && feature.properties.avgprice) {
        color = getColorFromPrice(feature.properties.avgprice);
    }

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
    const limit = useSelector(selectQueryLimit);

    const [cityInfos, setCityInfos] = useState(null)

    // get reducer dispatcher
    const dispatch = useDispatch();

    const map = useMapEvents({
        moveend(_event) {
            const bounds = {
                northEast: map.getBounds()._northEast,
                southWest: map.getBounds()._southWest
            }

            // need to reload data with new bounds
            service.post("api/cities?limit=" + limit, bounds)
                .then((response) => {
                    const infos = response.data;
                    setCityInfos(infos.cities);
             
                    if (infos.avgprice > 0) {
                        dispatch(changeAvgPrice(infos.avgprice));
                    } else {
                        dispatch(changeAvgPrice(-1));
                    }
                    if (infos.avgprice_sqm > 0) {
                        dispatch(changeAvgPriceSQM(infos.avgprice_sqm));
                    } else {
                        dispatch(changeAvgPriceSQM(-1));
                    }

                }).catch((error) => {
                    console.error('Failed to load cities:', error);
                });

        }

    })

    useEffect(() => {
        service.get("api/cities?limit=600")
            .then((response) => {
                if (Array.isArray(response.data)) {
                    setCityInfos(response.data);
                }
            }).catch((error) => {
                console.error('Failed to load city info:', error);
            });
    }, [limit]);

    return (
        <div>
            {cityInfos && cityInfos.map(item => (
                <GeoJSON key={item.name} data={item.contour} style={computeCityContourStyle}>
                    <Tooltip>
                        {`(${item.zip}) ${item.name}: ${item.avgprice.toFixed(0)}€`}<br /> {`Population: ${item.population}`}<br />
                        {item.stat && Object.keys(item.stat).map((k) => {
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