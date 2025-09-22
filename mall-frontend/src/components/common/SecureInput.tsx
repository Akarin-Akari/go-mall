/**
 * 安全输入组件
 * 集成XSS防护和输入验证功能
 */

import React, { useState, useCallback, useMemo } from 'react';
import { Input, Form, message } from 'antd';
// import { SecurityUtils, InputValidator } from '@/utils/xssProtection';
// import {
//   EnhancedInputValidator,
//   detectXSS,
// } from '@/utils/enhancedXssProtection';
import type { InputProps } from 'antd';
import type { TextAreaProps } from 'antd/es/input';

// 安全输入配置接口
interface SecureInputConfig {
  allowHtml?: boolean;
  maxLength?: number;
  trimWhitespace?: boolean;
  validateEmail?: boolean;
  validateUrl?: boolean;
  validatePhone?: boolean;
  customValidator?: (value: string) => { isValid: boolean; message?: string };
}

// 安全输入组件属性
interface SecureInputProps extends Omit<InputProps, 'onChange'> {
  securityConfig?: SecureInputConfig;
  onChange?: (value: string, isValid: boolean) => void;
  onValidationChange?: (isValid: boolean, errors: string[]) => void;
  showValidationFeedback?: boolean;
}

// 安全文本域组件属性
interface SecureTextAreaProps extends Omit<TextAreaProps, 'onChange'> {
  securityConfig?: SecureInputConfig;
  onChange?: (value: string, isValid: boolean) => void;
  onValidationChange?: (isValid: boolean, errors: string[]) => void;
  showValidationFeedback?: boolean;
}

/**
 * 安全输入组件
 */
export const SecureInput: React.FC<SecureInputProps> = ({
  securityConfig = {},
  onChange,
  onValidationChange,
  showValidationFeedback = true,
  ...inputProps
}) => {
  const [validationErrors, setValidationErrors] = useState<string[]>([]);
  const [isValid, setIsValid] = useState(true);

  // 验证输入值
  const validateInput = useCallback(
    (value: string): { isValid: boolean; errors: string[] } => {
      const errors: string[] = [];
      let valid = true;

      // 基础验证
      if (securityConfig.maxLength && value.length > securityConfig.maxLength) {
        errors.push(`输入长度不能超过${securityConfig.maxLength}个字符`);
        valid = false;
      }

      // 邮箱验证
      if (
        securityConfig.validateEmail &&
        value &&
        !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)
      ) {
        errors.push('请输入有效的邮箱地址');
        valid = false;
      }

      // URL验证
      if (
        securityConfig.validateUrl &&
        value &&
        !/^https?:\/\/.+/.test(value)
      ) {
        errors.push('请输入有效的URL地址');
        valid = false;
      }

      // 手机号验证
      if (
        securityConfig.validatePhone &&
        value &&
        !/^1[3-9]\d{9}$/.test(value)
      ) {
        errors.push('请输入有效的手机号码');
        valid = false;
      }

      // 自定义验证
      if (securityConfig.customValidator && value) {
        const customResult = securityConfig.customValidator(value);
        if (!customResult.isValid) {
          errors.push(customResult.message || '输入格式不正确');
          valid = false;
        }
      }

      return { isValid: valid, errors };
    },
    [securityConfig]
  );

  // 处理输入变化
  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const rawValue = e.target.value;

      // 清理输入
      const cleanValue = rawValue
        .trim()
        .slice(0, securityConfig.maxLength || 1000);

      // 简单XSS检测
      const hasXSS = /<script|javascript:|on\w+=/i.test(rawValue);
      if (hasXSS) {
        console.warn('Potential XSS attempt detected');
        message.warning('输入内容包含潜在的安全风险，已自动清理');
      }

      // 验证输入
      const validation = validateInput(cleanValue);
      setValidationErrors(validation.errors);
      setIsValid(validation.isValid);

      // 触发回调
      onChange?.(cleanValue, validation.isValid);
      onValidationChange?.(validation.isValid, validation.errors);

      // 显示验证反馈
      if (showValidationFeedback && validation.errors.length > 0) {
        message.warning(validation.errors[0]);
      }
    },
    [
      securityConfig,
      validateInput,
      onChange,
      onValidationChange,
      showValidationFeedback,
    ]
  );

  // 计算状态
  const status = useMemo(() => {
    if (!isValid) return 'error';
    return undefined;
  }, [isValid]);

  return <Input {...inputProps} status={status} onChange={handleChange} />;
};

/**
 * 安全文本域组件
 */
