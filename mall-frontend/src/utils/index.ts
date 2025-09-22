export const formatPrice = (price: number | string): string => {
  const numPrice = typeof price === 'string' ? parseFloat(price) : price;
  return numPrice.toFixed(2);
};

export const formatDate = (date: string | Date): string => {
  const d = new Date(date);
  return d.toLocaleDateString('zh-CN');
};

export const formatNumber = (num: number, decimals: number = 2): string => {
  return num.toFixed(decimals);
};

export const formatter = {
  date: (date: string | Date): string => {
    const d = new Date(date);
    return d.toLocaleDateString('zh-CN');
  },
};

// 错误处理已移至 @/utils/simpleErrorHandler
// 保留简单的工具函数用于向后兼容
export const errorHandler = {
  log: (error: Error, context?: string): void => {
    console.error('Error:', error);
    if (context) console.error('Context:', context);
  },
  notify: (message: string): void => {
    console.warn('Notification:', message);
  },
  handleBusinessError: (code: number, message: string): string => {
    const errorMap: Record<number, string> = {
      400: '请求参数错误',
      401: '未授权访问',
      403: '权限不足',
      404: '资源不存在',
      500: '服务器内部错误',
    };
    return errorMap[code] || message || '未知错误';
  },
};

export const debounce = (func: any, wait: number) => {
  let timeout: any;
  return (...args: any[]) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
};

export { formatDateTime } from './formatDateTime';
