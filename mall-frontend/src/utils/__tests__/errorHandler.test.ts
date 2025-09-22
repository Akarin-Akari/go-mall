/**
 * ErrorHandler 单元测试
 * 测试统一错误处理系统的各项功能
 */

import { ErrorHandler, ErrorType, ErrorLevel } from '../errorHandler';

// Mock console methods
const mockConsole = {
  error: jest.fn(),
  warn: jest.fn(),
  info: jest.fn(),
  log: jest.fn(),
};

// Mock fetch for error reporting
global.fetch = jest.fn();

describe('ErrorHandler', () => {
  let errorHandler: ErrorHandler;

  beforeEach(() => {
    // Reset singleton instance
    (ErrorHandler as any).instance = null;
    errorHandler = ErrorHandler.getInstance();

    // Clear mocks
    jest.clearAllMocks();

    // Mock console
    Object.assign(console, mockConsole);
  });

  afterEach(() => {
    // Clean up
    errorHandler.destroy();
  });

  describe('Singleton Pattern', () => {
    it('should return the same instance', () => {
      const instance1 = ErrorHandler.getInstance();
      const instance2 = ErrorHandler.getInstance();
      expect(instance1).toBe(instance2);
    });
  });

  describe('Error Handling', () => {
    it('should handle basic error', () => {
      const error = new Error('Test error');
      const result = errorHandler.handleError(error);

      expect(result).toBeDefined();
      expect(result.id).toBeDefined();
      expect(result.type).toBe(ErrorType.UNKNOWN);
      expect(result.level).toBe(ErrorLevel.ERROR);
    });

    it('should handle error with options', () => {
      const error = new Error('Test error');
      const options = {
        type: ErrorType.NETWORK,
        level: ErrorLevel.WARNING,
        context: { userId: '123' },
        code: 'NET_001',
      };

      const result = errorHandler.handleError(error, options);

      expect(result.type).toBe(ErrorType.NETWORK);
      expect(result.level).toBe(ErrorLevel.WARNING);
      expect(result.context).toEqual({ userId: '123' });
      expect(result.code).toBe('NET_001');
    });

    it('should handle string error', () => {
      const result = errorHandler.handleError('String error message');

      expect(result.message).toBe('String error message');
      expect(result.type).toBe(ErrorType.UNKNOWN);
    });

    it('should classify network errors', () => {
      const networkError = new Error('Failed to fetch');
      const result = errorHandler.handleError(networkError);

      expect(result.type).toBe(ErrorType.NETWORK);
    });

    it('should classify validation errors', () => {
      const validationError = new Error('Invalid input format');
      const result = errorHandler.handleError(validationError);

      expect(result.type).toBe(ErrorType.VALIDATION);
    });
  });

  describe('Error Listeners', () => {
    it('should add and trigger error listeners', () => {
      const listener = jest.fn();
      const removeListener = errorHandler.addListener(listener);

      const error = new Error('Test error');
      errorHandler.handleError(error);

      expect(listener).toHaveBeenCalledTimes(1);
      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({
          message: 'Test error',
        })
      );

      // Test remove listener
      removeListener();
      errorHandler.handleError(error);
      expect(listener).toHaveBeenCalledTimes(1); // Should not be called again
    });

    it('should handle multiple listeners', () => {
      const listener1 = jest.fn();
      const listener2 = jest.fn();

      errorHandler.addListener(listener1);
      errorHandler.addListener(listener2);

      const error = new Error('Test error');
      errorHandler.handleError(error);

      expect(listener1).toHaveBeenCalledTimes(1);
      expect(listener2).toHaveBeenCalledTimes(1);
    });
  });

  describe('Error Statistics', () => {
    it('should track error statistics', () => {
      // Handle different types of errors
      errorHandler.handleError(new Error('Error 1'), {
        type: ErrorType.NETWORK,
      });
      errorHandler.handleError(new Error('Error 2'), {
        type: ErrorType.NETWORK,
      });
      errorHandler.handleError(new Error('Error 3'), {
        type: ErrorType.VALIDATION,
      });
      errorHandler.handleError(new Error('Error 4'), {
        level: ErrorLevel.WARNING,
      });

      const stats = errorHandler.getErrorStats();

      expect(stats.totalErrors).toBe(4);
      expect(stats.errorsByType[ErrorType.NETWORK]).toBe(2);
      expect(stats.errorsByType[ErrorType.VALIDATION]).toBe(1);
      expect(stats.errorsByLevel[ErrorLevel.WARNING]).toBe(1);
      expect(stats.errorsByLevel[ErrorLevel.ERROR]).toBe(3);
    });

    it('should reset error count', () => {
      errorHandler.handleError(new Error('Test error'));
      expect(errorHandler.getErrorStats().totalErrors).toBe(1);

      errorHandler.resetErrorCount();
      expect(errorHandler.getErrorStats().totalErrors).toBe(0);
    });
  });

  describe('Error Queue Processing', () => {
    it('should process error queue', async () => {
      const mockFetch = fetch as jest.MockedFunction<typeof fetch>;
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      } as Response);

      // Configure remote reporting
      await errorHandler.updateConfig({
        enableRemoteReporting: true,
        remoteEndpoint: 'https://api.example.com/errors',
      });

      errorHandler.handleError(new Error('Test error'));

      // Wait for queue processing
      await new Promise(resolve => setTimeout(resolve, 100));

      expect(mockFetch).toHaveBeenCalledWith(
        'https://api.example.com/errors',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: expect.stringContaining('Test error'),
        })
      );
    });

    it('should handle remote reporting failure', async () => {
      const mockFetch = fetch as jest.MockedFunction<typeof fetch>;
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await errorHandler.updateConfig({
        enableRemoteReporting: true,
        remoteEndpoint: 'https://api.example.com/errors',
      });

      errorHandler.handleError(new Error('Test error'));

      // Wait for queue processing
      await new Promise(resolve => setTimeout(resolve, 100));

      expect(mockConsole.error).toHaveBeenCalledWith(
        'Failed to report error to remote endpoint:',
        expect.any(Error)
      );
    });
  });

  describe('Configuration', () => {
    it('should update configuration', async () => {
      const newConfig = {
        enableConsoleLogging: false,
        enableRemoteReporting: true,
        maxQueueSize: 200,
      };

      await errorHandler.updateConfig(newConfig);
      const config = errorHandler.getConfig();

      expect(config.enableConsoleLogging).toBe(false);
      expect(config.enableRemoteReporting).toBe(true);
      expect(config.maxQueueSize).toBe(200);
    });

    it('should reset to default configuration', async () => {
      await errorHandler.updateConfig({ enableConsoleLogging: false });
      await errorHandler.resetConfig();

      const config = errorHandler.getConfig();
      expect(config.enableConsoleLogging).toBe(true);
    });
  });

  describe('Lifecycle Management', () => {
    it('should initialize properly', async () => {
      await errorHandler.initialize();
      expect(errorHandler.getStatus()).toBe('initialized');
    });

    it('should perform health check', async () => {
      const isHealthy = await errorHandler.healthCheck();
      expect(isHealthy).toBe(true);
    });

    it('should destroy properly', async () => {
      await errorHandler.destroy();
      expect(errorHandler.getStatus()).toBe('destroyed');
    });
  });

  describe('Error Retry Mechanism', () => {
    it('should retry failed operations', async () => {
      const mockOperation = jest
        .fn()
        .mockRejectedValueOnce(new Error('First failure'))
        .mockRejectedValueOnce(new Error('Second failure'))
        .mockResolvedValueOnce('Success');

      const result = await errorHandler.withRetry(mockOperation, {
        maxRetries: 3,
        delay: 10,
      });

      expect(result).toBe('Success');
      expect(mockOperation).toHaveBeenCalledTimes(3);
    });

    it('should fail after max retries', async () => {
      const mockOperation = jest
        .fn()
        .mockRejectedValue(new Error('Always fails'));

      await expect(
        errorHandler.withRetry(mockOperation, { maxRetries: 2, delay: 10 })
      ).rejects.toThrow('Always fails');

      expect(mockOperation).toHaveBeenCalledTimes(3); // Initial + 2 retries
    });
  });

  describe('Error Context Enhancement', () => {
    it('should enhance error context with user info', () => {
      const error = new Error('Test error');
      const result = errorHandler.handleError(error, {
        context: { userId: '123', action: 'login' },
      });

      expect(result.context).toEqual(
        expect.objectContaining({
          userId: '123',
          action: 'login',
          timestamp: expect.any(Number),
          userAgent: expect.any(String),
          url: expect.any(String),
        })
      );
    });

    it('should include stack trace for errors', () => {
      const error = new Error('Test error');
      const result = errorHandler.handleError(error);

      expect(result.stack).toBeDefined();
      expect(result.stack).toContain('Test error');
    });
  });
});
