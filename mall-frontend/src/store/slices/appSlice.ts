import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { AppState } from '@/types';

// 初始状态
const initialState: AppState = {
  theme: 'light',
  collapsed: false,
  loading: false,
  locale: 'zh-CN',
};

// 创建slice
const appSlice = createSlice({
  name: 'app',
  initialState,
  reducers: {
    // 切换主题
    toggleTheme: (state) => {
      state.theme = state.theme === 'light' ? 'dark' : 'light';
    },
    
    // 设置主题
    setTheme: (state, action: PayloadAction<'light' | 'dark'>) => {
      state.theme = action.payload;
    },
    
    // 切换侧边栏折叠状态
    toggleCollapsed: (state) => {
      state.collapsed = !state.collapsed;
    },
    
    // 设置侧边栏折叠状态
    setCollapsed: (state, action: PayloadAction<boolean>) => {
      state.collapsed = action.payload;
    },
    
    // 设置全局加载状态
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    
    // 设置语言
    setLocale: (state, action: PayloadAction<'zh-CN' | 'en-US'>) => {
      state.locale = action.payload;
    },
  },
});

// 导出actions
export const {
  toggleTheme,
  setTheme,
  toggleCollapsed,
  setCollapsed,
  setLoading,
  setLocale,
} = appSlice.actions;

// 选择器
export const selectApp = (state: { app: AppState }) => state.app;
export const selectTheme = (state: { app: AppState }) => state.app.theme;
export const selectCollapsed = (state: { app: AppState }) => state.app.collapsed;
export const selectLoading = (state: { app: AppState }) => state.app.loading;
export const selectLocale = (state: { app: AppState }) => state.app.locale;

// 导出reducer
export default appSlice.reducer;
