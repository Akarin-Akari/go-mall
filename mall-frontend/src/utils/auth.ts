import { User } from '@/types';
import { STORAGE_KEYS, USER_ROLES } from '@/constants';
import Cookies from 'js-cookie';

// 本地存储工具 (避免循环依赖)
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

// Token管理工具 (避免循环依赖)
const tokenManager = {
  getToken: (): string | null => {
    return storage.get(STORAGE_KEYS.TOKEN) || Cookies.get(STORAGE_KEYS.TOKEN) || null;
  },

  setToken: (token: string, remember = false): void => {
    storage.set(STORAGE_KEYS.TOKEN, token);
    if (remember) {
      Cookies.set(STORAGE_KEYS.TOKEN, token, { expires: 7 });
    }
  },

  removeToken: (): void => {
    storage.remove(STORAGE_KEYS.TOKEN);
    Cookies.remove(STORAGE_KEYS.TOKEN);
  },

  getRefreshToken: (): string | null => {
    return storage.get(STORAGE_KEYS.REFRESH_TOKEN);
  },

  setRefreshToken: (token: string): void => {
    storage.set(STORAGE_KEYS.REFRESH_TOKEN, token);
  },

  removeRefreshToken: (): void => {
    storage.remove(STORAGE_KEYS.REFRESH_TOKEN);
  },

  clearAll: (): void => {
    tokenManager.removeToken();
    tokenManager.removeRefreshToken();
  },
};

// 认证状态管理
export class AuthManager {
  private static instance: AuthManager;
  private user: User | null = null;
  private listeners: ((user: User | null) => void)[] = [];
  private initialized: boolean = false;

  private constructor() {
    // 延迟初始化，避免循环依赖问题
    this.initializeAsync();
  }

  // 异步初始化方法
  private initializeAsync(): void {
    // 使用 setTimeout 确保所有模块都已加载完成
    setTimeout(() => {
      this.loadUserFromStorage();
      this.initialized = true;
    }, 0);
  }

  // 单例模式
  public static getInstance(): AuthManager {
    if (!AuthManager.instance) {
      AuthManager.instance = new AuthManager();
    }
    return AuthManager.instance;
  }

  // 检查是否已初始化
  public isInitialized(): boolean {
    return this.initialized;
  }

  // 等待初始化完成
  public async waitForInitialization(): Promise<void> {
    if (this.initialized) {
      return Promise.resolve();
    }

    return new Promise((resolve) => {
      const checkInitialized = () => {
        if (this.initialized) {
          resolve();
        } else {
          setTimeout(checkInitialized, 10);
        }
      };
      checkInitialized();
    });
  }

  // 从本地存储加载用户信息
  private loadUserFromStorage(): void {
    try {
      // 检查是否在浏览器环境中
      if (typeof window === 'undefined') {
        return;
      }

      // 安全检查：确保 storage 和 tokenManager 已初始化
      if (!storage || !tokenManager) {
        console.warn('Storage or tokenManager not initialized yet');
        return;
      }

      const userInfo = storage.getJSON<User>(STORAGE_KEYS.USER_INFO);
      const token = tokenManager.getToken();

      if (userInfo && token) {
        this.user = userInfo;
      }
    } catch (error) {
      console.error('Error loading user from storage:', error);
      // 清理可能损坏的数据
      this.clearUserData();
    }
  }

  // 设置用户信息
  setUser(user: User | null): void {
    this.user = user;

    try {
      // 安全检查：确保在浏览器环境中且storage已初始化
      if (typeof window !== 'undefined' && storage) {
        if (user) {
          storage.setJSON(STORAGE_KEYS.USER_INFO, user);
        } else {
          storage.remove(STORAGE_KEYS.USER_INFO);
        }
      }
    } catch (error) {
      console.error('Error setting user in storage:', error);
    }

    // 通知监听器
    this.notifyListeners();
  }

  // 获取用户信息
  getUser(): User | null {
    return this.user;
  }

  // 检查是否已登录
  isAuthenticated(): boolean {
    try {
      // 安全检查：确保tokenManager已初始化
      if (!tokenManager) {
        return false;
      }
      return !!(this.user && tokenManager.getToken());
    } catch (error) {
      console.error('Error checking authentication status:', error);
      return false;
    }
  }

