
import { createSlice } from '@reduxjs/toolkit';

export const uiParam = createSlice({
  name: 'uiParam',
  initialState: {
    queryLimit: localStorage.getItem('query_limit') || 100,
    queryDepartment: localStorage.getItem('query_department') || "29",
    theme: localStorage.getItem('ui_theme') || "dark",
    lang: localStorage.getItem('ui_lang') || "",
    showMark: localStorage.getItem('ui_mark') || false,
    avgPrice: 0
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
    changeShowMark(state, action) {
      state.showMark = action.payload;
      localStorage.setItem('ui_mark', state.showMark);
    },
    changeAvgPrice(state, action) {
      state.avgPrice = action.payload;
    }
  }
});

export const { changeQueryLimit, changeQueryDepartment, changeUITheme, changeUILanguage, changeShowMark, changeAvgPrice } = uiParam.actions;

// some selector
export const selectUITheme = state => state.uiParam.theme;
export const selectUILanguage = state => state.uiParam.lang;
export const selectQueryLimit = state => state.uiParam.queryLimit;
export const selectQueryDepartment = state => state.uiParam.queryDepartment;
export const selectUIShowMark = state => state.uiParam.showMark;
export const selectAvgPrice = state => state.uiParam.avgPrice;