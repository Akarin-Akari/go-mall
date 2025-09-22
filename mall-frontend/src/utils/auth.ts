import { User } from '@/types';
import { STORAGE_KEYS, USER_ROLES } from '@/constants';
import { secureTokenManager } from './secureTokenManager';

// 简化的本地存储工具
const storage = {
  get: (key: string): string | null => {
    if (typeof window === 'undefined') return null;
    try {
      return localStorage.getItem(key);
    } catch {
      return null;
    }
  },

  set: (key: string, value: string): void => {
    if (typeof window === 'undefined') return;
    try {
      localStorage.setItem(key, value);
    } catch {
      // 忽略存储错误
    }
  },

  remove: (key: string): void => {
    if (typeof window === 'undefined') return;
    try {
      localStorage.removeItem(key);
    } catch {
      // 忽略存储错误
    }
  },

  getJSON: <T>(key: string): T | null => {
    const value = storage.get(key);
    if (!value) return null;
    try {
      return JSON.parse(value) as T;
    } catch {
      return null;
    }
  },

  setJSON: <T>(key: string, value: T): void => {
    try {
      storage.set(key, JSON.stringify(value));
    } catch {
      // 忽略存储错误
    }
  },
};

// 简化的认证工具函数
export const authUtils = {
  // 获取当前用户
  getCurrentUser: (): User | null => {
    return storage.getJSON<User>(STORAGE_KEYS.USER_INFO);
  },

  // 保存用户信息
  saveUser: (user: User): void => {
    storage.setJSON(STORAGE_KEYS.USER_INFO, user);
  },

  // 清除用户信息
  clearUser: (): void => {
    storage.remove(STORAGE_KEYS.USER_INFO);
  },

  // 检查是否已登录
  isAuthenticated: (): boolean => {
    const token = secureTokenManager.getAccessToken();
    const user = storage.getJSON<User>(STORAGE_KEYS.USER_INFO);
    return !!(token && user);
  },

  // 检查用户角色
  hasRole: (role: string): boolean => {
    const user = storage.getJSON<User>(STORAGE_KEYS.USER_INFO);
    return user?.role === role;
  },

  // 检查是否为管理员
  isAdmin: (): boolean => {
    return authUtils.hasRole(USER_ROLES.ADMIN);
  },

  // 登录
  login: (user: User, token: string, refreshToken?: string): void => {
    secureTokenManager.setAccessToken(token);
    if (refreshToken) {
      secureTokenManager.setRefreshToken(refreshToken);
    }
    storage.setJSON(STORAGE_KEYS.USER_INFO, user);
  },

  // 登出
  logout: (): void => {
    secureTokenManager.clearTokens();
    storage.remove(STORAGE_KEYS.USER_INFO);
  },
};

// 简化的AuthManager类 - 向后兼容
export class AuthManager {
  private static instance: AuthManager;

  private constructor() {}

  public static getInstance(): AuthManager {
    if (!AuthManager.instance) {
      AuthManager.instance = new AuthManager();
    }
    return AuthManager.instance;
  }

  // 登录
  login(
    user: User,
    token: string,
    refreshToken?: string,
    remember = false
  ): void {
    authUtils.login(user, token, refreshToken);
  }

  // 登出
  logout(): void {
    authUtils.logout();
  }

  // 获取用户信息
  getUser(): User | null {
    return authUtils.getCurrentUser();
  }

  // 检查是否已登录
  isAuthenticated(): boolean {
    return authUtils.isAuthenticated();
  }

  // 检查用户角色
  hasRole(role: string): boolean {
    return authUtils.hasRole(role);
  }

  // 检查是否为管理员
  isAdmin(): boolean {
    return authUtils.isAdmin();
  }
}

// 创建全局认证管理器实例 - 向后兼容
export const authManager = AuthManager.getInstance();

// 导出默认的认证工具
export default authUtils;
