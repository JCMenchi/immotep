import React, { useState } from "react";
import { LayersControl, LayerGroup, MapContainer, TileLayer } from "react-leaflet";

import { CityStat } from "./CityStat";
import { LocationMarker } from "./LocationMarker";

const style = {
    height: '100%',
    width: '100%'
};

const mbbaseurl = "https://api.mapbox.com/styles/v1/jcmenchi/";

const mbtoken = "pk.eyJ1IjoiamNtZW5jaGkiLCJhIjoiY2tyaTQxOXZjMGM4YTJ1cnZ0ZGM0eWdlbSJ9.Cqy-UGrsUUWAGF8mFPUiGg";

export const MapViewer = () => {
    const [position] = useState([48.6007, -4.0451]);// initial position of map

    return (
        <div id="map" style={style}>
            <MapContainer center={position}
                zoom={10}

                scrollWheelZoom={true}
                style={style}>

                <LayersControl position="topright">
                    <LayersControl.BaseLayer checked name="Map">
                        <TileLayer
                            attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
                            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                        />
                    </LayersControl.BaseLayer>
                    <LayersControl.BaseLayer name="Satellite">
                        <TileLayer
                            attribution='&copy; <a href="https://www.mapbox.com/feedback/">Mapbox</a>'
                            url={mbbaseurl + "ckri48goh6zmf18q98pzmp1q4/tiles/{z}/{x}/{y}?access_token=" + mbtoken}
                        />
                    </LayersControl.BaseLayer>
                    <LayersControl.BaseLayer name="Light">
                        <TileLayer
                            url={'https://api.mapbox.com/styles/v1/{id}/tiles/{z}/{x}/{y}?access_token=' + mbtoken}
                            id='mapbox/light-v9'
                            attribution='&copy; <a href="https://www.mapbox.com/feedback/">Mapbox</a>'
                            tileSize={512}
                            zoomOffset={-1}
                        />

                    </LayersControl.BaseLayer>
                    <LayersControl.Overlay checked name="Communes Info">
                        <LayerGroup><CityStat /></LayerGroup>
                    </LayersControl.Overlay>
                    <LayersControl.Overlay checked name="Vente">
                        <LayerGroup>
                            <LocationMarker />
                        </LayerGroup>
                    </LayersControl.Overlay>
                </LayersControl>

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