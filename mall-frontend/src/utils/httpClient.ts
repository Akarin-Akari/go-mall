import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { message } from 'antd';
import { API_CONFIG, HTTP_STATUS, BUSINESS_CODE, getErrorMessage } from '@/config/api';
import { tokenManager } from './index';
import type { ApiResponse, RequestOptions } from '@/config/api';

// 请求队列管理
class RequestQueue {
  private queue: Map<string, Promise<any>> = new Map();

  // 生成请求key
  private generateKey(config: AxiosRequestConfig): string {
    const { method, url, params, data } = config;
    return `${method}:${url}:${JSON.stringify(params)}:${JSON.stringify(data)}`;
  }

  // 添加请求到队列
  add<T>(config: AxiosRequestConfig, request: () => Promise<T>): Promise<T> {
    const key = this.generateKey(config);
    
    if (this.queue.has(key)) {
      return this.queue.get(key);
    }

    const promise = request().finally(() => {
      this.queue.delete(key);
    });

    this.queue.set(key, promise);
    return promise;
  }

  // 清空队列
  clear(): void {
    this.queue.clear();
  }
}

// 重试机制
class RetryManager {
  private retryCount: number;
  private retryDelay: number;

  constructor(retryCount = API_CONFIG.RETRY_COUNT, retryDelay = API_CONFIG.RETRY_DELAY) {
    this.retryCount = retryCount;
    this.retryDelay = retryDelay;
  }

  // 判断是否应该重试
  shouldRetry(error: AxiosError, attempt: number): boolean {
    if (attempt >= this.retryCount) return false;

    // 网络错误或5xx错误才重试
    if (!error.response) return true;
    
    const status = error.response.status;
    return status >= 500 || status === HTTP_STATUS.TOO_MANY_REQUESTS;
  }

  // 计算重试延迟（指数退避）
  getRetryDelay(attempt: number): number {
    return this.retryDelay * Math.pow(2, attempt);
  }

  // 执行重试
  async retry<T>(
    request: () => Promise<T>,
    attempt = 0
  ): Promise<T> {
    try {
      return await request();
    } catch (error) {
      if (this.shouldRetry(error as AxiosError, attempt)) {
        const delay = this.getRetryDelay(attempt);
        await new Promise(resolve => setTimeout(resolve, delay));
        return this.retry(request, attempt + 1);
      }
      throw error;
    }
  }
}

// HTTP客户端类
class HttpClient {
  private instance: AxiosInstance;
  private requestQueue: RequestQueue;
  private retryManager: RetryManager;

  constructor() {
    this.requestQueue = new RequestQueue();
    this.retryManager = new RetryManager();
    this.instance = this.createInstance();
    this.setupInterceptors();
  }

