import { STORAGE_KEYS, DATE_FORMATS, REGEX } from '@/constants';
import dayjs from 'dayjs';
import Cookies from 'js-cookie';

// 本地存储工具
export const storage = {
  get: (key: string): string | null => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(key);
  },
  
  set: (key: string, value: string): void => {
    if (typeof window === 'undefined') return;
    localStorage.setItem(key, value);
  },
  
  remove: (key: string): void => {
    if (typeof window === 'undefined') return;
    localStorage.removeItem(key);
  },
  
  clear: (): void => {
    if (typeof window === 'undefined') return;
    localStorage.clear();
  },
  
  // JSON数据存储
  getJSON: <T>(key: string): T | null => {
    const value = storage.get(key);
    if (!value) return null;
    try {
      return JSON.parse(value) as T;
    } catch {
      return null;
    }
  },
  
  setJSON: <T>(key: string, value: T): void => {
    storage.set(key, JSON.stringify(value));
  },
};

// Cookie工具
export const cookie = {
  get: (key: string): string | undefined => {
    return Cookies.get(key);
  },
  
  set: (key: string, value: string, options?: Cookies.CookieAttributes): void => {
    Cookies.set(key, value, options);
  },
  
  remove: (key: string): void => {
    Cookies.remove(key);
  },
};

// Token管理
export const tokenManager = {
  getToken: (): string | null => {
    return storage.get(STORAGE_KEYS.TOKEN) || cookie.get(STORAGE_KEYS.TOKEN) || null;
  },
  
  setToken: (token: string, remember = false): void => {
    storage.set(STORAGE_KEYS.TOKEN, token);
    if (remember) {
      cookie.set(STORAGE_KEYS.TOKEN, token, { expires: 7 }); // 7天
    }
  },
  
  removeToken: (): void => {
    storage.remove(STORAGE_KEYS.TOKEN);
    cookie.remove(STORAGE_KEYS.TOKEN);
  },
  
  getRefreshToken: (): string | null => {
    return storage.get(STORAGE_KEYS.REFRESH_TOKEN);
  },
  
  setRefreshToken: (token: string): void => {
    storage.set(STORAGE_KEYS.REFRESH_TOKEN, token);
  },
  
  removeRefreshToken: (): void => {
    storage.remove(STORAGE_KEYS.REFRESH_TOKEN);
  },
  
  clearAll: (): void => {
    tokenManager.removeToken();
    tokenManager.removeRefreshToken();
  },
};

// 格式化工具
export const formatter = {
  // 价格格式化
  price: (price: string | number, currency = '¥'): string => {
    const num = typeof price === 'string' ? parseFloat(price) : price;
    if (isNaN(num)) return `${currency}0.00`;
    return `${currency}${num.toFixed(2)}`;
  },
  
  // 数字格式化（千分位）
  number: (num: number): string => {
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  },
  
  // 日期格式化
  date: (date: string | Date, format = DATE_FORMATS.DATETIME): string => {
    return dayjs(date).format(format);
  },
  
  // 相对时间
  relativeTime: (date: string | Date): string => {
    return dayjs(date).fromNow();
  },
  
  // 文件大小格式化
  fileSize: (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  },
  
  // 手机号脱敏
  phone: (phone: string): string => {
    if (!phone || phone.length !== 11) return phone;
    return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2');
  },
  
  // 邮箱脱敏
  email: (email: string): string => {
    if (!email) return email;
    const [username, domain] = email.split('@');
    if (!username || !domain) return email;
    const maskedUsername = username.length > 2 
      ? username.substring(0, 2) + '*'.repeat(username.length - 2)
      : username;
    return `${maskedUsername}@${domain}`;
  },
};

// 验证工具
export const validator = {
  // 邮箱验证
  email: (email: string): boolean => {
    return REGEX.EMAIL.test(email);
  },
  
  // 手机号验证
  phone: (phone: string): boolean => {
    return REGEX.PHONE.test(phone);
  },
  
  // 身份证验证
  idCard: (idCard: string): boolean => {
    return REGEX.ID_CARD.test(idCard);
  },
  
  // 密码强度验证
  password: (password: string): { valid: boolean; strength: 'weak' | 'medium' | 'strong' } => {
    if (password.length < 6) {
      return { valid: false, strength: 'weak' };
    }
    
    let score = 0;
    if (/[a-z]/.test(password)) score++;
    if (/[A-Z]/.test(password)) score++;
    if (/\d/.test(password)) score++;
    if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) score++;
    
    if (score < 2) return { valid: false, strength: 'weak' };
    if (score < 3) return { valid: true, strength: 'medium' };
    return { valid: true, strength: 'strong' };
  },
  
  // URL验证
  url: (url: string): boolean => {
    return REGEX.URL.test(url);
  },
  
  // 数字验证
  number: (value: string): boolean => {
    return REGEX.NUMBER.test(value);
  },
  
  // 小数验证
  decimal: (value: string): boolean => {
    return REGEX.DECIMAL.test(value);
  },
};

