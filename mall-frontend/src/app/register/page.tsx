'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Form,
  Input,
  Button,
  Card,
  Typography,
  Checkbox,
  message,
  Progress,
  Alert,
  Space,
  Tooltip,
  Steps,
} from 'antd';
import {
  UserOutlined,
  LockOutlined,
  MailOutlined,
  PhoneOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
  CheckCircleOutlined,
  InfoCircleOutlined,
  LoadingOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAppDispatch, useAppSelector } from '@/store';
import {
  registerAsync,
  selectAuth,
  clearError,
} from '@/store/slices/authSlice';
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

interface ValidationStatus {
  username: 'success' | 'error' | 'validating' | '';
  email: 'success' | 'error' | 'validating' | '';
  phone: 'success' | 'error' | 'validating' | '';
  password: 'success' | 'error' | 'validating' | '';
}

const RegisterPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);
  const [currentStep, setCurrentStep] = useState(0);
  const [validationStatus, setValidationStatus] = useState<ValidationStatus>({
    username: '',
    email: '',
    phone: '',
    password: '',
  });
  const [emailChecking, setEmailChecking] = useState(false);
  const [usernameChecking, setUsernameChecking] = useState(false);

  const router = useRouter();
  const dispatch = useAppDispatch();
  const {
    isAuthenticated,
    user,
    error,
    loading: authLoading,
  } = useAppSelector(selectAuth);

  // 如果已登录，重定向到首页
  useEffect(() => {
    if (isAuthenticated && user) {
      router.push(ROUTES.HOME);
    }
  }, [isAuthenticated, user, router]);

  // 清除错误状态
  useEffect(() => {
    return () => {
      dispatch(clearError());
    };
  }, [dispatch]);

  // 实时验证用户名
  const validateUsername = useCallback(
    async (username: string) => {
      if (!username || username.length < 3) return;

      setUsernameChecking(true);
      setValidationStatus(prev => ({ ...prev, username: 'validating' }));

      try {
        // 模拟API调用检查用户名是否存在
        await new Promise(resolve => setTimeout(resolve, 800));

        // 这里应该调用真实的API
        const isAvailable = !['admin', 'test', 'user'].includes(
          username.toLowerCase()
        );

        setValidationStatus(prev => ({
          ...prev,
          username: isAvailable ? 'success' : 'error',
        }));

        if (!isAvailable) {
          form.setFields([
            {
              name: 'username',
              errors: ['该用户名已被使用，请选择其他用户名'],
            },
          ]);
        }
      } catch (error) {
        setValidationStatus(prev => ({ ...prev, username: 'error' }));
      } finally {
        setUsernameChecking(false);
      }
    },
    [form]
  );

  // 实时验证邮箱
  const validateEmail = useCallback(
    async (email: string) => {
      if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return;

      setEmailChecking(true);
      setValidationStatus(prev => ({ ...prev, email: 'validating' }));

      try {
        // 模拟API调用检查邮箱是否存在
        await new Promise(resolve => setTimeout(resolve, 600));

        // 这里应该调用真实的API
        const isAvailable = !['test@example.com', 'admin@example.com'].includes(
          email.toLowerCase()
        );

        setValidationStatus(prev => ({
          ...prev,
          email: isAvailable ? 'success' : 'error',
        }));

        if (!isAvailable) {
          form.setFields([
            {
              name: 'email',
              errors: ['该邮箱已被注册，请使用其他邮箱或直接登录'],
            },
          ]);
        }
      } catch (error) {
        setValidationStatus(prev => ({ ...prev, email: 'error' }));
      } finally {
        setEmailChecking(false);
      }
    },
    [form]
  );

  // 密码强度检测
  const checkPasswordStrength = (password: string) => {
    let strength = 0;
    const checks = {
      length: password.length >= 8,
      lowercase: /[a-z]/.test(password),
      uppercase: /[A-Z]/.test(password),
      number: /[0-9]/.test(password),
      special: /[!@#$%^&*(),.?":{}|<>]/.test(password),
    };

    if (checks.length) strength += 20;
    if (checks.lowercase) strength += 20;
    if (checks.uppercase) strength += 20;
    if (checks.number) strength += 20;
    if (checks.special) strength += 20;

    return { strength, checks };
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const password = e.target.value;
    const result = checkPasswordStrength(password);
    setPasswordStrength(result.strength);

    // 更新步骤
    if (password) {
      setCurrentStep(2);
    }

    // 设置密码验证状态
    if (password.length > 0) {
      if (result.strength >= 60) {
        setValidationStatus(prev => ({ ...prev, password: 'success' }));
      } else if (result.strength >= 40) {
        setValidationStatus(prev => ({ ...prev, password: 'validating' }));
      } else {
        setValidationStatus(prev => ({ ...prev, password: 'error' }));
      }
    } else {
      setValidationStatus(prev => ({ ...prev, password: '' }));
    }
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
        nickname: values.username, // 使用用户名作为昵称
      };

      const result = await dispatch(registerAsync(registerData));

      if (registerAsync.fulfilled.match(result)) {
        message.success('注册成功！请登录您的账户');
        router.push(ROUTES.LOGIN);
      } else {
        message.error((result.payload as string) || '注册失败，请稍后重试');
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
          <Text type='secondary'>创建您的账户，开始购物之旅</Text>
        </div>

        {/* 注册步骤指示器 */}
        <Steps
          current={currentStep}
          size='small'
          style={{ marginBottom: 24 }}
          items={[
            {
              title: '基本信息',
              icon: <UserOutlined />,
            },
            {
              title: '安全设置',
              icon: <LockOutlined />,
            },
            {
              title: '完成注册',
              icon: <CheckCircleOutlined />,
            },
          ]}
        />

        {/* 错误提示 */}
        {error && (
          <Alert
            message='注册失败'
            description={error}
            type='error'
            showIcon
            closable
            style={{ marginBottom: 16 }}
            onClose={() => dispatch(clearError())}
          />
        )}

        <Form
          form={form}
          name='register'
          onFinish={handleSubmit}
          onFinishFailed={handleFormFailed}
          autoComplete='off'
          size='large'
          layout='vertical'
        >
          <Form.Item
            name='username'
            label={
              <Space>
                用户名
                <Tooltip title='用户名将作为您的唯一标识，3-20位字符，支持中文、英文、数字和下划线'>
                  <InfoCircleOutlined style={{ color: '#1890ff' }} />
                </Tooltip>
              </Space>
            }
            validateStatus={validationStatus.username}
            hasFeedback={validationStatus.username !== ''}
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3位字符' },
              { max: 20, message: '用户名最多20位字符' },
              {
                pattern: /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/,
                message: '用户名只能包含字母、数字、下划线和中文',
              },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              suffix={usernameChecking ? <LoadingOutlined /> : null}
              placeholder='请输入用户名'
              autoComplete='username'
              onChange={e => {
                const value = e.target.value;
                if (value.length >= 3) {
                  validateUsername(value);
                } else {
                  setValidationStatus(prev => ({ ...prev, username: '' }));
                }
                setCurrentStep(value ? 1 : 0);
              }}
            />
          </Form.Item>

          <Form.Item
            name='email'
            label={
              <Space>
                邮箱地址
                <Tooltip title='邮箱将用于账户验证、密码重置和重要通知'>
                  <InfoCircleOutlined style={{ color: '#1890ff' }} />
                </Tooltip>
              </Space>
            }
            validateStatus={validationStatus.email}
            hasFeedback={validationStatus.email !== ''}
            rules={[
              { required: true, message: '请输入邮箱地址' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input
              prefix={<MailOutlined />}
              suffix={emailChecking ? <LoadingOutlined /> : null}
              placeholder='请输入邮箱地址'
              autoComplete='email'
              onChange={e => {
                const value = e.target.value;
                if (value && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
                  validateEmail(value);
                } else {
                  setValidationStatus(prev => ({ ...prev, email: '' }));
                }
              }}
            />
          </Form.Item>

          <Form.Item
            name='phone'
            label='手机号码'
            rules={[
              { required: true, message: '请输入手机号码' },
              { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号码' },
            ]}
          >
            <Input
              prefix={<PhoneOutlined />}
              placeholder='请输入手机号码'
              autoComplete='tel'
            />
          </Form.Item>

          <Form.Item
            name='password'
            label={
              <Space>
                密码
                <Tooltip title='密码至少8位，建议包含大小写字母、数字和特殊字符'>
                  <InfoCircleOutlined style={{ color: '#1890ff' }} />
                </Tooltip>
              </Space>
            }
            validateStatus={validationStatus.password}
            hasFeedback={validationStatus.password !== ''}
            rules={[
              { required: true, message: '请输入密码' },
              { min: 8, message: '密码至少8位字符' },
              {
                validator: (_, value) => {
                  if (!value) return Promise.resolve();
                  const result = checkPasswordStrength(value);
                  if (result.strength < 40) {
                    return Promise.reject(
                      new Error('密码强度太弱，请增强密码复杂度')
                    );
                  }
                  return Promise.resolve();
                },
              },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder='请输入密码'
              autoComplete='new-password'
              onChange={handlePasswordChange}
              iconRender={visible =>
                visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>

          {passwordStrength > 0 && (
            <div style={{ marginBottom: 16 }}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  marginBottom: 4,
                }}
              >
                <Text style={{ fontSize: 12 }}>密码强度</Text>
                <Text
                  style={{ fontSize: 12, color: getPasswordStrengthColor() }}
                >
                  {getPasswordStrengthText()}
                </Text>
              </div>
              <Progress
                percent={passwordStrength}
                strokeColor={getPasswordStrengthColor()}
                showInfo={false}
                size='small'
              />

              {/* 密码要求检查 */}
              <div style={{ marginTop: 8 }}>
                <Space
                  direction='vertical'
                  size='small'
                  style={{ width: '100%' }}
                >
                  {(() => {
                    const password = form.getFieldValue('password') || '';
                    const result = checkPasswordStrength(password);
                    return (
                      <div
                        style={{
                          display: 'grid',
                          gridTemplateColumns: '1fr 1fr',
                          gap: '4px',
                          fontSize: '12px',
                        }}
                      >
                        <Text
                          type={result.checks.length ? 'success' : 'secondary'}
                        >
                          {result.checks.length ? '✓' : '○'} 至少8位字符
                        </Text>
                        <Text
                          type={
                            result.checks.lowercase ? 'success' : 'secondary'
                          }
                        >
                          {result.checks.lowercase ? '✓' : '○'} 包含小写字母
                        </Text>
                        <Text
                          type={
                            result.checks.uppercase ? 'success' : 'secondary'
                          }
                        >
                          {result.checks.uppercase ? '✓' : '○'} 包含大写字母
                        </Text>
                        <Text
                          type={result.checks.number ? 'success' : 'secondary'}
                        >
                          {result.checks.number ? '✓' : '○'} 包含数字
                        </Text>
                        <Text
                          type={result.checks.special ? 'success' : 'secondary'}
                          style={{ gridColumn: '1 / -1' }}
                        >
                          {result.checks.special ? '✓' : '○'} 包含特殊字符
                          (!@#$%^&*等)
                        </Text>
                      </div>
                    );
                  })()}
                </Space>
              </div>
            </div>
          )}

          <Form.Item
            name='confirmPassword'
            label='确认密码'
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
              placeholder='请再次输入密码'
              autoComplete='new-password'
              iconRender={visible =>
                visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>

          <Form.Item
            name='agreement'
            valuePropName='checked'
            rules={[
              {
                validator: (_, value) =>
                  value
                    ? Promise.resolve()
                    : Promise.reject(new Error('请同意用户协议')),
              },
            ]}
          >
            <Checkbox>
              我已阅读并同意{' '}
              <Link href='/terms' style={{ color: '#1890ff' }}>
                《用户协议》
              </Link>{' '}
              和{' '}
              <Link href='/privacy' style={{ color: '#1890ff' }}>
                《隐私政策》
              </Link>
            </Checkbox>
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
              {loading ? '注册中...' : '立即注册'}
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <Text type='secondary'>
              已有账户？{' '}
              <Link
                href={ROUTES.LOGIN}
                style={{ color: '#1890ff', fontWeight: 500 }}
              >
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
