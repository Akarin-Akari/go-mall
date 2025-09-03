'use client';

import React, { useState } from 'react';
import { 
  Layout, 
  Menu, 
  Button, 
  Space, 
  Badge, 
  Avatar, 
  Dropdown, 
  Input,
  Typography,
  Divider
} from 'antd';
import { 
  MenuOutlined,
  SearchOutlined,
  ShoppingCartOutlined,
  UserOutlined,
  HeartOutlined,
  BellOutlined,
  SettingOutlined,
  LogoutOutlined,
  HomeOutlined,
  ShopOutlined,
  AppstoreOutlined,
  CustomerServiceOutlined
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAppSelector, useAppDispatch } from '@/store';
import { selectAuth, logoutAsync } from '@/store/slices/authSlice';
import { ROUTES } from '@/constants';

const { Header, Content, Footer } = Layout;
const { Search } = Input;
const { Text } = Typography;

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const [mobileMenuVisible, setMobileMenuVisible] = useState(false);
  const router = useRouter();
  const dispatch = useAppDispatch();
  const { isAuthenticated, user } = useAppSelector(selectAuth);

  // 处理登出
  const handleLogout = async () => {
    await dispatch(logoutAsync());
    router.push(ROUTES.LOGIN);
  };

  // 用户菜单
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
      onClick: () => router.push(ROUTES.USER_CENTER),
    },
    {
      key: 'orders',
      icon: <ShopOutlined />,
      label: '我的订单',
      onClick: () => router.push(ROUTES.USER_ORDERS),
    },
    {
      key: 'favorites',
      icon: <HeartOutlined />,
      label: '我的收藏',
      onClick: () => router.push('/user/favorites'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '账户设置',
      onClick: () => router.push(ROUTES.USER_SETTINGS),
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  // 主导航菜单
  const navMenuItems = [
    {
      key: 'home',
      icon: <HomeOutlined />,
      label: <Link href={ROUTES.HOME}>首页</Link>,
    },
    {
      key: 'products',
      icon: <ShopOutlined />,
      label: <Link href={ROUTES.PRODUCTS}>商品</Link>,
    },
    {
      key: 'categories',
      icon: <AppstoreOutlined />,
      label: <Link href={ROUTES.CATEGORIES}>分类</Link>,
    },
    {
      key: 'service',
      icon: <CustomerServiceOutlined />,
      label: '客服',
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* 顶部导航 */}
      <Header style={{ 
        position: 'fixed', 
        zIndex: 1000, 
        width: '100%',
        backgroundColor: '#fff',
        borderBottom: '1px solid #f0f0f0',
        padding: '0 24px',
        height: 64,
        lineHeight: '64px'
      }}>
        <div style={{ 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'space-between',
          height: '100%'
        }}>
          {/* 左侧：Logo和导航 */}
          <div style={{ display: 'flex', alignItems: 'center' }}>
            {/* Logo */}
            <div style={{ 
              fontSize: 24, 
              fontWeight: 'bold', 
              color: '#1890ff',
              marginRight: 32,
              cursor: 'pointer'
            }} onClick={() => router.push(ROUTES.HOME)}>
              🛒 Go商城
            </div>

            {/* 桌面端导航菜单 */}
            <Menu
              mode="horizontal"
              items={navMenuItems}
              style={{ 
                border: 'none',
                backgroundColor: 'transparent',
                minWidth: 300
              }}
              className="hidden md:flex"
            />

            {/* 移动端菜单按钮 */}
            <Button
              type="text"
              icon={<MenuOutlined />}
              onClick={() => setMobileMenuVisible(!mobileMenuVisible)}
              className="md:hidden"
            />
          </div>

          {/* 中间：搜索框 */}
          <div style={{ flex: 1, maxWidth: 400, margin: '0 24px' }}>
            <Search
              placeholder="搜索商品..."
              onSearch={(value) => {
                if (value.trim()) {
                  router.push(`${ROUTES.PRODUCTS}?keyword=${encodeURIComponent(value)}`);
                }
              }}
              style={{ width: '100%' }}
              size="large"
            />
          </div>

          {/* 右侧：用户操作 */}
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Space size="large">
              {/* 购物车 */}
              <Badge count={3} size="small">
                <Button
                  type="text"
                  icon={<ShoppingCartOutlined style={{ fontSize: 20 }} />}
                  onClick={() => router.push(ROUTES.CART)}
                  style={{ height: 40, width: 40 }}
                />
              </Badge>

              {/* 通知 */}
              <Badge dot>
                <Button
                  type="text"
                  icon={<BellOutlined style={{ fontSize: 20 }} />}
                  style={{ height: 40, width: 40 }}
                />
              </Badge>

              {/* 用户信息 */}
              {isAuthenticated && user ? (
                <Dropdown
                  menu={{ items: userMenuItems }}
                  placement="bottomRight"
                  trigger={['click']}
                >
                  <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    cursor: 'pointer',
                    padding: '4px 8px',
                    borderRadius: 6,
                    transition: 'background-color 0.3s'
                  }}>
                    <Avatar 
                      src={user.avatar} 
                      icon={<UserOutlined />}
                      size="small"
                    />
                    <Text style={{ 
                      marginLeft: 8, 
                      color: '#333',
                      display: 'none'
                    }} className="sm:inline">
                      {user.nickname || user.username}
                    </Text>
                  </div>
                </Dropdown>
              ) : (
                <Space>
                  <Button 
                    type="text" 
                    onClick={() => router.push(ROUTES.LOGIN)}
                  >
                    登录
                  </Button>
                  <Button 
                    type="primary" 
                    onClick={() => router.push(ROUTES.REGISTER)}
                  >
                    注册
                  </Button>
                </Space>
              )}
            </Space>
          </div>
        </div>
      </Header>

      {/* 主要内容区域 */}
      <Content style={{ 
        marginTop: 64,
        minHeight: 'calc(100vh - 64px - 120px)',
        backgroundColor: '#f5f5f5'
      }}>
        {children}
      </Content>

      {/* 底部 */}
      <Footer style={{ 
        textAlign: 'center',
        backgroundColor: '#001529',
        color: '#fff',
        padding: '40px 24px 24px'
      }}>
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          <div style={{ 
            display: 'grid', 
            gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
            gap: 32,
            marginBottom: 24
          }}>
            <div>
              <Text strong style={{ color: '#fff', fontSize: 16 }}>
                关于我们
              </Text>
              <div style={{ marginTop: 12 }}>
                <div><Link href="/about" style={{ color: '#ccc' }}>公司介绍</Link></div>
                <div><Link href="/contact" style={{ color: '#ccc' }}>联系我们</Link></div>
                <div><Link href="/careers" style={{ color: '#ccc' }}>加入我们</Link></div>
              </div>
            </div>
            
            <div>
              <Text strong style={{ color: '#fff', fontSize: 16 }}>
                客户服务
              </Text>
              <div style={{ marginTop: 12 }}>
                <div><Link href="/help" style={{ color: '#ccc' }}>帮助中心</Link></div>
                <div><Link href="/service" style={{ color: '#ccc' }}>在线客服</Link></div>
                <div><Link href="/feedback" style={{ color: '#ccc' }}>意见反馈</Link></div>
              </div>
            </div>
            
            <div>
              <Text strong style={{ color: '#fff', fontSize: 16 }}>
                购物指南
              </Text>
              <div style={{ marginTop: 12 }}>
                <div><Link href="/guide" style={{ color: '#ccc' }}>购物流程</Link></div>
                <div><Link href="/payment" style={{ color: '#ccc' }}>支付方式</Link></div>
                <div><Link href="/delivery" style={{ color: '#ccc' }}>配送说明</Link></div>
              </div>
            </div>
            
            <div>
              <Text strong style={{ color: '#fff', fontSize: 16 }}>
                关注我们
              </Text>
              <div style={{ marginTop: 12 }}>
                <Space>
                  <Button type="link" style={{ color: '#ccc', padding: 0 }}>
                    微信
                  </Button>
                  <Button type="link" style={{ color: '#ccc', padding: 0 }}>
                    微博
                  </Button>
                  <Button type="link" style={{ color: '#ccc', padding: 0 }}>
                    抖音
                  </Button>
                </Space>
              </div>
            </div>
          </div>
          
          <Divider style={{ borderColor: '#434343' }} />
          
          <div style={{ color: '#ccc' }}>
            <Text style={{ color: '#ccc' }}>
              © 2024 Go商城. All rights reserved. | 
              <Link href="/privacy" style={{ color: '#ccc', marginLeft: 8 }}>隐私政策</Link> | 
              <Link href="/terms" style={{ color: '#ccc', marginLeft: 8 }}>服务条款</Link>
            </Text>
          </div>
        </div>
      </Footer>
    </Layout>
  );
};

export default MainLayout;