// URL工具
export const urlUtils = {
  // 获取查询参数
  getQuery: (key: string): string | null => {
    if (typeof window === 'undefined') return null;
    const params = new URLSearchParams(window.location.search);
    return params.get(key);
  },
  
  // 设置查询参数
  setQuery: (params: Record<string, string>): void => {
    if (typeof window === 'undefined') return;
    const url = new URL(window.location.href);
    Object.entries(params).forEach(([key, value]) => {
      url.searchParams.set(key, value);
    });
    window.history.replaceState({}, '', url.toString());
  },
  
  // 构建URL
  buildUrl: (base: string, params: Record<string, any>): string => {
    const url = new URL(base);
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        url.searchParams.set(key, String(value));
      }
    });
    return url.toString();
  },
};

// 防抖函数
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
};

// 节流函数
export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean;
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), wait);
    }
  };
};

// 深拷贝
export const deepClone = <T>(obj: T): T => {
  if (obj === null || typeof obj !== 'object') return obj;
  if (obj instanceof Date) return new Date(obj.getTime()) as unknown as T;
  if (obj instanceof Array) return obj.map(item => deepClone(item)) as unknown as T;
  if (typeof obj === 'object') {
    const clonedObj = {} as T;
    Object.keys(obj).forEach(key => {
      (clonedObj as any)[key] = deepClone((obj as any)[key]);
    });
    return clonedObj;
  }
  return obj;
};

// 数组去重
export const unique = <T>(arr: T[], key?: keyof T): T[] => {
  if (!key) {
    return [...new Set(arr)];
  }
  const seen = new Set();
  return arr.filter(item => {
    const value = item[key];
    if (seen.has(value)) {
      return false;
    }
    seen.add(value);
    return true;
  });
};

// 随机字符串生成
export const randomString = (length = 8): string => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
};

// 设备检测
export const device = {
  isMobile: (): boolean => {
    if (typeof window === 'undefined') return false;
    return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
      navigator.userAgent
    );
  },
  
  isIOS: (): boolean => {
    if (typeof window === 'undefined') return false;
    return /iPad|iPhone|iPod/.test(navigator.userAgent);
  },
  
  isAndroid: (): boolean => {
    if (typeof window === 'undefined') return false;
    return /Android/.test(navigator.userAgent);
  },
  
  isWechat: (): boolean => {
    if (typeof window === 'undefined') return false;
    return /MicroMessenger/i.test(navigator.userAgent);
  },
};

// 颜色工具
export const colorUtils = {
  // 十六进制转RGB
  hexToRgb: (hex: string): { r: number; g: number; b: number } | null => {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result
      ? {
          r: parseInt(result[1], 16),
          g: parseInt(result[2], 16),
          b: parseInt(result[3], 16),
        }
      : null;
  },
  
  // RGB转十六进制
  rgbToHex: (r: number, g: number, b: number): string => {
    return '#' + [r, g, b].map(x => {
      const hex = x.toString(16);
      return hex.length === 1 ? '0' + hex : hex;
    }).join('');
  },
};

// 错误处理工具
export const errorHandler = {
  // 网络错误处理
  handleNetworkError: (error: any): string => {
    if (!error.response) {
      return '网络连接失败，请检查网络设置';
    }
    
    const { status } = error.response;
    switch (status) {
      case 400:
        return '请求参数错误';
      case 401:
        return '登录已过期，请重新登录';
      case 403:
        return '没有权限访问该资源';
      case 404:
        return '请求的资源不存在';
      case 500:
        return '服务器内部错误';
      case 502:
        return '网关错误';
      case 503:
        return '服务暂时不可用';
      default:
        return '请求失败，请稍后重试';
    }
  },
  
  // 业务错误处理
  handleBusinessError: (code: number, message: string): string => {
    // 可以根据业务错误码进行特殊处理
    return message || '操作失败';
  },
};

// 环境检测
export const env = {
  isDev: process.env.NODE_ENV === 'development',
  isProd: process.env.NODE_ENV === 'production',
  isTest: process.env.NODE_ENV === 'test',
  isClient: typeof window !== 'undefined',
  isServer: typeof window === 'undefined',
};

// 商品相关工具函数
export const formatPrice = (price: string | number, currency = ''): string => {
  const num = typeof price === 'string' ? parseFloat(price) : price;
  if (isNaN(num)) return '0.00';
  return num.toFixed(2);
};

export const formatNumber = (num: number): string => {
  if (num < 1000) return num.toString();
  if (num < 10000) return (num / 1000).toFixed(1) + 'k';
  if (num < 100000) return (num / 10000).toFixed(1) + 'w';
  return (num / 10000).toFixed(0) + 'w';
};

// 导出所有工具
export * from './request';
export * from './auth';
export * from './upload';
