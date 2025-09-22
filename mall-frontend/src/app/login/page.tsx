'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Form,
  Input,
  Button,
  Card,
  Typography,
  Divider,
  Checkbox,
  message,
  Alert,
  Space,
} from 'antd';
import {
  UserOutlined,
  LockOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
  LoadingOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { useAppDispatch, useAppSelector } from '@/store';
import { loginAsync, selectAuth, clearError } from '@/store/slices/authSlice';
import { LoginRequest } from '@/types';
import { ROUTES, STORAGE_KEYS } from '@/constants';

const { Title, Text } = Typography;

interface LoginFormData {
  email: string;
  password: string;
  remember: boolean;
}

const LoginPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [loginAttempts, setLoginAttempts] = useState(0);
  const [isBlocked, setIsBlocked] = useState(false);
  const [blockTimeLeft, setBlockTimeLeft] = useState(0);

  const router = useRouter();
  const searchParams = useSearchParams();
  const dispatch = useAppDispatch();
  const {
    isAuthenticated,
    user,
    error,
    loading: authLoading,
  } = useAppSelector(selectAuth);

  // 获取重定向URL
  const redirectUrl = searchParams?.get('redirect') || ROUTES.HOME;

  // 如果已登录，重定向到首页
  useEffect(() => {
    if (isAuthenticated && user) {
      router.push(redirectUrl);
    }
  }, [isAuthenticated, user, router, redirectUrl]);

  // 清除错误状态
  useEffect(() => {
    return () => {
      dispatch(clearError());
    };
  }, [dispatch]);

  // 防暴力破解：检查登录尝试次数
  useEffect(() => {
    const attempts = parseInt(localStorage.getItem('login_attempts') || '0');
    const lastAttempt = parseInt(
      localStorage.getItem('last_login_attempt') || '0'
    );
    const now = Date.now();

    // 如果超过5次失败，且在30分钟内，则阻止登录
    if (attempts >= 5 && now - lastAttempt < 30 * 60 * 1000) {
      setIsBlocked(true);
      setLoginAttempts(attempts);
      const timeLeft = Math.ceil((30 * 60 * 1000 - (now - lastAttempt)) / 1000);
      setBlockTimeLeft(timeLeft);

      // 倒计时
      const timer = setInterval(() => {
        setBlockTimeLeft(prev => {
          if (prev <= 1) {
            setIsBlocked(false);
            setLoginAttempts(0);
            localStorage.removeItem('login_attempts');
            localStorage.removeItem('last_login_attempt');
            clearInterval(timer);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);

      return () => clearInterval(timer);
    } else if (now - lastAttempt > 30 * 60 * 1000) {
      // 超过30分钟，重置计数
      localStorage.removeItem('login_attempts');
      localStorage.removeItem('last_login_attempt');
      setLoginAttempts(0);
    } else {
      setLoginAttempts(attempts);
    }
  }, []);

  // 记住登录状态
  useEffect(() => {
    const rememberLogin = localStorage.getItem(STORAGE_KEYS.REMEMBER_LOGIN);
    if (rememberLogin === 'true') {
      const savedEmail = localStorage.getItem('saved_email');
      if (savedEmail) {
        form.setFieldsValue({
          email: savedEmail,
          remember: true,
        });
      }
    }
  }, [form]);

  const handleSubmit = async (values: LoginFormData) => {
    // 检查是否被阻止
    if (isBlocked) {
      message.error(
        `登录尝试过多，请等待 ${Math.ceil(blockTimeLeft / 60)} 分钟后再试`
      );
      return;
    }

    try {
      setLoading(true);
      dispatch(clearError()); // 清除之前的错误

      const loginData: LoginRequest = {
        email: values.email,
        password: values.password,
      };

      const result = await dispatch(loginAsync(loginData));

      if (loginAsync.fulfilled.match(result)) {
        message.success('登录成功！欢迎回来！');

        // 登录成功，重置失败计数
        localStorage.removeItem('login_attempts');
        localStorage.removeItem('last_login_attempt');
        setLoginAttempts(0);

        // 如果选择记住登录状态，保存到本地存储
        if (values.remember) {
          localStorage.setItem(STORAGE_KEYS.REMEMBER_LOGIN, 'true');
          localStorage.setItem('saved_email', values.email);
        } else {
          localStorage.removeItem(STORAGE_KEYS.REMEMBER_LOGIN);
          localStorage.removeItem('saved_email');
        }

        // 重定向到目标页面
        router.push(redirectUrl);
      } else {
        // 登录失败，增加失败计数
        const newAttempts = loginAttempts + 1;
        setLoginAttempts(newAttempts);
        localStorage.setItem('login_attempts', newAttempts.toString());
        localStorage.setItem('last_login_attempt', Date.now().toString());

        // 检查是否需要阻止
        if (newAttempts >= 5) {
          setIsBlocked(true);
          setBlockTimeLeft(30 * 60); // 30分钟
          message.error('登录失败次数过多，账户已被临时锁定30分钟');
        } else {
          const errorMsg =
            (result.payload as string) || '登录失败，请检查邮箱和密码';
          message.error(`${errorMsg} (剩余尝试次数: ${5 - newAttempts})`);
        }
      }
    } catch (error) {
      console.error('Login error:', error);
      message.error('网络错误，请检查网络连接后重试');
    } finally {
      setLoading(false);
    }
  };

  const handleFormFailed = (errorInfo: any) => {
    console.log('Form validation failed:', errorInfo);
    message.error('请检查表单输入');
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '20px',
      }}
    >
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
          <Text type='secondary'>欢迎回来，请登录您的账户</Text>
        </div>

        {/* 错误提示 */}
        {error && (
          <Alert
            message='登录失败'
            description={error}
            type='error'
            showIcon
            closable
            style={{ marginBottom: 16 }}
            onClose={() => dispatch(clearError())}
          />
        )}

        {/* 安全提示 */}
        {isBlocked && (
          <Alert
            message='账户临时锁定'
            description={
              <Space direction='vertical' size='small'>
                <Text>由于多次登录失败，您的账户已被临时锁定</Text>
                <Text strong>
                  剩余时间: {Math.floor(blockTimeLeft / 60)}分
                  {blockTimeLeft % 60}秒
                </Text>
                <Text type='secondary'>
                  为了您的账户安全，请稍后再试或联系客服
                </Text>
              </Space>
            }
            type='warning'
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        {/* 登录尝试提示 */}
        {loginAttempts > 0 && loginAttempts < 5 && !isBlocked && (
          <Alert
            message={`登录失败 ${loginAttempts} 次`}
            description={`还有 ${5 - loginAttempts} 次尝试机会，超过5次将被临时锁定30分钟`}
            type='info'
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <Form
          form={form}
          name='login'
          onFinish={handleSubmit}
          onFinishFailed={handleFormFailed}
          autoComplete='off'
          size='large'
          layout='vertical'
        >
          <Form.Item
            name='email'
            label='邮箱地址'
            rules={[
              { required: true, message: '请输入邮箱地址' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder='请输入邮箱地址'
              autoComplete='email'
            />
          </Form.Item>

          <Form.Item
            name='password'
            label='密码'
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6位字符' },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder='请输入密码'
              autoComplete='current-password'
              iconRender={visible =>
                visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>

          <Form.Item>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}
            >
              <Form.Item name='remember' valuePropName='checked' noStyle>
                <Checkbox>记住登录状态</Checkbox>
              </Form.Item>
              <Link href='/forgot-password' style={{ color: '#1890ff' }}>
                忘记密码？
              </Link>
            </div>
          </Form.Item>

          <Form.Item style={{ marginBottom: 16 }}>
            <Button
              type='primary'
              htmlType='submit'
              loading={loading || authLoading}
              disabled={isBlocked}
              block
              icon={
                loading || authLoading ? (
                  <LoadingOutlined />
                ) : (
                  <SafetyCertificateOutlined />
                )
              }
              style={{
                height: 48,
                fontSize: 16,
                fontWeight: 500,
              }}
            >
              {isBlocked
                ? `账户锁定中 (${Math.floor(blockTimeLeft / 60)}:${String(blockTimeLeft % 60).padStart(2, '0')})`
                : loading || authLoading
                  ? '登录中...'
                  : '安全登录'}
            </Button>
          </Form.Item>

          {/* 安全提示 */}
          <div style={{ textAlign: 'center', marginBottom: 16 }}>
            <Text type='secondary' style={{ fontSize: 12 }}>
              <SafetyCertificateOutlined style={{ marginRight: 4 }} />
              您的登录信息将被加密传输
            </Text>
          </div>

          <Divider>
            <Text type='secondary' style={{ fontSize: 12 }}>
              其他登录方式
            </Text>
          </Divider>

          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              gap: 16,
              marginBottom: 24,
            }}
          >
            <Button
              shape='circle'
              size='large'
              style={{ border: '1px solid #d9d9d9' }}
              title='微信登录'
            >
              💬
            </Button>
            <Button
              shape='circle'
              size='large'
              style={{ border: '1px solid #d9d9d9' }}
              title='QQ登录'
            >
              🐧
            </Button>
            <Button
              shape='circle'
              size='large'
              style={{ border: '1px solid #d9d9d9' }}
              title='支付宝登录'
            >
              💰
            </Button>
          </div>

          <div style={{ textAlign: 'center' }}>
            <Text type='secondary'>
              还没有账户？{' '}
              <Link
                href={ROUTES.REGISTER}
                style={{ color: '#1890ff', fontWeight: 500 }}
              >
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
