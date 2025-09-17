/**
 * 简化的前端错误处理器
 * 专注于前端错误日志记录和用户提示
 */

// 错误级别枚举
export enum ErrorLevel {
  DEBUG = 'debug',
  INFO = 'info',
  WARN = 'warn',
  ERROR = 'error',
  FATAL = 'fatal',
}

// 错误类型枚举
export enum ErrorType {
  NETWORK = 'network',
  VALIDATION = 'validation',
  BUSINESS = 'business',
  SYSTEM = 'system',
  SECURITY = 'security',
  PERFORMANCE = 'performance',
}

// 标准错误接口
export interface StandardError {
  id: string;
  message: string;
  type: ErrorType;
  level: ErrorLevel;
  timestamp: number;
  context?: Record<string, any>;
  stack?: string;
  userAgent?: string;
  url?: string;
}

// 简化的错误处理器配置
interface ErrorHandlerConfig {
  enableConsoleLogging: boolean;
  enableRemoteLogging: boolean;
  maxErrorsPerSession: number;
}

// 简化的错误处理器类
export class SimpleErrorHandler {
  private static instance: SimpleErrorHandler;
  private config: ErrorHandlerConfig;
  private errorCount = 0;

  private constructor() {
    this.config = {
      enableConsoleLogging: true,
      enableRemoteLogging: false,
      maxErrorsPerSession: 100,
    };
  }

  public static getInstance(): SimpleErrorHandler {
    if (!SimpleErrorHandler.instance) {
      SimpleErrorHandler.instance = new SimpleErrorHandler();
    }
    return SimpleErrorHandler.instance;
  }

  /**
   * 处理错误
   */
  public handleError(
    error: Error | string,
    context?: Record<string, any>
  ): void {
    if (this.errorCount >= this.config.maxErrorsPerSession) {
      return; // 防止错误过多
    }

    const standardError = this.createStandardError(error, context);
    this.logError(standardError);
    this.errorCount++;
  }

  /**
   * 创建标准错误对象
   */
  private createStandardError(
    error: Error | string,
    context?: Record<string, any>
  ): StandardError {
    const message = typeof error === 'string' ? error : error.message;
    const stack = typeof error === 'object' ? error.stack : undefined;

    return {
      id: this.generateErrorId(),
      message,
      type: this.determineErrorType(message),
      level: ErrorLevel.ERROR,
      timestamp: Date.now(),
      context,
      stack,
      userAgent: typeof window !== 'undefined' ? window.navigator.userAgent : undefined,
      url: typeof window !== 'undefined' ? window.location.href : undefined,
    };
  }

  /**
   * 记录错误日志
   */
  private logError(error: StandardError): void {
    if (this.config.enableConsoleLogging) {
      console.error('[Frontend Error]', {
        id: error.id,
        message: error.message,
        type: error.type,
        level: error.level,
        context: error.context,
        timestamp: new Date(error.timestamp).toISOString(),
      });

      if (error.stack) {
        console.error('Stack trace:', error.stack);
      }
    }

    // 如果需要远程日志记录，可以在这里发送到后端
    if (this.config.enableRemoteLogging) {
      this.sendToRemote(error);
    }
  }

  /**
   * 发送错误到远程服务器
   */
  private async sendToRemote(error: StandardError): Promise<void> {
    try {
      // 这里可以调用后端API记录错误
      await fetch('/api/v1/logs/error', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(error),
      });
    } catch (e) {
      // 静默处理远程日志失败
      console.warn('Failed to send error to remote:', e);
    }
  }

  /**
   * 确定错误类型
   */
  private determineErrorType(message: string): ErrorType {
    if (message.includes('network') || message.includes('fetch')) {
      return ErrorType.NETWORK;
    }
    if (message.includes('validation') || message.includes('invalid')) {
      return ErrorType.VALIDATION;
    }
    if (message.includes('permission') || message.includes('unauthorized')) {
      return ErrorType.SECURITY;
    }
    return ErrorType.SYSTEM;
  }

  /**
   * 生成错误ID
   */
  private generateErrorId(): string {
    return `fe_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * 处理业务错误
   */
  public handleBusinessError(code: number, message: string): string {
    const errorMap: Record<number, string> = {
      400: '请求参数错误',
      401: '未授权访问',
      403: '权限不足',
      404: '资源不存在',
      409: '资源冲突',
      422: '数据验证失败',
      429: '请求过于频繁',
      500: '服务器内部错误',
      502: '网关错误',
      503: '服务不可用',
    };

    return errorMap[code] || message || '未知错误';
  }

  /**
   * 重置错误计数
   */
  public resetErrorCount(): void {
    this.errorCount = 0;
  }

  /**
   * 获取错误统计
   */
  public getErrorStats(): { totalErrors: number } {
    return {
      totalErrors: this.errorCount,
    };
  }
}

// 创建全局实例
export const errorHandler = SimpleErrorHandler.getInstance();

// 导出便捷方法
export const logError = (error: Error | string, context?: Record<string, any>) => {
  errorHandler.handleError(error, context);
};

export const handleBusinessError = (code: number, message: string) => {
  return errorHandler.handleBusinessError(code, message);
};