  // 检查用户角色
  hasRole(role: string): boolean {
    return this.user?.role === role;
  }

  // 检查是否为管理员
  isAdmin(): boolean {
    return this.hasRole(USER_ROLES.ADMIN);
  }

  // 检查是否为商户
  isMerchant(): boolean {
    return this.hasRole(USER_ROLES.MERCHANT);
  }

  // 检查是否为普通用户
  isUser(): boolean {
    return this.hasRole(USER_ROLES.USER);
  }

  // 检查多个角色
  hasAnyRole(roles: string[]): boolean {
    return roles.some(role => this.hasRole(role));
  }

  // 登录
  login(user: User, token: string, refreshToken?: string, remember = false): void {
    try {
      this.setUser(user);

      if (tokenManager) {
        tokenManager.setToken(token, remember);

        if (refreshToken) {
          tokenManager.setRefreshToken(refreshToken);
        }
      }
    } catch (error) {
      console.error('Error during login:', error);
      throw error;
    }
  }

  // 清理用户数据
  private clearUserData(): void {
    try {
      this.user = null;
      if (typeof window !== 'undefined' && storage && tokenManager) {
        storage.remove(STORAGE_KEYS.USER_INFO);
        tokenManager.clearAll();
        storage.remove(STORAGE_KEYS.CART_ITEMS);
        storage.remove(STORAGE_KEYS.REMEMBER_LOGIN);
      }
    } catch (error) {
      console.error('Error clearing user data:', error);
    }
  }

  // 登出
  logout(): void {
    this.setUser(null);

    try {
      if (tokenManager) {
        tokenManager.clearAll();
      }

      // 清除其他相关存储
      if (storage) {
        storage.remove(STORAGE_KEYS.CART_ITEMS);
        storage.remove(STORAGE_KEYS.REMEMBER_LOGIN);
      }
    } catch (error) {
      console.error('Error during logout:', error);
    }
  }

  // 更新用户信息
  updateUser(updates: Partial<User>): void {
    if (this.user) {
      this.setUser({ ...this.user, ...updates });
    }
  }

  // 添加状态监听器
  addListener(listener: (user: User | null) => void): () => void {
    this.listeners.push(listener);
    
    // 返回取消监听的函数
    return () => {
      const index = this.listeners.indexOf(listener);
      if (index > -1) {
        this.listeners.splice(index, 1);
      }
    };
  }

  // 通知所有监听器
  private notifyListeners(): void {
    this.listeners.forEach(listener => listener(this.user));
  }

  // 检查token是否即将过期
  isTokenExpiringSoon(threshold = 5 * 60 * 1000): boolean {
    const token = tokenManager.getToken();
    if (!token) return false;

    try {
      // 解析JWT token
      const payload = JSON.parse(atob(token.split('.')[1]));
      const expirationTime = payload.exp * 1000;
      const currentTime = Date.now();
      
      return expirationTime - currentTime < threshold;
    } catch {
      return false;
    }
  }

  // 刷新token
  async refreshToken(): Promise<boolean> {
    const refreshToken = tokenManager.getRefreshToken();
    if (!refreshToken) return false;

    try {
      // 这里应该调用刷新token的API
      // const response = await refreshTokenAPI(refreshToken);
      // tokenManager.setToken(response.data.token);
      // if (response.data.refresh_token) {
      //   tokenManager.setRefreshToken(response.data.refresh_token);
      // }
      return true;
    } catch {
      this.logout();
      return false;
    }
  }
}

// 创建全局认证管理器实例
export const authManager = AuthManager.getInstance();

// 权限检查装饰器
export function requireAuth<T extends (...args: any[]) => any>(
  target: any,
  propertyName: string,
  descriptor: TypedPropertyDescriptor<T>
): TypedPropertyDescriptor<T> | void {
  const method = descriptor.value!;

  descriptor.value = ((...args: any[]) => {
    if (!authManager.isAuthenticated()) {
      throw new Error('需要登录才能执行此操作');
    }
    return method.apply(target, args);
  }) as T;
}

