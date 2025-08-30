import { User } from '@/types';
import { tokenManager, storage } from './index';
import { STORAGE_KEYS, USER_ROLES } from '@/constants';

// 认证状态管理
export class AuthManager {
  private static instance: AuthManager;
  private user: User | null = null;
  private listeners: ((user: User | null) => void)[] = [];

  private constructor() {
    this.loadUserFromStorage();
  }

  // 单例模式
  public static getInstance(): AuthManager {
    if (!AuthManager.instance) {
      AuthManager.instance = new AuthManager();
    }
    return AuthManager.instance;
  }

  // 从本地存储加载用户信息
  private loadUserFromStorage(): void {
    const userInfo = storage.getJSON<User>(STORAGE_KEYS.USER_INFO);
    const token = tokenManager.getToken();
    
    if (userInfo && token) {
      this.user = userInfo;
    }
  }

  // 设置用户信息
  setUser(user: User | null): void {
    this.user = user;
    
    if (user) {
      storage.setJSON(STORAGE_KEYS.USER_INFO, user);
    } else {
      storage.remove(STORAGE_KEYS.USER_INFO);
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
    return !!(this.user && tokenManager.getToken());
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
    this.setUser(user);
    tokenManager.setToken(token, remember);
    
    if (refreshToken) {
      tokenManager.setRefreshToken(refreshToken);
    }
  }

  // 登出
  logout(): void {
    this.setUser(null);
    tokenManager.clearAll();
    
    // 清除其他相关存储
    storage.remove(STORAGE_KEYS.CART_ITEMS);
    storage.remove(STORAGE_KEYS.REMEMBER_LOGIN);
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
