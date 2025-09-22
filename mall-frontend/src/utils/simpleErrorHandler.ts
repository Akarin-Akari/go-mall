/**
 * 简化的前端错误处理器
 * 只负责基本的错误日志记录
 */

// 错误级别枚举
export enum ErrorLevel {
  INFO = 'info',
  WARN = 'warn',
  ERROR = 'error',
}

// 错误类型枚举
export enum ErrorType {
  NETWORK = 'network',
  BUSINESS = 'business',
  SYSTEM = 'system',
  UNKNOWN = 'unknown',
}

// 简化的错误接口
export interface SimpleError {
  message: string;
  type: ErrorType;
  level: ErrorLevel;
  timestamp: number;
  context?: Record<string, any>;
}

// 简化的错误处理器类
export class SimpleErrorHandler {
  private static instance: SimpleErrorHandler;

  private constructor() {}

  public static getInstance(): SimpleErrorHandler {
    if (!SimpleErrorHandler.instance) {
      SimpleErrorHandler.instance = new SimpleErrorHandler();
    }
    return SimpleErrorHandler.instance;
  }

  /**
   * 处理错误 - 简化版本
   */
  public handleError(
    error: Error | string,
    options?: {
      type?: ErrorType;
      level?: ErrorLevel;
      context?: Record<string, any>;
    }
  ): SimpleError {
    const message = typeof error === 'string' ? error : error.message;
    const simpleError: SimpleError = {
      message,
      type: options?.type || this.determineErrorType(message),
      level: options?.level || ErrorLevel.ERROR,
      timestamp: Date.now(),
      context: options?.context,
    };

    // 简单的控制台日志记录
    this.logError(simpleError);

    return simpleError;
  }

  /**
   * 记录错误日志 - 简化版本
   */
  private logError(error: SimpleError): void {
    const logMethod =
      error.level === ErrorLevel.ERROR
        ? console.error
        : error.level === ErrorLevel.WARN
          ? console.warn
          : console.info;

    logMethod('[Frontend Error]', {
      message: error.message,
      type: error.type,
      level: error.level,
      context: error.context,
      timestamp: new Date(error.timestamp).toISOString(),
    });
  }

  /**
   * 确定错误类型 - 简化版本
   */
  private determineErrorType(message: string): ErrorType {
    if (message.includes('network') || message.includes('fetch')) {
      return ErrorType.NETWORK;
    }
    if (message.includes('business') || message.includes('validation')) {
      return ErrorType.BUSINESS;
    }
    return ErrorType.UNKNOWN;
  }
}

// 创建全局实例
export const errorHandler = SimpleErrorHandler.getInstance();

// 业务错误处理函数
export const handleBusinessError = (code: number, message: string): string => {
  const errorMap: Record<number, string> = {
    400: '请求参数错误',
    401: '未授权访问',
    403: '权限不足',
    404: '资源不存在',
    500: '服务器内部错误',
  };

  return errorMap[code] || message || '未知错误';
};

// 导出便捷方法
export const logError = (
  error: Error | string,
  context?: Record<string, any>
) => {
  errorHandler.handleError(error, { context });
};
