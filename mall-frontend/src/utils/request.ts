import axios, {
  AxiosInstance,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
  AxiosResponse,
  AxiosError,
} from 'axios';
import { message } from 'antd';
import {
  errorHandler,
  ErrorType,
  ErrorLevel,
  handleBusinessError,
} from './simpleErrorHandler';
import { ApiResponse } from '@/types';
import { HTTP_STATUS, BUSINESS_CODE, ROUTES } from '@/constants';
import { secureTokenManager } from './secureTokenManager';

// 安全Token获取 (避免循环依赖)
const getToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  try {
    return secureTokenManager.getAccessToken();
  } catch {
    return null;
  }
};

// Token刷新状态管理
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value: any) => void;
  reject: (reason: any) => void;
}> = [];

// 处理队列中的请求
const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error);
    } else {
      resolve(token);
    }
  });

  failedQueue = [];
};

// 刷新Token
const refreshToken = async (): Promise<string | null> => {
  try {
    const refreshTokenValue = secureTokenManager.getRefreshToken();
    if (!refreshTokenValue) {
      throw new Error('No refresh token available');
    }

    // 调用刷新Token API
    const response = await axios.post('/api/v1/auth/refresh', {
      refresh_token: refreshTokenValue,
    });

    const { token, refresh_token } = response.data.data;

    // 保存新的Token
    secureTokenManager.setAccessToken(token);
    if (refresh_token) {
      secureTokenManager.setRefreshToken(refresh_token);
    }

    return token;
  } catch (error) {
    // 刷新失败，清除所有Token
    secureTokenManager.clearTokens();
    throw error;
  }
};

// 请求配置接口
interface RequestConfig extends InternalAxiosRequestConfig {
  skipAuth?: boolean;
  skipErrorHandler?: boolean;
  showLoading?: boolean;
  showSuccessMessage?: boolean;
  successMessage?: string;
}

// 创建axios实例
const createAxiosInstance = (): AxiosInstance => {
  const instance = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8081',
    timeout: parseInt(process.env.NEXT_PUBLIC_API_TIMEOUT || '10000'),
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // 请求拦截器
  instance.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      const customConfig = config as RequestConfig;
      // 添加认证token
      if (!customConfig.skipAuth) {
        const token = getToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
      }

      // 添加请求时间戳（防止缓存）
      if (config.method === 'get') {
        config.params = {
          ...config.params,
          _t: Date.now(),
        };
      }

      // 显示加载状态
      if (customConfig.showLoading) {
        // 可以在这里显示全局loading
        console.log('Loading...');
      }

      return config;
    },
    (error: AxiosError) => {
      return Promise.reject(error);
    }
  );

  // 响应拦截器
  instance.interceptors.response.use(
    (response: AxiosResponse<ApiResponse>) => {
      const { config } = response;
      const customConfig = config as RequestConfig;

      // 隐藏加载状态
      if (customConfig.showLoading) {
        console.log('Loading finished');
      }

      // 检查业务状态码
      const { code, message: msg, data } = response.data;

      if (code === BUSINESS_CODE.SUCCESS) {
        // 显示成功消息
        if (customConfig.showSuccessMessage) {
          message.success(customConfig.successMessage || msg || '操作成功');
        }
        return response;
      } else {
        // 业务错误处理
        const errorMessage = handleBusinessError(code, msg);

        // 使用统一错误处理
        errorHandler.handleError(new Error(errorMessage), {
          type: ErrorType.BUSINESS,
          level: ErrorLevel.WARN,
          context: {
            businessCode: code,
            url: response.config?.url,
            method: response.config?.method,
          },
        });

        if (!customConfig.skipErrorHandler) {
          message.error(errorMessage);
        }

        // 特殊错误码处理
        if (code === BUSINESS_CODE.UNAUTHORIZED) {
          return handleTokenExpired(error.config);
        }

        return Promise.reject(new Error(errorMessage));
      }
    },
    async (error: AxiosError) => {
      const { config, response } = error;
      const customConfig = config as RequestConfig;

      // 隐藏加载状态
      if (customConfig?.showLoading) {
        console.log('Loading finished');
      }

      // 网络错误处理 - 生成错误消息
      const errorMessage = getNetworkErrorMessage(error);

      // 使用统一错误处理
      errorHandler.handleError(error, {
        type: ErrorType.NETWORK,
        level: ErrorLevel.ERROR,
        context: {
          status: response?.status,
          url: error.config?.url,
          method: error.config?.method,
        },
      });

      if (!customConfig?.skipErrorHandler) {
        message.error(errorMessage);
      }

      // 特殊状态码处理
      if (response?.status === HTTP_STATUS.UNAUTHORIZED) {
        return handleTokenExpired(error.config);
      }

      return Promise.reject(error);
    }
  );

  return instance;
};

