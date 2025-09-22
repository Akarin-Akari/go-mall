import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { User, LoginRequest, RegisterRequest, AuthState } from '@/types';
import { authAPI } from '@/services/api';
import { secureTokenManager } from '@/utils/secureTokenManager';
import { message } from 'antd';

// 鍒濆鐘舵€?
const initialState: AuthState = {
  user: null,
  token: null,
  isAuthenticated: false,
  loading: false,
  error: null,
};

// 寮傛actions
export const loginAsync = createAsyncThunk(
  'auth/login',
  async (
    loginData: LoginRequest & { remember?: boolean },
    { rejectWithValue }
  ) => {
    try {
      const response = await authAPI.login(loginData);
      const { user, token, refresh_token } = response.data;

      // 保存用户信息和token
      secureTokenManager.setAccessToken(token);
      if (refresh_token) {
        secureTokenManager.setRefreshToken(refresh_token);
      }

      return { user, token };
    } catch (error: unknown) {
      return rejectWithValue(
        error instanceof Error ? error.message : '鐧诲綍澶辫触'
      );
    }
  }
);

export const registerAsync = createAsyncThunk(
  'auth/register',
  async (registerData: RegisterRequest, { rejectWithValue }) => {
    try {
      const response = await authAPI.register(registerData);
      const { user, token, refresh_token } = response.data;

      // 保存用户信息和token
      secureTokenManager.setAccessToken(token);
      if (refresh_token) {
        secureTokenManager.setRefreshToken(refresh_token);
      }

      return { user, token };
    } catch (error: unknown) {
      return rejectWithValue(
        error instanceof Error ? error.message : '娉ㄥ唽澶辫触'
      );
    }
  }
);

export const logoutAsync = createAsyncThunk(
  'auth/logout',
  async (_, { rejectWithValue }) => {
    try {
      await authAPI.logout();
      secureTokenManager.clearTokens();
      return null;
    } catch (error: any) {
      // 即使API调用失败，也要清除本地token
      secureTokenManager.clearTokens();
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
      return rejectWithValue(error.message || '鑾峰彇鐢ㄦ埛淇℃伅澶辫触');
    }
  }
);

export const refreshTokenAsync = createAsyncThunk(
  'auth/refreshToken',
  async (_, { rejectWithValue }) => {
    try {
      const refreshToken = secureTokenManager.getRefreshToken();
      if (!refreshToken) {
        throw new Error('娌℃湁鍒锋柊token');
      }

      const response = await authAPI.refreshToken(refreshToken);
      const { token, refresh_token } = response.data;

      secureTokenManager.setAccessToken(token);
      if (refresh_token) {
        secureTokenManager.setRefreshToken(refresh_token);
      }

      return { token };
    } catch (error: any) {
      secureTokenManager.clearTokens();
      return rejectWithValue(error.message || 'Token鍒锋柊澶辫触');
    }
  }
);

// 鍒涘缓slice
const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    // 璁剧疆鐢ㄦ埛淇℃伅
    setUser: (state, action: PayloadAction<User | null>) => {
      state.user = action.payload;
      state.isAuthenticated = !!action.payload;
    },

    // 璁剧疆token
    setToken: (state, action: PayloadAction<string | null>) => {
      state.token = action.payload;
      state.isAuthenticated = !!action.payload && !!state.user;
    },

    // 鏇存柊鐢ㄦ埛淇℃伅
    updateUser: (state, action: PayloadAction<Partial<User>>) => {
      if (state.user) {
        state.user = { ...state.user, ...action.payload };
      }
    },

    // 娓呴櫎璁よ瘉鐘舵€?
    clearAuth: state => {
      state.user = null;
      state.token = null;
      state.isAuthenticated = false;
      state.error = null;
    },

    // 清除错误状态
    clearError: state => {
      state.error = null;
    },

    // 浠庢湰鍦板瓨鍌ㄦ仮澶嶇姸鎬?
    restoreAuth: state => {
      const token = secureTokenManager.getAccessToken();
      if (token) {
        state.token = token;
        // 娉ㄦ剰锛氳繖閲屼笉璁剧疆isAuthenticated涓簍rue锛岄渶瑕侀獙璇乼oken鏈夋晥鎬?
      }
    },
  },
  extraReducers: builder => {
    // 鐧诲綍
    builder
      .addCase(loginAsync.pending, state => {
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

    // 娉ㄥ唽
    builder
      .addCase(registerAsync.pending, state => {
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

    // 鐧诲嚭
    builder
      .addCase(logoutAsync.pending, state => {
        state.loading = true;
      })
      .addCase(logoutAsync.fulfilled, state => {
        state.loading = false;
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
      })
      .addCase(logoutAsync.rejected, state => {
        state.loading = false;
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
      });

    // 鑾峰彇鐢ㄦ埛淇℃伅
    builder
      .addCase(getProfileAsync.pending, state => {
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
        secureTokenManager.clearTokens();
      });

    // 鍒锋柊token
    builder
      .addCase(refreshTokenAsync.fulfilled, (state, action) => {
        state.token = action.payload.token;
      })
      .addCase(refreshTokenAsync.rejected, state => {
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
      });
  },
});

// 瀵煎嚭actions
export const {
  setUser,
  setToken,
  updateUser,
  clearAuth,
  clearError,
  restoreAuth,
} = authSlice.actions;

// 閫夋嫨鍣?
export const selectAuth = (state: { auth: AuthState }) => state.auth;
export const selectUser = (state: { auth: AuthState }) => state.auth.user;
export const selectIsAuthenticated = (state: { auth: AuthState }) =>
  state.auth.isAuthenticated;
export const selectAuthLoading = (state: { auth: AuthState }) =>
  state.auth.loading;

// 瀵煎嚭reducer
export default authSlice.reducer;
