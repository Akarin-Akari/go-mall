'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import { Result, Button } from 'antd';
import { useAuth } from '@/hooks/useAuth';
import { ROUTES } from '@/constants';
import Loading from '@/components/common/Loading';

export interface WithAuthOptions {
  requireAuth?: boolean; // 是否需要认证
  requireRoles?: string[]; // 需要的角色
  redirectTo?: string; // 自定义重定向路径
  fallback?: React.ComponentType; // 自定义加载组件
  unauthorizedComponent?: React.ComponentType; // 自定义未授权组件
}

/**
 * 高阶组件：为组件添加认证和权限检查
 * @param WrappedComponent 要包装的组件
 * @param options 认证选项
 */
export function withAuth<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  options: WithAuthOptions = {}
) {
  const {
    requireAuth = true,
    requireRoles = [],
    redirectTo,
    fallback: FallbackComponent,
    unauthorizedComponent: UnauthorizedComponent,
  } = options;

  const AuthenticatedComponent: React.FC<P> = (props) => {
    const router = useRouter();
    const { user, isAuthenticated, loading, hasRole, redirectToLogin } = useAuth();

    // 如果正在加载，显示加载状态
    if (loading) {
      if (FallbackComponent) {
        return <FallbackComponent />;
      }
      return <Loading fullScreen text="正在验证身份..." />;
    }

    // 如果需要认证但用户未登录
    if (requireAuth && !isAuthenticated) {
      if (redirectTo) {
        router.push(redirectTo);
        return <Loading fullScreen text="正在跳转..." />;
      }
      
      if (UnauthorizedComponent) {
        return <UnauthorizedComponent />;
      }

      // 自动重定向到登录页
      redirectToLogin();
      return <Loading fullScreen text="正在跳转到登录页..." />;
    }

    // 如果需要特定角色但用户没有权限
    if (requireRoles.length > 0 && user) {
      const hasRequiredRole = requireRoles.some(role => hasRole(role));
      
      if (!hasRequiredRole) {
        if (UnauthorizedComponent) {
          return <UnauthorizedComponent />;
        }

        return (
          <Result
            status="403"
            title="权限不足"
            subTitle={`您需要以下角色之一才能访问此页面: ${requireRoles.join(', ')}`}
            extra={
              <Button type="primary" onClick={() => router.push(ROUTES.HOME)}>
                返回首页
              </Button>
            }
          />
        );
      }
    }

    // 权限检查通过，渲染原组件
    return <WrappedComponent {...props} />;
  };

  // 设置显示名称，便于调试
  AuthenticatedComponent.displayName = `withAuth(${WrappedComponent.displayName || WrappedComponent.name})`;

  return AuthenticatedComponent;
}

/**
 * 预定义的认证HOC
 */

// 需要登录的组件
export const withRequireAuth = <P extends object>(
  WrappedComponent: React.ComponentType<P>
) => withAuth(WrappedComponent, { requireAuth: true });

// 需要管理员权限的组件
export const withRequireAdmin = <P extends object>(
  WrappedComponent: React.ComponentType<P>
) => withAuth(WrappedComponent, { 
  requireAuth: true, 
  requireRoles: ['admin'] 
});

// 需要商户权限的组件
export const withRequireMerchant = <P extends object>(
  WrappedComponent: React.ComponentType<P>
) => withAuth(WrappedComponent, { 
  requireAuth: true, 
  requireRoles: ['merchant', 'admin'] 
});

// 可选认证的组件（登录后显示更多功能）
export const withOptionalAuth = <P extends object>(
  WrappedComponent: React.ComponentType<P>
) => withAuth(WrappedComponent, { requireAuth: false });

/**
 * 权限检查组件
 * 用于在JSX中进行条件渲染
 */
interface PermissionCheckProps {
  children: React.ReactNode;
  requireAuth?: boolean;
  requireRoles?: string[];
  fallback?: React.ReactNode;
}

export const PermissionCheck: React.FC<PermissionCheckProps> = ({
  children,
  requireAuth = false,
  requireRoles = [],
  fallback = null,
}) => {
  const { user, isAuthenticated, hasRole } = useAuth();

  // 如果需要认证但用户未登录
  if (requireAuth && !isAuthenticated) {
    return <>{fallback}</>;
  }

  // 如果需要特定角色但用户没有权限
  if (requireRoles.length > 0) {
    if (!user || !requireRoles.some(role => hasRole(role))) {
      return <>{fallback}</>;
    }
  }

  return <>{children}</>;
};

/**
 * 角色检查组件
 * 根据用户角色显示不同内容
 */
interface RoleBasedRenderProps {
  roles: {
    [role: string]: React.ReactNode;
  };
  fallback?: React.ReactNode;
}

export const RoleBasedRender: React.FC<RoleBasedRenderProps> = ({
  roles,
  fallback = null,
}) => {
  const { user, hasRole } = useAuth();

  if (!user) {
    return <>{fallback}</>;
  }

  // 按优先级检查角色（admin > merchant > user）
  const roleOrder = ['admin', 'merchant', 'user'];
  
  for (const role of roleOrder) {
    if (roles[role] && hasRole(role)) {
      return <>{roles[role]}</>;
    }
  }

  return <>{fallback}</>;
};

export default withAuth;
