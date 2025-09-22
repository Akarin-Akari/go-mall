export interface ResourceMetrics {
  memory: {
    used: number;
    limit: number;
  };
  cpu: {
    usage: number;
  };
  network: {
    requests: number;
    bandwidth: number;
  };
}

export class ResourceManager {
  private static instance: ResourceManager;
  private metrics: ResourceMetrics;
  private observers: Set<(metrics: ResourceMetrics) => void> = new Set();

  private constructor() {
    this.metrics = {
      memory: { used: 0, limit: 0 },
      cpu: { usage: 0 },
      network: { requests: 0, bandwidth: 0 }
    };
    this.startMonitoring();
  }

  static getInstance(): ResourceManager {
    if (!ResourceManager.instance) {
      ResourceManager.instance = new ResourceManager();
    }
    return ResourceManager.instance;
  }

  private startMonitoring(): void {
    // 模拟资源监控
    if (typeof window !== 'undefined' && 'performance' in window) {
      setInterval(() => {
        this.updateMetrics();
      }, 5000);
    }
  }

  private updateMetrics(): void {
    if (typeof window !== 'undefined') {
      // 更新内存使用情况
      if ('memory' in performance) {
        const memory = (performance as any).memory;
        this.metrics.memory.used = memory.usedJSHeapSize;
        this.metrics.memory.limit = memory.jsHeapSizeLimit;
      }

      // 更新网络请求计数
      const entries = performance.getEntriesByType('resource');
      this.metrics.network.requests = entries.length;

      // 通知观察者
      this.notifyObservers();
    }
  }

  getMetrics(): ResourceMetrics {
    return { ...this.metrics };
  }

  addObserver(callback: (metrics: ResourceMetrics) => void): void {
    this.observers.add(callback);
  }

  removeObserver(callback: (metrics: ResourceMetrics) => void): void {
    this.observers.delete(callback);
  }

  private notifyObservers(): void {
    this.observers.forEach(callback => callback(this.metrics));
  }

  checkMemoryUsage(): boolean {
    if (this.metrics.memory.limit > 0) {
      const usage = this.metrics.memory.used / this.metrics.memory.limit;
      return usage < 0.9; // 返回true表示内存使用正常
    }
    return true;
  }

  cleanup(): void {
    // 清理资源
    this.observers.clear();
  }
}

export const resourceManager = ResourceManager.getInstance();

export default resourceManager;