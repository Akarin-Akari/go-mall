'use client';

import React, { useState, useEffect } from 'react';
import {
  Form,
  Input,
  Button,
  Card,
  Typography,
  message,
  Alert,
  Space,
  Steps,
  Result,
} from 'antd';
import {
  MailOutlined,
  SafetyCertificateOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  LockOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { ROUTES } from '@/constants';
import { authAPI } from '@/services/api';

const { Title, Text } = Typography;

interface ForgotPasswordFormData {
  email: string;
}

interface ResetPasswordFormData {
  code: string;
  newPassword: string;
  confirmPassword: string;
}

const ForgotPasswordPage: React.FC = () => {
  const [form] = Form.useForm();
  const [resetForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [currentStep, setCurrentStep] = useState(0);
  const [email, setEmail] = useState('');
  const [countdown, setCountdown] = useState(0);
  const [canResend, setCanResend] = useState(true);

  const router = useRouter();

  // 倒计时逻辑
  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (countdown > 0) {
      timer = setTimeout(() => {
        setCountdown(countdown - 1);
      }, 1000);
    } else if (countdown === 0 && !canResend) {
      setCanResend(true);
    }
    return () => clearTimeout(timer);
  }, [countdown, canResend]);

  // 发送验证码
  const handleSendCode = async (values: ForgotPasswordFormData) => {
    try {
      setLoading(true);

      await authAPI.sendResetCode({ email: values.email });

      setEmail(values.email);
      setCurrentStep(1);
      setCountdown(60);
      setCanResend(false);
      message.success('验证码已发送到您的邮箱，请查收');
    } catch (error) {
      message.error('发送失败，请检查邮箱地址是否正确');
    } finally {
      setLoading(false);
    }
  };

  // 重新发送验证码
  const handleResendCode = async () => {
    if (!canResend || !email) return;

    try {
      setLoading(true);
      await authAPI.sendResetCode({ email });
      setCountdown(60);
      setCanResend(false);
      message.success('验证码已重新发送');
    } catch (error) {
      message.error('重发失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  // 重置密码
  const handleResetPassword = async (values: ResetPasswordFormData) => {
    try {
      setLoading(true);

      await authAPI.resetPassword({
        email,
        code: values.code,
        new_password: values.newPassword,
      });

      setCurrentStep(2);
      message.success('密码重置成功！');
    } catch (error) {
      message.error('重置失败，请检查验证码是否正确');
    } finally {
      setLoading(false);
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <Form
            form={form}
            name='forgot-password'
            onFinish={handleSendCode}
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
                prefix={<MailOutlined />}
                placeholder='请输入注册时使用的邮箱地址'
                autoComplete='email'
              />
            </Form.Item>

            <Form.Item style={{ marginBottom: 16 }}>
              <Button
                type='primary'
                htmlType='submit'
                loading={loading}
                block
                icon={<SafetyCertificateOutlined />}
                style={{
                  height: 48,
                  fontSize: 16,
                  fontWeight: 500,
                }}
              >
                {loading ? '发送中...' : '发送验证码'}
              </Button>
            </Form.Item>

            <div style={{ textAlign: 'center' }}>
              <Text type='secondary'>
                想起密码了？{' '}
                <Link
                  href={ROUTES.LOGIN}
                  style={{ color: '#1890ff', fontWeight: 500 }}
                >
                  返回登录
                </Link>
              </Text>
            </div>
          </Form>
        );

      case 1:
        return (
          <Form
            form={resetForm}
            name='reset-password'
            onFinish={handleResetPassword}
            autoComplete='off'
            size='large'
            layout='vertical'
          >
            <Alert
              message='验证码已发送'
              description={
                <Space direction='vertical' size='small'>
                  <Text>验证码已发送到 {email}</Text>
                  <Text type='secondary'>请查收邮件并输入6位验证码</Text>
                </Space>
              }
              type='info'
              showIcon
              style={{ marginBottom: 16 }}
            />

            <Form.Item
              name='code'
              label='验证码'
              rules={[
                { required: true, message: '请输入验证码' },
                { len: 6, message: '验证码为6位数字' },
                { pattern: /^\d{6}$/, message: '验证码只能包含数字' },
              ]}
            >
              <Input
                placeholder='请输入6位验证码'
                maxLength={6}
                style={{ textAlign: 'center', fontSize: 18, letterSpacing: 4 }}
              />
            </Form.Item>

            <Form.Item
              name='newPassword'
              label='新密码'
              rules={[
                { required: true, message: '请输入新密码' },
                { min: 8, message: '密码至少8位字符' },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder='请输入新密码'
                autoComplete='new-password'
                iconRender={visible =>
                  visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
                }
              />
            </Form.Item>

            <Form.Item
              name='confirmPassword'
              label='确认新密码'
              dependencies={['newPassword']}
              rules={[
                { required: true, message: '请确认新密码' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('newPassword') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('两次输入的密码不一致'));
                  },
                }),
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder='请再次输入新密码'
                autoComplete='new-password'
                iconRender={visible =>
                  visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
                }
              />
            </Form.Item>

            <Form.Item style={{ marginBottom: 16 }}>
              <Button
                type='primary'
                htmlType='submit'
                loading={loading}
                block
                style={{
                  height: 48,
                  fontSize: 16,
                  fontWeight: 500,
                }}
              >
                {loading ? '重置中...' : '重置密码'}
              </Button>
            </Form.Item>

            <div style={{ textAlign: 'center' }}>
              <Space>
                <Text type='secondary'>没收到验证码？</Text>
                <Button
                  type='link'
                  size='small'
                  disabled={!canResend}
                  onClick={handleResendCode}
                  loading={loading}
                  style={{ padding: 0, height: 'auto' }}
                >
                  {canResend ? '重新发送' : `${countdown}秒后重发`}
                </Button>
              </Space>
            </div>
          </Form>
        );

      case 2:
        return (
          <Result
            status='success'
            title='密码重置成功！'
            subTitle='您的密码已成功重置，现在可以使用新密码登录了'
            extra={[
              <Button
                type='primary'
                key='login'
                onClick={() => router.push(ROUTES.LOGIN)}
              >
                立即登录
              </Button>,
              <Button key='home' onClick={() => router.push(ROUTES.HOME)}>
                返回首页
              </Button>,
            ]}
          />
        );

      default:
        return null;
    }
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
          <Text type='secondary'>
            {currentStep === 0 && '重置您的密码'}
            {currentStep === 1 && '验证您的身份'}
            {currentStep === 2 && '重置完成'}
          </Text>
        </div>

        {/* 步骤指示器 */}
        {currentStep < 2 && (
          <Steps
            current={currentStep}
            size='small'
            style={{ marginBottom: 32 }}
            items={[
              {
                title: '输入邮箱',
                icon: <MailOutlined />,
              },
              {
                title: '验证身份',
                icon: <SafetyCertificateOutlined />,
              },
              {
                title: '重置密码',
                icon: <CheckCircleOutlined />,
              },
            ]}
          />
        )}

        {renderStepContent()}
      </Card>
    </div>
  );
};

export default ForgotPasswordPage;
