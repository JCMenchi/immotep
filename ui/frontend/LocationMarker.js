import React, { useState } from "react";
import { useSelector, useDispatch } from 'react-redux';

import { Marker, useMapEvents, Popup } from "react-leaflet";
import service from './poi_service'

import {
    changeAvgPrice,
    changeAvgPriceSQM,
    selectQueryLimit,
    selectQueryDepartment
} from './store/uiparamSlice';

export const LocationMarker = () => {

    // state from redux global store
    const limit = useSelector(selectQueryLimit);
    const department = useSelector(selectQueryDepartment);

    // get reducer dispatcher
    const dispatch = useDispatch();

    const [lastPos, setLastPos] = useState(null)
    const [info, setInfo] = useState("")

    const [transactions, setTransactions] = useState(null)

    const map = useMapEvents({
        click(event) {
            //map.locate()
            console.log('map point:', event.latlng)
            console.log('map center:', map.getCenter())
            console.log('map bounds:', map.getBounds())
            console.log('map zoom:', map.getZoom())
            setLastPos(event.latlng)
            setInfo(`latlong: ${event.latlng.lat}, ${event.latlng.lng}`)
        },

        moveend(_event) {
            console.log('moveend map bounds:', map.getBounds())
            const bounds = {
                northEast: map.getBounds()._northEast,
                southWest: map.getBounds()._southWest,
                code: department
            }

            // need to reload data with new bounds
            service.post("api/pois/filter?limit=" + limit, bounds)
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


    return (
        <div>
            {lastPos !== null &&
                <Marker position={lastPos}>
                    <Popup><div>{info}</div></Popup>
                </Marker>
            }
            {transactions !== null && transactions.map(item => (
                <Marker key={item.id} riseOnHover={true} position={[item.lat, item.long]} >
                    <Popup> {`${item.date.split('T')[0]}: ${item.price}€`} <br /> {`${item.area}m²`}<br /> {item.address} <br /> {item.city} </Popup>
                </Marker>
            ))}
        </div>
    )
}
