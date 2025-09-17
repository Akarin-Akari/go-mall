/**
 * 简化的应用程序引导文件
 * 专注于前端应用初始化，移除复杂的依赖注入系统
 */

import { errorHandler } from '@/utils/simpleErrorHandler';
import { CONFIG, isFeatureEnabled } from '@/config/app';

// 应用初始化状态
interface AppInitState {
  isInitialized: boolean;
  initStartTime: number;
  initEndTime?: number;
  errors: string[];
}

// 全局应用状态
let appState: AppInitState = {
  isInitialized: false,
  initStartTime: 0,
  errors: [],
};

/**
 * 初始化应用程序
 */
export async function initializeApp(): Promise<void> {
  try {
    appState.initStartTime = Date.now();
    console.log('🚀 开始初始化 Mall Frontend 应用...');

    // 1. 初始化错误处理
    await initializeErrorHandling();

    // 2. 初始化性能监控（如果启用）
    if (isFeatureEnabled('ENABLE_PERFORMANCE_MONITORING')) {
      await initializePerformanceMonitoring();
    }

    // 3. 初始化分析工具（如果启用）
    if (isFeatureEnabled('ENABLE_ANALYTICS')) {
      await initializeAnalytics();
    }

    // 4. 初始化用户会话
    await initializeUserSession();

    // 5. 初始化应用状态
    await initializeAppState();

    appState.isInitialized = true;
    appState.initEndTime = Date.now();
    
    const initTime = appState.initEndTime - appState.initStartTime;
    console.log(`✅ Mall Frontend 应用初始化完成 (耗时: ${initTime}ms)`);

  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    appState.errors.push(errorMessage);
    
    errorHandler.handleError(error as Error, {
      phase: 'app_initialization',
      config: CONFIG.APP,
    });
    
    console.error('❌ 应用初始化失败:', error);
    throw error;
  }
}

/**
 * 初始化错误处理
 */
async function initializeErrorHandling(): Promise<void> {
  try {
    // 设置全局错误处理
    if (typeof window !== 'undefined') {
      window.addEventListener('error', (event) => {
        errorHandler.handleError(event.error, {
          type: 'global_error',
          filename: event.filename,
          lineno: event.lineno,
          colno: event.colno,
        });
      });

      window.addEventListener('unhandledrejection', (event) => {
        errorHandler.handleError(new Error(event.reason), {
          type: 'unhandled_promise_rejection',
        });
      });
    }

    console.log('✓ 错误处理初始化完成');
  } catch (error) {
    console.error('错误处理初始化失败:', error);
    throw error;
  }
}

/**
 * 初始化性能监控
 */
async function initializePerformanceMonitoring(): Promise<void> {
  try {
    if (typeof window !== 'undefined' && 'performance' in window) {
      // 监控页面加载性能
      window.addEventListener('load', () => {
        const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        if (perfData) {
          console.log('📊 页面性能数据:', {
            domContentLoaded: perfData.domContentLoadedEventEnd - perfData.domContentLoadedEventStart,
            loadComplete: perfData.loadEventEnd - perfData.loadEventStart,
            totalTime: perfData.loadEventEnd - perfData.fetchStart,
          });
        }
      });
    }

    console.log('✓ 性能监控初始化完成');
  } catch (error) {
    console.error('性能监控初始化失败:', error);
    // 性能监控失败不应该阻止应用启动
  }
}

/**
 * 初始化分析工具
 */
async function initializeAnalytics(): Promise<void> {
  try {
    // 这里可以初始化Google Analytics、百度统计等
    console.log('✓ 分析工具初始化完成');
  } catch (error) {
    console.error('分析工具初始化失败:', error);
    // 分析工具失败不应该阻止应用启动
  }
}

/**
 * 初始化用户会话
 */
async function initializeUserSession(): Promise<void> {
  try {
    // 检查本地存储的用户信息
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem(CONFIG.SECURITY.TOKEN_STORAGE_KEY);
      if (token) {
        console.log('✓ 发现已保存的用户会话');
      }
    }

    console.log('✓ 用户会话初始化完成');
  } catch (error) {
    console.error('用户会话初始化失败:', error);
    // 用户会话初始化失败不应该阻止应用启动
  }
}

/**
 * 初始化应用状态
 */
async function initializeAppState(): Promise<void> {
  try {
    // 初始化Redux store或其他状态管理
    console.log('✓ 应用状态初始化完成');
  } catch (error) {
    console.error('应用状态初始化失败:', error);
    throw error;
  }
}

/**
 * 获取应用初始化状态
 */
export function getAppInitState(): AppInitState {
  return { ...appState };
}

/**
 * 检查应用是否已初始化
 */
export function isAppInitialized(): boolean {
  return appState.isInitialized;
}

/**
 * 重置应用状态（主要用于测试）
 */
export function resetAppState(): void {
  appState = {
    isInitialized: false,
    initStartTime: 0,
    errors: [],
  };
}

/**
 * 应用健康检查
 */
export function healthCheck(): {
  status: 'healthy' | 'unhealthy';
  details: Record<string, any>;
} {
  const details = {
    initialized: appState.isInitialized,
    initTime: appState.initEndTime ? appState.initEndTime - appState.initStartTime : null,
    errors: appState.errors,
    config: {
      apiBaseUrl: CONFIG.API.BASE_URL,
      environment: process.env.NODE_ENV,
      features: CONFIG.FEATURES,
    },
  };

  const status = appState.isInitialized && appState.errors.length === 0 ? 'healthy' : 'unhealthy';

  return { status, details };
}

// 导出默认初始化函数
export default initializeApp;
