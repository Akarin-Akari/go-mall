/**
 * 安全和性能组件集成测试
 * 测试各个组件之间的协作和整体系统功能
 */

import { ErrorHandler } from '../../utils/errorHandler';
import { ConfigManager } from '../../utils/configManager';
import { ResourceManager } from '../../utils/resourceManager';
import { ImageOptimizer } from '../../utils/imageOptimizer';

// Mock DOM APIs
const mockElement = {
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
};

const mockObserver = {
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
};

global.MutationObserver = jest.fn().mockImplementation(() => mockObserver);
global.IntersectionObserver = jest.fn().mockImplementation(() => mockObserver);

// Mock localStorage
const mockLocalStorage = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};

Object.defineProperty(window, 'localStorage', {
  value: mockLocalStorage,
});

// Mock fetch
global.fetch = jest.fn();

describe('Security and Performance Integration Tests', () => {
  let errorHandler: ErrorHandler;
  let configManager: ConfigManager;
  let resourceManager: ResourceManager;
  let imageOptimizer: ImageOptimizer;

  beforeEach(() => {
    // Reset all singleton instances
    (ErrorHandler as any).instance = null;
    (ConfigManager as any).instance = null;
    (ResourceManager as any).instance = null;
    (ImageOptimizer as any).instance = null;

    // Get fresh instances
    errorHandler = ErrorHandler.getInstance();
    configManager = ConfigManager.getInstance();
    resourceManager = ResourceManager.getInstance();
    imageOptimizer = ImageOptimizer.getInstance();

    // Clear mocks
    jest.clearAllMocks();
  });

  afterEach(async () => {
    // Clean up all instances
    await errorHandler.destroy();
    await configManager.destroy();
    await resourceManager.destroy();
    await imageOptimizer.destroy();
  });

  describe('System Initialization', () => {
    it('should initialize all components successfully', async () => {
      await Promise.all([
        errorHandler.initialize(),
        configManager.initialize(),
        resourceManager.initialize(),
        imageOptimizer.initialize(),
      ]);

      expect(errorHandler.getStatus()).toBe('initialized');
      expect(configManager.getStatus()).toBe('initialized');
      expect(resourceManager.getStatus()).toBe('initialized');
      expect(imageOptimizer.getStatus()).toBe('initialized');
    });

    it('should handle initialization failures gracefully', async () => {
      // Mock one component to fail initialization
      const originalInitialize = configManager.initialize;
      configManager.initialize = jest
        .fn()
        .mockRejectedValue(new Error('Init failed'));

      const results = await Promise.allSettled([
        errorHandler.initialize(),
        configManager.initialize(),
        resourceManager.initialize(),
        imageOptimizer.initialize(),
      ]);

      // Should have one failure and three successes
      const failures = results.filter(r => r.status === 'rejected');
      const successes = results.filter(r => r.status === 'fulfilled');

      expect(failures).toHaveLength(1);
      expect(successes).toHaveLength(3);

      // Restore original method
      configManager.initialize = originalInitialize;
    });
  });

  describe('Error Handling Integration', () => {
    it('should handle errors from all components', async () => {
      await errorHandler.initialize();

      const errorListener = jest.fn();
      errorHandler.addListener(errorListener);

      // Simulate errors from different components
      const configError = new Error('Config validation failed');
      const resourceError = new Error('Resource cleanup failed');
      const imageError = new Error('Image load failed');

      errorHandler.handleError(configError, {
        context: { component: 'ConfigManager' },
      });
      errorHandler.handleError(resourceError, {
        context: { component: 'ResourceManager' },
      });
      errorHandler.handleError(imageError, {
        context: { component: 'ImageOptimizer' },
      });

      expect(errorListener).toHaveBeenCalledTimes(3);

      const stats = errorHandler.getErrorStats();
      expect(stats.totalErrors).toBe(3);
    });

    it('should integrate with configuration for error reporting', async () => {
      await Promise.all([
        errorHandler.initialize(),
        configManager.initialize(),
      ]);

      // Configure error reporting
      configManager.registerGroup('errorReporting');
      configManager.registerConfig('errorReporting', {
        enabled: { value: true, type: 'boolean' },
        endpoint: { value: 'https://api.example.com/errors', type: 'string' },
      });

      await errorHandler.updateConfig({
        enableRemoteReporting: configManager.get('errorReporting', 'enabled'),
        remoteEndpoint: configManager.get('errorReporting', 'endpoint'),
      });

      const mockFetch = fetch as jest.MockedFunction<typeof fetch>;
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      } as Response);

      errorHandler.handleError(new Error('Test error'));

      // Wait for async error reporting
      await new Promise(resolve => setTimeout(resolve, 100));

      expect(mockFetch).toHaveBeenCalledWith(
        'https://api.example.com/errors',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });
  });

  describe('Resource Management Integration', () => {
    it('should manage resources across all components', async () => {
      await Promise.all([
        resourceManager.initialize(),
        imageOptimizer.initialize(),
      ]);

      // Create a resource group for image optimization
      const groupId = resourceManager.createGroup('image-optimization', true);

      // Register some resources
      const callback = jest.fn();
      const timerId = resourceManager.registerTimer(callback, 1000);
      const listenerId = resourceManager.registerEventListener(
        mockElement as any,
        'click',
        callback
      );

      resourceManager.addToGroup(groupId, timerId);
      resourceManager.addToGroup(groupId, listenerId);

      let stats = resourceManager.getResourceStats();
      expect(stats.totalResources).toBe(2);

      // Cleanup group should remove all resources
      resourceManager.cleanupGroup(groupId);

      stats = resourceManager.getResourceStats();
      expect(stats.totalResources).toBe(0);
    });

    it('should handle resource cleanup on component destruction', async () => {
      await resourceManager.initialize();

      const callback = jest.fn();
      resourceManager.registerTimer(callback, 1000);
      resourceManager.registerInterval(callback, 500);

      let stats = resourceManager.getResourceStats();
      expect(stats.totalResources).toBe(2);

      // Destroy should cleanup all resources
      await resourceManager.destroy();

      stats = resourceManager.getResourceStats();
      expect(stats.totalResources).toBe(0);
    });
  });

  describe('Configuration Management Integration', () => {
    it('should share configuration across components', async () => {
      await configManager.initialize();

      // Register shared configuration
      configManager.registerGroup('performance');
      configManager.registerConfig('performance', {
        imageQuality: { value: 85, type: 'number' },
        cacheSize: { value: 100, type: 'number' },
        enableOptimization: { value: true, type: 'boolean' },
      });

      // Components should be able to access shared config
      const imageQuality = configManager.get('performance', 'imageQuality');
      const cacheSize = configManager.get('performance', 'cacheSize');
      const enableOptimization = configManager.get(
        'performance',
        'enableOptimization'
      );

      expect(imageQuality).toBe(85);
      expect(cacheSize).toBe(100);
      expect(enableOptimization).toBe(true);
    });

    it('should notify components of configuration changes', async () => {
      await configManager.initialize();

      configManager.registerGroup('test');
      configManager.registerConfig('test', {
        setting: { value: 'initial', type: 'string' },
      });

      const listener = jest.fn();
      configManager.addListener('setting', listener);

      // Change configuration
      configManager.set('test', 'setting', 'updated');

      expect(listener).toHaveBeenCalledWith('setting', 'updated', 'initial');
    });
  });

  describe('Image Optimization Integration', () => {
    it('should integrate with resource management for cleanup', async () => {
      await Promise.all([
        resourceManager.initialize(),
        imageOptimizer.initialize(),
      ]);

      // Image optimizer should register its observers with resource manager
      const mockImg = document.createElement('img');
      imageOptimizer.observe(mockImg);

      // Resource manager should track the observer
      const stats = resourceManager.getResourceStats();
      expect(stats.totalResources).toBeGreaterThan(0);
    });

    it('should integrate with error handling for load failures', async () => {
      await Promise.all([
        errorHandler.initialize(),
        imageOptimizer.initialize(),
      ]);

      const errorListener = jest.fn();
      errorHandler.addListener(errorListener);

      // Mock image load failure
      const mockImage = {
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        onerror: null as any,
      };

      global.Image = jest.fn().mockImplementation(() => mockImage);

      const loadPromise = imageOptimizer.loadImage('/images/error.jpg');

      // Simulate error
      setTimeout(() => {
        if (mockImage.onerror) {
          mockImage.onerror(new Error('Load failed'));
        }
      }, 0);

      await expect(loadPromise).rejects.toThrow();

      // Error should be recorded
      expect(errorListener).toHaveBeenCalled();
    });
  });

  describe('Performance Monitoring Integration', () => {
    it('should collect performance metrics from all components', async () => {
      await Promise.all([
        errorHandler.initialize(),
        configManager.initialize(),
        resourceManager.initialize(),
        imageOptimizer.initialize(),
      ]);

      // Simulate some activity
      errorHandler.handleError(new Error('Test error'));
      configManager.set('test', 'value', 'new');
      resourceManager.registerTimer(() => {}, 1000);
      imageOptimizer.recordLoadSuccess('/images/test.jpg', 150);

      // Collect metrics
      const errorStats = errorHandler.getErrorStats();
      const resourceStats = resourceManager.getResourceStats();
      const imageStats = imageOptimizer.getOptimizationStats();

      expect(errorStats.totalErrors).toBe(1);
      expect(resourceStats.totalResources).toBeGreaterThan(0);
      expect(imageStats.totalImages).toBe(1);
    });

    it('should handle health checks across all components', async () => {
      await Promise.all([
        errorHandler.initialize(),
        configManager.initialize(),
        resourceManager.initialize(),
        imageOptimizer.initialize(),
      ]);

      const healthChecks = await Promise.all([
        errorHandler.healthCheck(),
        configManager.healthCheck(),
        resourceManager.healthCheck(),
        imageOptimizer.healthCheck(),
      ]);

      expect(healthChecks.every(healthy => healthy === true)).toBe(true);
    });
  });

  describe('System Shutdown', () => {
    it('should gracefully shutdown all components', async () => {
      await Promise.all([
        errorHandler.initialize(),
        configManager.initialize(),
        resourceManager.initialize(),
        imageOptimizer.initialize(),
      ]);

      // Add some resources and activity
      resourceManager.registerTimer(() => {}, 1000);
      errorHandler.handleError(new Error('Test error'));

      // Shutdown all components
      await Promise.all([
        errorHandler.destroy(),
        configManager.destroy(),
        resourceManager.destroy(),
        imageOptimizer.destroy(),
      ]);

      // All components should be destroyed
      expect(errorHandler.getStatus()).toBe('destroyed');
      expect(configManager.getStatus()).toBe('destroyed');
      expect(resourceManager.getStatus()).toBe('destroyed');
      expect(imageOptimizer.getStatus()).toBe('destroyed');

      // Resources should be cleaned up
      const resourceStats = resourceManager.getResourceStats();
      expect(resourceStats.totalResources).toBe(0);
    });

    it('should handle shutdown errors gracefully', async () => {
      await Promise.all([
        errorHandler.initialize(),
        configManager.initialize(),
        resourceManager.initialize(),
        imageOptimizer.initialize(),
      ]);

      // Mock one component to fail during shutdown
      const originalDestroy = configManager.destroy;
      configManager.destroy = jest
        .fn()
        .mockRejectedValue(new Error('Shutdown failed'));

      const results = await Promise.allSettled([
        errorHandler.destroy(),
        configManager.destroy(),
        resourceManager.destroy(),
        imageOptimizer.destroy(),
      ]);

      // Should handle failures gracefully
      const failures = results.filter(r => r.status === 'rejected');
      expect(failures).toHaveLength(1);

      // Restore original method
      configManager.destroy = originalDestroy;
    });
  });
});
