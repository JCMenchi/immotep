import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'

import { http, HttpResponse } from 'msw'
import { server } from '../mock/httpserver'

import { DepartmentStat, computeDepartmentContourStyle } from './DepartmentStat'

import service from './poi_service'

// mock react-leaflet components to avoid requiring a real Leaflet map in tests

vi.mock('react-leaflet', () => {
  const React = require('react')
  return {
    GeoJSON: ({ children }) => React.createElement('div', { 'data-testid': 'geojson' }, children),
    Tooltip: ({ children }) => React.createElement('div', { 'data-testid': 'tooltip' }, children),
  }
})


beforeEach(() => {
  vi.clearAllMocks()
})

describe('DepartmentStat Component', () => {
  it('renders department tooltip when service returns data', async () => {
    const getSpy = vi.spyOn(service, 'get')

    render(
      <DepartmentStat />
    )

    const tooltip = await screen.findByTestId('tooltip')
    const text = tooltip.textContent || ''

    //screen.debug()

    // avgprice is formatted with toFixed(0) in the component -> 1235€
    expect(text).toContain('(01) DeptA: 1235€')
    expect(text).toContain('2020: 10')
    expect(text).toContain('2021: 5')

    // verify service is called correctly
    expect(getSpy).toHaveBeenCalledWith('api/departments')
  })

  it('renders nothing (no GeoJSON) when service returns empty array', async () => {
    server.use(
      http.get('api/departments', () => {
        return HttpResponse.json([])
      })
    )

    render(
      <DepartmentStat />
    )

    // wait a short while for the effect to complete and assert no geojson was rendered
    await waitFor(() => {
      expect(screen.queryByTestId('geojson')).toBeNull()
    })
  })

  it('renders nothing (no GeoJSON) when service returns an error', async () => {
    server.use(
      http.get('api/departments', () => {
        return HttpResponse.error(500)
      })
    )

    render(
      <DepartmentStat />
    )

    // wait a short while for the effect to complete and assert no geojson was rendered
    await waitFor(() => {
      expect(screen.queryByTestId('geojson')).toBeNull()
    })
  })

})


describe('DepartmentStat Styling', () => {
  it('choose color from feature', () => {
   
    const st = computeDepartmentContourStyle({ type: 'Feature', geometry: null, properties: { avgprice: 2500} })

    // check some style values
    expect(st.weight).toBe(1)
    expect(st.color).toBe('red')
    expect(st.fillColor).toBe('#ef6548')

  })

  it('choose min color from feature', () => {
   
    const st = computeDepartmentContourStyle({ type: 'Feature', geometry: null, properties: { avgprice: 10} })

    // check some style values
    expect(st.weight).toBe(1)
    expect(st.color).toBe('red')
    expect(st.fillColor).toBe('#fff7ec')

  })

  it('choose max color from feature', () => {
   
    const st = computeDepartmentContourStyle({ type: 'Feature', geometry: null, properties: { avgprice: 3100} })

    // check some style values
    expect(st.weight).toBe(1)
    expect(st.color).toBe('red')
    expect(st.fillColor).toBe('#7f0000')

  })

  it('choose default color if feature not defined', () => {
   
    const st = computeDepartmentContourStyle()

    // check some style values
    expect(st.weight).toBe(1)
    expect(st.color).toBe('red')
    expect(st.fillColor).toBe('grey')

  })
})