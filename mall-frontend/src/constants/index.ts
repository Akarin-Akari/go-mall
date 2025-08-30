// API相关常量
export const API_ENDPOINTS = {
  // 认证相关
  AUTH: {
    LOGIN: '/api/v1/users/login',
    REGISTER: '/api/v1/users/register',
    LOGOUT: '/api/v1/users/logout',
    PROFILE: '/api/v1/users/profile',
    REFRESH_TOKEN: '/api/v1/auth/refresh',
  },
  
  // 用户相关
  USERS: {
    LIST: '/api/v1/users',
    DETAIL: (id: number) => `/api/v1/users/${id}`,
    UPDATE: (id: number) => `/api/v1/users/${id}`,
    DELETE: (id: number) => `/api/v1/users/${id}`,
  },
  
  // 商品相关
  PRODUCTS: {
    LIST: '/api/v1/products',
    DETAIL: (id: number) => `/api/v1/products/${id}`,
    CREATE: '/api/v1/products',
    UPDATE: (id: number) => `/api/v1/products/${id}`,
    DELETE: (id: number) => `/api/v1/products/${id}`,
    CATEGORIES: '/api/v1/products/categories',
  },
  
  // 购物车相关
  CART: {
    LIST: '/api/v1/cart',
    ADD: '/api/v1/cart/add',
    UPDATE: '/api/v1/cart/update',
    REMOVE: '/api/v1/cart/remove',
    CLEAR: '/api/v1/cart/clear',
    SYNC: '/api/v1/cart/sync',
  },
  
  // 订单相关
  ORDERS: {
    LIST: '/api/v1/orders',
    DETAIL: (id: number) => `/api/v1/orders/${id}`,
    CREATE: '/api/v1/orders',
    UPDATE_STATUS: (id: number) => `/api/v1/orders/${id}/status`,
    CANCEL: (id: number) => `/api/v1/orders/${id}/cancel`,
  },
  
  // 支付相关
  PAYMENT: {
    CREATE: '/api/v1/payment/create',
    QUERY: (id: number) => `/api/v1/payment/${id}`,
    CALLBACK: '/api/v1/payment/callback',
    REFUND: '/api/v1/payment/refund',
  },
  
  // 文件上传
  UPLOAD: {
    IMAGE: '/api/v1/files/upload/image',
    FILE: '/api/v1/files/upload',
    DELETE: (id: string) => `/api/v1/files/${id}`,
  },
  
  // 地址管理
  ADDRESS: {
    LIST: '/api/v1/address',
    CREATE: '/api/v1/address',
    UPDATE: (id: number) => `/api/v1/address/${id}`,
    DELETE: (id: number) => `/api/v1/address/${id}`,
    SET_DEFAULT: (id: number) => `/api/v1/address/${id}/default`,
  },
} as const;

// 状态常量
export const ORDER_STATUS = {
  PENDING: 'pending',
  PAID: 'paid',
  SHIPPED: 'shipped',
  DELIVERED: 'delivered',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
} as const;

export const PAYMENT_STATUS = {
  PENDING: 'pending',
  PAID: 'paid',
  FAILED: 'failed',
  REFUNDED: 'refunded',
} as const;

export const PAYMENT_METHODS = {
  ALIPAY: 'alipay',
  WECHAT: 'wechat',
  BALANCE: 'balance',
  UNIONPAY: 'unionpay',
} as const;

export const USER_ROLES = {
  ADMIN: 'admin',
  USER: 'user',
  MERCHANT: 'merchant',
} as const;

export const PRODUCT_STATUS = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
  DRAFT: 'draft',
} as const;

// 本地存储键名
export const STORAGE_KEYS = {
  TOKEN: 'mall_token',
  REFRESH_TOKEN: 'mall_refresh_token',
  USER_INFO: 'mall_user_info',
  CART_ITEMS: 'mall_cart_items',
  THEME: 'mall_theme',
  LANGUAGE: 'mall_language',
  REMEMBER_LOGIN: 'mall_remember_login',
} as const;

// 分页配置
export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
  MAX_PAGE_SIZE: 100,
} as const;

// 文件上传配置
export const UPLOAD_CONFIG = {
  MAX_SIZE: 10 * 1024 * 1024, // 10MB
  ALLOWED_IMAGE_TYPES: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
  ALLOWED_FILE_TYPES: ['application/pdf', 'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'],
} as const;

