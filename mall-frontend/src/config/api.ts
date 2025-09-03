// API配置文件
export const API_CONFIG = {
  // 基础配置
  BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
  TIMEOUT: parseInt(process.env.NEXT_PUBLIC_API_TIMEOUT || '10000'),
  
  // 重试配置
  RETRY_COUNT: 3,
  RETRY_DELAY: 1000,
  
  // 缓存配置
  CACHE_TTL: 5 * 60 * 1000, // 5分钟
  
  // 认证配置
  TOKEN_HEADER: 'Authorization',
  TOKEN_PREFIX: 'Bearer ',
  
  // 环境配置
  IS_DEVELOPMENT: process.env.NODE_ENV === 'development',
  IS_PRODUCTION: process.env.NODE_ENV === 'production',
  
  // 调试配置
  ENABLE_REQUEST_LOG: process.env.NODE_ENV === 'development',
  ENABLE_RESPONSE_LOG: process.env.NODE_ENV === 'development',
  ENABLE_ERROR_LOG: true,
};

// API端点配置
export const API_ENDPOINTS = {
  // 认证相关
  AUTH: {
    LOGIN: '/api/v1/auth/login',
    REGISTER: '/api/v1/auth/register',
    LOGOUT: '/api/v1/auth/logout',
    REFRESH: '/api/v1/auth/refresh',
    PROFILE: '/api/v1/auth/profile',
    VERIFY_EMAIL: '/api/v1/auth/verify-email',
    RESET_PASSWORD: '/api/v1/auth/reset-password',
  },
  
  // 用户相关
  USER: {
    PROFILE: '/api/v1/users/profile',
    UPDATE_PROFILE: '/api/v1/users/profile',
    CHANGE_PASSWORD: '/api/v1/users/change-password',
    UPLOAD_AVATAR: '/api/v1/users/avatar',
    ADDRESSES: '/api/v1/users/addresses',
    FAVORITES: '/api/v1/users/favorites',
  },
  
  // 商品相关
  PRODUCT: {
    LIST: '/api/v1/products',
    DETAIL: (id: number) => `/api/v1/products/${id}`,
    SEARCH: '/api/v1/products/search',
    CATEGORIES: '/api/v1/products/categories',
    BRANDS: '/api/v1/products/brands',
    REVIEWS: (id: number) => `/api/v1/products/${id}/reviews`,
    RELATED: (id: number) => `/api/v1/products/${id}/related`,
  },
  
  // 购物车相关
  CART: {
    LIST: '/api/v1/cart',
    ADD: '/api/v1/cart/items',
    UPDATE: (id: number) => `/api/v1/cart/items/${id}`,
    REMOVE: (id: number) => `/api/v1/cart/items/${id}`,
    CLEAR: '/api/v1/cart/clear',
    SYNC: '/api/v1/cart/sync',
    COUNT: '/api/v1/cart/count',
  },
  
  // 订单相关
  ORDER: {
    LIST: '/api/v1/orders',
    DETAIL: (id: number) => `/api/v1/orders/${id}`,
    CREATE: '/api/v1/orders',
    CANCEL: (id: number) => `/api/v1/orders/${id}/cancel`,
    CONFIRM: (id: number) => `/api/v1/orders/${id}/confirm`,
    REFUND: (id: number) => `/api/v1/orders/${id}/refund`,
  },
  
  // 支付相关
  PAYMENT: {
    CREATE: '/api/v1/payments',
    QUERY: (id: string) => `/api/v1/payments/${id}`,
    CALLBACK: '/api/v1/payments/callback',
    REFUND: '/api/v1/payments/refund',
    METHODS: '/api/v1/payments/methods',
  },
  
  // 优惠券相关
  COUPON: {
    LIST: '/api/v1/coupons',
    AVAILABLE: '/api/v1/coupons/available',
    USE: (id: number) => `/api/v1/coupons/${id}/use`,
    MY_COUPONS: '/api/v1/users/coupons',
  },
  
  // 地址相关
  ADDRESS: {
    LIST: '/api/v1/addresses',
    CREATE: '/api/v1/addresses',
    UPDATE: (id: number) => `/api/v1/addresses/${id}`,
    DELETE: (id: number) => `/api/v1/addresses/${id}`,
    SET_DEFAULT: (id: number) => `/api/v1/addresses/${id}/default`,
    REGIONS: '/api/v1/addresses/regions',
  },
  
  // 文件上传
  UPLOAD: {
    IMAGE: '/api/v1/upload/image',
    FILE: '/api/v1/upload/file',
    AVATAR: '/api/v1/upload/avatar',
  },
  
  // 系统相关
  SYSTEM: {
    CONFIG: '/api/v1/system/config',
    BANNERS: '/api/v1/system/banners',
    NOTICES: '/api/v1/system/notices',
    FEEDBACK: '/api/v1/system/feedback',
  },
};

