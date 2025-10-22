import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, within, waitFor } from '@testing-library/react'

import { Provider } from 'react-redux';
import Menubar from './Menubar'
import { store } from './store';
import * as uiparamSlice from './store/uiparamSlice';

describe('Menubar', () => {
  it('renders correctly with children', () => {
    render(<Provider store={store}><Menubar/></Provider>)

    expect(screen.getByText('Go')).toBeInTheDocument()
  })

  it('updates the address when typing in the address field', () => {
    render(<Provider store={store}><Menubar /></Provider>);
    const addressInput = screen.getByLabelText('Address');
    fireEvent.change(addressInput, { target: { value: '123 Main St' } });
    expect(addressInput.value).toBe('123 Main St');
  });

  it('dispatches changeQueryLimit action when limit is entered and Enter key is pressed', () => {
    const dispatch = vi.spyOn(store, 'dispatch');
    render(<Provider store={store}><Menubar /></Provider>);
    const limitInput = screen.getByLabelText('Limit');
    fireEvent.change(limitInput, { target: { value: '100' } });
    fireEvent.keyUp(limitInput, { key: 'Enter' });
    expect(dispatch).toHaveBeenCalledWith(uiparamSlice.changeQueryLimit(100));
  });

  it('dispatches changeQueryDepartment action when department is entered and Enter key is pressed', () => {
    const dispatch = vi.spyOn(store, 'dispatch');
    render(<Provider store={store}><Menubar /></Provider>);
    const departmentInput = screen.getByLabelText('Departement');
    fireEvent.change(departmentInput, { target: { value: '75' } });
    fireEvent.keyUp(departmentInput, { key: 'Enter' });
    expect(dispatch).toHaveBeenCalledWith(uiparamSlice.changeQueryDepartment('75'));
  });

  it('dispatches changeQueryLimit action when limit is entered and Enter key is pressed', () => {
    const dispatch = vi.spyOn(store, 'dispatch');
    render(<Provider store={store}><Menubar /></Provider>);
    const limitInput = screen.getByLabelText('Limit');
    fireEvent.change(limitInput, { target: { value: '200' } });
    fireEvent.keyUp(limitInput, { key: 'Enter' });
    expect(dispatch).toHaveBeenCalledWith(uiparamSlice.changeQueryLimit(200));
  });

  it('search address bar', async () => {
    const dispatch = vi.spyOn(store, 'dispatch');
    render(<Provider store={store}><Menubar /></Provider>);
    const addressInput = screen.getByLabelText('Address');
    fireEvent.change(addressInput, { target: { value: 'Avenue Gustave Eiffel Paris' } });
    fireEvent.keyUp(addressInput, { key: 'Enter' });
    await waitFor(() => expect(dispatch).toHaveBeenCalledTimes(1))
    expect(dispatch).toHaveBeenCalledWith(uiparamSlice.changePosition([ 48.857739, 2.294844 ]));
  });

  it('dispatches changeUITheme action when theme button is clicked', () => {
    const dispatch = vi.spyOn(store, 'dispatch');
    render(<Provider store={store}><Menubar /></Provider>);
    const themeButton = document.getElementById('change-theme-button');
    fireEvent.click(themeButton);
    expect(dispatch).toHaveBeenCalledWith(uiparamSlice.changeUITheme('light'));

    // toggle back to dark
    fireEvent.click(themeButton);
    expect(dispatch).toHaveBeenCalledWith(uiparamSlice.changeUITheme('dark'));
  });

  it('dispatches changeYear action when year is selected', () => {
    const dispatch = vi.spyOn(store, 'dispatch');

    render(<Provider store={store}><Menubar /></Provider>);

    // click to open the select dropdown
    const yearSelect = document.getElementById('select-year');
    fireEvent.mouseDown(yearSelect);
    
    // get dropdown listbox and click on 2020 item
    const listbox = screen.getByRole('listbox');
    const item = within(listbox).getByText('2020');
    fireEvent.click(item);

    expect(dispatch).toHaveBeenCalledWith(uiparamSlice.changeYear(2020));
  });
})
