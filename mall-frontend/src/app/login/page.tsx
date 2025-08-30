'use client';

import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Typography, Divider, Checkbox, message } from 'antd';
import { UserOutlined, LockOutlined, EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAppDispatch, useAppSelector } from '@/store';
import { loginAsync, selectAuth } from '@/store/slices/authSlice';
import { LoginRequest } from '@/types';
import { ROUTES } from '@/constants';

const { Title, Text } = Typography;

interface LoginFormData {
  email: string;
  password: string;
  remember: boolean;
}

const LoginPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const router = useRouter();
  const dispatch = useAppDispatch();
  const { isAuthenticated, user } = useAppSelector(selectAuth);

  // 如果已登录，重定向到首页
  useEffect(() => {
    if (isAuthenticated && user) {
      router.push(ROUTES.HOME);
    }
  }, [isAuthenticated, user, router]);

  const handleSubmit = async (values: LoginFormData) => {
    try {
      setLoading(true);
      
      const loginData: LoginRequest = {
        email: values.email,
        password: values.password,
      };

      const result = await dispatch(loginAsync(loginData));
      
      if (loginAsync.fulfilled.match(result)) {
        message.success('登录成功！');
        
        // 如果选择记住登录状态，保存到本地存储
        if (values.remember) {
          localStorage.setItem('mall_remember_login', 'true');
        }
        
        // 重定向到首页或之前访问的页面
        const redirectUrl = new URLSearchParams(window.location.search).get('redirect') || ROUTES.HOME;
        router.push(redirectUrl);
      } else {
        message.error(result.payload as string || '登录失败，请检查用户名和密码');
      }
    } catch (error) {
      console.error('Login error:', error);
      message.error('登录失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  const handleFormFailed = (errorInfo: any) => {
    console.log('Form validation failed:', errorInfo);
    message.error('请检查表单输入');
  };

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      padding: '20px'
    }}>
      <Card
        style={{
          width: '100%',
          maxWidth: 400,
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.1)',
          borderRadius: 12,
        }}
        bodyStyle={{ padding: '40px 32px' }}
      >
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Title level={2} style={{ color: '#1890ff', marginBottom: 8 }}>
            🛒 Go商城
          </Title>
          <Text type="secondary">
            欢迎回来，请登录您的账户
          </Text>
        </div>

        <Form
          form={form}
          name="login"
          onFinish={handleSubmit}
          onFinishFailed={handleFormFailed}
          autoComplete="off"
          size="large"
          layout="vertical"
        >
          <Form.Item
            name="email"
            label="邮箱地址"
            rules={[
              { required: true, message: '请输入邮箱地址' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="请输入邮箱地址"
              autoComplete="email"
            />
          </Form.Item>

          <Form.Item
            name="password"
            label="密码"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6位字符' },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入密码"
              autoComplete="current-password"
              iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
            />
          </Form.Item>

          <Form.Item>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Form.Item name="remember" valuePropName="checked" noStyle>
                <Checkbox>记住登录状态</Checkbox>
              </Form.Item>
              <Link href="/forgot-password" style={{ color: '#1890ff' }}>
                忘记密码？
              </Link>
            </div>
          </Form.Item>

          <Form.Item style={{ marginBottom: 16 }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              style={{
                height: 48,
                fontSize: 16,
                fontWeight: 500,
              }}
            >
              {loading ? '登录中...' : '登录'}
            </Button>
          </Form.Item>

          <Divider>
            <Text type="secondary" style={{ fontSize: 12 }}>
              其他登录方式
            </Text>
          </Divider>

          <div style={{ display: 'flex', justifyContent: 'center', gap: 16, marginBottom: 24 }}>
            <Button
              shape="circle"
              size="large"
              style={{ border: '1px solid #d9d9d9' }}
              title="微信登录"
            >
              💬
            </Button>
            <Button
              shape="circle"
              size="large"
              style={{ border: '1px solid #d9d9d9' }}
              title="QQ登录"
            >
              🐧
            </Button>
            <Button
              shape="circle"
              size="large"
              style={{ border: '1px solid #d9d9d9' }}
              title="支付宝登录"
            >
              💰
            </Button>
          </div>

          <div style={{ textAlign: 'center' }}>
            <Text type="secondary">
              还没有账户？{' '}
              <Link href={ROUTES.REGISTER} style={{ color: '#1890ff', fontWeight: 500 }}>
                立即注册
              </Link>
            </Text>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default LoginPage;
