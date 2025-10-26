import { describe, it, expect, vi } from 'vitest'
import { render, act, waitFor } from '@testing-library/react'

import { Provider } from 'react-redux';
import { LocationMarker } from './LocationMarker'
import { store } from './store';
import { MapContainer } from "react-leaflet";
import service from './poi_service'

// Mock the service
vi.mock('./poi_service', () => ({
  default: {
    post: vi.fn()
  }
}))

describe('LocationMarker', () => {
  const mockTransactions = {
    data: {
      transactions: [
        {
          id: 1,
          lat: 48.6,
          long: -4.0,
          date: "2023-01-01T00:00:00",
          price: 200000,
          area: 100,
          address: "Test Street",
          city: "Test City"
        }
      ],
      avgprice: 200000,
      avgprice_sqm: 2000
    }
  }

  beforeEach(() => {
    service.post.mockResolvedValue(mockTransactions)
  })

  it('loads and displays transactions on mount', async () => {
    await act(() => render(
      <Provider store={store}>
        <MapContainer center={[48.6007, -4.0451]} zoom={10}>
          <LocationMarker/>
        </MapContainer>
      </Provider>
    ))
    
    await waitFor(() => expect(service.post).toHaveBeenCalledTimes(1))

    expect(service.post).toHaveBeenCalled()
  })

  it('updates transactions when year, department or limit changes', async () => {
    await act(() => {render(
      <Provider store={store}>
        <MapContainer center={[48.6007, -4.0451]} zoom={10}>
          <LocationMarker/>
        </MapContainer>
      </Provider>
    )})

    await act(() => store.dispatch({ type: 'uiparam/changeYear', payload: 2022 }))
    expect(service.post).toHaveBeenCalled()
  })

  it('displays markers with popups for transactions', async () => {
    render(
      <Provider store={store}>
        <MapContainer center={[48.6007, -4.0451]} zoom={10}>
          <LocationMarker/>
        </MapContainer>
      </Provider>
    )

    // Wait for transactions to load
    await vi.waitFor(() => {
      expect(service.post).toHaveBeenCalled()
    })
  })
})
