'use client';

import React, { useState, useEffect } from 'react';
import { Layout, Menu, Avatar, Dropdown, Badge, Button, theme, message } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
  ShoppingCartOutlined,
  BellOutlined,
  LogoutOutlined,
  SettingOutlined,
  LoginOutlined,
  HomeOutlined,
  AppstoreOutlined,
  ShopOutlined,
  OrderedListOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import { useAppSelector, useAppDispatch } from '@/store';
import { selectAuth, logoutAsync, restoreAuth } from '@/store/slices/authSlice';
import { selectCartItemCount } from '@/store/slices/cartSlice';
import { toggleCollapsed, selectCollapsed } from '@/store/slices/appSlice';
import { useAuth } from '@/hooks/useAuth';
import { ROUTES } from '@/constants';
import Link from 'next/link';

const { Header, Sider, Content } = Layout;

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const { logout, restoreAuthState } = useAuth();

  const { user, isAuthenticated } = useAppSelector(selectAuth);
  const cartItemCount = useAppSelector(selectCartItemCount);
  const collapsed = useAppSelector(selectCollapsed);

  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // 初始化时恢复认证状态
  useEffect(() => {
    restoreAuthState();
  }, [restoreAuthState]);

  // 菜单项配置
  const menuItems = [
    {
      key: '/',
      icon: <HomeOutlined />,
      label: <Link href='/'>首页</Link>,
    },
    {
      key: '/products',
      icon: <ShopOutlined />,
      label: <Link href='/products'>商品</Link>,
    },
    {
      key: '/categories',
      icon: <AppstoreOutlined />,
      label: <Link href='/categories'>分类</Link>,
    },
    ...(isAuthenticated ? [
      {
        key: '/orders',
        icon: <OrderedListOutlined />,
        label: <Link href='/orders'>我的订单</Link>,
      },
    ] : []),
  ];

  // 用户下拉菜单
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
      onClick: () => router.push(ROUTES.PROFILE),
    },
    {
      key: 'orders',
      icon: <OrderedListOutlined />,
      label: '我的订单',
      onClick: () => router.push(ROUTES.USER_ORDERS),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '账户设置',
      onClick: () => router.push(ROUTES.USER_SETTINGS),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: async () => {
        try {
          await logout();
        } catch (error) {
          message.error('退出登录失败');
        }
      },
    },
  ];

  const handleToggleCollapsed = () => {
    dispatch(toggleCollapsed());
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed}>
        <div
          className='demo-logo-vertical'
          style={{
            height: 32,
            margin: 16,
            background: 'rgba(255, 255, 255, 0.3)',
            borderRadius: 6,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'white',
            fontWeight: 'bold',
          }}
        >
          {collapsed ? 'M' : 'Mall'}
        </div>
        <Menu
          theme='dark'
          mode='inline'
          defaultSelectedKeys={['/']}
          items={menuItems}
        />
      </Sider>

      <Layout>
        <Header
          style={{
            padding: 0,
            background: colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            paddingRight: 24,
          }}
        >
          <Button
            type='text'
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={handleToggleCollapsed}
            style={{
              fontSize: '16px',
              width: 64,
              height: 64,
            }}
          />

          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {/* 购物车图标 */}
            <Badge count={cartItemCount} size='small'>
              <Button
                type='text'
                icon={<ShoppingCartOutlined />}
                onClick={() => router.push('/cart')}
                style={{ fontSize: '16px' }}
              />
            </Badge>

            {/* 通知图标 */}
            <Badge count={0} size='small'>
              <Button
                type='text'
                icon={<BellOutlined />}
                style={{ fontSize: '16px' }}
              />
            </Badge>

            {/* 用户信息 */}
            {isAuthenticated && user ? (
              <Dropdown menu={{ items: userMenuItems }} placement='bottomRight'>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 8,
                    cursor: 'pointer',
                    padding: '4px 8px',
                    borderRadius: 6,
                    transition: 'background-color 0.2s',
                  }}
                >
                  <Avatar
                    size='small'
                    src={user.avatar}
                    icon={<UserOutlined />}
                  />
                  <span>{user.nickname || user.username}</span>
                </div>
              </Dropdown>
            ) : (
              <div style={{ display: 'flex', gap: 8 }}>
                <Button
                  type='text'
                  icon={<LoginOutlined />}
                  onClick={() => router.push(ROUTES.LOGIN)}
                >
                  登录
                </Button>
                <Button
                  type='primary'
                  onClick={() => router.push(ROUTES.REGISTER)}
                >
                  注册
                </Button>
              </div>
            )}
          </div>
        </Header>

        <Content
          style={{
            margin: '24px 16px',
            padding: 24,
            minHeight: 280,
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
          }}
        >
          {children}
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