// 表单验证规则
export const VALIDATION_RULES = {
  USERNAME: {
    MIN_LENGTH: 3,
    MAX_LENGTH: 20,
    PATTERN: /^[a-zA-Z0-9_]+$/,
  },
  PASSWORD: {
    MIN_LENGTH: 6,
    MAX_LENGTH: 20,
    PATTERN: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]/,
  },
  EMAIL: {
    PATTERN: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  },
  PHONE: {
    PATTERN: /^1[3-9]\d{9}$/,
  },
} as const;

// 主题配置
export const THEME_CONFIG = {
  LIGHT: {
    primaryColor: '#1890ff',
    borderRadius: 6,
    colorBgContainer: '#ffffff',
  },
  DARK: {
    primaryColor: '#1890ff',
    borderRadius: 6,
    colorBgContainer: '#141414',
  },
} as const;

// 路由路径
export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  PROFILE: '/profile',
  
  // 商品相关
  PRODUCTS: '/products',
  PRODUCT_DETAIL: (id: number) => `/products/${id}`,
  CATEGORIES: '/categories',
  
  // 购物车和订单
  CART: '/cart',
  CHECKOUT: '/checkout',
  ORDERS: '/orders',
  ORDER_DETAIL: (id: number) => `/orders/${id}`,
  
  // 用户中心
  USER_CENTER: '/user',
  USER_ORDERS: '/user/orders',
  USER_ADDRESS: '/user/address',
  USER_SETTINGS: '/user/settings',
  
  // 管理后台
  ADMIN: '/admin',
  ADMIN_PRODUCTS: '/admin/products',
  ADMIN_ORDERS: '/admin/orders',
  ADMIN_USERS: '/admin/users',
  
  // 支付相关
  PAYMENT: '/payment',
  PAYMENT_SUCCESS: '/payment/success',
  PAYMENT_FAILED: '/payment/failed',
  
  // 错误页面
  NOT_FOUND: '/404',
  SERVER_ERROR: '/500',
  UNAUTHORIZED: '/401',
  FORBIDDEN: '/403',
} as const;

// 消息类型
export const MESSAGE_TYPES = {
  SUCCESS: 'success',
  ERROR: 'error',
  WARNING: 'warning',
  INFO: 'info',
} as const;

// 响应状态码
export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  CONFLICT: 409,
  UNPROCESSABLE_ENTITY: 422,
  TOO_MANY_REQUESTS: 429,
  INTERNAL_SERVER_ERROR: 500,
  BAD_GATEWAY: 502,
  SERVICE_UNAVAILABLE: 503,
} as const;

// 业务状态码
export const BUSINESS_CODE = {
  SUCCESS: 200,
  ERROR: 500,
  INVALID_PARAM: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  CONFLICT: 409,
  TOO_MANY_REQ: 429,
} as const;

// 缓存配置
export const CACHE_CONFIG = {
  DEFAULT_TTL: 5 * 60 * 1000, // 5分钟
  USER_INFO_TTL: 30 * 60 * 1000, // 30分钟
  PRODUCT_LIST_TTL: 10 * 60 * 1000, // 10分钟
  CART_TTL: 60 * 60 * 1000, // 1小时
} as const;

// 外卖相关常量（为未来扩展准备）
export const DELIVERY_CONFIG = {
  DEFAULT_RADIUS: 5000, // 5公里
  MAX_RADIUS: 20000, // 20公里
  DEFAULT_DELIVERY_FEE: '5.00',
  FREE_DELIVERY_THRESHOLD: '30.00',
} as const;

// 平台检测
export const PLATFORM = {
  WEB: 'web',
  IOS: 'ios',
  ANDROID: 'android',
} as const;

// 设备类型
export const DEVICE_TYPE = {
  MOBILE: 'mobile',
  TABLET: 'tablet',
  DESKTOP: 'desktop',
} as const;

// 语言配置
export const LOCALES = {
  ZH_CN: 'zh-CN',
  EN_US: 'en-US',
} as const;

// 时间格式
export const DATE_FORMATS = {
  DATE: 'YYYY-MM-DD',
  DATETIME: 'YYYY-MM-DD HH:mm:ss',
  TIME: 'HH:mm:ss',
  MONTH: 'YYYY-MM',
  YEAR: 'YYYY',
} as const;

// 正则表达式
export const REGEX = {
  PHONE: /^1[3-9]\d{9}$/,
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  ID_CARD: /^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$/,
  POSTAL_CODE: /^\d{6}$/,
  URL: /^https?:\/\/.+/,
  NUMBER: /^\d+$/,
  DECIMAL: /^\d+(\.\d{1,2})?$/,
} as const;