// 获取网络错误消息
const getNetworkErrorMessage = (error: AxiosError): string => {
  if (error.code === 'ECONNABORTED') {
    return '请求超时，请稍后重试';
  }
  if (error.message === 'Network Error') {
    return '网络连接失败，请检查网络设置';
  }
  if (error.response?.status) {
    const status = error.response.status;
    const statusMessages: Record<number, string> = {
      400: '请求参数错误',
      401: '未授权访问',
      403: '权限不足',
      404: '请求的资源不存在',
      500: '服务器内部错误',
      502: '网关错误',
      503: '服务不可用',
    };
    return statusMessages[status] || `请求失败 (${status})`;
  }
  return error.message || '请求失败';
};

// 处理未授权错误
const handleUnauthorized = () => {
  // 清理token
  if (typeof window !== 'undefined') {
    try {
      secureTokenManager.clearTokens();
    } catch {
      // 忽略清理错误
    }
  }

  // 跳转到登录页面
  if (typeof window !== 'undefined') {
    const currentPath = window.location.pathname;
    if (currentPath !== '/login') {
      window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`;
    }
  }
};

// 处理Token过期
const handleTokenExpired = async (originalRequest: any): Promise<any> => {
  if (originalRequest._retry) {
    // 已经重试过，直接跳转到登录页
    handleUnauthorized();
    return Promise.reject(new Error('Token refresh failed'));
  }

  if (isRefreshing) {
    // 正在刷新Token，将请求加入队列
    return new Promise((resolve, reject) => {
      failedQueue.push({ resolve, reject });
    }).then(token => {
      originalRequest.headers['Authorization'] = `Bearer ${token}`;
      return axios(originalRequest);
    }).catch(err => {
      return Promise.reject(err);
    });
  }

  originalRequest._retry = true;
  isRefreshing = true;

  try {
    const newToken = await refreshToken();
    processQueue(null, newToken);

    // 重新发送原始请求
    originalRequest.headers['Authorization'] = `Bearer ${newToken}`;
    return axios(originalRequest);
  } catch (error) {
    processQueue(error, null);
    handleUnauthorized();
    return Promise.reject(error);
  } finally {
    isRefreshing = false;
  }
};

// 创建请求实例
const request = createAxiosInstance();

// 请求方法封装
export const http = {
  // GET请求
  get: <T = unknown>(
    url: string,
    config?: RequestConfig
  ): Promise<ApiResponse<T>> => {
    return request.get(url, config).then(res => res.data);
  },

  // POST请求
  post: <T = unknown>(
    url: string,
    data?: unknown,
    config?: RequestConfig
  ): Promise<ApiResponse<T>> => {
    return request.post(url, data, config).then(res => res.data);
  },

  // PUT请求
  put: <T = unknown>(
    url: string,
    data?: unknown,
    config?: RequestConfig
  ): Promise<ApiResponse<T>> => {
    return request.put(url, data, config).then(res => res.data);
  },

  // DELETE请求
  delete: <T = unknown>(
    url: string,
    config?: RequestConfig
  ): Promise<ApiResponse<T>> => {
    return request.delete(url, config).then(res => res.data);
  },

  // PATCH请求
  patch: <T = unknown>(
    url: string,
    data?: unknown,
    config?: RequestConfig
  ): Promise<ApiResponse<T>> => {
    return request.patch(url, data, config).then(res => res.data);
  },

  // 文件上传
  upload: <T = unknown>(
    url: string,
    formData: FormData,
    config?: RequestConfig
  ): Promise<ApiResponse<T>> => {
    return request
      .post(url, formData, {
        ...config,
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
      .then(res => res.data);
  },

  // 下载文件
  download: (
    url: string,
    filename?: string,
    config?: RequestConfig
  ): Promise<void> => {
    return request
      .get(url, {
        ...config,
        responseType: 'blob',
      })
      .then(response => {
        const blob = new Blob([response.data]);
        const downloadUrl = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = filename || 'download';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(downloadUrl);
      });
  },
};

// 请求取消token管理
export class RequestCancelManager {
  private cancelTokens: Map<string, () => void> = new Map();

  // 创建取消token
  createCancelToken(key: string): { cancelToken: any; cancel: () => void } {
    // 如果已存在相同key的请求，先取消它
    this.cancel(key);

    const source = axios.CancelToken.source();
    this.cancelTokens.set(key, source.cancel);

    return {
      cancelToken: source.token,
      cancel: () => this.cancel(key),
    };
  }

  // 取消指定请求
  cancel(key: string): void {
    const cancel = this.cancelTokens.get(key);
    if (cancel) {
      cancel();
      this.cancelTokens.delete(key);
    }
  }

  // 取消所有请求
  cancelAll(): void {
    this.cancelTokens.forEach(cancel => cancel());
    this.cancelTokens.clear();
  }
}

// 创建全局请求取消管理器
export const requestCancelManager = new RequestCancelManager();

// 重试机制
export const retryRequest = async <T>(
  requestFn: () => Promise<T>,
  maxRetries = 3,
  delay = 1000
): Promise<T> => {
  let lastError: Error;

  for (let i = 0; i <= maxRetries; i++) {
    try {
      return await requestFn();
    } catch (error) {
      lastError = error as Error;

      if (i === maxRetries) {
        throw lastError;
      }

      // 等待指定时间后重试
      await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i)));
    }
  }

  throw lastError!;
};

// 并发请求控制
export const concurrentRequest = async <T>(
  requests: (() => Promise<T>)[],
  limit = 5
): Promise<T[]> => {
  const results: T[] = [];
  const executing: Promise<void>[] = [];

  for (const request of requests) {
    const promise = request().then(result => {
      results.push(result);
    });

    executing.push(promise);

    if (executing.length >= limit) {
      await Promise.race(executing);
      executing.splice(
        executing.findIndex(p => p === promise),
        1
      );
    }
  }

  await Promise.all(executing);
  return results;
};

// 请求缓存管理
class RequestCache {
  private cache: Map<string, { data: any; timestamp: number; ttl: number }> =
    new Map();

  // 生成缓存key
  private generateKey(url: string, params?: any): string {
    return `${url}_${JSON.stringify(params || {})}`;
  }

  // 获取缓存
  get(url: string, params?: any): any | null {
    const key = this.generateKey(url, params);
    const cached = this.cache.get(key);

    if (!cached) return null;

    // 检查是否过期
    if (Date.now() - cached.timestamp > cached.ttl) {
      this.cache.delete(key);
      return null;
    }

    return cached.data;
  }

  // 设置缓存
  set(url: string, data: any, ttl = 5 * 60 * 1000, params?: any): void {
    const key = this.generateKey(url, params);
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
  }

  // 清除缓存
  clear(url?: string, params?: any): void {
    if (url) {
      const key = this.generateKey(url, params);
      this.cache.delete(key);
    } else {
      this.cache.clear();
    }
  }

  // 清除过期缓存
  clearExpired(): void {
    const now = Date.now();
    for (const [key, cached] of this.cache.entries()) {
      if (now - cached.timestamp > cached.ttl) {
        this.cache.delete(key);
      }
    }
  }
}

// 创建全局请求缓存管理器
export const requestCache = new RequestCache();

// 带缓存的GET请求
export const cachedGet = async <T = any>(
  url: string,
  params?: any,
  ttl = 5 * 60 * 1000,
  config?: RequestConfig
): Promise<ApiResponse<T>> => {
  // 尝试从缓存获取
  const cached = requestCache.get(url, params);
  if (cached) {
    return cached;
  }

  // 发起请求
  const response = await http.get<T>(url, { ...config, params });

  // 缓存响应
  requestCache.set(url, response, ttl, params);

  return response;
};

export default request;
