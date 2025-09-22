'use client';

import React, { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { Spin, Result, Button } from 'antd';
import { useAppSelector, useAppDispatch } from '@/store';
import { selectAuth, restoreAuth, getProfileAsync } from '@/store/slices/authSlice';
import { secureTokenManager } from '@/utils/secureTokenManager';
import { ROUTES } from '@/constants';

interface AuthGuardProps {
  children: React.ReactNode;
  requireAuth?: boolean; // 是否需要认证
  requireRoles?: string[]; // 需要的角色
  fallback?: React.ReactNode; // 自定义加载组件
}

// 不需要认证的公开路由
const PUBLIC_ROUTES = [
  ROUTES.HOME,
  ROUTES.LOGIN,
  ROUTES.REGISTER,
  '/forgot-password',
  '/products',
  '/categories',
];

// 需要认证的路由
const PROTECTED_ROUTES = [
  ROUTES.CART,
  ROUTES.CHECKOUT,
  ROUTES.ORDERS,
  ROUTES.PROFILE,
  ROUTES.USER_CENTER,
  ROUTES.USER_ORDERS,
  ROUTES.USER_ADDRESS,
  ROUTES.USER_SETTINGS,
];

// 管理员路由
const ADMIN_ROUTES = [
  ROUTES.ADMIN,
  ROUTES.ADMIN_PRODUCTS,
  ROUTES.ADMIN_ORDERS,
  ROUTES.ADMIN_USERS,
];

const AuthGuard: React.FC<AuthGuardProps> = ({
  children,
  requireAuth,
  requireRoles = [],
  fallback,
}) => {
  const router = useRouter();
  const pathname = usePathname();
  const dispatch = useAppDispatch();
  
  const { user, isAuthenticated, loading } = useAppSelector(selectAuth);
  const [isInitializing, setIsInitializing] = useState(true);
  const [authChecked, setAuthChecked] = useState(false);

  // 检查路由是否需要认证
  const isProtectedRoute = (path: string): boolean => {
    if (requireAuth !== undefined) {
      return requireAuth;
    }
    
    // 检查是否是受保护的路由
    return PROTECTED_ROUTES.some(route => path.startsWith(route)) ||
           ADMIN_ROUTES.some(route => path.startsWith(route));
  };

  // 检查是否是管理员路由
  const isAdminRoute = (path: string): boolean => {
    return ADMIN_ROUTES.some(route => path.startsWith(route));
  };

  // 检查用户角色
  const hasRequiredRole = (userRoles: string[], requiredRoles: string[]): boolean => {
    if (requiredRoles.length === 0) return true;
    return requiredRoles.some(role => userRoles.includes(role));
  };

  // 初始化认证状态
  useEffect(() => {
    const initializeAuth = async () => {
      try {
        const token = secureTokenManager.getAccessToken();
        
        if (token) {
          // 恢复认证状态
          dispatch(restoreAuth());
          
          // 如果没有用户信息，尝试获取
          if (!user) {
            try {
              await dispatch(getProfileAsync()).unwrap();
            } catch (error) {
              console.error('获取用户信息失败:', error);
              // Token可能已过期，清除本地存储
              secureTokenManager.clearTokens();
            }
          }
        }
      } catch (error) {
        console.error('认证初始化失败:', error);
      } finally {
        setIsInitializing(false);
        setAuthChecked(true);
      }
    };

    initializeAuth();
  }, [dispatch, user]);

  // 监听路由变化，进行权限检查
  useEffect(() => {
    if (!authChecked || isInitializing) return;

    const checkAccess = () => {
      const needsAuth = isProtectedRoute(pathname);
      const needsAdmin = isAdminRoute(pathname);
      
      // 如果需要认证但用户未登录
      if (needsAuth && !isAuthenticated) {
        const redirectUrl = encodeURIComponent(pathname);
        router.push(`${ROUTES.LOGIN}?redirect=${redirectUrl}`);
        return;
      }

      // 如果需要管理员权限但用户不是管理员
      if (needsAdmin && (!user || user.role !== 'admin')) {
        router.push(ROUTES.UNAUTHORIZED);
        return;
      }

      // 检查特定角色要求
      if (requireRoles.length > 0 && user) {
        const userRoles = [user.role]; // 假设用户只有一个角色
        if (!hasRequiredRole(userRoles, requireRoles)) {
          router.push(ROUTES.FORBIDDEN);
          return;
        }
      }
    };

    checkAccess();
  }, [pathname, isAuthenticated, user, authChecked, isInitializing, requireRoles, router]);

  // 如果正在初始化，显示加载状态
  if (isInitializing || loading) {
    return (
      fallback || (
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: '100vh',
            flexDirection: 'column',
          }}
        >
          <Spin size="large" />
          <div style={{ marginTop: 16, color: '#666' }}>
            正在验证身份...
          </div>
        </div>
      )
    );
  }

  // 如果需要认证但用户未登录，显示未授权页面
  if (isProtectedRoute(pathname) && !isAuthenticated) {
    return (
      <Result
        status="403"
        title="需要登录"
        subTitle="请先登录后再访问此页面"
        extra={
          <Button type="primary" onClick={() => router.push(ROUTES.LOGIN)}>
            去登录
          </Button>
        }
      />
    );
  }

  // 如果需要管理员权限但用户不是管理员
  if (isAdminRoute(pathname) && (!user || user.role !== 'admin')) {
    return (
      <Result
        status="403"
        title="权限不足"
        subTitle="您没有访问此页面的权限"
        extra={
          <Button type="primary" onClick={() => router.push(ROUTES.HOME)}>
            返回首页
          </Button>
        }
      />
    );
  }

  // 检查特定角色要求
  if (requireRoles.length > 0 && user) {
    const userRoles = [user.role];
    if (!hasRequiredRole(userRoles, requireRoles)) {
      return (
        <Result
          status="403"
          title="权限不足"
          subTitle="您的角色权限不足以访问此页面"
          extra={
            <Button type="primary" onClick={() => router.push(ROUTES.HOME)}>
              返回首页
            </Button>
          }
        />
      );
    }
  }

  // 权限检查通过，渲染子组件
  return <>{children}</>;
};

export default AuthGuard;
