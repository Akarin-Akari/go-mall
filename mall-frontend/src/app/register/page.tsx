'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
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
import { authAPI } from '@/services/api';

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
  const usernameDebounceRef = useRef<ReturnType<typeof setTimeout> | null>(
    null
  );
  const emailDebounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastValuesRef = useRef<{ username: string; email: string }>({
    username: '',
    email: '',
  });
  const validateUsernameRef = useRef<((v: string) => void) | undefined>(undefined);
  const validateEmailRef = useRef<((v: string) => void) | undefined>(undefined);

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

  // 组件挂载日志（用于诊断渲染是否来自当前文件）
  useEffect(() => {
    if (process.env.NODE_ENV === 'development') {
      console.debug('🎯 RegisterPage 组件已挂载');
    }
  }, []);

  // 自动化测试钩子：当 URL 带有 ?autotest=1 时，自动填充并触发实时校验
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const sp = new URLSearchParams(window.location.search);
      if (sp.get('autotest') === '1') {
        const usernameVal = 'testuser_e2e';
        const emailVal = 'test@example.com';
        form.setFieldsValue({ username: usernameVal, email: emailVal });
        if (process.env.NODE_ENV === 'development') {
          console.debug('🤖 自动化测试：填充值并触发校验', {
            usernameVal,
            emailVal,
          });
        }
        // 稍作延迟，确保受控值与表单状态同步
        setTimeout(() => {
          validateUsername(usernameVal);
          validateEmail(emailVal);
        }, 300);
      }
    }
    // 仅在首次挂载检查一次
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 兜底轮询（占位，具体实现见下方定义完成后）

  // 实时验证用户名（接入真实 API）
  const validateUsername = useCallback(
    async (username: string) => {
      if (process.env.NODE_ENV === 'development') {
        console.debug('🔍 validateUsername 被调用，参数:', username);
      }
      if (!username || username.length < 3) return;

      setUsernameChecking(true);
      setValidationStatus(prev => ({ ...prev, username: 'validating' }));

      try {
        if (process.env.NODE_ENV === 'development') {
          console.debug(
            '🌐 准备调用 API:',
            'authAPI.checkUsernameAvailable',
            username
          );
        }
        const res = await authAPI.checkUsernameAvailable(username);
        const isAvailable = !!res.data?.available;

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

  // 实时验证邮箱（接入真实 API）
  const validateEmail = useCallback(
    async (email: string) => {
      if (process.env.NODE_ENV === 'development') {
        console.debug('🔍 validateEmail 被调用，参数:', email);
      }
      if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return;

      setEmailChecking(true);
      setValidationStatus(prev => ({ ...prev, email: 'validating' }));

      try {
        if (process.env.NODE_ENV === 'development') {
          console.debug(
            '🌐 准备调用 API:',
            'authAPI.checkEmailAvailable',
            email
          );
        }
        const res = await authAPI.checkEmailAvailable(email);
        const isAvailable = !!res.data?.available;

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

  // 兜底轮询：监听表单值变化（用户名/邮箱），满足条件时以800ms防抖触发校验
  useEffect(() => {
    // 将回调函数引用同步到 ref，避免初始化时序导致的引用错误
    validateUsernameRef.current = validateUsername;
    validateEmailRef.current = validateEmail;

    const intervalId = setInterval(() => {
      const usernameVal: string = form.getFieldValue('username') || '';
      const emailVal: string = form.getFieldValue('email') || '';

      if (usernameVal !== lastValuesRef.current.username) {
        lastValuesRef.current.username = usernameVal;
        if (usernameDebounceRef.current)
          clearTimeout(usernameDebounceRef.current);
        if (usernameVal && usernameVal.length >= 3) {
          usernameDebounceRef.current = setTimeout(() => {
            if (process.env.NODE_ENV === 'development') {
              console.debug('⏲️ 轮询触发用户名校验:', usernameVal);
            }
            validateUsernameRef.current &&
              validateUsernameRef.current(usernameVal);
          }, 800);
        }
      }

      if (emailVal !== lastValuesRef.current.email) {
        lastValuesRef.current.email = emailVal;
        if (emailDebounceRef.current) clearTimeout(emailDebounceRef.current);
        if (emailVal && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailVal)) {
          emailDebounceRef.current = setTimeout(() => {
            if (process.env.NODE_ENV === 'development') {
              console.debug('⏲️ 轮询触发邮箱校验:', emailVal);
            }
            validateEmailRef.current && validateEmailRef.current(emailVal);
          }, 800);
        }
      }
    }, 300);

    return () => {
      clearInterval(intervalId);
      if (usernameDebounceRef.current)
        clearTimeout(usernameDebounceRef.current);
      if (emailDebounceRef.current) clearTimeout(emailDebounceRef.current);
    };
  }, [form, validateUsername, validateEmail]);

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
        const errorMsg =
          typeof result.payload === 'string'
            ? result.payload
            : (result.payload as any)?.message ||
              (result.payload as any)?.error ||
              JSON.stringify(result.payload);
        message.error(errorMsg || '注册失败，请稍后重试');
      }
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Register error:', error);
      }
      message.error('注册失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  const handleFormFailed = (errorInfo: any) => {
    const firstMsg =
      errorInfo?.errorFields?.[0]?.errors?.[0] ||
      errorInfo?.errorFields
        ?.map((f: any) => f.errors?.[0])
        .filter(Boolean)
        .join('；') ||
      '请检查表单输入';

    if (process.env.NODE_ENV === 'development') {
      const errorSummary =
        errorInfo?.errorFields?.map((f: any) => ({
          field: f.name?.[0] || 'unknown',
          errors: f.errors || [],
        })) || [];
      console.debug('表单验证失败:', firstMsg);
      if (errorSummary.length > 0) {
        console.debug('验证错误详情:', JSON.stringify(errorSummary, null, 2));
      }
    }

    message.error(firstMsg);
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
            description={
              typeof error === 'string'
                ? error
                : (error as any)?.message ||
                  (error as any)?.error ||
                  JSON.stringify(error) ||
                  '请求失败，请稍后重试'
            }
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
          onValuesChange={(changedValues, allValues) => {
            if (process.env.NODE_ENV === 'development') {
              console.debug('📋 Form 值变化:', { changedValues, allValues });
            }

            // 用户名防抖触发
            if (
              Object.prototype.hasOwnProperty.call(changedValues, 'username')
            ) {
              const value = (changedValues as any).username as string;
              if (usernameDebounceRef.current) {
                clearTimeout(usernameDebounceRef.current);
              }
              usernameDebounceRef.current = setTimeout(() => {
                if (value && value.length >= 3) {
                  if (process.env.NODE_ENV === 'development') {
                    console.debug('⏱️ 用户名防抖触发校验:', value);
                  }
                  validateUsername(value);
                }
              }, 600);
            }

            // 邮箱防抖触发
            if (Object.prototype.hasOwnProperty.call(changedValues, 'email')) {
              const value = (changedValues as any).email as string;
              if (emailDebounceRef.current) {
                clearTimeout(emailDebounceRef.current);
              }
              emailDebounceRef.current = setTimeout(() => {
                if (value && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
                  if (process.env.NODE_ENV === 'development') {
                    console.debug('⏱️ 邮箱防抖触发校验:', value);
                  }
                  validateEmail(value);
                }
              }, 600);
            }
          }}
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
              {
                // 兜底触发：在校验规则阶段也调用实时校验
                validator: async (_, value) => {
                  if (process.env.NODE_ENV === 'development') {
                    console.debug('🔧 用户名 validator 被触发，值:', value);
                  }
                  if (value && value.length >= 3) {
                    await validateUsername(value);
                  }
                  return Promise.resolve();
                },
              },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              suffix={usernameChecking ? <LoadingOutlined /> : null}
              placeholder='请输入用户名'
              autoComplete='username'
              value={form.getFieldValue('username')}
              onBlur={e => {
                const value = e.target.value;
                if (process.env.NODE_ENV === 'development') {
                  console.debug('👆 用户名失焦，值:', value);
                }
                if (value && value.length >= 3) {
                  validateUsername(value);
                }
              }}
              onChange={e => {
                if (process.env.NODE_ENV === 'development') {
                  console.debug('📝 用户名输入变化:', e.target.value);
                }
                const value = e.target.value;
                form.setFieldsValue({ username: value });
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
              {
                // 兜底触发：在校验规则阶段也调用实时校验
                validator: async (_, value) => {
                  if (process.env.NODE_ENV === 'development') {
                    console.debug('🔧 邮箱 validator 被触发，值:', value);
                  }
                  if (value && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
                    await validateEmail(value);
                  }
                  return Promise.resolve();
                },
              },
            ]}
          >
            <Input
              prefix={<MailOutlined />}
              suffix={emailChecking ? <LoadingOutlined /> : null}
              placeholder='请输入邮箱地址'
              autoComplete='email'
              value={form.getFieldValue('email')}
              onBlur={e => {
                const value = e.target.value;
                if (process.env.NODE_ENV === 'development') {
                  console.debug('👆 邮箱失焦，值:', value);
                }
                if (value && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
                  validateEmail(value);
                }
              }}
              onChange={e => {
                if (process.env.NODE_ENV === 'development') {
                  console.debug('📝 邮箱输入变化:', e.target.value);
                }
                const value = e.target.value;
                form.setFieldsValue({ email: value });
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
              {
                validator: (_, value) => {
                  if (!value)
                    return Promise.reject(new Error('请输入手机号码'));
                  const normalized = String(value).replace(/[\s-]/g, '');
                  const e164 = /^\+?[1-9]\d{6,14}$/; // 国际E.164: 7-15位
                  const cn = /^1[3-9]\d{9}$/; // 中国大陆手机号
                  return e164.test(normalized) || cn.test(normalized)
                    ? Promise.resolve()
                    : Promise.reject(
                        new Error(
                          '请输入有效的手机号码（支持国际格式，如 +1 202-555-0123）'
                        )
                      );
                },
              },
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
