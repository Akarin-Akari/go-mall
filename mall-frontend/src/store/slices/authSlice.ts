import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { User, LoginRequest, RegisterRequest, AuthState } from '@/types';
import { authAPI } from '@/services/api';
import { tokenManager } from '@/utils';
import { message } from 'antd';

// 初始状态
const initialState: AuthState = {
  user: null,
  token: null,
  isAuthenticated: false,
  loading: false,
};

// 异步actions
export const loginAsync = createAsyncThunk(
  'auth/login',
  async (loginData: LoginRequest & { remember?: boolean }, { rejectWithValue }) => {
    try {
      const response = await authAPI.login(loginData);
      const { user, token, refresh_token } = response.data;
      
      // 保存token
      tokenManager.setToken(token, loginData.remember);
      if (refresh_token) {
        tokenManager.setRefreshToken(refresh_token);
      }
      
      return { user, token };
    } catch (error: any) {
      return rejectWithValue(error.message || '登录失败');
    }
  }
);

export const registerAsync = createAsyncThunk(
  'auth/register',
  async (registerData: RegisterRequest, { rejectWithValue }) => {
    try {
      const response = await authAPI.register(registerData);
      const { user, token, refresh_token } = response.data;
      
      // 保存token
      tokenManager.setToken(token);
      if (refresh_token) {
        tokenManager.setRefreshToken(refresh_token);
      }
      
      return { user, token };
    } catch (error: any) {
      return rejectWithValue(error.message || '注册失败');
    }
  }
);

export const logoutAsync = createAsyncThunk(
  'auth/logout',
  async (_, { rejectWithValue }) => {
    try {
      await authAPI.logout();
      tokenManager.clearAll();
      return null;
    } catch (error: any) {
      // 即使API调用失败，也要清除本地token
      tokenManager.clearAll();
      return null;
    }
  }
);

export const getProfileAsync = createAsyncThunk(
  'auth/getProfile',
  async (_, { rejectWithValue }) => {
    try {
      const response = await authAPI.getProfile();
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取用户信息失败');
    }
  }
);

export const refreshTokenAsync = createAsyncThunk(
  'auth/refreshToken',
  async (_, { rejectWithValue }) => {
    try {
      const refreshToken = tokenManager.getRefreshToken();
      if (!refreshToken) {
        throw new Error('没有刷新token');
      }
      
      const response = await authAPI.refreshToken(refreshToken);
      const { token, refresh_token } = response.data;
      
      tokenManager.setToken(token);
      if (refresh_token) {
        tokenManager.setRefreshToken(refresh_token);
      }
      
      return { token };
    } catch (error: any) {
      tokenManager.clearAll();
      return rejectWithValue(error.message || 'Token刷新失败');
    }
  }
);

// 创建slice
const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    // 设置用户信息
    setUser: (state, action: PayloadAction<User | null>) => {
      state.user = action.payload;
      state.isAuthenticated = !!action.payload;
    },
    
    // 设置token
    setToken: (state, action: PayloadAction<string | null>) => {
      state.token = action.payload;
      state.isAuthenticated = !!action.payload && !!state.user;
    },
    
    // 更新用户信息
    updateUser: (state, action: PayloadAction<Partial<User>>) => {
      if (state.user) {
        state.user = { ...state.user, ...action.payload };
      }
    },
    
    // 清除认证状态
    clearAuth: (state) => {
      state.user = null;
      state.token = null;
      state.isAuthenticated = false;
      state.loading = false;
    },
    
    // 从本地存储恢复状态
    restoreAuth: (state) => {
      const token = tokenManager.getToken();
      if (token) {
        state.token = token;
        // 注意：这里不设置isAuthenticated为true，需要验证token有效性
      }
    },
  },
  extraReducers: (builder) => {
    // 登录
    builder
      .addCase(loginAsync.pending, (state) => {
        state.loading = true;
      })
      .addCase(loginAsync.fulfilled, (state, action) => {
        state.loading = false;
        state.user = action.payload.user;
        state.token = action.payload.token;
        state.isAuthenticated = true;
      })
      .addCase(loginAsync.rejected, (state, action) => {
        state.loading = false;
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
        message.error(action.payload as string);
      });
    
    // 注册
    builder
      .addCase(registerAsync.pending, (state) => {
        state.loading = true;
      })
      .addCase(registerAsync.fulfilled, (state, action) => {
        state.loading = false;
        state.user = action.payload.user;
        state.token = action.payload.token;
        state.isAuthenticated = true;
      })
      .addCase(registerAsync.rejected, (state, action) => {
        state.loading = false;
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
        message.error(action.payload as string);
      });
    
    // 登出
    builder
      .addCase(logoutAsync.pending, (state) => {
        state.loading = true;
      })
      .addCase(logoutAsync.fulfilled, (state) => {
        state.loading = false;
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
      })
      .addCase(logoutAsync.rejected, (state) => {
        state.loading = false;
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
      });
    
    // 获取用户信息
    builder
      .addCase(getProfileAsync.pending, (state) => {
        state.loading = true;
      })
      .addCase(getProfileAsync.fulfilled, (state, action) => {
        state.loading = false;
        state.user = action.payload;
        state.isAuthenticated = true;
      })
      .addCase(getProfileAsync.rejected, (state, action) => {
        state.loading = false;
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
        tokenManager.clearAll();
      });
    
    // 刷新token
    builder
      .addCase(refreshTokenAsync.fulfilled, (state, action) => {
        state.token = action.payload.token;
      })
      .addCase(refreshTokenAsync.rejected, (state) => {
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
      });
  },
});

// 导出actions
export const { setUser, setToken, updateUser, clearAuth, restoreAuth } = authSlice.actions;

// 选择器
export const selectAuth = (state: { auth: AuthState }) => state.auth;
export const selectUser = (state: { auth: AuthState }) => state.auth.user;
export const selectIsAuthenticated = (state: { auth: AuthState }) => state.auth.isAuthenticated;
export const selectAuthLoading = (state: { auth: AuthState }) => state.auth.loading;

// 导出reducer
export default authSlice.reducer;
