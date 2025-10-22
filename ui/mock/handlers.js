import { http, HttpResponse } from 'msw'

export const handlers = [

  http.get('api/regions', () => {
    return HttpResponse.json([
      {
        name: 'RegA',
        code: '111',
        avgprice: 1234.56,
        contour: { type: 'Feature', geometry: null, properties: { avgprice: 2000 } },
        stat: { '2020': 10, '2021': 5 }
      }
    ])
  }),

  http.get('api/departments', () => {
    return HttpResponse.json([
      {
        name: 'DeptA',
        code: '01',
        avgprice: 1234.56,
        contour: { type: 'Feature', geometry: null, properties: { avgprice: 2000 } },
        stat: { '2020': 10, '2021': 5 }
      }
    ])
  }),

  http.get('api/cities?limit=600&dep=29', () => {
    return HttpResponse.json([
      {
        name: 'Ville',
        zip: '12345',
        avgprice: 1234.56,
        population: 10000,
        contour: { type: 'Feature', geometry: null, properties: { avgprice: 2000 } },
        stat: { '2020': 10, '2021': 5 }
      }
    ])
  }),

  http.get('https://data.geopf.fr/geocodage/search/?q=Avenue%20Gustave%20Eiffel%20Paris', () => {
    return HttpResponse.json({
      "type": "FeatureCollection",
      "features": [
        {
          "type": "Feature",
          "geometry": {
            "type": "Point",
            "coordinates": [
              2.294844,
              48.857739
            ]
          },
          "properties": {
            "label": "Avenue Gustave Eiffel 75007 Paris",
            "score": 0.9594127272727271,
            "id": "75107_4409",
            "name": "Avenue Gustave Eiffel",
            "postcode": "75007",
            "citycode": "75107",
            "x": 648261.88,
            "y": 6862197.96,
            "city": "Paris",
            "district": "Paris 7e Arrondissement",
            "context": "75, Paris, Île-de-France",
            "type": "street",
            "importance": 0.55354,
            "street": "Avenue Gustave Eiffel",
            "_type": "address"
          }
        }
      ],
      "query": "Avenue Gustave Eiffel Paris"
    })

  }),
]