// HTTP状态码
export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  METHOD_NOT_ALLOWED: 405,
  CONFLICT: 409,
  UNPROCESSABLE_ENTITY: 422,
  TOO_MANY_REQUESTS: 429,
  INTERNAL_SERVER_ERROR: 500,
  BAD_GATEWAY: 502,
  SERVICE_UNAVAILABLE: 503,
  GATEWAY_TIMEOUT: 504,
} as const;

// 业务错误码
export const BUSINESS_CODE = {
  SUCCESS: 0,
  UNKNOWN_ERROR: 1000,
  INVALID_PARAMS: 1001,
  UNAUTHORIZED: 1002,
  FORBIDDEN: 1003,
  NOT_FOUND: 1004,
  CONFLICT: 1005,
  RATE_LIMITED: 1006,
  
  // 用户相关错误
  USER_NOT_FOUND: 2001,
  USER_ALREADY_EXISTS: 2002,
  INVALID_CREDENTIALS: 2003,
  EMAIL_NOT_VERIFIED: 2004,
  PASSWORD_TOO_WEAK: 2005,
  
  // 商品相关错误
  PRODUCT_NOT_FOUND: 3001,
  PRODUCT_OUT_OF_STOCK: 3002,
  PRODUCT_UNAVAILABLE: 3003,
  INVALID_PRODUCT_SPEC: 3004,
  
  // 购物车相关错误
  CART_ITEM_NOT_FOUND: 4001,
  CART_ITEM_OUT_OF_STOCK: 4002,
  CART_EMPTY: 4003,
  CART_ITEM_LIMIT_EXCEEDED: 4004,
  
  // 订单相关错误
  ORDER_NOT_FOUND: 5001,
  ORDER_CANNOT_CANCEL: 5002,
  ORDER_ALREADY_PAID: 5003,
  ORDER_EXPIRED: 5004,
  INSUFFICIENT_STOCK: 5005,
  
  // 支付相关错误
  PAYMENT_FAILED: 6001,
  PAYMENT_CANCELLED: 6002,
  PAYMENT_EXPIRED: 6003,
  INVALID_PAYMENT_METHOD: 6004,
  
  // 优惠券相关错误
  COUPON_NOT_FOUND: 7001,
  COUPON_EXPIRED: 7002,
  COUPON_USED: 7003,
  COUPON_NOT_APPLICABLE: 7004,
  COUPON_LIMIT_EXCEEDED: 7005,
} as const;

