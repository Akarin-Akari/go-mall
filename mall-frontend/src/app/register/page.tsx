'use client';

import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Typography, Divider, Checkbox, message, Progress } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined, PhoneOutlined, EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAppDispatch, useAppSelector } from '@/store';
import { registerAsync, selectAuth } from '@/store/slices/authSlice';
import { RegisterRequest } from '@/types';
import { ROUTES } from '@/constants';

const { Title, Text } = Typography;

interface RegisterFormData {
  username: string;
  email: string;
  phone: string;
  password: string;
  confirmPassword: string;
  agreement: boolean;
}

const RegisterPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);
  const router = useRouter();
  const dispatch = useAppDispatch();
  const { isAuthenticated, user } = useAppSelector(selectAuth);

  // 如果已登录，重定向到首页
  useEffect(() => {
    if (isAuthenticated && user) {
      router.push(ROUTES.HOME);
    }
  }, [isAuthenticated, user, router]);

  // 密码强度检测
  const checkPasswordStrength = (password: string) => {
    let strength = 0;
    if (password.length >= 8) strength += 25;
    if (/[a-z]/.test(password)) strength += 25;
    if (/[A-Z]/.test(password)) strength += 25;
    if (/[0-9]/.test(password)) strength += 25;
    return strength;
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const password = e.target.value;
    setPasswordStrength(checkPasswordStrength(password));
  };

  const getPasswordStrengthColor = () => {
    if (passwordStrength < 25) return '#ff4d4f';
    if (passwordStrength < 50) return '#faad14';
    if (passwordStrength < 75) return '#1890ff';
    return '#52c41a';
  };

  const getPasswordStrengthText = () => {
    if (passwordStrength < 25) return '弱';
    if (passwordStrength < 50) return '一般';
    if (passwordStrength < 75) return '良好';
    return '强';
  };

  const handleSubmit = async (values: RegisterFormData) => {
    try {
      setLoading(true);
      
      const registerData: RegisterRequest = {
        username: values.username,
        email: values.email,
        phone: values.phone,
        password: values.password,
      };

      const result = await dispatch(registerAsync(registerData));
      
      if (registerAsync.fulfilled.match(result)) {
        message.success('注册成功！请登录您的账户');
        router.push(ROUTES.LOGIN);
      } else {
        message.error(result.payload as string || '注册失败，请稍后重试');
      }
    } catch (error) {
      console.error('Register error:', error);
      message.error('注册失败，请稍后重试');
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
          maxWidth: 450,
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
            创建您的账户，开始购物之旅
          </Text>
        </div>

        <Form
          form={form}
          name="register"
          onFinish={handleSubmit}
          onFinishFailed={handleFormFailed}
          autoComplete="off"
          size="large"
          layout="vertical"
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3位字符' },
              { max: 20, message: '用户名最多20位字符' },
              { pattern: /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/, message: '用户名只能包含字母、数字、下划线和中文' },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="请输入用户名"
              autoComplete="username"
            />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱地址"
            rules={[
              { required: true, message: '请输入邮箱地址' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input
              prefix={<MailOutlined />}
              placeholder="请输入邮箱地址"
              autoComplete="email"
            />
          </Form.Item>

          <Form.Item
            name="phone"
            label="手机号码"
            rules={[
              { required: true, message: '请输入手机号码' },
              { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号码' },
            ]}
          >
            <Input
              prefix={<PhoneOutlined />}
              placeholder="请输入手机号码"
              autoComplete="tel"
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
              autoComplete="new-password"
              onChange={handlePasswordChange}
              iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
            />
          </Form.Item>

          {passwordStrength > 0 && (
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                <Text style={{ fontSize: 12 }}>密码强度</Text>
                <Text style={{ fontSize: 12, color: getPasswordStrengthColor() }}>
                  {getPasswordStrengthText()}
                </Text>
              </div>
              <Progress
                percent={passwordStrength}
                strokeColor={getPasswordStrengthColor()}
                showInfo={false}
                size="small"
              />
            </div>
          )}

          <Form.Item
            name="confirmPassword"
            label="确认密码"
            dependencies={['password']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请再次输入密码"
              autoComplete="new-password"
              iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
            />
          </Form.Item>

          <Form.Item
            name="agreement"
            valuePropName="checked"
            rules={[
              { validator: (_, value) => value ? Promise.resolve() : Promise.reject(new Error('请同意用户协议')) },
            ]}
          >
            <Checkbox>
              我已阅读并同意{' '}
              <Link href="/terms" style={{ color: '#1890ff' }}>
                《用户协议》
              </Link>
              {' '}和{' '}
              <Link href="/privacy" style={{ color: '#1890ff' }}>
                《隐私政策》
              </Link>
            </Checkbox>
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
              {loading ? '注册中...' : '立即注册'}
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <Text type="secondary">
              已有账户？{' '}
              <Link href={ROUTES.LOGIN} style={{ color: '#1890ff', fontWeight: 500 }}>
                立即登录
              </Link>
            </Text>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default RegisterPage;
