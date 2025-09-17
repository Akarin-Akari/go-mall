/**
 * 应用配置文件
 * 简化的配置管理，基于环境变量
 */

// API配置
export const API_CONFIG = {
  BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
  TIMEOUT: parseInt(process.env.NEXT_PUBLIC_API_TIMEOUT || '10000'),
  VERSION: 'v1',
} as const;

// 应用配置
export const APP_CONFIG = {
  NAME: 'Mall Frontend',
  VERSION: '1.0.0',
  DESCRIPTION: 'Next.js电商前端应用',
  AUTHOR: 'Mall Team',
} as const;

// 功能开关配置
export const FEATURE_FLAGS = {
  ENABLE_DEBUG: process.env.NODE_ENV === 'development',
  ENABLE_ANALYTICS: process.env.NEXT_PUBLIC_ENABLE_ANALYTICS === 'true',
  ENABLE_ERROR_REPORTING: process.env.NEXT_PUBLIC_ENABLE_ERROR_REPORTING === 'true',
  ENABLE_PERFORMANCE_MONITORING: process.env.NEXT_PUBLIC_ENABLE_PERFORMANCE_MONITORING === 'true',
} as const;

// 缓存配置
export const CACHE_CONFIG = {
  DEFAULT_TTL: 5 * 60 * 1000, // 5分钟
  USER_INFO_TTL: 30 * 60 * 1000, // 30分钟
  PRODUCT_LIST_TTL: 10 * 60 * 1000, // 10分钟
  CART_TTL: 60 * 60 * 1000, // 1小时
} as const;

// 安全配置
export const SECURITY_CONFIG = {
  TOKEN_STORAGE_KEY: 'mall_token',
  REFRESH_TOKEN_STORAGE_KEY: 'mall_refresh_token',
  TOKEN_EXPIRY_BUFFER: 5 * 60 * 1000, // 5分钟缓冲时间
  MAX_LOGIN_ATTEMPTS: 5,
  LOCKOUT_DURATION: 15 * 60 * 1000, // 15分钟
} as const;

// UI配置
export const UI_CONFIG = {
  THEME: {
    PRIMARY_COLOR: '#1890ff',
    BORDER_RADIUS: 6,
    ANIMATION_DURATION: 300,
  },
  PAGINATION: {
    DEFAULT_PAGE_SIZE: 10,
    PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
    MAX_PAGE_SIZE: 100,
  },
  UPLOAD: {
    MAX_FILE_SIZE: 10 * 1024 * 1024, // 10MB
    ALLOWED_IMAGE_TYPES: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
    ALLOWED_FILE_TYPES: [
      'application/pdf',
      'application/msword',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    ],
  },
} as const;

// 业务配置
export const BUSINESS_CONFIG = {
  ORDER: {
    AUTO_CANCEL_TIMEOUT: 30 * 60 * 1000, // 30分钟自动取消未支付订单
    DELIVERY_TIME_SLOTS: [
      '09:00-12:00',
      '12:00-15:00',
      '15:00-18:00',
      '18:00-21:00',
    ],
  },
  CART: {
    MAX_ITEMS: 99,
    SYNC_INTERVAL: 30 * 1000, // 30秒同步一次
  },
  PRODUCT: {
    IMAGES_PER_PRODUCT: 10,
    MAX_DESCRIPTION_LENGTH: 5000,
  },
} as const;

// 开发配置
export const DEV_CONFIG = {
  ENABLE_MOCK_DATA: process.env.NEXT_PUBLIC_ENABLE_MOCK === 'true',
  ENABLE_DEBUG_LOGS: process.env.NODE_ENV === 'development',
  SHOW_PERFORMANCE_METRICS: process.env.NODE_ENV === 'development',
} as const;

// 导出所有配置
export const CONFIG = {
  API: API_CONFIG,
  APP: APP_CONFIG,
  FEATURES: FEATURE_FLAGS,
  CACHE: CACHE_CONFIG,
  SECURITY: SECURITY_CONFIG,
  UI: UI_CONFIG,
  BUSINESS: BUSINESS_CONFIG,
  DEV: DEV_CONFIG,
} as const;

// 便捷的配置获取函数
export const getConfig = <T extends keyof typeof CONFIG>(
  section: T
): typeof CONFIG[T] => {
  return CONFIG[section];
};

// 检查功能是否启用
export const isFeatureEnabled = (feature: keyof typeof FEATURE_FLAGS): boolean => {
  return FEATURE_FLAGS[feature];
};

// 获取API端点
export const getApiUrl = (path: string): string => {
  const baseUrl = API_CONFIG.BASE_URL.replace(/\/$/, '');
  const cleanPath = path.replace(/^\//, '');
  return `${baseUrl}/api/${API_CONFIG.VERSION}/${cleanPath}`;
};

// 环境检查
export const isDevelopment = (): boolean => process.env.NODE_ENV === 'development';
export const isProduction = (): boolean => process.env.NODE_ENV === 'production';
export const isTest = (): boolean => process.env.NODE_ENV === 'test';

export default CONFIG;
