import React, { useState, useEffect } from "react";
import { useSelector, useDispatch } from 'react-redux';

import { MapContainer, Marker, TileLayer, useMapEvents, Popup } from "react-leaflet";
import service from './poi_service'

import {
    changeQueryLimit,
    changeQueryDepartment,
    changeShowMark,
    selectQueryLimit,
    selectQueryDepartment,
    selectUIShowMark
} from './store/uiparamSlice';


const style = {
    height: '400px',
    width: '100%'
};

function LocationMarker() {

    // state from redux global store
    const showMark = useSelector(selectUIShowMark);
    const limit = useSelector(selectQueryLimit);
    const department = useSelector(selectQueryDepartment);

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
                    setTransactions(response.data);
                }).catch((error) => {
                    console.error('Failed to load pois:', error);
                });
        }

    })

    useEffect(() => {
    }, []);

    return (
        <div>
            { lastPos !== null &&
            <Marker position={lastPos}>
                <Popup><div>{info}</div></Popup>
            </Marker>
            }
            {showMark && transactions !== null && transactions.map(item => (
                <Marker key={item.id} riseOnHover={true} position={[item.lat, item.long]} >
                    <Popup> {`${item.date.split('T')[0]}: ${item.price}€`} <br /> {`${item.area}m²`}<br /> {item.address} <br /> {item.city} </Popup>
                </Marker>
            ))}
        </div>
    )
}


export const MapViewer = () => {
    const [position] = useState([48.6007, -4.0451]);// initial position of map

    return (
        <div id="map">
            <MapContainer center={position}
                zoom={10}
                scrollWheelZoom={true}
                style={style}>

                <TileLayer
                    attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
                    url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                />
                <LocationMarker />
            </MapContainer>
        </div>
        )
}

/*
Result of call to GET /api/pois
[
    {
        "id": 2145605,
        "date": "2020-01-03T00:00:00Z",
        "address": "14  HAM DES HAUTS DU GUERN",
        "city": "LA FORET-FOUESNANT",
        "price": 650000,
        "area": 218,
        "lat": 47.906863,
        "long": -3.954772
    }
]

*/