  // 创建axios实例
  private createInstance(): AxiosInstance {
    return axios.create({
      baseURL: API_CONFIG.BASE_URL,
      timeout: API_CONFIG.TIMEOUT,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  // 设置拦截器
  private setupInterceptors(): void {
    // 请求拦截器
    this.instance.interceptors.request.use(
      (config: any) => {
        // 添加认证token
        if (!config.skipAuth) {
          const token = tokenManager.getToken();
          if (token) {
            config.headers[API_CONFIG.TOKEN_HEADER] = `${API_CONFIG.TOKEN_PREFIX}${token}`;
          }
        }

        // 添加请求ID
        config.headers['X-Request-ID'] = this.generateRequestId();

        // 添加时间戳防缓存
        if (config.method === 'get' && !config.cache) {
          config.params = {
            ...config.params,
            _t: Date.now(),
          };
        }

        // 请求日志
        if (API_CONFIG.ENABLE_REQUEST_LOG) {
          console.log('🚀 Request:', {
            method: config.method?.toUpperCase(),
            url: config.url,
            params: config.params,
            data: config.data,
          });
        }

        return config;
      },
      (error) => {
        if (API_CONFIG.ENABLE_ERROR_LOG) {
          console.error('❌ Request Error:', error);
        }
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.instance.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        // 响应日志
        if (API_CONFIG.ENABLE_RESPONSE_LOG) {
          console.log('✅ Response:', {
            status: response.status,
            data: response.data,
          });
        }

        // 检查业务状态码
        const { code, message: msg, data } = response.data;
        
        if (code !== BUSINESS_CODE.SUCCESS) {
          const errorMessage = getErrorMessage(code, msg);
          
          // 特殊错误处理
          if (code === BUSINESS_CODE.UNAUTHORIZED) {
            this.handleUnauthorized();
          }
          
          throw new Error(errorMessage);
        }

        return response;
      },
      async (error: AxiosError) => {
        // 错误日志
        if (API_CONFIG.ENABLE_ERROR_LOG) {
          console.error('❌ Response Error:', {
            status: error.response?.status,
            message: error.message,
            data: error.response?.data,
          });
        }

        // 处理网络错误
        if (!error.response) {
          throw new Error('网络连接失败，请检查网络设置');
        }

        // 处理HTTP状态码错误
        const status = error.response.status;
        switch (status) {
          case HTTP_STATUS.UNAUTHORIZED:
            this.handleUnauthorized();
            throw new Error('登录已过期，请重新登录');
          case HTTP_STATUS.FORBIDDEN:
            throw new Error('没有权限访问该资源');
          case HTTP_STATUS.NOT_FOUND:
            throw new Error('请求的资源不存在');
          case HTTP_STATUS.TOO_MANY_REQUESTS:
            throw new Error('请求过于频繁，请稍后重试');
          case HTTP_STATUS.INTERNAL_SERVER_ERROR:
            throw new Error('服务器内部错误');
          case HTTP_STATUS.BAD_GATEWAY:
            throw new Error('网关错误');
          case HTTP_STATUS.SERVICE_UNAVAILABLE:
            throw new Error('服务暂时不可用');
          default:
            throw new Error('请求失败，请稍后重试');
        }
      }
    );
  }

  // 生成请求ID
  private generateRequestId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // 处理未授权错误
  private handleUnauthorized(): void {
    tokenManager.clearAll();
    
    // 跳转到登录页面
    if (typeof window !== 'undefined') {
      window.location.href = '/login';
    }
  }

  // GET请求
  async get<T = any>(
    url: string,
    params?: any,
    options: RequestOptions = {}
  ): Promise<ApiResponse<T>> {
    const config: AxiosRequestConfig = {
      method: 'GET',
      url,
      params,
      ...options,
    };

    return this.requestQueue.add(config, () =>
      this.retryManager.retry(() =>
        this.instance.request(config).then(res => res.data)
      )
    );
  }

  // POST请求
  async post<T = any>(
    url: string,
    data?: any,
    options: RequestOptions = {}
  ): Promise<ApiResponse<T>> {
    const config: AxiosRequestConfig = {
      method: 'POST',
      url,
      data,
      ...options,
    };

    const response = await this.retryManager.retry(() =>
      this.instance.request(config).then(res => res.data)
    );

    // 显示成功消息
    if (options.showSuccessMessage) {
      message.success(options.successMessage || '操作成功');
    }

    return response;
  }

  // PUT请求
  async put<T = any>(
    url: string,
    data?: any,
    options: RequestOptions = {}
  ): Promise<ApiResponse<T>> {
    const config: AxiosRequestConfig = {
      method: 'PUT',
      url,
      data,
      ...options,
    };

    const response = await this.retryManager.retry(() =>
      this.instance.request(config).then(res => res.data)
    );

    if (options.showSuccessMessage) {
      message.success(options.successMessage || '更新成功');
    }

    return response;
  }

  // DELETE请求
  async delete<T = any>(
    url: string,
    options: RequestOptions = {}
  ): Promise<ApiResponse<T>> {
    const config: AxiosRequestConfig = {
      method: 'DELETE',
      url,
      ...options,
    };

    const response = await this.retryManager.retry(() =>
      this.instance.request(config).then(res => res.data)
    );

    if (options.showSuccessMessage) {
      message.success(options.successMessage || '删除成功');
    }

    return response;
  }

  // PATCH请求
  async patch<T = any>(
    url: string,
    data?: any,
    options: RequestOptions = {}
  ): Promise<ApiResponse<T>> {
    const config: AxiosRequestConfig = {
      method: 'PATCH',
      url,
      data,
      ...options,
    };

    const response = await this.retryManager.retry(() =>
      this.instance.request(config).then(res => res.data)
    );

    if (options.showSuccessMessage) {
      message.success(options.successMessage || '更新成功');
    }

    return response;
  }

  // 文件上传
  async upload<T = any>(
    url: string,
    formData: FormData,
    options: RequestOptions = {}
  ): Promise<ApiResponse<T>> {
    const config: AxiosRequestConfig = {
      method: 'POST',
      url,
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      ...options,
    };

    const response = await this.instance.request(config).then(res => res.data);

    if (options.showSuccessMessage) {
      message.success(options.successMessage || '上传成功');
    }

    return response;
  }

  // 取消所有请求
  cancelAllRequests(): void {
    this.requestQueue.clear();
  }

  // 获取axios实例（用于特殊需求）
  getInstance(): AxiosInstance {
    return this.instance;
  }
}

// 创建全局HTTP客户端实例
export const httpClient = new HttpClient();

// 导出便捷方法
export const { get, post, put, delete: del, patch, upload } = httpClient;

export default httpClient;
