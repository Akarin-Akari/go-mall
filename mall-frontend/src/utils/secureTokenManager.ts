/**
 * 简化的Token管理器
 * 使用localStorage进行基本的token存储
 */

const TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';

/**
 * 简化Token管理器类
 */
export class SecureTokenManager {
  private static instance: SecureTokenManager;

  private constructor() {}

  /**
   * 获取单例实例
   */
  public static getInstance(): SecureTokenManager {
    if (!SecureTokenManager.instance) {
      SecureTokenManager.instance = new SecureTokenManager();
    }
    return SecureTokenManager.instance;
  }

  /**
   * 设置访问令牌
   */
  public setAccessToken(token: string): void {
    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem(TOKEN_KEY, token);
      } catch (error) {
        console.error('设置访问令牌失败:', error);
      }
    }
  }

  /**
   * 获取访问令牌
   */
  public getAccessToken(): string | null {
    if (typeof window !== 'undefined') {
      try {
        return localStorage.getItem(TOKEN_KEY);
      } catch (error) {
        console.error('获取访问令牌失败:', error);
        return null;
      }
    }
    return null;
  }

  /**
   * 设置刷新令牌
   */
  public setRefreshToken(token: string): void {
    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem(REFRESH_TOKEN_KEY, token);
      } catch (error) {
        console.error('设置刷新令牌失败:', error);
      }
    }
  }

  /**
   * 获取刷新令牌
   */
  public getRefreshToken(): string | null {
    if (typeof window !== 'undefined') {
      try {
        return localStorage.getItem(REFRESH_TOKEN_KEY);
      } catch (error) {
        console.error('获取刷新令牌失败:', error);
        return null;
      }
    }
    return null;
  }

  /**
   * 清除所有令牌
   */
  public clearTokens(): void {
    if (typeof window !== 'undefined') {
      try {
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(REFRESH_TOKEN_KEY);
      } catch (error) {
        console.error('清除令牌失败:', error);
      }
    }
  }
}

// 创建全局实例
export const secureTokenManager = SecureTokenManager.getInstance();
