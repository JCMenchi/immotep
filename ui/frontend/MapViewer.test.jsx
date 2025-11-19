// ui/frontend/MapViewer.test.jsx
import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'

// controllable mock position for useSelector
let mockPosition = [47.906863, -3.954772]

// mock react-redux useSelector to return mockPosition
vi.mock('react-redux', () => {
  return {
    useSelector: (_) => {
      return mockPosition
    }
  }
})

// mock react-leaflet with lightweight test-friendly components
vi.mock('react-leaflet', () => {
  const React = require('react')
  const fakeMap = {
      flyTo: (...args) => {
        // ensure array exists
        if (!globalThis.__flyCalls) globalThis.__flyCalls = []
        globalThis.__flyCalls.push(args)
      }
    }

  // MapContainer: calls ref with a fake map that records flyTo calls on window.__flyCalls
  const MapContainer = ({ children, ref: refProp, ...props }) => {
    
    // call ref function if provided (the component uses ref={setMap})
    if (typeof refProp === 'function') {
      refProp(fakeMap)
    } else if (refProp && typeof refProp === 'object') {
      refProp = fakeMap
    }
    return React.createElement('div', { 'data-testid': 'mapcontainer', ...props }, children)
  }

  const LayersControl = ({ children }) =>
    React.createElement('div', { 'data-testid': 'layerscontrol' }, children)

  LayersControl.BaseLayer = ({ children, name }) =>
    React.createElement('div', { 'data-testid': `baselayer-${name || ''}` }, children)

  LayersControl.Overlay = ({ children, name }) =>
    React.createElement('div', { 'data-testid': `overlay-${name || ''}` }, children)

  const LayerGroup = ({ children }) =>
    React.createElement('div', { 'data-testid': 'layergroup' }, children)

  const TileLayer = (props) =>
    React.createElement('div', { 'data-testid': 'tilelayer', 'data-url': props.url || '' }, null)

  return {
    MapContainer,
    LayersControl,
    LayerGroup,
    TileLayer
  }
})

// mock child components used by MapViewer
vi.mock('./RegionStat', () => {
  const React = require('react')
  return { RegionStat: () => React.createElement('div', { 'data-testid': 'regionstat' }) }
})
vi.mock('./DepartmentStat', () => {
  const React = require('react')
  return { DepartmentStat: () => React.createElement('div', { 'data-testid': 'departmentstat' }) }
})
vi.mock('./CityStat', () => {
  const React = require('react')
  return { CityStat: () => React.createElement('div', { 'data-testid': 'citystat' }) }
})
vi.mock('./LocationMarker', () => {
  const React = require('react')
  return { LocationMarker: () => React.createElement('div', { 'data-testid': 'locationmarker' }) }
})

// import component under test (mocks are hoisted by vitest)
import { MapViewer } from './MapViewer'

beforeEach(() => {
  // reset recorded flyTo calls and default mock position
  globalThis.__flyCalls = []
  mockPosition = [47.906863, -3.954772]
  vi.clearAllMocks()
})

describe('MapViewer Component', () => {
  it('renders map container, layers and child components and includes expected tile URLs', async () => {
    render(<MapViewer initposition={[0, 0]} initZoom={5} />)

    // basic elements present
    expect(screen.getByTestId('mapcontainer')).toBeDefined()
    expect(screen.getByTestId('regionstat')).toBeDefined()
    expect(screen.getByTestId('departmentstat')).toBeDefined()
    expect(screen.getByTestId('citystat')).toBeDefined()
    expect(screen.getByTestId('locationmarker')).toBeDefined()

    // tile layers rendered and contain expected URL substrings
    const tiles = screen.getAllByTestId('tilelayer')
    const urls = tiles.map(t => t.getAttribute('data-url') || '')

    // expect at least one OpenStreetMap tile
    expect(urls.some(u => u.includes('openstreetmap') || u.includes('tile.openstreetmap.org'))).toBe(true)
    // expect at least one Mapbox url (constructed in component)
    expect(urls.some(u => u.includes('api.mapbox.com') || u.includes('mapbox'))).toBe(true)
  })

  it('calls flyTo again when selected position changes (rerender with new position)', async () => {
    const { rerender } = render(<MapViewer initposition={[0, 0]} initZoom={5} />)


    // change mockPosition to a new value and rerender
    mockPosition = [48.8566, 2.3522] // Paris coords
    rerender(<MapViewer initposition={mockPosition} initZoom={5} />)

    await waitFor(() => {
      // expect at least one more call recorded
      expect(globalThis.__flyCalls.length).toBeGreaterThanOrEqual(1)
    })

    const lastCallArgs = globalThis.__flyCalls[globalThis.__flyCalls.length - 1][0]
    expect(lastCallArgs).toEqual(mockPosition)
  })
})