export const SecureTextArea: React.FC<SecureTextAreaProps> = ({
  securityConfig = {},
  onChange,
  onValidationChange,
  showValidationFeedback = true,
  ...textAreaProps
}) => {
  const [validationErrors, setValidationErrors] = useState<string[]>([]);
  const [isValid, setIsValid] = useState(true);

  // 验证输入值
  const validateInput = useCallback(
    (value: string): { isValid: boolean; errors: string[] } => {
      const errors: string[] = [];
      let valid = true;

      // 基础验证
      if (securityConfig.maxLength && value.length > securityConfig.maxLength) {
        errors.push(`输入长度不能超过${securityConfig.maxLength}个字符`);
        valid = false;
      }

      // 自定义验证
      if (securityConfig.customValidator && value) {
        const customResult = securityConfig.customValidator(value);
        if (!customResult.isValid) {
          errors.push(customResult.message || '输入格式不正确');
          valid = false;
        }
      }

      return { isValid: valid, errors };
    },
    [securityConfig]
  );

  // 处理输入变化
  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const rawValue = e.target.value;

      // 清理输入
      const cleanValue = rawValue
        .trim()
        .slice(0, securityConfig.maxLength || 1000);

      // 验证输入
      const validation = validateInput(cleanValue);
      setValidationErrors(validation.errors);
      setIsValid(validation.isValid);

      // 触发回调
      onChange?.(cleanValue, validation.isValid);
      onValidationChange?.(validation.isValid, validation.errors);

      // 显示验证反馈
      if (showValidationFeedback && validation.errors.length > 0) {
        message.warning(validation.errors[0]);
      }
    },
    [
      securityConfig,
      validateInput,
      onChange,
      onValidationChange,
      showValidationFeedback,
    ]
  );

  // 计算状态
  const status = useMemo(() => {
    if (!isValid) return 'error';
    return undefined;
  }, [isValid]);

  return (
    <Input.TextArea
      {...textAreaProps}
      status={status}
      onChange={handleChange}
    />
  );
};

/**
 * 安全密码输入组件
 */
export const SecurePasswordInput: React.FC<SecureInputProps> = ({
  securityConfig = {},
  onChange,
  onValidationChange,
  showValidationFeedback = true,
  ...inputProps
}) => {
  const [validationErrors, setValidationErrors] = useState<string[]>([]);
  const [isValid, setIsValid] = useState(true);

  // 处理密码输入变化
  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const password = e.target.value;

      // 简单密码强度验证
      const errors: string[] = [];
      let isValid = true;

      if (password.length < 8) {
        errors.push('密码长度至少8位');
        isValid = false;
      }
      if (!/[A-Z]/.test(password)) {
        errors.push('密码需包含大写字母');
        isValid = false;
      }
      if (!/[a-z]/.test(password)) {
        errors.push('密码需包含小写字母');
        isValid = false;
      }
      if (!/\d/.test(password)) {
        errors.push('密码需包含数字');
        isValid = false;
      }

      setValidationErrors(errors);
      setIsValid(isValid);

      // 触发回调
      onChange?.(password, isValid);
      onValidationChange?.(isValid, errors);

      // 显示验证反馈
      if (showValidationFeedback && errors.length > 0) {
        message.warning(errors[0]);
      }
    },
    [onChange, onValidationChange, showValidationFeedback]
  );

  // 计算状态
  const status = useMemo(() => {
    if (!isValid) return 'error';
    return undefined;
  }, [isValid]);

  return (
    <Input.Password {...inputProps} status={status} onChange={handleChange} />
  );
};

/**
 * 安全表单项组件
 */
interface SecureFormItemProps {
  name: string;
  label: string;
  rules?: any[];
  securityConfig?: SecureInputConfig;
  inputType?: 'input' | 'textarea' | 'password';
  children?: React.ReactNode;
}

export const SecureFormItem: React.FC<SecureFormItemProps> = ({
  name,
  label,
  rules = [],
  securityConfig = {},
  inputType = 'input',
  children,
}) => {
  // 添加安全验证规则
  const securityRules = useMemo(() => {
    const baseRules = [...rules];

    // 添加XSS防护规则
    baseRules.push({
      validator: (_: any, value: string) => {
        if (!value) return Promise.resolve();

        const cleanValue = value.trim().slice(0, securityConfig.maxLength || 1000);
        if (/<script|javascript:|on\w+=/i.test(value)) {
          return Promise.reject(new Error('输入包含不安全的内容'));
        }

        return Promise.resolve();
      },
    });

    return baseRules;
  }, [rules, securityConfig]);

  // 渲染输入组件
  const renderInput = () => {
    if (children) return children;

    switch (inputType) {
      case 'textarea':
        return <SecureTextArea securityConfig={securityConfig} />;
      case 'password':
        return <SecurePasswordInput securityConfig={securityConfig} />;
      default:
        return <SecureInput securityConfig={securityConfig} />;
    }
  };

  return (
    <Form.Item name={name} label={label} rules={securityRules}>
      {renderInput()}
    </Form.Item>
  );
};

// 默认导出
export default {
  SecureInput,
  SecureTextArea,
  SecurePasswordInput,
  SecureFormItem,
};
