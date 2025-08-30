import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { message } from 'antd';
import { tokenManager, errorHandler } from './index';
import { ApiResponse } from '@/types';
import { HTTP_STATUS, BUSINESS_CODE } from '@/constants';

// 请求配置接口
interface RequestConfig extends AxiosRequestConfig {
  skipAuth?: boolean;
  skipErrorHandler?: boolean;
  showLoading?: boolean;
  showSuccessMessage?: boolean;
  successMessage?: string;
}

// 创建axios实例
const createAxiosInstance = (): AxiosInstance => {
  const instance = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
    timeout: parseInt(process.env.NEXT_PUBLIC_API_TIMEOUT || '10000'),
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // 请求拦截器
  instance.interceptors.request.use(
    (config: any) => {
      // 添加认证token
      if (!config.skipAuth) {
        const token = tokenManager.getToken();
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
      if (config.showLoading) {
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
        const errorMessage = errorHandler.handleBusinessError(code, msg);
        
        if (!customConfig.skipErrorHandler) {
          message.error(errorMessage);
        }
        
        // 特殊错误码处理
        if (code === BUSINESS_CODE.UNAUTHORIZED) {
          handleUnauthorized();
        }
        
        return Promise.reject(new Error(errorMessage));
      }
    },
    (error: AxiosError) => {
      const { config, response } = error;
      const customConfig = config as RequestConfig;

      // 隐藏加载状态
      if (customConfig?.showLoading) {
        console.log('Loading finished');
      }

      // 网络错误处理
      if (!customConfig?.skipErrorHandler) {
        const errorMessage = errorHandler.handleNetworkError(error);
        message.error(errorMessage);
      }

      // 特殊状态码处理
      if (response?.status === HTTP_STATUS.UNAUTHORIZED) {
        handleUnauthorized();
      }

      return Promise.reject(error);
    }
  );

  return instance;
};

// 处理未授权错误
const handleUnauthorized = () => {
  tokenManager.clearAll();
  
  // 跳转到登录页面
  if (typeof window !== 'undefined') {
    const currentPath = window.location.pathname;
    if (currentPath !== '/login') {
      window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`;
    }
  }
};

// 创建请求实例
const request = createAxiosInstance();

// 请求方法封装
export const http = {
  // GET请求
  get: <T = any>(url: string, config?: RequestConfig): Promise<ApiResponse<T>> => {
    return request.get(url, config).then(res => res.data);
  },

  // POST请求
  post: <T = any>(url: string, data?: any, config?: RequestConfig): Promise<ApiResponse<T>> => {
    return request.post(url, data, config).then(res => res.data);
  },

  // PUT请求
  put: <T = any>(url: string, data?: any, config?: RequestConfig): Promise<ApiResponse<T>> => {
    return request.put(url, data, config).then(res => res.data);
  },

  // DELETE请求
  delete: <T = any>(url: string, config?: RequestConfig): Promise<ApiResponse<T>> => {
    return request.delete(url, config).then(res => res.data);
  },

  // PATCH请求
  patch: <T = any>(url: string, data?: any, config?: RequestConfig): Promise<ApiResponse<T>> => {
    return request.patch(url, data, config).then(res => res.data);
  },

  // 文件上传
  upload: <T = any>(url: string, formData: FormData, config?: RequestConfig): Promise<ApiResponse<T>> => {
    return request.post(url, formData, {
      ...config,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }).then(res => res.data);
  },

  // 下载文件
  download: (url: string, filename?: string, config?: RequestConfig): Promise<void> => {
    return request.get(url, {
      ...config,
      responseType: 'blob',
    }).then(response => {
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
      cancel('Request canceled');
      this.cancelTokens.delete(key);
    }
  }

  // 取消所有请求
  cancelAll(): void {
    this.cancelTokens.forEach(cancel => cancel('All requests canceled'));
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
      executing.splice(executing.findIndex(p => p === promise), 1);
    }
  }

  await Promise.all(executing);
  return results;
};

// 请求缓存管理
class RequestCache {
  private cache: Map<string, { data: any; timestamp: number; ttl: number }> = new Map();

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
