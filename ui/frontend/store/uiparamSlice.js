
import { createSlice } from '@reduxjs/toolkit';

export const uiParam = createSlice({
  name: 'uiParam',
  initialState: {
    queryLimit: localStorage.getItem('query_limit') || 50,
    queryDepartment: localStorage.getItem('query_department') || "",
    theme: localStorage.getItem('ui_theme') || "dark",
    lang: localStorage.getItem('ui_lang') || "",
    currentPosition: JSON.parse(localStorage.getItem('ui_center')) || [48.6007, -4.0451],
    zoom: localStorage.getItem('ui_zoom') || 10,
    avgPrice: -1,
    avgPriceSQM: -1,
    year: -1
  },
  reducers: {
    changeQueryLimit(state, action) {
      state.queryLimit = action.payload;
      localStorage.setItem('query_limit', state.queryLimit);
    },
    changeQueryDepartment(state, action) {
      state.queryDepartment = action.payload;
      localStorage.setItem('query_department', state.queryDepartment);
    },
    changeUITheme(state, action) {
      state.theme = action.payload;
      localStorage.setItem('ui_theme', state.theme);
    },
    changeUILanguage(state, action) {
      state.lang = action.payload;
      localStorage.setItem('ui_lang', state.lang);
    },
    changePosition(state, action) {
      state.currentPosition = action.payload;
      localStorage.setItem('ui_center', JSON.stringify(state.currentPosition));
    },
    changeZoom(state, action) {
      state.zoom = action.payload;
      localStorage.setItem('ui_zoom', state.zoom);
    },
    changeAvgPrice(state, action) {
      state.avgPrice = action.payload;
    },
    changeAvgPriceSQM(state, action) {
      state.avgPriceSQM = action.payload;
    },
    changeYear(state, action) {
      state.year = action.payload;
    }
  }
});

export const { changeQueryLimit, changeQueryDepartment,
  changeUITheme, changeUILanguage,
  changePosition, changeZoom, changeYear,
  changeAvgPrice, changeAvgPriceSQM } = uiParam.actions;

// some selector
export const selectUITheme = state => state.uiParam.theme;
export const selectUILanguage = state => state.uiParam.lang;
export const selectCenterPosition = state => state.uiParam.currentPosition;
export const selectZoom = state => state.uiParam.zoom;
export const selectQueryLimit = state => state.uiParam.queryLimit;
export const selectQueryDepartment = state => state.uiParam.queryDepartment;
export const selectAvgPrice = state => state.uiParam.avgPrice;
export const selectAvgPriceSQM = state => state.uiParam.avgPriceSQM;
export const selectYear= state => state.uiParam.year;