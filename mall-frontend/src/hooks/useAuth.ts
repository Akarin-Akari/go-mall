import { useCallback, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { message } from 'antd';
import { useAppSelector, useAppDispatch } from '@/store';
import {
  selectAuth,
  loginAsync,
  registerAsync,
  logoutAsync,
  getProfileAsync,
  refreshTokenAsync,
  clearAuth,
  restoreAuth,
} from '@/store/slices/authSlice';
import { secureTokenManager } from '@/utils/secureTokenManager';
import { authUtils } from '@/utils/auth';
import { LoginRequest, RegisterRequest, User } from '@/types';
import { ROUTES } from '@/constants';

export interface UseAuthReturn {
  // 状态
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  
  // 方法
  login: (data: LoginRequest) => Promise<boolean>;
  register: (data: RegisterRequest) => Promise<boolean>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<boolean>;
  getProfile: () => Promise<User | null>;
  checkAuth: () => boolean;
  hasRole: (role: string) => boolean;
  isAdmin: () => boolean;
  
  // 工具方法
  redirectToLogin: (returnUrl?: string) => void;
  restoreAuthState: () => void;
  clearAuthState: () => void;
}

/**
 * 认证钩子
 * 提供完整的认证功能和状态管理
 */
export const useAuth = (): UseAuthReturn => {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const { user, isAuthenticated, loading, error } = useAppSelector(selectAuth);

  // 登录
  const login = useCallback(async (data: LoginRequest): Promise<boolean> => {
    try {
      const result = await dispatch(loginAsync(data));
      
      if (loginAsync.fulfilled.match(result)) {
        message.success('登录成功！');
        return true;
      } else {
        const errorMsg = result.payload as string || '登录失败';
        message.error(errorMsg);
        return false;
      }
    } catch (error) {
      console.error('Login error:', error);
      message.error('登录过程中发生错误');
      return false;
    }
  }, [dispatch]);

  // 注册
  const register = useCallback(async (data: RegisterRequest): Promise<boolean> => {
    try {
      const result = await dispatch(registerAsync(data));
      
      if (registerAsync.fulfilled.match(result)) {
        message.success('注册成功！欢迎加入！');
        return true;
      } else {
        const errorMsg = result.payload as string || '注册失败';
        message.error(errorMsg);
        return false;
      }
    } catch (error) {
      console.error('Register error:', error);
      message.error('注册过程中发生错误');
      return false;
    }
  }, [dispatch]);

  // 登出
  const logout = useCallback(async (): Promise<void> => {
    try {
      await dispatch(logoutAsync());
      message.success('已安全退出');
      router.push(ROUTES.HOME);
    } catch (error) {
      console.error('Logout error:', error);
      // 即使API调用失败，也要清除本地状态
      dispatch(clearAuth());
      secureTokenManager.clearTokens();
      authUtils.clearUser();
      router.push(ROUTES.HOME);
    }
  }, [dispatch, router]);

  // 刷新Token
  const refreshToken = useCallback(async (): Promise<boolean> => {
    try {
      const result = await dispatch(refreshTokenAsync());
      
      if (refreshTokenAsync.fulfilled.match(result)) {
        return true;
      } else {
        // Token刷新失败，清除认证状态
        dispatch(clearAuth());
        secureTokenManager.clearTokens();
        authUtils.clearUser();
        return false;
      }
    } catch (error) {
      console.error('Refresh token error:', error);
      dispatch(clearAuth());
      secureTokenManager.clearTokens();
      authUtils.clearUser();
      return false;
    }
  }, [dispatch]);

  // 获取用户信息
  const getProfile = useCallback(async (): Promise<User | null> => {
    try {
      const result = await dispatch(getProfileAsync());
      
      if (getProfileAsync.fulfilled.match(result)) {
        return result.payload;
      } else {
        console.error('Get profile failed:', result.payload);
        return null;
      }
    } catch (error) {
      console.error('Get profile error:', error);
      return null;
    }
  }, [dispatch]);

  // 检查认证状态
  const checkAuth = useCallback((): boolean => {
    const token = secureTokenManager.getAccessToken();
    const currentUser = authUtils.getCurrentUser();
    return !!(token && currentUser && isAuthenticated);
  }, [isAuthenticated]);

  // 检查用户角色
  const hasRole = useCallback((role: string): boolean => {
    return user?.role === role || false;
  }, [user]);

  // 检查是否为管理员
  const isAdmin = useCallback((): boolean => {
    return hasRole('admin');
  }, [hasRole]);

  // 重定向到登录页
  const redirectToLogin = useCallback((returnUrl?: string) => {
    const currentUrl = returnUrl || window.location.pathname;
    const loginUrl = `${ROUTES.LOGIN}?redirect=${encodeURIComponent(currentUrl)}`;
    router.push(loginUrl);
  }, [router]);

  // 恢复认证状态
  const restoreAuthState = useCallback(() => {
    dispatch(restoreAuth());
  }, [dispatch]);

  // 清除认证状态
  const clearAuthState = useCallback(() => {
    dispatch(clearAuth());
    secureTokenManager.clearTokens();
    authUtils.clearUser();
  }, [dispatch]);

  // 自动刷新Token
  useEffect(() => {
    if (!isAuthenticated) return;

    const token = secureTokenManager.getAccessToken();
    if (!token) return;

    // 解析Token过期时间（这里简化处理，实际应该解析JWT）
    // 在Token过期前5分钟尝试刷新
    const refreshInterval = setInterval(async () => {
      const refreshTokenValue = secureTokenManager.getRefreshToken();
      if (refreshTokenValue) {
        const success = await refreshToken();
        if (!success) {
          // 刷新失败，重定向到登录页
          redirectToLogin();
        }
      }
    }, 25 * 60 * 1000); // 25分钟检查一次

    return () => clearInterval(refreshInterval);
  }, [isAuthenticated, refreshToken, redirectToLogin]);

  // 监听存储变化（多标签页同步）
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'access_token') {
        if (!e.newValue) {
          // Token被清除，同步登出状态
          dispatch(clearAuth());
        } else if (e.newValue && !isAuthenticated) {
          // Token被设置，尝试恢复认证状态
          restoreAuthState();
        }
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, [dispatch, isAuthenticated, restoreAuthState]);

  return {
    // 状态
    user,
    isAuthenticated,
    loading,
    error,
    
    // 方法
    login,
    register,
    logout,
    refreshToken,
    getProfile,
    checkAuth,
    hasRole,
    isAdmin,
    
    // 工具方法
    redirectToLogin,
    restoreAuthState,
    clearAuthState,
  };
};