// 角色检查装饰器
export function requireRole(roles: string | string[]) {
  return function <T extends (...args: any[]) => any>(
    target: any,
    propertyName: string,
    descriptor: TypedPropertyDescriptor<T>
  ): TypedPropertyDescriptor<T> | void {
    const method = descriptor.value!;
    const roleArray = Array.isArray(roles) ? roles : [roles];

    descriptor.value = ((...args: any[]) => {
      if (!authManager.isAuthenticated()) {
        throw new Error('需要登录才能执行此操作');
      }
      
      if (!authManager.hasAnyRole(roleArray)) {
        throw new Error('没有权限执行此操作');
      }
      
      return method.apply(target, args);
    }) as T;
  };
}

// 权限检查函数
export const checkPermission = {
  // 检查是否已登录
  isAuthenticated: (): boolean => {
    return authManager.isAuthenticated();
  },

  // 检查角色权限
  hasRole: (role: string): boolean => {
    return authManager.hasRole(role);
  },

  // 检查多个角色权限
  hasAnyRole: (roles: string[]): boolean => {
    return authManager.hasAnyRole(roles);
  },

  // 检查是否为资源所有者
  isOwner: (resourceUserId: number): boolean => {
    const user = authManager.getUser();
    return user?.id === resourceUserId;
  },

  // 检查是否为管理员或资源所有者
  isAdminOrOwner: (resourceUserId: number): boolean => {
    return authManager.isAdmin() || checkPermission.isOwner(resourceUserId);
  },

  // 检查页面访问权限
  canAccessPage: (requiredRoles?: string[]): boolean => {
    if (!requiredRoles || requiredRoles.length === 0) {
      return true;
    }
    
    if (!authManager.isAuthenticated()) {
      return false;
    }
    
    return authManager.hasAnyRole(requiredRoles);
  },

  // 检查操作权限
  canPerformAction: (action: string, resource?: any): boolean => {
    if (!authManager.isAuthenticated()) {
      return false;
    }

    const user = authManager.getUser();
    if (!user) return false;

    // 管理员拥有所有权限
    if (authManager.isAdmin()) {
      return true;
    }

    // 根据不同的操作和资源类型检查权限
    switch (action) {
      case 'create':
        return true; // 登录用户都可以创建
      
      case 'read':
        return true; // 登录用户都可以读取
      
      case 'update':
      case 'delete':
        // 只能操作自己的资源
        return resource?.user_id === user.id;
      
      default:
        return false;
    }
  },
};

// 路由守卫
export const routeGuard = {
  // 检查路由访问权限
  canActivate: (requiredRoles?: string[]): boolean => {
    return checkPermission.canAccessPage(requiredRoles);
  },

  // 获取重定向URL
  getRedirectUrl: (currentPath: string): string => {
    if (!authManager.isAuthenticated()) {
      return `/login?redirect=${encodeURIComponent(currentPath)}`;
    }
    
    // 权限不足，跳转到403页面
    return '/403';
  },

  // 处理路由守卫
  handleGuard: (requiredRoles?: string[]): string | null => {
    if (routeGuard.canActivate(requiredRoles)) {
      return null; // 允许访问
    }
    
    return routeGuard.getRedirectUrl(window.location.pathname);
  },
};

// React Hook：使用认证状态
export const useAuth = () => {
  const [user, setUser] = React.useState<User | null>(authManager.getUser());

  React.useEffect(() => {
    const unsubscribe = authManager.addListener(setUser);
    return unsubscribe;
  }, []);

  return {
    user,
    isAuthenticated: authManager.isAuthenticated(),
    isAdmin: authManager.isAdmin(),
    isMerchant: authManager.isMerchant(),
    isUser: authManager.isUser(),
    hasRole: authManager.hasRole.bind(authManager),
    hasAnyRole: authManager.hasAnyRole.bind(authManager),
    login: authManager.login.bind(authManager),
    logout: authManager.logout.bind(authManager),
    updateUser: authManager.updateUser.bind(authManager),
  };
};

// 导入React（用于Hook）
import React from 'react';
