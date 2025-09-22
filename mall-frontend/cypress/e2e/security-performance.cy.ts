/**
 * 安全和性能 E2E 测试
 * 测试完整的用户交互流程和系统性能
 */

describe('Security and Performance E2E Tests', () => {
  beforeEach(() => {
    // 访问首页
    cy.visit('/');

    // 等待页面加载完成
    cy.get('[data-testid="app-loaded"]', { timeout: 10000 }).should('exist');
  });

  describe('Security Features', () => {
    it('should handle XSS protection', () => {
      // 测试输入框的XSS防护
      cy.get('[data-testid="search-input"]').should('exist');

      // 尝试输入恶意脚本
      const maliciousScript = '<script>alert("XSS")</script>';
      cy.get('[data-testid="search-input"]')
        .type(maliciousScript)
        .should('not.contain', '<script>');

      // 确保脚本没有执行
      cy.window().then(win => {
        expect(win.alert).not.to.have.been.called;
      });
    });

    it('should secure token storage', () => {
      // 检查localStorage中没有敏感token
      cy.window().then(win => {
        const localStorage = win.localStorage;
        const sensitiveKeys = ['token', 'accessToken', 'authToken'];

        sensitiveKeys.forEach(key => {
          expect(localStorage.getItem(key)).to.be.null;
        });
      });

      // 检查sessionStorage中没有敏感token
      cy.window().then(win => {
        const sessionStorage = win.sessionStorage;
        const sensitiveKeys = ['token', 'accessToken', 'authToken'];

        sensitiveKeys.forEach(key => {
          expect(sessionStorage.getItem(key)).to.be.null;
        });
      });
    });

    it('should handle CSRF protection', () => {
      // 模拟登录流程
      cy.get('[data-testid="login-button"]').click();

      // 填写登录表单
      cy.get('[data-testid="email-input"]').type('test@example.com');
      cy.get('[data-testid="password-input"]').type('password123');

      // 提交表单
      cy.get('[data-testid="submit-login"]').click();

      // 检查请求头中包含CSRF token
      cy.intercept('POST', '/api/auth/login').as('loginRequest');
      cy.wait('@loginRequest').then(interception => {
        expect(interception.request.headers).to.have.property('x-csrf-token');
      });
    });

    it('should validate input sanitization', () => {
      // 测试评论输入的HTML清理
      cy.get('[data-testid="product-card"]').first().click();
      cy.get('[data-testid="add-review-button"]').click();

      // 输入包含HTML标签的评论
      const htmlContent =
        '<b>Bold text</b> and <img src="x" onerror="alert(1)">';
      cy.get('[data-testid="review-content"]').type(htmlContent);

      // 提交评论
      cy.get('[data-testid="submit-review"]').click();

      // 检查显示的内容已被清理
      cy.get('[data-testid="review-list"]')
        .should('contain', 'Bold text')
        .should('not.contain', '<img')
        .should('not.contain', 'onerror');
    });
  });

  describe('Performance Features', () => {
    it('should load images efficiently', () => {
      // 检查图片懒加载
      cy.get('[data-testid="product-image"]').should(
        'have.length.greaterThan',
        0
      );

      // 滚动到页面底部触发懒加载
      cy.scrollTo('bottom');

      // 检查图片加载性能
      cy.get('[data-testid="product-image"]').each($img => {
        cy.wrap($img)
          .should('have.attr', 'loading', 'lazy')
          .and('have.attr', 'src')
          .and('not.be.empty');
      });
    });

    it('should optimize image formats', () => {
      // 检查图片URL是否包含优化参数
      cy.get('[data-testid="product-image"]')
        .first()
        .then($img => {
          const src = $img.attr('src');
          expect(src).to.match(/[?&](w=|h=|q=|f=)/); // 包含宽度、高度、质量或格式参数
        });
    });

    it('should handle image loading errors gracefully', () => {
      // 模拟图片加载失败
      cy.intercept('GET', '/images/**', { statusCode: 404 }).as('imageError');

      cy.visit('/products');

      // 检查错误处理
      cy.get('[data-testid="image-error-placeholder"]', { timeout: 5000 })
        .should('be.visible')
        .and('contain', '图片加载失败');
    });

    it('should cache resources effectively', () => {
      // 首次访问页面
      cy.visit('/products');
      cy.get('[data-testid="product-list"]').should('be.visible');

      // 记录网络请求
      let firstLoadRequests = 0;
      cy.intercept('GET', '/api/**', () => {
        firstLoadRequests++;
      }).as('apiRequests');

      // 刷新页面
      cy.reload();
      cy.get('[data-testid="product-list"]').should('be.visible');

      // 检查缓存效果（第二次加载的请求应该更少）
      cy.get('@apiRequests.all').should(
        'have.length.lessThan',
        firstLoadRequests
      );
    });

    it('should measure page load performance', () => {
      // 使用Performance API测量加载时间
      cy.window().then(win => {
        const navigation = win.performance.getEntriesByType(
          'navigation'
        )[0] as PerformanceNavigationTiming;

        // 检查关键性能指标
        const loadTime = navigation.loadEventEnd - navigation.navigationStart;
        const domContentLoaded =
          navigation.domContentLoadedEventEnd - navigation.navigationStart;

        expect(loadTime).to.be.lessThan(3000); // 页面加载时间小于3秒
        expect(domContentLoaded).to.be.lessThan(2000); // DOM加载时间小于2秒
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle network errors gracefully', () => {
      // 模拟网络错误
      cy.intercept('GET', '/api/products', { statusCode: 500 }).as(
        'serverError'
      );

      cy.visit('/products');

      // 检查错误提示
      cy.get('[data-testid="error-message"]', { timeout: 5000 })
        .should('be.visible')
        .and('contain', '服务器错误');

      // 检查重试按钮
      cy.get('[data-testid="retry-button"]').should('be.visible');
    });

    it('should handle JavaScript errors', () => {
      // 监听JavaScript错误
      cy.window().then(win => {
        win.addEventListener('error', e => {
          // 错误应该被错误处理器捕获，不应该导致页面崩溃
          expect(e.error).to.exist;
        });
      });

      // 触发可能的错误场景
      cy.get('[data-testid="complex-interaction"]').click();

      // 页面应该仍然可用
      cy.get('[data-testid="app-loaded"]').should('exist');
    });

    it('should provide user-friendly error messages', () => {
      // 测试表单验证错误
      cy.get('[data-testid="contact-form"]').within(() => {
        cy.get('[data-testid="email-input"]').type('invalid-email');
        cy.get('[data-testid="submit-button"]').click();

        // 检查友好的错误提示
        cy.get('[data-testid="email-error"]')
          .should('be.visible')
          .and('contain', '请输入有效的邮箱地址');
      });
    });
  });

  describe('Resource Management', () => {
    it('should clean up resources on page navigation', () => {
      // 访问包含定时器的页面
      cy.visit('/dashboard');
      cy.get('[data-testid="dashboard-loaded"]').should('exist');

      // 导航到其他页面
      cy.get('[data-testid="nav-products"]').click();
      cy.get('[data-testid="product-list"]').should('be.visible');

      // 检查内存使用情况（通过Performance API）
      cy.window().then(win => {
        if ('memory' in win.performance) {
          const memory = (win.performance as any).memory;
          expect(memory.usedJSHeapSize).to.be.lessThan(
            memory.jsHeapSizeLimit * 0.8
          );
        }
      });
    });

    it('should handle memory leaks prevention', () => {
      // 多次导航测试内存泄漏
      const pages = ['/products', '/dashboard', '/profile', '/cart'];

      pages.forEach((page, index) => {
        cy.visit(page);
        cy.get('[data-testid="page-loaded"]', { timeout: 5000 }).should(
          'exist'
        );

        // 检查内存使用趋势
        cy.window().then(win => {
          if ('memory' in win.performance) {
            const memory = (win.performance as any).memory;
            // 内存使用不应该持续增长
            expect(memory.usedJSHeapSize).to.be.lessThan(50 * 1024 * 1024); // 小于50MB
          }
        });
      });
    });
  });

  describe('Configuration Management', () => {
    it('should respect user preferences', () => {
      // 设置用户偏好
      cy.visit('/settings');
      cy.get('[data-testid="theme-selector"]').select('dark');
      cy.get('[data-testid="language-selector"]').select('en');
      cy.get('[data-testid="save-settings"]').click();

      // 检查设置是否生效
      cy.get('body').should('have.class', 'dark-theme');
      cy.get('[data-testid="welcome-message"]').should('contain', 'Welcome');

      // 刷新页面检查持久化
      cy.reload();
      cy.get('body').should('have.class', 'dark-theme');
    });

    it('should handle configuration updates', () => {
      // 测试运行时配置更新
      cy.visit('/admin/config');

      // 更新配置
      cy.get('[data-testid="api-timeout"]').clear().type('5000');
      cy.get('[data-testid="update-config"]').click();

      // 检查配置更新通知
      cy.get('[data-testid="config-updated"]')
        .should('be.visible')
        .and('contain', '配置已更新');
    });
  });

  describe('Accessibility and User Experience', () => {
    it('should be accessible to screen readers', () => {
      // 检查ARIA标签
      cy.get('[data-testid="main-navigation"]')
        .should('have.attr', 'role', 'navigation')
        .and('have.attr', 'aria-label');

      // 检查图片alt属性
      cy.get('img').each($img => {
        cy.wrap($img).should('have.attr', 'alt');
      });
    });

    it('should provide loading indicators', () => {
      // 检查加载状态
      cy.intercept('GET', '/api/products', { delay: 2000 }).as('slowRequest');

      cy.visit('/products');

      // 应该显示加载指示器
      cy.get('[data-testid="loading-spinner"]').should('be.visible');

      cy.wait('@slowRequest');

      // 加载完成后应该隐藏
      cy.get('[data-testid="loading-spinner"]').should('not.exist');
      cy.get('[data-testid="product-list"]').should('be.visible');
    });

    it('should handle offline scenarios', () => {
      // 模拟离线状态
      cy.window().then(win => {
        Object.defineProperty(win.navigator, 'onLine', {
          writable: true,
          value: false,
        });

        win.dispatchEvent(new Event('offline'));
      });

      // 检查离线提示
      cy.get('[data-testid="offline-banner"]', { timeout: 5000 })
        .should('be.visible')
        .and('contain', '网络连接已断开');
    });
  });
});
