/**
 * 简化的接口定义入口文件
 * 只保留前端必需的基础类型定义
 */

// 基础类型定义
export type ServiceIdentifier = string | symbol;

// 简化的应用配置类型
export type ApplicationConfig = {
  apiBaseUrl?: string;
  timeout?: number;
  enableLogging?: boolean;
};

// 默认导出空对象，保持兼容性
export default {};