// 错误消息映射
export const ERROR_MESSAGES = {
  [BUSINESS_CODE.UNKNOWN_ERROR]: '未知错误',
  [BUSINESS_CODE.INVALID_PARAMS]: '参数错误',
  [BUSINESS_CODE.UNAUTHORIZED]: '未授权访问',
  [BUSINESS_CODE.FORBIDDEN]: '禁止访问',
  [BUSINESS_CODE.NOT_FOUND]: '资源不存在',
  [BUSINESS_CODE.CONFLICT]: '资源冲突',
  [BUSINESS_CODE.RATE_LIMITED]: '请求过于频繁',
  
  // 用户相关错误
  [BUSINESS_CODE.USER_NOT_FOUND]: '用户不存在',
  [BUSINESS_CODE.USER_ALREADY_EXISTS]: '用户已存在',
  [BUSINESS_CODE.INVALID_CREDENTIALS]: '用户名或密码错误',
  [BUSINESS_CODE.EMAIL_NOT_VERIFIED]: '邮箱未验证',
  [BUSINESS_CODE.PASSWORD_TOO_WEAK]: '密码强度不够',
  
  // 商品相关错误
  [BUSINESS_CODE.PRODUCT_NOT_FOUND]: '商品不存在',
  [BUSINESS_CODE.PRODUCT_OUT_OF_STOCK]: '商品库存不足',
  [BUSINESS_CODE.PRODUCT_UNAVAILABLE]: '商品暂时不可用',
  [BUSINESS_CODE.INVALID_PRODUCT_SPEC]: '商品规格无效',
  
  // 购物车相关错误
  [BUSINESS_CODE.CART_ITEM_NOT_FOUND]: '购物车商品不存在',
  [BUSINESS_CODE.CART_ITEM_OUT_OF_STOCK]: '购物车商品库存不足',
  [BUSINESS_CODE.CART_EMPTY]: '购物车为空',
  [BUSINESS_CODE.CART_ITEM_LIMIT_EXCEEDED]: '购物车商品数量超限',
  
  // 订单相关错误
  [BUSINESS_CODE.ORDER_NOT_FOUND]: '订单不存在',
  [BUSINESS_CODE.ORDER_CANNOT_CANCEL]: '订单无法取消',
  [BUSINESS_CODE.ORDER_ALREADY_PAID]: '订单已支付',
  [BUSINESS_CODE.ORDER_EXPIRED]: '订单已过期',
  [BUSINESS_CODE.INSUFFICIENT_STOCK]: '库存不足',
  
  // 支付相关错误
  [BUSINESS_CODE.PAYMENT_FAILED]: '支付失败',
  [BUSINESS_CODE.PAYMENT_CANCELLED]: '支付已取消',
  [BUSINESS_CODE.PAYMENT_EXPIRED]: '支付已过期',
  [BUSINESS_CODE.INVALID_PAYMENT_METHOD]: '支付方式无效',
  
  // 优惠券相关错误
  [BUSINESS_CODE.COUPON_NOT_FOUND]: '优惠券不存在',
  [BUSINESS_CODE.COUPON_EXPIRED]: '优惠券已过期',
  [BUSINESS_CODE.COUPON_USED]: '优惠券已使用',
  [BUSINESS_CODE.COUPON_NOT_APPLICABLE]: '优惠券不适用',
  [BUSINESS_CODE.COUPON_LIMIT_EXCEEDED]: '优惠券使用次数超限',
} as const;

// 请求方法类型
export type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

// API响应类型
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  timestamp: number;
  request_id?: string;
}

// 分页响应类型
export interface PageResponse<T = any> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// 请求配置类型
export interface RequestOptions {
  skipAuth?: boolean;
  skipErrorHandler?: boolean;
  showLoading?: boolean;
  showSuccessMessage?: boolean;
  successMessage?: string;
  retryCount?: number;
  timeout?: number;
  cache?: boolean;
  cacheTTL?: number;
}

// 环境检测工具
export const isProduction = () => API_CONFIG.IS_PRODUCTION;
export const isDevelopment = () => API_CONFIG.IS_DEVELOPMENT;

// 获取完整的API URL
export const getApiUrl = (endpoint: string): string => {
  return `${API_CONFIG.BASE_URL}${endpoint}`;
};

// 获取错误消息
export const getErrorMessage = (code: number, defaultMessage = '操作失败'): string => {
  return ERROR_MESSAGES[code as keyof typeof ERROR_MESSAGES] || defaultMessage;
};
