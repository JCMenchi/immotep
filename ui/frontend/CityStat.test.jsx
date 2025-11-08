import { describe, it, expect } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../mock/httpserver'

import { Provider } from 'react-redux';
import { CityStat, computeCityContourStyle } from './CityStat'
import { store } from './store';

import service from './poi_service'

// mock react-leaflet components to avoid requiring a real Leaflet map in tests

vi.mock('react-leaflet', () => {
  const React = require('react')
  return {
    GeoJSON: ({ children }) => React.createElement('div', { 'data-testid': 'geojson' }, children),
    Tooltip: ({ children }) => React.createElement('div', { 'data-testid': 'tooltip', 'key': 1 }, children),
    useMapEvents: () => { console.log('useMapEvents') }
  }
})

describe('CityStat Component', () => {
  it('renders correctly with children', async () => {

    const getSpy = vi.spyOn(service, 'get')

    render(<Provider store={store}>
      <CityStat />
    </Provider>)


    const tooltip = await screen.findByTestId('tooltip')
    const text = tooltip.textContent || ''

    expect(text).toContain('(12345) Ville: 1235€')
    expect(text).toContain('2020: 10')
    expect(text).toContain('2021: 5')

    // verify service is called correctly
    expect(getSpy).toHaveBeenCalledWith('api/cities?limit=600')
  })

  it('renders nothing (no GeoJSON) when service returns an error', async () => {
    server.use(
      http.get('api/departments', () => {
        return HttpResponse.error(500)
      })
    )

    render(
      <Provider store={store}>
        <CityStat />
      </Provider>
    )

    // wait a short while for the effect to complete and assert no geojson was rendered
    await waitFor(() => {
      expect(screen.queryByTestId('geojson')).toBeNull()
    })
  })
})

describe('CityStat Styling', () => {
  it('choose color from feature', () => {

    const st = computeCityContourStyle({ type: 'Feature', geometry: null, properties: { avgprice: 2500 } })

    // check some style values
    expect(st.weight).toBe(1)
    expect(st.color).toBe('red')
    expect(st.fillColor).toBe('#ef6548')

  })

  it('choose default color if feature not defined', () => {

    const st = computeCityContourStyle()

    // check some style values
    expect(st.weight).toBe(1)
    expect(st.color).toBe('red')
    expect(st.fillColor).toBe('grey')

  })
})