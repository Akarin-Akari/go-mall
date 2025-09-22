export interface AppConfig {
  apiBaseUrl: string;
  apiTimeout: number;
  maxUploadSize: number;
  supportedImageTypes: string[];
  features: {
    enableCache: boolean;
    enableCompression: boolean;
    enablePerformanceMonitoring: boolean;
  };
}

export class ConfigManager {
  private static instance: ConfigManager;
  private config: AppConfig;

  private constructor() {
    this.config = {
      apiBaseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8081',
      apiTimeout: Number(process.env.NEXT_PUBLIC_API_TIMEOUT) || 30000,
      maxUploadSize: 5 * 1024 * 1024, // 5MB
      supportedImageTypes: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
      features: {
        enableCache: true,
        enableCompression: true,
        enablePerformanceMonitoring: true
      }
    };
  }

  static getInstance(): ConfigManager {
    if (!ConfigManager.instance) {
      ConfigManager.instance = new ConfigManager();
    }
    return ConfigManager.instance;
  }

  getConfig(): AppConfig {
    return this.config;
  }

  get(key: keyof AppConfig): any {
    return this.config[key];
  }

  set(key: keyof AppConfig, value: any): void {
    this.config[key] = value;
  }

  getApiBaseUrl(): string {
    return this.config.apiBaseUrl;
  }

  getApiTimeout(): number {
    return this.config.apiTimeout;
  }
}

export const configManager = ConfigManager.getInstance();

export default configManager;