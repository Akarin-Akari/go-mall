import { message } from 'antd';

export interface ErrorInfo {
  message: string;
  code?: string | number;
  details?: any;
}

export class ErrorHandler {
  private static instance: ErrorHandler;
  private errorQueue: ErrorInfo[] = [];

  private constructor() {}

  static getInstance(): ErrorHandler {
    if (!ErrorHandler.instance) {
      ErrorHandler.instance = new ErrorHandler();
    }
    return ErrorHandler.instance;
  }

  handleError(error: any): void {
    const errorInfo: ErrorInfo = {
      message: error?.message || '未知错误',
      code: error?.code,
      details: error
    };

    this.errorQueue.push(errorInfo);
    
    // 显示错误提示
    message.error(errorInfo.message);

    // 控制台输出
    console.error('Error handled:', errorInfo);
  }

  getErrors(): ErrorInfo[] {
    return [...this.errorQueue];
  }

  clearErrors(): void {
    this.errorQueue = [];
  }
}

export const errorHandler = ErrorHandler.getInstance();

export default errorHandler;